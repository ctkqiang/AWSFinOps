package worker

import (
	"aws_fin_ops/internal/billing"
	"aws_fin_ops/internal/services"
	"aws_fin_ops/internal/utilities"
	"context"
	"fmt"
	"os"
	"strings"
	"time"
)

const (
	component = "FinOpsWorker"
)

// WorkerResult 表示单个任务步骤的执行结果。
type WorkerResult struct {
	Step     string        `json:"step"`
	Status   string        `json:"status"`
	Duration time.Duration `json:"duration"`
	Message  string        `json:"message,omitempty"`
	Data     interface{}   `json:"data,omitempty"`
	Error    string        `json:"error,omitempty"`
}

// WorkerReport 表示一次完整的 Worker 执行报告，包含所有步骤的结果。
type WorkerReport struct {
	RunID     string         `json:"run_id"`
	StartTime time.Time      `json:"start_time"`
	EndTime   time.Time      `json:"end_time"`
	Duration  time.Duration  `json:"duration"`
	Status    string         `json:"status"`
	Steps     []WorkerResult `json:"steps"`
	Summary   string         `json:"summary"`
}

// WorkerConfig 控制 Worker 执行哪些步骤以及相关参数。
type WorkerConfig struct {
	BudgetThreshold   float64                 `json:"budget_threshold"`
	AnalysisConfig    *billing.AnalysisConfig `json:"analysis_config,omitempty"`
	EnableCostQuery   bool                    `json:"enable_cost_query"`
	EnableForecast    bool                    `json:"enable_forecast"`
	EnableComparison  bool                    `json:"enable_comparison"`
	EnableRightsizing bool                    `json:"enable_rightsizing"`
	EnableAnalysis    bool                    `json:"enable_analysis"`
	EnableBroadcast   bool                    `json:"enable_broadcast"`
	CostQueryDays     int                     `json:"cost_query_days"`
	ForecastDays      int                     `json:"forecast_days"`
	AWSAccountID      string                  `json:"aws_account_id,omitempty"`
	EnableReportExport bool                   `json:"enable_report_export"`
	ReportOutputDir   string                  `json:"report_output_dir,omitempty"`
	ExportPDF         bool                    `json:"export_pdf"`
	ExportJSON        bool                    `json:"export_json"`
	ExportCSV         bool                    `json:"export_csv"`
}

// DefaultWorkerConfig 返回包含合理默认值的 WorkerConfig。
// 所有开关均可通过环境变量覆盖，环境变量未设置时使用内置默认值。
//
// 返回：
//   - WorkerConfig : 默认启用所有步骤的配置
func DefaultWorkerConfig() WorkerConfig {
	return WorkerConfig{
		BudgetThreshold:    0.8,
		EnableCostQuery:    true,
		EnableForecast:     true,
		EnableComparison:   true,
		EnableRightsizing:  true,
		EnableAnalysis:     true,
		EnableBroadcast:    true,
		CostQueryDays:      30,
		ForecastDays:       30,
		AWSAccountID:       resolveAWSAccountID(),
		EnableReportExport: envBool("EXPORT_REPORT", true),
		ReportOutputDir:    envString("REPORT_OUTPUT_DIR", "./reports"),
		ExportPDF:          envBool("EXPORT_PDF", true),
		ExportJSON:         envBool("EXPORT_JSON", true),
		ExportCSV:          envBool("EXPORT_CSV", true),
	}
}

// Engine 是 FinOps Worker 的核心引擎，负责按序编排所有分析步骤。
type Engine struct {
	budgetState *billing.BudgetState
	config      WorkerConfig
}

