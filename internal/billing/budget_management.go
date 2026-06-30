package billing

import "time"

// BudgetTimeUnit 定义预算的时间单位。
type BudgetTimeUnit string

const (
	// TimeUnitDaily 每日预算。
	TimeUnitDaily BudgetTimeUnit = "DAILY"

	// TimeUnitMonthly 每月预算。
	TimeUnitMonthly BudgetTimeUnit = "MONTHLY"

	// TimeUnitQuarterly 每季度预算。
	TimeUnitQuarterly BudgetTimeUnit = "QUARTERLY"

	// TimeUnitAnnually 每年预算。
	TimeUnitAnnually BudgetTimeUnit = "ANNUALLY"
)

// NotificationType 定义预算通知的触发依据。
type NotificationType string

const (
	// NotificationActual 基于实际花费触发。
	NotificationActual NotificationType = "ACTUAL"

	// NotificationForecasted 基于预测花费触发。
	NotificationForecasted NotificationType = "FORECASTED"
)

// ComparisonOperator 定义预算阈值的比较运算符。
type ComparisonOperator string

const (
	// ComparisonGreaterThan 实际或预测金额大于阈值时触发。
	ComparisonGreaterThan ComparisonOperator = "GREATER_THAN"

	// ComparisonLessThan 实际或预测金额小于阈值时触发。
	ComparisonLessThan ComparisonOperator = "LESS_THAN"

	// ComparisonEqualTo 实际或预测金额等于阈值时触发。
	ComparisonEqualTo ComparisonOperator = "EQUAL_TO"
)

// ThresholdType 定义预算阈值的类型。
type ThresholdType string

const (
	// ThresholdPercentage 以百分比衡量（0-100+）。
	ThresholdPercentage ThresholdType = "PERCENTAGE"

	// ThresholdAbsoluteValue 以绝对金额衡量。
	ThresholdAbsoluteValue ThresholdType = "ABSOLUTE_VALUE"
)

// NotificationState 定义通知的发送状态。
type NotificationState string

const (
	// NotificationOK 未触发阈值，状态正常。
	NotificationOK NotificationState = "OK"

	// NotificationAlarm 已触发阈值，处于告警状态。
	NotificationAlarm NotificationState = "ALARM"
)

// SubscriptionType 定义通知订阅者的接收方式。
type SubscriptionType string

const (
	// SubscriptionEmail 通过电子邮件接收。
	SubscriptionEmail SubscriptionType = "EMAIL"

	// SubscriptionSNS 通过 SNS 主题接收。
	SubscriptionSNS SubscriptionType = "SNS"
)

// AutoAdjustType 定义预算金额的自动调整策略。
type AutoAdjustType string

const (
	// AutoAdjustForecast 根据费用预测自动调整预算。
	AutoAdjustForecast AutoAdjustType = "FORECAST"

	// AutoAdjustHistorical 根据历史费用自动调整预算。
	AutoAdjustHistorical AutoAdjustType = "HISTORICAL"
)

// Spend 表示金额与货币单位的组合，是 AWS Budgets API 的核心计量结构。
//
// 字段：
//   - Amount : 金额数值（字符串以保持精度，最多 15 位）
//   - Unit   : 货币代码（如 "USD"、"CNY"）
type Spend struct {
	Amount string `json:"amount"`
	Unit   string `json:"unit"`
}

// CalculatedSpend 表示当前预算周期内的实际和预测花费。
//
// 字段：
//   - ActualSpend     : 已产生的实际花费
//   - ForecastedSpend : 根据当前趋势预测到周期末的花费
type CalculatedSpend struct {
	ActualSpend     Spend  `json:"actual_spend"`
	ForecastedSpend *Spend `json:"forecasted_spend,omitempty"`
}

// CostTypes 控制预算金额计算时包含或排除的费用类型。
// 每个布尔字段对应一类 AWS 费用，默认值各不相同，需根据业务场景明确设置。
//
// 字段：
//   - IncludeCredit             : 是否包含抵扣额度（默认 true）
//   - IncludeDiscount           : 是否包含折扣（默认 true）
//   - IncludeOtherSubscription  : 是否包含其他订阅费（默认 true）
//   - IncludeRecurring          : 是否包含周期性费用（默认 true）
//   - IncludeRefund             : 是否包含退款（默认 true）
//   - IncludeSubscription       : 是否包含订阅（默认 true）
//   - IncludeSupport            : 是否包含 AWS Support 费用（默认 true）
//   - IncludeTax                : 是否包含税费（默认 true）
//   - IncludeUpfront            : 是否包含预付费用（默认 true）
//   - UseAmortized              : 是否使用摊销成本（默认 false）
//   - UseBlended                : 是否使用混合成本（默认 false）
type CostTypes struct {
	IncludeCredit            *bool `json:"include_credit,omitempty"`
	IncludeDiscount          *bool `json:"include_discount,omitempty"`
	IncludeOtherSubscription *bool `json:"include_other_subscription,omitempty"`
	IncludeRecurring         *bool `json:"include_recurring,omitempty"`
	IncludeRefund            *bool `json:"include_refund,omitempty"`
	IncludeSubscription      *bool `json:"include_subscription,omitempty"`
	IncludeSupport           *bool `json:"include_support,omitempty"`
	IncludeTax               *bool `json:"include_tax,omitempty"`
	IncludeUpfront           *bool `json:"include_upfront,omitempty"`
	UseAmortized             *bool `json:"use_amortized,omitempty"`
	UseBlended               *bool `json:"use_blended,omitempty"`
}

