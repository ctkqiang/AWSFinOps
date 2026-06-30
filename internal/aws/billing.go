package aws

import "time"

type (
	// BillingAccountType 标识 AWS 账户在组织中的角色。
	BillingAccountType string

	// LineItemType 标识 CUR（Cost and Usage Report）中的行项目类型。
	LineItemType string
)

const (
	// BillingAccountManagement 管理账户（付款方账户）。
	BillingAccountManagement BillingAccountType = "MANAGEMENT"

	// BillingAccountMember 成员账户（关联账户）。
	BillingAccountMember BillingAccountType = "MEMBER"
)

const (
	// LineItemUsage 常规使用费。
	LineItemUsage LineItemType = "Usage"

	// LineItemTax 税费。
	LineItemTax LineItemType = "Tax"

	// LineItemFee 服务费。
	LineItemFee LineItemType = "Fee"

	// LineItemRefund 退款。
	LineItemRefund LineItemType = "Refund"

	// LineItemCredit 抵扣额度。
	LineItemCredit LineItemType = "Credit"

	// LineItemRIFee 预留实例费用。
	LineItemRIFee LineItemType = "RIFee"

	// LineItemDiscountedUsage RI 折扣使用。
	LineItemDiscountedUsage LineItemType = "DiscountedUsage"

	// LineItemSavingsPlanUpfrontFee Savings Plans 预付费。
	LineItemSavingsPlanUpfrontFee LineItemType = "SavingsPlanUpfrontFee"

	// LineItemSavingsPlanRecurringFee Savings Plans 周期费。
	LineItemSavingsPlanRecurringFee LineItemType = "SavingsPlanRecurringFee"

	// LineItemSavingsPlanCoveredUsage Savings Plans 覆盖的使用量。
	LineItemSavingsPlanCoveredUsage LineItemType = "SavingsPlanCoveredUsage"

	// LineItemSavingsPlanNegation Savings Plans 抵销项。
	LineItemSavingsPlanNegation LineItemType = "SavingsPlanNegation"

	// LineItemBundledDiscount 捆绑折扣。
	LineItemBundledDiscount LineItemType = "BundledDiscount"

	// LineItemPrivateRateDiscount 私有费率折扣。
	LineItemPrivateRateDiscount LineItemType = "PrivateRateDiscount"
)

// PricingModel 定义 AWS 资源的定价模式。
type PricingModel string

const (
	// PricingOnDemand 按需计费。
	PricingOnDemand PricingModel = "ON_DEMAND"

	// PricingReserved 预留实例。
	PricingReserved PricingModel = "RESERVED"

	// PricingSavingsPlans Savings Plans。
	PricingSavingsPlans PricingModel = "SAVINGS_PLANS"

	// PricingSpot 竞价实例。
	PricingSpot PricingModel = "SPOT"
)

// CurrencyCode 定义支持的货币代码。
type CurrencyCode string

const (
	// CurrencyUSD 美元。
	CurrencyUSD CurrencyCode = "USD"

	// CurrencyCNY 人民币。
	CurrencyCNY CurrencyCode = "CNY"

	// CurrencyJPY 日元。
	CurrencyJPY CurrencyCode = "JPY"

	// CurrencyEUR 欧元。
	CurrencyEUR CurrencyCode = "EUR"

	// CurrencyGBP 英镑。
	CurrencyGBP CurrencyCode = "GBP"

	// CurrencyMYR 马来西亚林吉特。
	CurrencyMYR CurrencyCode = "MYR"
)

// BillingPeriod 表示一个计费周期的时间范围和摘要信息。
//
// 字段：
//   - Start          : 计费周期开始时间
//   - End            : 计费周期结束时间
//   - TotalCost      : 总费用
//   - Currency       : 货币代码
//   - AccountID      : 账户 ID
//   - IsConsolidated : 是否为合并账单
type BillingPeriod struct {
	Start          time.Time    `json:"start"`
	End            time.Time    `json:"end"`
	TotalCost      float64      `json:"total_cost"`
	Currency       CurrencyCode `json:"currency"`
	AccountID      string       `json:"account_id"`
	IsConsolidated bool         `json:"is_consolidated"`
}
