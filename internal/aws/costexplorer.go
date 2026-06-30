package aws

import "time"

// Granularity 定义费用查询的时间粒度。
type Granularity string

const (
	// GranularityDaily 按天聚合费用数据。
	GranularityDaily Granularity = "DAILY"

	// GranularityMonthly 按月聚合费用数据。
	GranularityMonthly Granularity = "MONTHLY"

	// GranularityHourly 按小时聚合费用数据（仅对最近 14 天可用）。
	GranularityHourly Granularity = "HOURLY"
)

// MetricName 定义 Cost Explorer 支持的费用指标名称。
type MetricName string

const (
	// MetricAmortizedCost 摊销成本，将预付费用均摊到每个计费周期。
	MetricAmortizedCost MetricName = "AmortizedCost"

	// MetricBlendedCost 混合成本，按组织内所有账户的加权平均费率计算。
	MetricBlendedCost MetricName = "BlendedCost"

	// MetricNetAmortizedCost 净摊销成本，扣除折扣和抵扣后的摊销金额。
	MetricNetAmortizedCost MetricName = "NetAmortizedCost"

	// MetricNetUnblendedCost 净未混合成本，扣除折扣后各账户的实际费率。
	MetricNetUnblendedCost MetricName = "NetUnblendedCost"

	// MetricNormalizedUsageAmount 归一化用量，将不同实例大小换算为统一单位。
	MetricNormalizedUsageAmount MetricName = "NormalizedUsageAmount"

	// MetricUnblendedCost 未混合成本，各账户按自身实际费率计算。
	MetricUnblendedCost MetricName = "UnblendedCost"

	// MetricUsageQuantity 资源使用量（如 EC2 小时数、S3 存储 GB 数）。
	MetricUsageQuantity MetricName = "UsageQuantity"
)

// DimensionKey 定义 Cost Explorer 支持的筛选与分组维度。
type DimensionKey string

const (
	// DimensionService AWS 服务名称（如 Amazon EC2、Amazon S3）。
	DimensionService DimensionKey = "SERVICE"

	// DimensionLinkedAccount 关联账户 ID。
	DimensionLinkedAccount DimensionKey = "LINKED_ACCOUNT"

	// DimensionRegion AWS 区域（如 us-east-1、ap-southeast-1）。
	DimensionRegion DimensionKey = "REGION"

	// DimensionInstanceType 实例类型（如 m5.xlarge、t3.micro）。
	DimensionInstanceType DimensionKey = "INSTANCE_TYPE"

	// DimensionUsageType 使用类型，精确到具体计费维度。
	DimensionUsageType DimensionKey = "USAGE_TYPE"

	// DimensionPlatform 操作系统平台（如 Linux、Windows）。
	DimensionPlatform DimensionKey = "PLATFORM"

	// DimensionTenancy 租户类型（共享、专用宿主机、专用实例）。
	DimensionTenancy DimensionKey = "TENANCY"

	// DimensionPurchaseType 购买方式（按需、预留、Savings Plans）。
	DimensionPurchaseType DimensionKey = "PURCHASE_TYPE"

	// DimensionDatabase 数据库引擎（如 Aurora MySQL、PostgreSQL）。
	DimensionDatabase DimensionKey = "DATABASE_ENGINE"

	// DimensionAZ 可用区（如 us-east-1a）。
	DimensionAZ DimensionKey = "AZ"

	// DimensionOperatingSystem 操作系统。
	DimensionOperatingSystem DimensionKey = "OPERATING_SYSTEM"

	// DimensionRecordType 记录类型。
	DimensionRecordType DimensionKey = "RECORD_TYPE"

	// DimensionLegalEntity 法律实体名称。
	DimensionLegalEntity DimensionKey = "LEGAL_ENTITY_NAME"
)

// MatchOption 定义维度值匹配方式。
type MatchOption string

const (
	// MatchEquals 精确匹配。
	MatchEquals MatchOption = "EQUALS"

	// MatchAbsent 维度值不存在。
	MatchAbsent MatchOption = "ABSENT"

	// MatchStartsWith 前缀匹配。
	MatchStartsWith MatchOption = "STARTS_WITH"

	// MatchEndsWith 后缀匹配。
	MatchEndsWith MatchOption = "ENDS_WITH"

	// MatchContains 包含匹配。
	MatchContains MatchOption = "CONTAINS"

	// MatchCaseInsensitive 不区分大小写匹配。
	MatchCaseInsensitive MatchOption = "CASE_INSENSITIVE"

	// MatchCaseSensitive 区分大小写匹配。
	MatchCaseSensitive MatchOption = "CASE_SENSITIVE"

	// MatchGreaterThanOrEqual 大于等于（数值维度）。
	MatchGreaterThanOrEqual MatchOption = "GREATER_THAN_OR_EQUAL"
)