// BudgetTimePeriod 定义预算的生效时间范围。
//
// 字段：
//   - Start : 预算生效的起始时间
//   - End   : 预算失效的结束时间（可选，不设置则持续有效）
type BudgetTimePeriod struct {
	Start time.Time  `json:"start"`
	End   *time.Time `json:"end,omitempty"`
}

// HistoricalOptions 定义基于历史数据自动调整时回溯的周期数。
//
// 字段：
//   - BudgetAdjustmentPeriod : 回溯的预算周期数量（1-60）
//   - LookBackAvailablePeriods : 实际可用的历史周期数量
type HistoricalOptions struct {
	BudgetAdjustmentPeriod   int `json:"budget_adjustment_period"`
	LookBackAvailablePeriods int `json:"look_back_available_periods,omitempty"`
}

// AutoAdjustData 定义预算金额自动调整的策略配置。
//
// 字段：
//   - AutoAdjustType    : 调整策略类型（FORECAST / HISTORICAL）
//   - HistoricalOptions : 历史调整的回溯配置（仅 HISTORICAL 类型使用）
//   - LastAutoAdjustTime: 上次自动调整的时间
type AutoAdjustData struct {
	AutoAdjustType     AutoAdjustType     `json:"auto_adjust_type"`
	HistoricalOptions  *HistoricalOptions `json:"historical_options,omitempty"`
	LastAutoAdjustTime *time.Time         `json:"last_auto_adjust_time,omitempty"`
}

// ManagedBudget 表示 AWS Budgets API 中的完整预算对象。
// 这是 CreateBudget / UpdateBudget / DescribeBudget 操作的核心数据结构。
//
// 字段：
//   - BudgetName         : 预算名称（账户内唯一，必填）
//   - BudgetType         : 预算类型（COST / USAGE / RI / SP，必填）
//   - BudgetLimit        : 预算金额上限（非自动调整时必填）
//   - TimeUnit           : 预算周期单位（DAILY / MONTHLY / QUARTERLY / ANNUALLY，必填）
//   - TimePeriod         : 预算生效的时间范围
//   - CalculatedSpend    : 当前周期的实际与预测花费
//   - CostTypes          : 费用类型包含/排除配置
//   - CostFilters        : 维度筛选（键为维度名，值为维度值列表）
//   - PlannedBudgetLimits: 按周期计划的预算金额（键为月份 yyyy-MM 或 BudgetLimit）
//   - AutoAdjustData     : 自动调整配置（使用后 BudgetLimit 由系统管理）
//   - LastUpdatedTime    : 最后更新时间
type ManagedBudget struct {
	BudgetName          string              `json:"budget_name"`
	BudgetType          BudgetType          `json:"budget_type"`
	BudgetLimit         *Spend              `json:"budget_limit,omitempty"`
	TimeUnit            BudgetTimeUnit      `json:"time_unit"`
	TimePeriod          *BudgetTimePeriod   `json:"time_period,omitempty"`
	CalculatedSpend     *CalculatedSpend    `json:"calculated_spend,omitempty"`
	CostTypes           *CostTypes          `json:"cost_types,omitempty"`
	CostFilters         map[string][]string `json:"cost_filters,omitempty"`
	PlannedBudgetLimits map[string]Spend    `json:"planned_budget_limits,omitempty"`
	AutoAdjustData      *AutoAdjustData     `json:"auto_adjust_data,omitempty"`
	LastUpdatedTime     *time.Time          `json:"last_updated_time,omitempty"`
}

// BudgetNotification 表示预算告警通知规则。
//
// 字段：
//   - ComparisonOperator : 比较运算符（GREATER_THAN / LESS_THAN / EQUAL_TO）
//   - NotificationType   : 通知触发依据（ACTUAL / FORECASTED）
//   - Threshold          : 触发阈值
//   - ThresholdType      : 阈值类型（PERCENTAGE / ABSOLUTE_VALUE）
//   - NotificationState  : 当前通知状态
type BudgetNotification struct {
	ComparisonOperator ComparisonOperator `json:"comparison_operator"`
	NotificationType   NotificationType   `json:"notification_type"`
	Threshold          float64            `json:"threshold"`
	ThresholdType      ThresholdType      `json:"threshold_type"`
	NotificationState  NotificationState  `json:"notification_state,omitempty"`
}