// NewEngine 创建一个新的 Worker 引擎实例。
//
// 参数：
//   - budgetState : 当前预算状态
//   - cfg         : Worker 配置，传 nil 则使用默认配置
//
// 返回：
//   - *Engine : Worker 引擎实例
func NewEngine(budgetState *billing.BudgetState, cfg *WorkerConfig) *Engine {
	c := DefaultWorkerConfig()
	if cfg != nil {
		if cfg.BudgetThreshold > 0 {
			c.BudgetThreshold = cfg.BudgetThreshold
		}
		if cfg.CostQueryDays > 0 {
			c.CostQueryDays = cfg.CostQueryDays
		}
		if cfg.ForecastDays > 0 {
			c.ForecastDays = cfg.ForecastDays
		}
		c.EnableCostQuery = cfg.EnableCostQuery
		c.EnableForecast = cfg.EnableForecast
		c.EnableComparison = cfg.EnableComparison
		c.EnableRightsizing = cfg.EnableRightsizing
		c.EnableAnalysis = cfg.EnableAnalysis
		c.EnableBroadcast = cfg.EnableBroadcast
		if cfg.AnalysisConfig != nil {
			c.AnalysisConfig = cfg.AnalysisConfig
		}
		if cfg.AWSAccountID != "" {
			c.AWSAccountID = cfg.AWSAccountID
		}
		c.EnableReportExport = cfg.EnableReportExport
		if cfg.ReportOutputDir != "" {
			c.ReportOutputDir = cfg.ReportOutputDir
		}
		c.ExportPDF = cfg.ExportPDF
		c.ExportJSON = cfg.ExportJSON
		c.ExportCSV = cfg.ExportCSV
	}
	return &Engine{
		budgetState: budgetState,
		config:      c,
	}
}

// Run 执行完整的 FinOps 分析流水线。
// 按照以下顺序逐步执行各任务，并将结果记录到日志：
//  1. 预算状态检查
//  2. 成本查询（过去 N 天）
//  3. 成本预测（未来 N 天）
//  4. 用量对比（本期 vs 上期）
//  5. 规格优化建议（Rightsizing）
//  6. 综合分析报告
//  7. 广播通知（Telegram 等）
//
// 参数：
//   - ctx : 上下文，用于超时和取消控制
//
// 返回：
//   - *WorkerReport : 完整的执行报告
//   - error         : 流水线级别的致命错误，单步失败不会导致整体终止
func (e *Engine) Run(ctx context.Context) (*WorkerReport, error) {
	runStart := time.Now()
	runID := fmt.Sprintf("finops-%s", runStart.Format("20060102-150405"))

	utilities.LogStart(component, "Run")
	utilities.LogProgress(component, "Run",
		fmt.Sprintf("run_id=%s", runID),
		fmt.Sprintf("config.cost_query_days=%d", e.config.CostQueryDays),
		fmt.Sprintf("config.forecast_days=%d", e.config.ForecastDays),
	)

	report := &WorkerReport{
		RunID:     runID,
		StartTime: runStart,
		Steps:     make([]WorkerResult, 0, 7),
	}

	report.Steps = append(report.Steps, e.stepCheckBudget(ctx))

	if e.config.EnableCostQuery {
		report.Steps = append(report.Steps, e.stepGetCost(ctx))
	}

	if e.config.EnableForecast {
		report.Steps = append(report.Steps, e.stepGetForecast(ctx))
	}

	if e.config.EnableComparison {
		report.Steps = append(report.Steps, e.stepCompareUsage(ctx))
	}

	if e.config.EnableRightsizing {
		report.Steps = append(report.Steps, e.stepGetRightsizing(ctx))
	}

	if e.config.EnableAnalysis {
		report.Steps = append(report.Steps, e.stepRunAnalysis(ctx))
	}

	if e.config.EnableBroadcast {
		report.Steps = append(report.Steps, e.stepBroadcast(ctx, report))
	}

	report.EndTime = time.Now()
	report.Duration = report.EndTime.Sub(runStart)
	report.Status = e.evaluateOverallStatus(report.Steps)
	report.Summary = e.buildSummary(report)

	utilities.LogSuccess(component, "Run", report.Duration,
		fmt.Sprintf("run_id=%s", runID),
		fmt.Sprintf("status=%s", report.Status),
		fmt.Sprintf("steps=%d", len(report.Steps)),
	)

	e.logReportStructured(report)

	if e.config.EnableReportExport {
		e.exportReport(report)
	}

	return report, nil
}

