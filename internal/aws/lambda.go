package aws

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"aws_fin_ops/internal/billing"
	"aws_fin_ops/internal/utilities"
	"aws_fin_ops/internal/worker"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
)

const (
	lambdaComponent = "LambdaHandler"
)

// StartLambda 注册 Lambda handler 并启动 Lambda 运行时循环。
// 接收 EventBridge 定时事件（CloudWatch Scheduled Event），触发 FinOps Worker 流水线。
// 仅在 AWS Lambda 环境下由 main.go 调用。
//
// 参数：
//   - budgetState : 由 main 初始化的预算状态实例，在 Lambda 冷启动时创建，
//     后续 warm invocation 复用同一实例
func StartLambda(budgetState *billing.BudgetState) {
	utilities.LogStart(lambdaComponent, "StartLambda")
	lambda.Start(newHandler(budgetState))
}

// newHandler 返回一个接收 EventBridge 定时事件的 Lambda handler 函数。
// EventBridge 的 ScheduleExpression（如 rate(1 day) 或 cron(...)）触发此函数，
// 函数内部创建 Worker 引擎并执行完整的 FinOps 分析流水线。
// 所有执行结果通过结构化日志输出到 CloudWatch Logs。
//
// 参数：
//   - budgetState : 预算状态实例
//
// 返回：
//   - func(ctx, event) error : Lambda handler 函数
func newHandler(budgetState *billing.BudgetState) func(context.Context, events.CloudWatchEvent) error {
	return func(ctx context.Context, event events.CloudWatchEvent) error {
		start := time.Now()
		utilities.LogStart(lambdaComponent, "Invoke")

		utilities.LogProgress(lambdaComponent, "Invoke",
			fmt.Sprintf("source=%s", event.Source),
			fmt.Sprintf("detail_type=%s", event.DetailType),
			fmt.Sprintf("event_id=%s", event.ID),
			fmt.Sprintf("time=%s", event.Time.Format(time.RFC3339)),
		)

		var cfg *worker.WorkerConfig
		if len(event.Detail) > 0 {
			var customCfg worker.WorkerConfig
			if err := json.Unmarshal(event.Detail, &customCfg); err != nil {
				utilities.LogWarn(lambdaComponent, "Invoke",
					fmt.Sprintf("EventBridge detail 解析失败，使用默认配置: %v", err),
					time.Since(start),
				)
			} else {
				cfg = &customCfg
			}
		}

		engine := worker.NewEngine(budgetState, cfg)
		report, err := engine.Run(ctx)
		if err != nil {
			utilities.LogError(lambdaComponent, "Invoke", err, time.Since(start))
			return fmt.Errorf("worker 执行失败: %w", err)
		}

		utilities.LogSuccess(lambdaComponent, "Invoke", time.Since(start),
			fmt.Sprintf("run_id=%s", report.RunID),
			fmt.Sprintf("status=%s", report.Status),
			fmt.Sprintf("steps=%d", len(report.Steps)),
		)

		return nil
	}
}