// GroupDefinitionType 定义分组定义的类型。
type GroupDefinitionType string

const (
	// GroupByDimension 按 AWS 费用维度分组。
	GroupByDimension GroupDefinitionType = "DIMENSION"

	// GroupByTag 按用户自定义标签分组。
	GroupByTag GroupDefinitionType = "TAG"

	// GroupByCostCategory 按成本类别分组。
	GroupByCostCategory GroupDefinitionType = "COST_CATEGORY"
)

// DateInterval 表示费用查询的时间范围。
// Start 和 End 使用 yyyy-MM-dd 格式，End 为开区间（不包含当天）。
//
// 字段：
//   - Start : 起始日期，格式 yyyy-MM-dd
//   - End   : 结束日期（不含），格式 yyyy-MM-dd
type DateInterval struct {
	Start string `json:"start"`
	End   string `json:"end"`
}

// DimensionValues 表示按维度筛选的条件。
//
// 字段：
//   - Key          : 维度键
//   - Values       : 维度值列表
//   - MatchOptions : 值匹配方式列表
type DimensionValues struct {
	Key          DimensionKey  `json:"key"`
	Values       []string      `json:"values"`
	MatchOptions []MatchOption `json:"match_options,omitempty"`
}

// TagValues 表示按标签筛选的条件。
//
// 字段：
//   - Key          : 标签键
//   - Values       : 标签值列表
//   - MatchOptions : 值匹配方式列表
type TagValues struct {
	Key          string        `json:"key"`
	Values       []string      `json:"values"`
	MatchOptions []MatchOption `json:"match_options,omitempty"`
}

// CostCategoryValues 表示按成本类别筛选的条件。
//
// 字段：
//   - Key          : 成本类别名称
//   - Values       : 类别值列表
//   - MatchOptions : 值匹配方式列表
type CostCategoryValues struct {
	Key          string        `json:"key"`
	Values       []string      `json:"values"`
	MatchOptions []MatchOption `json:"match_options,omitempty"`
}

// Expression 表示 Cost Explorer 的过滤表达式，支持嵌套组合。
// 逻辑运算符 And、Or 为列表形式，Not 为单个 Expression 指针。
// 叶子节点为 Dimensions、Tags 或 CostCategories。
//
// 字段：
//   - And            : 所有子表达式都必须满足（逻辑与）
//   - Or             : 任一子表达式满足即可（逻辑或）
//   - Not            : 取反
//   - Dimensions     : 维度筛选条件
//   - Tags           : 标签筛选条件
//   - CostCategories : 成本类别筛选条件
type Expression struct {
	And            []Expression         `json:"and,omitempty"`
	Or             []Expression         `json:"or,omitempty"`
	Not            *Expression          `json:"not,omitempty"`
	Dimensions     *DimensionValues     `json:"dimensions,omitempty"`
	Tags           *TagValues           `json:"tags,omitempty"`
	CostCategories *CostCategoryValues  `json:"cost_categories,omitempty"`
}

// GroupDefinition 定义 Cost Explorer 查询结果的分组方式。
//
// 字段：
//   - Key  : 分组键名（维度名、标签键或成本类别名）
//   - Type : 分组类型（DIMENSION / TAG / COST_CATEGORY）
type GroupDefinition struct {
	Key  string              `json:"key"`
	Type GroupDefinitionType `json:"type"`
}

// MetricValue 表示某个指标的数值和单位。
//
// 字段：
//   - Amount : 数值金额（字符串以保持精度）
//   - Unit   : 单位（如 "USD"、"Hours"、"GB"）
type MetricValue struct {
	Amount string `json:"amount"`
	Unit   string `json:"unit"`
}

// Group 表示按分组维度聚合后的一条结果记录。
//
// 字段：
//   - Keys    : 分组键值列表（对应 GroupDefinition 的顺序）
//   - Metrics : 各指标的数值映射，键为 MetricName
type Group struct {
	Keys    []string               `json:"keys"`
	Metrics map[string]MetricValue `json:"metrics"`
}

