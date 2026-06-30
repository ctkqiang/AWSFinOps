package billing

import "time"

// AnalysisPeriod 定义分析涵盖的时间范围。
//
// 字段：
//   - Start : 起始时间
//   - End   : 结束时间
//   - Label : 可读标签（如 "2025年6月"、"Q2 2025"）
type AnalysisPeriod struct {
	Start time.Time `json:"start"`
	End   time.Time `json:"end"`
	Label string    `json:"label"`
}

// TrendDirection 定义费用趋势方向。
type TrendDirection string

const (
	// TrendIncreasing 费用呈上升趋势。
	TrendIncreasing TrendDirection = "INCREASING"

	// TrendDecreasing 费用呈下降趋势。
	TrendDecreasing TrendDirection = "DECREASING"

	// TrendStable 费用保持稳定（波动在阈值内）。
	TrendStable TrendDirection = "STABLE"

	// TrendSpike 费用出现异常飙升。
	TrendSpike TrendDirection = "SPIKE"
)

// SeverityLevel 定义分析发现的严重程度。
type SeverityLevel string

const (
	// SeverityCritical 关键级别：需要立即处理（如费用超预算 50%+）。
	SeverityCritical SeverityLevel = "CRITICAL"

	// SeverityHigh 高级别：需要近期关注（如费用超预算 20%-50%）。
	SeverityHigh SeverityLevel = "HIGH"

	// SeverityMedium 中级别：值得关注但不紧急（如费用环比增长 10%-20%）。
	SeverityMedium SeverityLevel = "MEDIUM"

	// SeverityLow 低级别：仅供参考的优化建议。
	SeverityLow SeverityLevel = "LOW"

	// SeverityInfo 信息级别：统计信息，无需行动。
	SeverityInfo SeverityLevel = "INFO"
)

// CostOptimizationCategory 定义费用优化建议的分类。
type CostOptimizationCategory string

const (
	// OptRightsizing 资源规格调整（降配、升配）。
	OptRightsizing CostOptimizationCategory = "RIGHTSIZING"

	// OptIdleResources 闲置资源清理。
	OptIdleResources CostOptimizationCategory = "IDLE_RESOURCES"

	// OptReservedInstances 预留实例购买建议。
	OptReservedInstances CostOptimizationCategory = "RESERVED_INSTANCES"

	// OptSavingsPlans Savings Plans 购买建议。
	OptSavingsPlans CostOptimizationCategory = "SAVINGS_PLANS"

	// OptStorageOptimization 存储优化（S3 生命周期、EBS 卷类型）。
	OptStorageOptimization CostOptimizationCategory = "STORAGE_OPTIMIZATION"

	// OptNetworkOptimization 网络优化（NAT Gateway、数据传输）。
	OptNetworkOptimization CostOptimizationCategory = "NETWORK_OPTIMIZATION"

	// OptLicenseOptimization 许可证优化（BYOL、开源替代）。
	OptLicenseOptimization CostOptimizationCategory = "LICENSE_OPTIMIZATION"

	// OptArchitectureChange 架构变更（无服务器化、容器化）。
	OptArchitectureChange CostOptimizationCategory = "ARCHITECTURE_CHANGE"
)

// ServiceCostBreakdown 表示单个 AWS 服务的费用明细。
//
// 字段：
//   - ServiceName    : AWS 服务名称（如 "Amazon EC2"、"Amazon S3"）
//   - CurrentCost    : 当前周期费用
//   - PreviousCost   : 上一周期费用（用于对比）
//   - CostChange     : 费用变化金额（当前 - 上一周期）
//   - ChangePct      : 费用变化百分比
//   - CostPercentage : 占总费用的百分比
//   - Trend          : 费用趋势方向
//   - Unit           : 货币单位
type ServiceCostBreakdown struct {
	ServiceName    string         `json:"service_name"`
	CurrentCost    float64        `json:"current_cost"`
	PreviousCost   float64        `json:"previous_cost"`
	CostChange     float64        `json:"cost_change"`
	ChangePct      float64        `json:"change_pct"`
	CostPercentage float64        `json:"cost_percentage"`
	Trend          TrendDirection `json:"trend"`
	Unit           string         `json:"unit"`
}

// RegionCostBreakdown 表示按区域划分的费用明细。
//
// 字段：
//   - Region       : AWS 区域
//   - CurrentCost  : 当前周期费用
//   - PreviousCost : 上一周期费用
//   - CostChange   : 费用变化金额
//   - ChangePct    : 费用变化百分比
//   - Unit         : 货币单位
type RegionCostBreakdown struct {
	Region       string  `json:"region"`
	CurrentCost  float64 `json:"current_cost"`
	PreviousCost float64 `json:"previous_cost"`
	CostChange   float64 `json:"cost_change"`
	ChangePct    float64 `json:"change_pct"`
	Unit         string  `json:"unit"`
}

