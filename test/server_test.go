package test

import (
	"aws_fin_ops/internal/billing"
	"aws_fin_ops/internal/worker"
	"context"
	"testing"
	"time"
)

// TestDefaultWorkerConfig 验证默认 WorkerConfig 各字段的默认值。
func TestDefaultWorkerConfig(t *testing.T) {
	cfg := worker.DefaultWorkerConfig()

	if cfg.BudgetThreshold != 0.8 {
		t.Fatalf("期望 BudgetThreshold=0.8，实际得到 %f", cfg.BudgetThreshold)
	}
	if cfg.CostQueryDays != 30 {
		t.Fatalf("期望 CostQueryDays=30，实际得到 %d", cfg.CostQueryDays)
	}
	if cfg.ForecastDays != 30 {
		t.Fatalf("期望 ForecastDays=30，实际得到 %d", cfg.ForecastDays)
	}
	if !cfg.EnableCostQuery {
		t.Fatal("期望 EnableCostQuery=true")
	}
	if !cfg.EnableForecast {
		t.Fatal("期望 EnableForecast=true")
	}
	if !cfg.EnableComparison {
		t.Fatal("期望 EnableComparison=true")
	}
	if !cfg.EnableRightsizing {
		t.Fatal("期望 EnableRightsizing=true")
	}
	if !cfg.EnableAnalysis {
		t.Fatal("期望 EnableAnalysis=true")
	}
	if !cfg.EnableBroadcast {
		t.Fatal("期望 EnableBroadcast=true")
	}
}

// TestWorkerReport_Timing 验证执行报告中的时间字段正确记录。
func TestWorkerReport_Timing(t *testing.T) {
	bs, _ := billing.NewBudgetState(500, "USD", billing.BudgetTypeCost)

	cfg := &worker.WorkerConfig{
		EnableCostQuery:   false,
		EnableForecast:    false,
		EnableComparison:  false,
		EnableRightsizing: false,
		EnableAnalysis:    false,
		EnableBroadcast:   false,
	}

	before := time.Now()
	engine := worker.NewEngine(bs, cfg)
	report, err := engine.Run(context.Background())
	after := time.Now()

	if err != nil {
		t.Fatalf("Worker 执行失败: %v", err)
	}

	if report.StartTime.Before(before) {
		t.Fatal("StartTime 不应早于测试开始时间")
	}
	if report.EndTime.After(after) {
		t.Fatal("EndTime 不应晚于测试结束时间")
	}
	if report.EndTime.Before(report.StartTime) {
		t.Fatal("EndTime 不应早于 StartTime")
	}
	if report.Duration <= 0 {
		t.Fatalf("期望 Duration > 0，实际得到 %v", report.Duration)
	}
}

// TestWorkerReport_RunID 验证 RunID 格式以 finops- 开头。
func TestWorkerReport_RunID(t *testing.T) {
	bs, _ := billing.NewBudgetState(500, "USD", billing.BudgetTypeCost)

	cfg := &worker.WorkerConfig{
		EnableCostQuery:   false,
		EnableForecast:    false,
		EnableComparison:  false,
		EnableRightsizing: false,
		EnableAnalysis:    false,
		EnableBroadcast:   false,
	}

	engine := worker.NewEngine(bs, cfg)
	report, _ := engine.Run(context.Background())

	if len(report.RunID) < 7 || report.RunID[:7] != "finops-" {
		t.Fatalf("期望 RunID 以 'finops-' 开头，实际得到 %s", report.RunID)
	}
}

// TestWorkerEngine_ContextCancellation 验证引擎在上下文取消前仍能正常完成。
func TestWorkerEngine_ContextCancellation(t *testing.T) {
	bs, _ := billing.NewBudgetState(500, "USD", billing.BudgetTypeCost)

	cfg := &worker.WorkerConfig{
		EnableCostQuery:   false,
		EnableForecast:    false,
		EnableComparison:  false,
		EnableRightsizing: false,
		EnableAnalysis:    false,
		EnableBroadcast:   false,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	engine := worker.NewEngine(bs, cfg)
	report, err := engine.Run(ctx)
	if err != nil {
		t.Fatalf("Worker 执行失败: %v", err)
	}
	if report.Status != "completed" {
		t.Fatalf("期望 status=completed，实际得到 %s", report.Status)
	}
}

// TestWorkerEngine_StepDurations 验证每个步骤的执行时间都已记录。
func TestWorkerEngine_StepDurations(t *testing.T) {
	bs, _ := billing.NewBudgetState(500, "USD", billing.BudgetTypeCost)

	cfg := worker.DefaultWorkerConfig()
	cfg.EnableBroadcast = false
	engine := worker.NewEngine(bs, &cfg)

	report, err := engine.Run(context.Background())
	if err != nil {
		t.Fatalf("Worker 执行失败: %v", err)
	}

	for i, step := range report.Steps {
		if step.Duration < 0 {
			t.Fatalf("步骤 %d (%s) 的 Duration 不应为负值: %v", i+1, step.Step, step.Duration)
		}
	}
}
