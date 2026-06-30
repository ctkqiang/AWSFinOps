package test

import (
	"aws_fin_ops/internal/billing"
	"aws_fin_ops/internal/worker"
	"context"
	"testing"
)

// TestWorkerEngine_DefaultConfig 验证默认配置创建的引擎能成功执行。
func TestWorkerEngine_DefaultConfig(t *testing.T) {
	bs, err := billing.NewBudgetState(500, "USD", billing.BudgetTypeCost)
	if err != nil {
		t.Fatalf("预算初始化失败: %v", err)
	}

	engine := worker.NewEngine(bs, nil)
	report, err := engine.Run(context.Background())
	if err != nil {
		t.Fatalf("Worker 引擎执行失败: %v", err)
	}

	if report.RunID == "" {
		t.Fatal("期望 RunID 非空")
	}
	if report.Status == "" {
		t.Fatal("期望 Status 非空")
	}
	if len(report.Steps) == 0 {
		t.Fatal("期望至少有一个执行步骤")
	}
}

// TestWorkerEngine_BudgetCheckStep 验证预算检查步骤返回正确的预算数据。
func TestWorkerEngine_BudgetCheckStep(t *testing.T) {
	bs, _ := billing.NewBudgetState(1000, "USD", billing.BudgetTypeCost)

	cfg := worker.DefaultWorkerConfig()
	cfg.EnableCostQuery = false
	cfg.EnableForecast = false
	cfg.EnableComparison = false
	cfg.EnableRightsizing = false
	cfg.EnableAnalysis = false
	cfg.EnableBroadcast = false

	engine := worker.NewEngine(bs, &cfg)
	report, err := engine.Run(context.Background())
	if err != nil {
		t.Fatalf("Worker 执行失败: %v", err)
	}

	if len(report.Steps) != 1 {
		t.Fatalf("期望仅 1 个步骤（check_budget），实际得到 %d", len(report.Steps))
	}

	step := report.Steps[0]
	if step.Step != "check_budget" {
		t.Fatalf("期望步骤名为 check_budget，实际得到 %s", step.Step)
	}
	if step.Status != "ok" {
		t.Fatalf("期望状态为 ok，实际得到 %s", step.Status)
	}

	data, ok := step.Data.(map[string]interface{})
	if !ok {
		t.Fatal("期望 Data 为 map[string]interface{} 类型")
	}
	if data["amount"] != float64(1000) {
		t.Fatalf("期望 amount=1000，实际得到 %v", data["amount"])
	}
	if data["currency"] != "USD" {
		t.Fatalf("期望 currency=USD，实际得到 %v", data["currency"])
	}
}

// TestWorkerEngine_AllStepsEnabled 验证启用所有步骤时报告包含完整的 7 个步骤。
func TestWorkerEngine_AllStepsEnabled(t *testing.T) {
	bs, _ := billing.NewBudgetState(500, "USD", billing.BudgetTypeCost)
	engine := worker.NewEngine(bs, nil)

	report, err := engine.Run(context.Background())
	if err != nil {
		t.Fatalf("Worker 执行失败: %v", err)
	}

	expectedSteps := 7
	if len(report.Steps) != expectedSteps {
		t.Fatalf("期望 %d 个步骤，实际得到 %d", expectedSteps, len(report.Steps))
	}

	stepNames := []string{
		"check_budget",
		"get_cost",
		"get_forecast",
		"compare_usage",
		"get_rightsizing",
		"run_analysis",
		"broadcast",
	}
	for i, expected := range stepNames {
		if report.Steps[i].Step != expected {
			t.Fatalf("步骤 %d 期望名称 %s，实际得到 %s", i+1, expected, report.Steps[i].Step)
		}
	}
}

// TestWorkerEngine_CustomConfig 验证自定义配置能正确应用。
func TestWorkerEngine_CustomConfig(t *testing.T) {
	bs, _ := billing.NewBudgetState(500, "USD", billing.BudgetTypeCost)

	cfg := worker.DefaultWorkerConfig()
	cfg.CostQueryDays = 7
	cfg.ForecastDays = 14
	cfg.EnableRightsizing = false
	cfg.EnableBroadcast = false

	engine := worker.NewEngine(bs, &cfg)
	report, err := engine.Run(context.Background())
	if err != nil {
		t.Fatalf("Worker 执行失败: %v", err)
	}

	expectedSteps := 5
	if len(report.Steps) != expectedSteps {
		t.Fatalf("期望 %d 个步骤，实际得到 %d", expectedSteps, len(report.Steps))
	}
}

// TestWorkerEngine_ReportSummary 验证执行报告摘要非空且包含 RunID。
func TestWorkerEngine_ReportSummary(t *testing.T) {
	bs, _ := billing.NewBudgetState(500, "USD", billing.BudgetTypeCost)

	cfg := worker.DefaultWorkerConfig()
	cfg.EnableBroadcast = false
	engine := worker.NewEngine(bs, &cfg)

	report, err := engine.Run(context.Background())
	if err != nil {
		t.Fatalf("Worker 执行失败: %v", err)
	}

	if report.Summary == "" {
		t.Fatal("期望 Summary 非空")
	}
	if report.Duration <= 0 {
		t.Fatalf("期望 Duration > 0，实际得到 %v", report.Duration)
	}
}

// TestWorkerEngine_DisableAllOptional 验证禁用所有可选步骤时仅执行预算检查。
func TestWorkerEngine_DisableAllOptional(t *testing.T) {
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
	report, err := engine.Run(context.Background())
	if err != nil {
		t.Fatalf("Worker 执行失败: %v", err)
	}

	if len(report.Steps) != 1 {
		t.Fatalf("期望仅 1 个步骤，实际得到 %d", len(report.Steps))
	}
	if report.Steps[0].Step != "check_budget" {
		t.Fatalf("期望步骤为 check_budget，实际得到 %s", report.Steps[0].Step)
	}
	if report.Status != "completed" {
		t.Fatalf("期望状态为 completed，实际得到 %s", report.Status)
	}
}