// stepCheckBudget 检查当前预算状态并记录到日志。
func (e *Engine) stepCheckBudget(ctx context.Context) WorkerResult {
	start := time.Now()
	stepName := "check_budget"
	utilities.LogStart(component, stepName)

	amount, err := e.budgetState.GetAmount()
	if err != nil {
		utilities.LogError(component, stepName, err, time.Since(start))
		return WorkerResult{
			Step:     stepName,
			Status:   "error",
			Duration: time.Since(start),
			Error:    fmt.Sprintf("读取预算金额失败: %v", err),
		}
	}

	currency, _ := e.budgetState.GetCurrency()
	budgetType, _ := e.budgetState.GetBudgetType()

	data := map[string]interface{}{
		"amount":    float64(amount),
		"currency":  string(currency),
		"type":      string(budgetType),
		"threshold": e.config.BudgetThreshold,
	}

	utilities.LogSuccess(component, stepName, time.Since(start),
		fmt.Sprintf("amount=%.2f", float64(amount)),
		fmt.Sprintf("currency=%s", currency),
		fmt.Sprintf("type=%s", budgetType),
	)

	return WorkerResult{
		Step:     stepName,
		Status:   "ok",
		Duration: time.Since(start),
		Message:  fmt.Sprintf("预算: %.2f %s (%s)", float64(amount), currency, budgetType),
		Data:     data,
	}
}

// stepGetCost 查询过去 N 天的成本数据。
func (e *Engine) stepGetCost(ctx context.Context) WorkerResult {
	start := time.Now()
	stepName := "get_cost"
	utilities.LogStart(component, stepName)

	now := time.Now()
	startDate := now.AddDate(0, 0, -e.config.CostQueryDays).Format("2006-01-02")
	endDate := now.Format("2006-01-02")

	utilities.LogProgress(component, stepName,
		fmt.Sprintf("period=%s~%s", startDate, endDate),
		fmt.Sprintf("days=%d", e.config.CostQueryDays),
	)

	// TODO: 接入 AWS Cost Explorer SDK 实现真实查询
	utilities.LogWarn(component, stepName,
		"成本查询尚未接入 AWS SDK，当前为桩实现",
		time.Since(start),
	)

	return WorkerResult{
		Step:     stepName,
		Status:   "pending_implementation",
		Duration: time.Since(start),
		Message:  fmt.Sprintf("成本查询（%s ~ %s）待接入 AWS Cost Explorer SDK", startDate, endDate),
	}
}

// stepGetForecast 预测未来 N 天的成本趋势。
func (e *Engine) stepGetForecast(ctx context.Context) WorkerResult {
	start := time.Now()
	stepName := "get_forecast"
	utilities.LogStart(component, stepName)

	now := time.Now()
	forecastStart := now.Format("2006-01-02")
	forecastEnd := now.AddDate(0, 0, e.config.ForecastDays).Format("2006-01-02")

	utilities.LogProgress(component, stepName,
		fmt.Sprintf("forecast_period=%s~%s", forecastStart, forecastEnd),
		fmt.Sprintf("days=%d", e.config.ForecastDays),
	)

	// TODO: 接入 AWS Cost Explorer GetCostForecast SDK 实现真实预测
	utilities.LogWarn(component, stepName,
		"成本预测尚未接入 AWS SDK，当前为桩实现",
		time.Since(start),
	)

	return WorkerResult{
		Step:     stepName,
		Status:   "pending_implementation",
		Duration: time.Since(start),
		Message:  fmt.Sprintf("成本预测（%s ~ %s）待接入 AWS Cost Explorer SDK", forecastStart, forecastEnd),
	}
}