// UsageComparison 表示两个时间段之间的费用使用量对比结果。
//
// 字段：
//   - BaselinePeriod    : 基线（历史参照）时间段
//   - ComparisonPeriod  : 对比（当前）时间段
//   - BaselineTotalCost : 基线期总费用
//   - ComparisonTotalCost : 对比期总费用
//   - TotalDifference   : 总差额
//   - TotalChangePct    : 总变化百分比
//   - ServiceBreakdowns : 按服务的费用明细列表
//   - RegionBreakdowns  : 按区域的费用明细列表
//   - TopIncreases      : 费用增长最多的前 N 项
//   - TopDecreases      : 费用降低最多的前 N 项
//   - Unit              : 货币单位
type UsageComparison struct {
	BaselinePeriod      AnalysisPeriod         `json:"baseline_period"`
	ComparisonPeriod    AnalysisPeriod         `json:"comparison_period"`
	BaselineTotalCost   float64                `json:"baseline_total_cost"`
	ComparisonTotalCost float64                `json:"comparison_total_cost"`
	TotalDifference     float64                `json:"total_difference"`
	TotalChangePct      float64                `json:"total_change_pct"`
	ServiceBreakdowns   []ServiceCostBreakdown `json:"service_breakdowns"`
	RegionBreakdowns    []RegionCostBreakdown  `json:"region_breakdowns,omitempty"`
	TopIncreases        []ServiceCostBreakdown `json:"top_increases"`
	TopDecreases        []ServiceCostBreakdown `json:"top_decreases"`
	Unit                string                 `json:"unit"`
}

// CostAnomaly 表示费用异常检测的结果。
//
// 字段：
//   - DetectedAt    : 检测时间
//   - Service       : 异常所属服务
//   - Region        : 异常所属区域
//   - ExpectedCost  : 预期费用
//   - ActualCost    : 实际费用
//   - Deviation     : 偏差金额
//   - DeviationPct  : 偏差百分比
//   - Severity      : 严重程度
//   - Description   : 异常描述
//   - Unit          : 货币单位
type CostAnomaly struct {
	DetectedAt   time.Time     `json:"detected_at"`
	Service      string        `json:"service"`
	Region       string        `json:"region"`
	ExpectedCost float64       `json:"expected_cost"`
	ActualCost   float64       `json:"actual_cost"`
	Deviation    float64       `json:"deviation"`
	DeviationPct float64       `json:"deviation_pct"`
	Severity     SeverityLevel `json:"severity"`
	Description  string        `json:"description"`
	Unit         string        `json:"unit"`
}

// OptimizationRecommendation 表示一条费用优化建议。
//
// 字段：
//   - Category              : 优化类别
//   - Severity              : 严重程度 / 优先级
//   - Title                 : 建议标题
//   - Description           : 详细描述
//   - EstimatedMonthlySavings : 预估月度节省金额
//   - EstimatedAnnualSavings  : 预估年度节省金额
//   - Effort                : 实施难度（LOW / MEDIUM / HIGH）
//   - AffectedResources     : 受影响的资源 ID 列表
//   - AffectedServices      : 受影响的服务名称列表
//   - Unit                  : 货币单位
type OptimizationRecommendation struct {
	Category               CostOptimizationCategory `json:"category"`
	Severity               SeverityLevel            `json:"severity"`
	Title                  string                   `json:"title"`
	Description            string                   `json:"description"`
	EstimatedMonthlySavings float64                 `json:"estimated_monthly_savings"`
	EstimatedAnnualSavings  float64                 `json:"estimated_annual_savings"`
	Effort                 string                   `json:"effort"`
	AffectedResources      []string                 `json:"affected_resources,omitempty"`
	AffectedServices       []string                 `json:"affected_services,omitempty"`
	Unit                   string                   `json:"unit"`
}

// ForecastSummary 表示费用预测的汇总信息。
//
// 字段：
//   - ForecastPeriod    : 预测覆盖的时间范围
//   - ForecastedTotal   : 预测总费用
//   - LowerBound        : 预测下界
//   - UpperBound        : 预测上界
//   - ConfidenceLevel   : 置信度百分比（51-99）
//   - ComparedToBudget  : 与预算的差额（正数表示预测超预算）
//   - BudgetUtilization : 预测费用占预算的百分比
//   - Unit              : 货币单位
type ForecastSummary struct {
	ForecastPeriod    AnalysisPeriod `json:"forecast_period"`
	ForecastedTotal   float64        `json:"forecasted_total"`
	LowerBound        float64        `json:"lower_bound"`
	UpperBound        float64        `json:"upper_bound"`
	ConfidenceLevel   int            `json:"confidence_level"`
	ComparedToBudget  float64        `json:"compared_to_budget"`
	BudgetUtilization float64        `json:"budget_utilization"`
	Unit              string         `json:"unit"`
}