// ResultByTime 表示按时间周期聚合的费用结果。
//
// 字段：
//   - TimePeriod : 对应的时间区间
//   - Total      : 无分组时的汇总指标
//   - Groups     : 有分组时的各组结果
//   - Estimated  : 是否为估算值（月末尚未出账时为 true）
type ResultByTime struct {
	TimePeriod DateInterval           `json:"time_period"`
	Total      map[string]MetricValue `json:"total,omitempty"`
	Groups     []Group                `json:"groups,omitempty"`
	Estimated  bool                   `json:"estimated"`
}

// GetCostAndUsageRequest 是 GetCostAndUsage API 的请求结构体。
//
// 字段：
//   - TimePeriod       : 查询时间范围（必填）
//   - Granularity      : 时间粒度（DAILY / MONTHLY / HOURLY，必填）
//   - Metrics          : 需要返回的指标列表（必填）
//   - Filter           : 过滤表达式（可选）
//   - GroupBy          : 分组定义，最多 2 个（可选）
//   - NextPageToken    : 分页令牌，用于获取下一页（可选）
type GetCostAndUsageRequest struct {
	TimePeriod    DateInterval      `json:"time_period"`
	Granularity   Granularity       `json:"granularity"`
	Metrics       []MetricName      `json:"metrics"`
	Filter        *Expression       `json:"filter,omitempty"`
	GroupBy       []GroupDefinition  `json:"group_by,omitempty"`
	NextPageToken string            `json:"next_page_token,omitempty"`
}

// GetCostAndUsageResponse 是 GetCostAndUsage API 的响应结构体。
//
// 字段：
//   - ResultsByTime    : 按时间周期排列的费用结果列表
//   - DimensionValueAttributes : 维度值的扩展属性
//   - GroupDefinitions : 实际使用的分组定义（与请求中的对应）
//   - NextPageToken    : 下一页令牌，为空表示无更多数据
type GetCostAndUsageResponse struct {
	ResultsByTime            []ResultByTime  `json:"results_by_time"`
	DimensionValueAttributes []DimensionValueAttribute `json:"dimension_value_attributes,omitempty"`
	GroupDefinitions         []GroupDefinition          `json:"group_definitions,omitempty"`
	NextPageToken            string                     `json:"next_page_token,omitempty"`
}

// DimensionValueAttribute 表示维度值的扩展属性。
//
// 字段：
//   - Value      : 维度值
//   - Attributes : 附加属性映射（如服务名称的全称）
type DimensionValueAttribute struct {
	Value      string            `json:"value"`
	Attributes map[string]string `json:"attributes,omitempty"`
}

// ForecastResult 表示单个时间段的费用预测结果。
//
// 字段：
//   - TimePeriod                     : 预测覆盖的时间区间
//   - MeanValue                      : 预测均值
//   - PredictionIntervalLowerBound   : 预测区间下界
//   - PredictionIntervalUpperBound   : 预测区间上界
type ForecastResult struct {
	TimePeriod                   DateInterval `json:"time_period"`
	MeanValue                    string       `json:"mean_value"`
	PredictionIntervalLowerBound string       `json:"prediction_interval_lower_bound"`
	PredictionIntervalUpperBound string       `json:"prediction_interval_upper_bound"`
}

// GetCostForecastRequest 是 GetCostForecast API 的请求结构体。
//
// 字段：
//   - TimePeriod              : 预测时间范围（必填，Start 不可早于今天）
//   - Granularity             : 时间粒度（DAILY / MONTHLY，必填）
//   - Metric                  : 预测的费用指标（必填）
//   - PredictionIntervalLevel : 预测置信区间水平，取值 51-99（可选，默认 80）
//   - Filter                  : 过滤表达式（可选）
type GetCostForecastRequest struct {
	TimePeriod              DateInterval `json:"time_period"`
	Granularity             Granularity  `json:"granularity"`
	Metric                  MetricName   `json:"metric"`
	PredictionIntervalLevel int          `json:"prediction_interval_level,omitempty"`
	Filter                  *Expression  `json:"filter,omitempty"`
}

// GetCostForecastResponse 是 GetCostForecast API 的响应结构体。
//
// 字段：
//   - ForecastResultsByTime : 按时间周期排列的预测结果列表
//   - Total                 : 整个预测时间范围的汇总值
type GetCostForecastResponse struct {
	ForecastResultsByTime []ForecastResult `json:"forecast_results_by_time"`
	Total                 MetricValue      `json:"total"`
}

// ComparisonMetricValue 表示基线期与对比期之间的费用对比指标。
//
// 字段：
//   - BaselineTimePeriodAmount   : 基线期金额
//   - ComparisonTimePeriodAmount : 对比期金额
//   - Difference                 : 差额（对比期 - 基线期）
//   - Unit                       : 单位
type ComparisonMetricValue struct {
	BaselineTimePeriodAmount   string `json:"baseline_time_period_amount"`
	ComparisonTimePeriodAmount string `json:"comparison_time_period_amount"`
	Difference                 string `json:"difference"`
	Unit                       string `json:"unit"`
}