// stepCompareUsage 对比本期与上期的用量数据。
func (e *Engine) stepCompareUsage(ctx context.Context) WorkerResult {
	start := time.Now()
	stepName := "compare_usage"
	utilities.LogStart(component, stepName)

	now := time.Now()
	currentEnd := now.Format("2006-01-02")
	currentStart := now.AddDate(0, 0, -e.config.CostQueryDays).Format("2006-01-02")
	previousEnd := currentStart
	previousStart := now.AddDate(0, 0, -e.config.CostQueryDays*2).Format("2006-01-02")

	utilities.LogProgress(component, stepName,
		fmt.Sprintf("current=%s~%s", currentStart, currentEnd),
		fmt.Sprintf("previous=%s~%s", previousStart, previousEnd),
	)

	// TODO: 接入 AWS Cost Explorer GetCostAndUsageComparisons SDK 实现真实对比
	utilities.LogWarn(component, stepName,
		"用量对比尚未接入 AWS SDK，当前为桩实现",
		time.Since(start),
	)

	return WorkerResult{
		Step:     stepName,
		Status:   "pending_implementation",
		Duration: time.Since(start),
		Message:  fmt.Sprintf("用量对比（当前: %s~%s vs 上期: %s~%s）待接入 AWS SDK", currentStart, currentEnd, previousStart, previousEnd),
	}
}

// stepGetRightsizing 获取 EC2 规格优化建议。
func (e *Engine) stepGetRightsizing(ctx context.Context) WorkerResult {
	start := time.Now()
	stepName := "get_rightsizing"
	utilities.LogStart(component, stepName)

	// TODO: 接入 AWS Cost Explorer GetRightsizingRecommendation SDK 实现真实建议
	utilities.LogWarn(component, stepName,
		"规格优化建议尚未接入 AWS SDK，当前为桩实现",
		time.Since(start),
	)

	return WorkerResult{
		Step:     stepName,
		Status:   "pending_implementation",
		Duration: time.Since(start),
		Message:  "EC2 规格优化建议待接入 AWS Cost Explorer SDK",
	}
}

// stepRunAnalysis 执行综合成本分析并生成报告。
func (e *Engine) stepRunAnalysis(ctx context.Context) WorkerResult {
	start := time.Now()
	stepName := "run_analysis"
	utilities.LogStart(component, stepName)

	cfg := billing.DefaultAnalysisConfig()
	if e.config.AnalysisConfig != nil {
		cfg = *e.config.AnalysisConfig
	}

	utilities.LogProgress(component, stepName,
		fmt.Sprintf("metric=%s", cfg.Metric),
		fmt.Sprintf("granularity=%s", cfg.Granularity),
		fmt.Sprintf("forecast_days=%d", cfg.ForecastDays),
		fmt.Sprintf("anomaly_threshold=%.1f%%", cfg.AnomalyThresholdPct),
	)

	// TODO: 聚合上述步骤结果，生成 CostAnalysisReport
	utilities.LogWarn(component, stepName,
		"综合分析引擎尚未实现，当前为桩实现",
		time.Since(start),
	)

	return WorkerResult{
		Step:     stepName,
		Status:   "pending_implementation",
		Duration: time.Since(start),
		Message:  "综合成本分析报告待实现",
	}
}

// stepBroadcast 将分析结果通过配置的渠道（Telegram 等）发送通知。
func (e *Engine) stepBroadcast(ctx context.Context, report *WorkerReport) WorkerResult {
	start := time.Now()
	stepName := "broadcast"
	utilities.LogStart(component, stepName)

	summary := e.buildBroadcastMessage(report)

	err := services.TelegramSendMessage(summary)
	if err != nil {
		utilities.LogWarn(component, stepName,
			fmt.Sprintf("Telegram 通知发送失败: %v", err),
			time.Since(start),
		)
		return WorkerResult{
			Step:     stepName,
			Status:   "warn",
			Duration: time.Since(start),
			Message:  fmt.Sprintf("Telegram 通知发送失败（非致命）: %v", err),
		}
	}

	utilities.LogSuccess(component, stepName, time.Since(start), "channel=telegram")

	return WorkerResult{
		Step:     stepName,
		Status:   "ok",
		Duration: time.Since(start),
		Message:  "通知已通过 Telegram 发送",
	}
}

