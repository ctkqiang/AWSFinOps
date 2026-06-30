package aws

type (
	BillingAccountType string
	LineItemType       string
)

const (
	BillingAccountManagement BillingAccountType = "MANAGEMENT" // payer account
	BillingAccountMember     BillingAccountType = "MEMBER"     // linked account
)

const (
	LineItemUsage                   LineItemType = "Usage"
	LineItemTax                     LineItemType = "Tax"
	LineItemFee                     LineItemType = "Fee"
	LineItemRefund                  LineItemType = "Refund"
	LineItemCredit                  LineItemType = "Credit"
	LineItemRIFee                   LineItemType = "RIFee"
	LineItemDiscountedUsage         LineItemType = "DiscountedUsage"
	LineItemSavingsPlanUpfrontFee   LineItemType = "SavingsPlanUpfrontFee"
	LineItemSavingsPlanRecurringFee LineItemType = "SavingsPlanRecurringFee"
	LineItemSavingsPlanCoveredUsage LineItemType = "SavingsPlanCoveredUsage"
	LineItemSavingsPlanNegation     LineItemType = "SavingsPlanNegation"
	LineItemBundledDiscount         LineItemType = "BundledDiscount"
	LineItemPrivateRateDiscount     LineItemType = "PrivateRateDiscount"
)