// CostAndUsageSelector 表示费用对比中的选择器（标识哪组数据在对比）。
//
// 字段：
//   - Key   : 选择器键（如 SERVICE、REGION）
//   - Value : 选择器值（如 "Amazon EC2"）
type CostAndUsageSelector struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

// CostAndUsageComparison 表示一条费用对比结果。
//
// 字段：
//   - CostAndUsageSelector : 对比维度的选择器
//   - Metrics              : 各指标的对比值映射
type CostAndUsageComparison struct {
	CostAndUsageSelector CostAndUsageSelector          `json:"cost_and_usage_selector"`
	Metrics              map[string]ComparisonMetricValue `json:"metrics"`
}

// ComparisonMetricForComparison 定义对比查询中使用的指标。
type ComparisonMetricForComparison string

const (
	// CompareByUnblendedCost 使用未混合成本进行对比。
	CompareByUnblendedCost ComparisonMetricForComparison = "UnblendedCost"

	// CompareByNetUnblendedCost 使用净未混合成本进行对比。
	CompareByNetUnblendedCost ComparisonMetricForComparison = "NetUnblendedCost"

	// CompareByAmortizedCost 使用摊销成本进行对比。
	CompareByAmortizedCost ComparisonMetricForComparison = "AmortizedCost"

	// CompareByNetAmortizedCost 使用净摊销成本进行对比。
	CompareByNetAmortizedCost ComparisonMetricForComparison = "NetAmortizedCost"
)

// GetCostAndUsageComparisonsRequest 是 GetCostAndUsageComparisons API 的请求结构体。
// 用于将基线期费用与对比期费用进行对比分析。
//
// 字段：
//   - BaselineTimePeriod       : 基线（历史参照）时间范围（必填）
//   - ComparisonTimePeriod     : 对比（当前或目标）时间范围（必填）
//   - MetricForComparison      : 对比使用的指标（必填）
//   - Filter                   : 过滤表达式（可选）
//   - GroupBy                  : 分组定义，最多 1 个（可选）
//   - MaxResults               : 最大返回条数（可选）
//   - NextPageToken            : 分页令牌（可选）
type GetCostAndUsageComparisonsRequest struct {
	BaselineTimePeriod   DateInterval                  `json:"baseline_time_period"`
	ComparisonTimePeriod DateInterval                  `json:"comparison_time_period"`
	MetricForComparison  ComparisonMetricForComparison `json:"metric_for_comparison"`
	Filter               *Expression                   `json:"filter,omitempty"`
	GroupBy              []GroupDefinition              `json:"group_by,omitempty"`
	MaxResults           int                           `json:"max_results,omitempty"`
	NextPageToken        string                        `json:"next_page_token,omitempty"`
}

// GetCostAndUsageComparisonsResponse 是 GetCostAndUsageComparisons API 的响应结构体。
//
// 字段：
//   - CostAndUsageComparisons : 各维度的对比结果列表
//   - TotalCostAndUsage       : 汇总的对比结果
//   - NextPageToken           : 下一页令牌
type GetCostAndUsageComparisonsResponse struct {
	CostAndUsageComparisons []CostAndUsageComparison `json:"cost_and_usage_comparisons"`
	TotalCostAndUsage       *CostAndUsageComparison  `json:"total_cost_and_usage,omitempty"`
	NextPageToken           string                   `json:"next_page_token,omitempty"`
}

// CostRecord 表示从 Cost Explorer 查询回来后存储在内存或数据库中的单条费用记录。
// 这是从 AWS API 响应转换为内部业务模型的中间结构。
//
// 字段：
//   - Date        : 费用归属日期
//   - Service     : AWS 服务名称
//   - Region      : AWS 区域
//   - AccountID   : 关联账户 ID
//   - MetricName  : 指标名称
//   - Amount      : 金额（浮点数，内部计算用）
//   - Unit        : 单位
//   - IsEstimated : 是否为估算值
type CostRecord struct {
	Date        time.Time  `json:"date"`
	Service     string     `json:"service"`
	Region      string     `json:"region"`
	AccountID   string     `json:"account_id"`
	MetricName  MetricName `json:"metric_name"`
	Amount      float64    `json:"amount"`
	Unit        string     `json:"unit"`
	IsEstimated bool       `json:"is_estimated"`
}