// evaluateOverallStatus 根据所有步骤结果判定整体状态。
func (e *Engine) evaluateOverallStatus(steps []WorkerResult) string {
	hasError := false
	hasWarn := false
	for _, s := range steps {
		switch s.Status {
		case "error":
			hasError = true
		case "warn":
			hasWarn = true
		}
	}
	if hasError {
		return "partial_failure"
	}
	if hasWarn {
		return "completed_with_warnings"
	}
	return "completed"
}

// buildSummary 构建一份简洁的执行摘要。
func (e *Engine) buildSummary(report *WorkerReport) string {
	total := len(report.Steps)
	ok := 0
	pending := 0
	failed := 0
	warned := 0

	for _, s := range report.Steps {
		switch s.Status {
		case "ok":
			ok++
		case "pending_implementation":
			pending++
		case "error":
			failed++
		case "warn":
			warned++
		}
	}

	return fmt.Sprintf(
		"[%s] 执行完成: 共 %d 步, 成功 %d, 待实现 %d, 警告 %d, 失败 %d, 耗时 %v",
		report.RunID, total, ok, pending, warned, failed, report.Duration.Round(time.Millisecond),
	)
}

// buildBroadcastMessage 构建发送到通知渠道的消息文本。
func (e *Engine) buildBroadcastMessage(report *WorkerReport) string {
	msg := fmt.Sprintf("AWSFinOps Worker 执行报告\n"+
		"━━━━━━━━━━━━━━━━━━━━━\n"+
		"运行 ID: %s\n"+
		"时间: %s\n"+
		"耗时: %v\n\n",
		report.RunID,
		report.StartTime.Format("2006-01-02 15:04:05 MST"),
		report.Duration.Round(time.Millisecond),
	)

	for _, step := range report.Steps {
		icon := "[OK]"
		switch step.Status {
		case "error":
			icon = "[FAIL]"
		case "warn":
			icon = "[WARN]"
		case "pending_implementation":
			icon = "[STUB]"
		}
		msg += fmt.Sprintf("%s %s: %s\n", icon, step.Step, step.Message)
	}

	msg += fmt.Sprintf("\n━━━━━━━━━━━━━━━━━━━━━\n状态: %s", report.Status)

	return msg
}

// logReportStructured 将执行报告输出为 CloudWatch 兼容的结构化日志。
// 每一步一条独立的结构化日志条目，便于 CloudWatch Logs Insights 过滤和聚合。
//
// 参数：
//   - report : 完整的执行报告
func (e *Engine) logReportStructured(report *WorkerReport) {
	utilities.LogProgress(component, "Report", "执行报告汇总",
		fmt.Sprintf("run_id=%s", report.RunID),
		fmt.Sprintf("aws_account_id=%s", strings.ToUpper(e.config.AWSAccountID)),
		fmt.Sprintf("start_time=%s", report.StartTime.Format(time.RFC3339)),
		fmt.Sprintf("end_time=%s", report.EndTime.Format(time.RFC3339)),
		fmt.Sprintf("duration_ms=%d", report.Duration.Milliseconds()),
		fmt.Sprintf("status=%s", report.Status),
		fmt.Sprintf("total_steps=%d", len(report.Steps)),
		fmt.Sprintf("summary=%s", report.Summary),
	)

	for _, step := range report.Steps {
		switch step.Status {
		case "ok":
			utilities.LogSuccess(component, "ReportStep", step.Duration,
				fmt.Sprintf("step=%s", step.Step),
				fmt.Sprintf("message=%s", step.Message),
			)
		case "warn", "pending_implementation":
			utilities.LogWarn(component, "ReportStep",
				fmt.Sprintf("step=%s status=%s", step.Step, step.Status),
				step.Duration,
				fmt.Sprintf("message=%s", step.Message),
			)
		case "error":
			utilities.LogError(component, "ReportStep",
				fmt.Errorf("%s", step.Error),
				step.Duration,
				fmt.Sprintf("step=%s", step.Step),
			)
		default:
			utilities.LogProgress(component, "ReportStep",
				fmt.Sprintf("step=%s status=%s", step.Step, step.Status),
				fmt.Sprintf("duration=%v", step.Duration),
			)
		}
	}
}