// BudgetSubscriber 表示接收预算通知的订阅者。
//
// 字段：
//   - Address          : 接收地址（邮箱或 SNS ARN）
//   - SubscriptionType : 订阅方式（EMAIL / SNS）
type BudgetSubscriber struct {
	Address          string           `json:"address"`
	SubscriptionType SubscriptionType `json:"subscription_type"`
}

// NotificationWithSubscribers 将通知规则与其订阅者列表绑定。
// 在 CreateBudget 时一起提交，最多 5 个通知，每个通知最多 10 个订阅者。
//
// 字段：
//   - Notification : 通知规则
//   - Subscribers  : 订阅者列表
type NotificationWithSubscribers struct {
	Notification BudgetNotification `json:"notification"`
	Subscribers  []BudgetSubscriber `json:"subscribers"`
}

// CreateBudgetRequest 是 CreateBudget API 的请求结构体。
//
// 字段：
//   - AccountID                   : AWS 账户 ID（必填）
//   - Budget                      : 预算定义（必填）
//   - NotificationsWithSubscribers: 通知规则及订阅者列表（可选，最多 5 个）
//   - ResourceTags                : 预算资源标签
type CreateBudgetRequest struct {
	AccountID                    string                        `json:"account_id"`
	Budget                       ManagedBudget                 `json:"budget"`
	NotificationsWithSubscribers []NotificationWithSubscribers `json:"notifications_with_subscribers,omitempty"`
	ResourceTags                 map[string]string             `json:"resource_tags,omitempty"`
}

// UpdateBudgetRequest 是 UpdateBudget API 的请求结构体。
//
// 字段：
//   - AccountID : AWS 账户 ID（必填）
//   - NewBudget : 更新后的完整预算定义（必填）
type UpdateBudgetRequest struct {
	AccountID string        `json:"account_id"`
	NewBudget ManagedBudget `json:"new_budget"`
}

// DescribeBudgetRequest 是 DescribeBudget API 的请求结构体。
//
// 字段：
//   - AccountID  : AWS 账户 ID（必填）
//   - BudgetName : 预算名称（必填）
type DescribeBudgetRequest struct {
	AccountID  string `json:"account_id"`
	BudgetName string `json:"budget_name"`
}

// DescribeBudgetResponse 是 DescribeBudget API 的响应结构体。
//
// 字段：
//   - Budget : 查询到的预算完整信息
type DescribeBudgetResponse struct {
	Budget ManagedBudget `json:"budget"`
}

// DescribeBudgetsRequest 是 DescribeBudgets（列出所有预算）API 的请求结构体。
//
// 字段：
//   - AccountID  : AWS 账户 ID（必填）
//   - MaxResults : 最大返回条数（1-100）
//   - NextToken  : 分页令牌
type DescribeBudgetsRequest struct {
	AccountID  string `json:"account_id"`
	MaxResults int    `json:"max_results,omitempty"`
	NextToken  string `json:"next_token,omitempty"`
}

// DescribeBudgetsResponse 是 DescribeBudgets API 的响应结构体。
//
// 字段：
//   - Budgets   : 预算列表
//   - NextToken : 下一页令牌
type DescribeBudgetsResponse struct {
	Budgets   []ManagedBudget `json:"budgets"`
	NextToken string          `json:"next_token,omitempty"`
}

// DeleteBudgetRequest 是 DeleteBudget API 的请求结构体。
//
// 字段：
//   - AccountID  : AWS 账户 ID（必填）
//   - BudgetName : 预算名称（必填）
type DeleteBudgetRequest struct {
	AccountID  string `json:"account_id"`
	BudgetName string `json:"budget_name"`
}

// BudgetPerformance 表示预算执行情况的汇总统计。
// 用于内部分析引擎，衡量预算健康度。
//
// 字段：
//   - BudgetName       : 预算名称
//   - BudgetLimit      : 预算上限金额
//   - ActualSpend      : 实际花费金额
//   - ForecastedSpend  : 预测到周期末的花费金额
//   - UtilizationPct   : 预算使用率（ActualSpend / BudgetLimit * 100）
//   - ForecastPct      : 预测使用率（ForecastedSpend / BudgetLimit * 100）
//   - RemainingBudget  : 剩余预算额度
//   - Status           : 预算状态标签（UNDER_BUDGET / NEAR_LIMIT / OVER_BUDGET）
//   - EvaluatedAt      : 评估时间
type BudgetPerformance struct {
	BudgetName      string    `json:"budget_name"`
	BudgetLimit     float64   `json:"budget_limit"`
	ActualSpend     float64   `json:"actual_spend"`
	ForecastedSpend float64   `json:"forecasted_spend"`
	UtilizationPct  float64   `json:"utilization_pct"`
	ForecastPct     float64   `json:"forecast_pct"`
	RemainingBudget float64   `json:"remaining_budget"`
	Status          string    `json:"status"`
	EvaluatedAt     time.Time `json:"evaluated_at"`
}