// DailyCostTrend 表示按天追踪的费用趋势数据点。
//
// 字段：
//   - Date      : 日期
//   - Cost      : 当日费用
//   - Cumulative: 累计费用（从周期起始到当日）
//   - Unit      : 货币单位
type DailyCostTrend struct {
	Date       time.Time `json:"date"`
	Cost       float64   `json:"cost"`
	Cumulative float64   `json:"cumulative"`
	Unit       string    `json:"unit"`
}

// CostAnalysisReport 是完整的费用分析报告，整合所有分析维度。
// 这是分析引擎的最终输出结构，前端展示或通知推送均基于此结构。
//
// 字段：
//   - ReportID           : 报告唯一标识
//   - GeneratedAt        : 报告生成时间
//   - AnalysisPeriod     : 分析涵盖的时间范围
//   - AccountID          : AWS 账户 ID
//   - TotalCurrentCost   : 当前周期总费用
//   - TotalPreviousCost  : 上一周期总费用
//   - CostTrend          : 费用趋势方向
//   - Comparison         : 周期间对比结果
//   - Forecast           : 费用预测汇总
//   - BudgetPerformances : 各预算的执行情况
//   - Anomalies          : 检测到的费用异常列表
//   - Recommendations    : 费用优化建议列表
//   - DailyTrend         : 按天的费用趋势数据
//   - TotalEstimatedSavings : 所有优化建议的预估总节省金额
//   - Unit               : 货币单位
type CostAnalysisReport struct {
	ReportID              string                       `json:"report_id"`
	GeneratedAt           time.Time                    `json:"generated_at"`
	AnalysisPeriod        AnalysisPeriod               `json:"analysis_period"`
	AccountID             string                       `json:"account_id"`
	TotalCurrentCost      float64                      `json:"total_current_cost"`
	TotalPreviousCost     float64                      `json:"total_previous_cost"`
	CostTrend             TrendDirection               `json:"cost_trend"`
	Comparison            *UsageComparison             `json:"comparison,omitempty"`
	Forecast              *ForecastSummary             `json:"forecast,omitempty"`
	BudgetPerformances    []BudgetPerformance          `json:"budget_performances,omitempty"`
	Anomalies             []CostAnomaly                `json:"anomalies,omitempty"`
	Recommendations       []OptimizationRecommendation `json:"recommendations,omitempty"`
	DailyTrend            []DailyCostTrend             `json:"daily_trend,omitempty"`
	TotalEstimatedSavings float64                      `json:"total_estimated_savings"`
	Unit                  string                       `json:"unit"`
}

// AnalysisConfig 是分析引擎的配置参数。
//
// 字段：
//   - ComparisonMonths       : 对比月份数（默认 1，即上月 vs 当月）
//   - ForecastDays           : 预测天数（默认 30）
//   - AnomalyThresholdPct    : 异常检测阈值百分比（默认 20%）
//   - TopNServices           : Top N 服务数量（默认 10）
//   - IncludeRightsizing     : 是否包含 Rightsizing 建议
//   - IncludeForecast        : 是否包含费用预测
//   - IncludeAnomalyDetection: 是否包含异常检测
//   - ConfidenceLevel        : 预测置信度（默认 80）
//   - Metric                 : 使用的费用指标（默认 UnblendedCost）
//   - Granularity            : 趋势数据的时间粒度
type AnalysisConfig struct {
	ComparisonMonths        int     `json:"comparison_months"`
	ForecastDays            int     `json:"forecast_days"`
	AnomalyThresholdPct     float64 `json:"anomaly_threshold_pct"`
	TopNServices            int     `json:"top_n_services"`
	IncludeRightsizing      bool    `json:"include_rightsizing"`
	IncludeForecast         bool    `json:"include_forecast"`
	IncludeAnomalyDetection bool    `json:"include_anomaly_detection"`
	ConfidenceLevel         int     `json:"confidence_level"`
	Metric                  string  `json:"metric"`
	Granularity             string  `json:"granularity"`
}

// DefaultAnalysisConfig 返回分析引擎的默认配置。
//
// 返回：
//   - AnalysisConfig : 包含推荐默认值的配置实例
func DefaultAnalysisConfig() AnalysisConfig {
	return AnalysisConfig{
		ComparisonMonths:        1,
		ForecastDays:            30,
		AnomalyThresholdPct:    20.0,
		TopNServices:            10,
		IncludeRightsizing:      true,
		IncludeForecast:         true,
		IncludeAnomalyDetection: true,
		ConfidenceLevel:         80,
		Metric:                  "UnblendedCost",
		Granularity:             "DAILY",
	}
}
