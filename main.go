package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"

	"aws_fin_ops/internal"
	internalaws "aws_fin_ops/internal/aws"
	"aws_fin_ops/internal/billing"
	"aws_fin_ops/internal/utilities"
	"aws_fin_ops/internal/worker"
)

const (
	// mainComponent 是本文件在日志中使用的组件名称标识。
	mainComponent = "FinOps"

	// defaultWorkerInterval 是本地模式下 Worker 循环执行的默认间隔（6 小时）。
	// 与 EventBridge 排程对齐，保证本地测试行为与生产一致。
	defaultWorkerInterval = 6 * time.Hour
)

var (
	// Budget 是初始预算金额（500 美元）。
	// 这是一个示例值，实际使用时应根据业务需求调整。
	Budget billing.Budget = 500

	// BudgetCurrency 是预算使用的货币代码。
	BudgetCurrency billing.BudgetCurrency = "USD"

	// BudgetType 是预算类型，使用基于费用的预算。
	BudgetType = billing.BudgetTypeCost
)

func main() {
	ctx := context.Background()

	if err := internal.LoadConfig(ctx); err != nil {
		utilities.Error("配置加载失败: %v", err)
		os.Exit(1)
	}

	budgetState, err := billing.NewBudgetState(Budget, BudgetCurrency, BudgetType)
	if err != nil {
		utilities.Error("预算初始化失败: %v", err)
		os.Exit(1)
	}

	amount, _ := budgetState.GetAmount()
	currency, _ := budgetState.GetCurrency()
	utilities.Info("预算已就绪: %.2f %s", float64(amount), currency)

	if internal.IsLocalMode() {
		runLocal(ctx, budgetState)
	} else {
		utilities.LogProgress(mainComponent, "main", "Lambda 模式 — 启动 EventBridge 事件处理器")
		internalaws.StartLambda(budgetState)
	}
}

// runLocal 在本地开发环境中运行 Worker。
//
// 运行模式由环境变量控制：
//   - WORKER_ONCE=true        : 执行一次后退出（适合调试、CI/CD）
//   - WORKER_ONCE 未设置/false : 进入循环调度模式，按 WORKER_INTERVAL 间隔重复执行
//
// 循环模式行为：
//  1. 启动后立即执行首轮 Worker 流水线（不等第一个 tick）
//  2. 每隔 WORKER_INTERVAL 执行下一轮
//  3. 监听 SIGINT / SIGTERM 信号，收到后等待当前轮次结束并优雅退出
//
// 参数：
//   - ctx         : 根上下文
//   - budgetState : 预算状态实例
func runLocal(ctx context.Context, budgetState *billing.BudgetState) {
	if isWorkerOnce() {
		utilities.LogProgress(mainComponent, "runLocal", "单次模式（WORKER_ONCE=true）— 执行一次后退出")
		runOnce(ctx, budgetState)
		return
	}

	interval := resolveWorkerInterval()
	utilities.LogProgress(mainComponent, "runLocal",
		fmt.Sprintf("循环模式 — 每 %v 执行一次 Worker 流水线", interval),
		"提示: 按 Ctrl+C 优雅退出",
	)

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	round := 0
	runWorkerRound(ctx, budgetState, &round)

	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			runWorkerRound(ctx, budgetState, &round)
		case sig := <-sigCh:
			utilities.LogProgress(mainComponent, "runLocal",
				fmt.Sprintf("收到信号 %v，正在优雅退出……", sig),
			)
			cancel()
			utilities.LogSuccess(mainComponent, "runLocal", 0,
				fmt.Sprintf("已完成 %d 轮执行，进程退出", round),
			)
			return
		}
	}
}

// runWorkerRound 执行一轮 Worker 流水线并记录结果。
//
// 参数：
//   - ctx         : 上下文，用于取消控制
//   - budgetState : 预算状态实例
//   - round       : 当前轮次计数器（会被自增）
func runWorkerRound(ctx context.Context, budgetState *billing.BudgetState, round *int) {
	*round++
	utilities.LogProgress(mainComponent, "runWorkerRound",
		fmt.Sprintf("═══ 第 %d 轮开始 ═══", *round),
	)

	engine := worker.NewEngine(budgetState, nil)
	report, err := engine.Run(ctx)
	if err != nil {
		utilities.Error("第 %d 轮 Worker 执行失败: %v", *round, err)
		return
	}

	if report.Status == "partial_failure" {
		utilities.LogWarn(mainComponent, "runWorkerRound",
			fmt.Sprintf("第 %d 轮执行部分失败，请检查日志", *round),
			report.Duration,
		)
		return
	}

	utilities.LogSuccess(mainComponent, "runWorkerRound", report.Duration,
		fmt.Sprintf("round=%d", *round),
		"status="+report.Status,
	)
}

// runOnce 执行一次 Worker 流水线后退出进程。
//
// 参数：
//   - ctx         : 上下文
//   - budgetState : 预算状态实例
func runOnce(ctx context.Context, budgetState *billing.BudgetState) {
	engine := worker.NewEngine(budgetState, nil)
	report, err := engine.Run(ctx)
	if err != nil {
		utilities.Error("Worker 执行失败: %v", err)
		os.Exit(1)
	}

	if report.Status == "partial_failure" {
		utilities.LogWarn(mainComponent, "runOnce", "Worker 执行部分失败，请检查日志", report.Duration)
		os.Exit(1)
	}

	utilities.LogSuccess(mainComponent, "runOnce", report.Duration, "mode=once", "status="+report.Status)
}

// resolveWorkerInterval 从环境变量 WORKER_INTERVAL 解析本地调度间隔。
//
// 支持格式：
//   - 纯数字（秒）: "60" → 60s
//   - 带单位后缀  : "30s" → 30s, "5m" → 5m, "3h" → 3h
//
// 环境变量未设置或解析失败时回退到 defaultWorkerInterval（3 小时）。
//
// 返回：
//   - time.Duration : 解析后的调度间隔
func resolveWorkerInterval() time.Duration {
	raw := os.Getenv("WORKER_INTERVAL")
	if raw == "" {
		return defaultWorkerInterval
	}

	raw = strings.TrimSpace(raw)

	if d, err := time.ParseDuration(raw); err == nil {
		if d < 10*time.Second {
			utilities.LogWarn(mainComponent, "resolveWorkerInterval",
				fmt.Sprintf("WORKER_INTERVAL=%s 过短（最小 10s），回退到默认值 %v", raw, defaultWorkerInterval),
				0,
			)
			return defaultWorkerInterval
		}
		return d
	}

	if secs, err := strconv.Atoi(raw); err == nil && secs > 0 {
		d := time.Duration(secs) * time.Second
		if d < 10*time.Second {
			utilities.LogWarn(mainComponent, "resolveWorkerInterval",
				fmt.Sprintf("WORKER_INTERVAL=%s 过短（最小 10s），回退到默认值 %v", raw, defaultWorkerInterval),
				0,
			)
			return defaultWorkerInterval
		}
		return d
	}

	utilities.LogWarn(mainComponent, "resolveWorkerInterval",
		fmt.Sprintf("WORKER_INTERVAL=%q 格式无效，回退到默认值 %v", raw, defaultWorkerInterval),
		0,
	)
	return defaultWorkerInterval
}

// isWorkerOnce 检查环境变量 WORKER_ONCE 是否为 true。
// 为 true 时本地模式仅执行一次后退出。
//
// 返回：
//   - bool : WORKER_ONCE=true 时返回 true
func isWorkerOnce() bool {
	return strings.EqualFold(strings.TrimSpace(os.Getenv("WORKER_ONCE")), "true")
}