// exportReport 将执行报告导出为 PDF / JSON / CSV 三种格式。
// 导出失败仅记录警告，不影响主流程。
//
// 参数：
//   - report : 完整的执行报告
func (e *Engine) exportReport(report *WorkerReport) {
	start := time.Now()
	utilities.LogStart(component, "ExportReport")

	steps := make([]utilities.ReportStepData, len(report.Steps))
	for i, s := range report.Steps {
		steps[i] = utilities.ReportStepData{
			Index:    i + 1,
			Step:     s.Step,
			Status:   s.Status,
			Duration: s.Duration,
			Message:  s.Message,
			Error:    s.Error,
		}
	}

	data := &utilities.ReportData{
		RunID:        report.RunID,
		AWSAccountID: strings.ToUpper(e.config.AWSAccountID),
		StartTime:    report.StartTime,
		EndTime:      report.EndTime,
		Duration:     report.Duration,
		Status:       report.Status,
		Steps:        steps,
		Summary:      report.Summary,
	}

	cfg := &utilities.ExportConfig{
		OutputDir:    e.config.ReportOutputDir,
		EnablePDF:    e.config.ExportPDF,
		EnableJSON:   e.config.ExportJSON,
		EnableCSV:    e.config.ExportCSV,
		AWSAccountID: e.config.AWSAccountID,
	}

	paths, err := utilities.ExportAll(data, cfg)
	if err != nil {
		utilities.LogWarn(component, "ExportReport",
			fmt.Sprintf("报告导出失败: %v", err),
			time.Since(start),
		)
		return
	}

	utilities.LogSuccess(component, "ExportReport", time.Since(start),
		fmt.Sprintf("files=%d", len(paths)),
		fmt.Sprintf("paths=%s", strings.Join(paths, ",")),
	)

	// 根据 .env 配置自动广播到已启用的平台
	if !envBool("BROADCAST_ENABLED", true) {
		utilities.LogProgress(component, "ExportReport",
			"广播已禁用（BROADCAST_ENABLED=false）",
		)
		return
	}

	sentPlatforms := services.BroadcastReport(data, paths)
	if len(sentPlatforms) > 0 {
		utilities.LogProgress(component, "ExportReport",
			fmt.Sprintf("报告已发送至: %s", strings.Join(sentPlatforms, ", ")),
		)
	}
}

// envString 读取字符串类型的环境变量，未设置或为空时返回默认值。
//
// 参数：
//   - key      : 环境变量名称
//   - fallback : 环境变量未设置时的回退值
//
// 返回：
//   - string : 环境变量的值或回退值
func envString(key string, fallback string) string {
	if v := strings.TrimSpace(os.Getenv(key)); v != "" {
		return v
	}
	return fallback
}

// envBool 读取布尔类型的环境变量，未设置或解析失败时返回默认值。
// 可识别的值：true / false / 1 / 0 / yes / no（不区分大小写）。
//
// 参数：
//   - key      : 环境变量名称
//   - fallback : 环境变量未设置或无效时的回退值
//
// 返回：
//   - bool : 解析后的布尔值或回退值
func envBool(key string, fallback bool) bool {
	v := strings.TrimSpace(os.Getenv(key))
	if v == "" {
		return fallback
	}
	switch strings.ToLower(v) {
	case "true", "1", "yes", "y", "on":
		return true
	case "false", "0", "no", "n", "off":
		return false
	default:
		return fallback
	}
}

// resolveAWSAccountID 从环境变量读取 AWS 账号 ID。
// 优先级：AWS_ACCOUNT_ID > ACCOUNT_ID > 空字符串
// 全部为空时返回空字符串，PDF 水印会回退为 "CONFIDENTIAL"。
//
// 返回：
//   - string : AWS 账号 ID
func resolveAWSAccountID() string {
	if v := strings.TrimSpace(os.Getenv("AWS_ACCOUNT_ID")); v != "" {
		return v
	}
	if v := strings.TrimSpace(os.Getenv("ACCOUNT_ID")); v != "" {
		return v
	}
	return ""
}
