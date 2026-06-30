package billing

import (
	"aws_fin_ops/internal/utilities"
	"fmt"
	"sync"
	"time"
)

type (
	Budget         float64
	BudgetCurrency string
	BudgetType     string
)

const (
	// 基于费用的预算
	BudgetTypeCost BudgetType = "COST"

	// 基于用量的预算
	BudgetTypeUsage BudgetType = "USAGE"

	// 预留实例（RI）预算
	BudgetTypeRIUtilization BudgetType = "RI_UTILIZATION"
	BudgetTypeRICoverage    BudgetType = "RI_COVERAGE"

	// Savings Plans 预算
	BudgetTypeSPUtilization BudgetType = "SAVINGS_PLANS_UTILIZATION"
	BudgetTypeSPCoverage    BudgetType = "SAVINGS_PLANS_COVERAGE"
)

const (
	// billingComponent 是本文件在日志中使用的组件名称标识。
	billingComponent = "Billing"

	// maxBudget 是允许设置的最大预算金额（100 万美元），防止误输入。
	maxBudget Budget = 1_000_000
)

// BudgetState 管理单个预算的完整生命周期状态。
// 所有字段通过 mu 互斥锁保护，支持并发安全的读写操作。
type BudgetState struct {
	mu          sync.RWMutex
	amount      Budget
	currency    BudgetCurrency
	budgetType  BudgetType
	initialized bool
	updatedAt   time.Time
}

// NewBudgetState 创建并初始化一个 BudgetState 实例。
//
// 参数：
//   - amount     : 预算金额，必须 > 0 且 <= maxBudget
//   - currency   : 货币代码（如 "USD"、"CNY"），不可为空
//   - budgetType : 预算类型，必须是已定义的 BudgetType 常量之一
//
// 返回：
//   - *BudgetState : 初始化完成的预算状态实例
//   - error        : 参数校验失败时返回描述性错误
func NewBudgetState(amount Budget, currency BudgetCurrency, budgetType BudgetType) (*BudgetState, error) {
	start := time.Now()
	utilities.LogStart(billingComponent, "NewBudgetState")

	if err := validateBudgetAmount(amount); err != nil {
		utilities.LogError(billingComponent, "NewBudgetState", err, time.Since(start))
		return nil, err
	}

	if currency == "" {
		err := fmt.Errorf("currency 不可为空")
		utilities.LogError(billingComponent, "NewBudgetState", err, time.Since(start))
		return nil, err
	}

	if !isValidBudgetType(budgetType) {
		err := fmt.Errorf("无效的预算类型: %q", budgetType)
		utilities.LogError(billingComponent, "NewBudgetState", err, time.Since(start))
		return nil, err
	}

	bs := &BudgetState{
		amount:      amount,
		currency:    currency,
		budgetType:  budgetType,
		initialized: true,
		updatedAt:   time.Now(),
	}

	utilities.LogSuccess(billingComponent, "NewBudgetState", time.Since(start),
		fmt.Sprintf("amount=%.2f", float64(amount)),
		fmt.Sprintf("currency=%s", currency),
		fmt.Sprintf("type=%s", budgetType),
	)
	return bs, nil
}

// SetAmount 更新预算金额。
// 调用前 BudgetState 必须已通过 NewBudgetState 初始化，否则返回错误。
//
// 参数：
//   - amount : 新的预算金额，必须 > 0 且 <= maxBudget
//
// 返回：
//   - error : 未初始化或金额校验失败时返回错误，成功时返回 nil
func (bs *BudgetState) SetAmount(amount Budget) error {
	start := time.Now()

	bs.mu.Lock()
	defer bs.mu.Unlock()

	if !bs.initialized {
		err := fmt.Errorf("BudgetState 未初始化，请先调用 NewBudgetState")
		utilities.LogError(billingComponent, "SetAmount", err, time.Since(start))
		return err
	}

	if err := validateBudgetAmount(amount); err != nil {
		utilities.LogError(billingComponent, "SetAmount", err, time.Since(start),
			fmt.Sprintf("current=%.2f", float64(bs.amount)),
			fmt.Sprintf("requested=%.2f", float64(amount)),
		)
		return err
	}

	old := bs.amount
	bs.amount = amount
	bs.updatedAt = time.Now()

	utilities.LogSuccess(billingComponent, "SetAmount", time.Since(start),
		fmt.Sprintf("old=%.2f", float64(old)),
		fmt.Sprintf("new=%.2f", float64(amount)),
		fmt.Sprintf("currency=%s", bs.currency),
	)
	return nil
}

// GetAmount 返回当前预算金额。并发安全。
//
// 返回：
//   - Budget : 当前预算金额
//   - error  : 未初始化时返回错误
func (bs *BudgetState) GetAmount() (Budget, error) {
	bs.mu.RLock()
	defer bs.mu.RUnlock()

	if !bs.initialized {
		return 0, fmt.Errorf("BudgetState 未初始化，请先调用 NewBudgetState")
	}
	return bs.amount, nil
}

// GetCurrency 返回当前预算的货币代码。并发安全。
//
// 返回：
//   - BudgetCurrency : 货币代码
//   - error          : 未初始化时返回错误
func (bs *BudgetState) GetCurrency() (BudgetCurrency, error) {
	bs.mu.RLock()
	defer bs.mu.RUnlock()

	if !bs.initialized {
		return "", fmt.Errorf("BudgetState 未初始化，请先调用 NewBudgetState")
	}
	return bs.currency, nil
}

// GetBudgetType 返回当前预算类型。并发安全。
//
// 返回：
//   - BudgetType : 预算类型
//   - error      : 未初始化时返回错误
func (bs *BudgetState) GetBudgetType() (BudgetType, error) {
	bs.mu.RLock()
	defer bs.mu.RUnlock()

	if !bs.initialized {
		return "", fmt.Errorf("BudgetState 未初始化，请先调用 NewBudgetState")
	}
	return bs.budgetType, nil
}

// GetUpdatedAt 返回最近一次状态更新的时间戳。并发安全。
//
// 返回：
//   - time.Time : 最近更新时间
//   - error     : 未初始化时返回错误
func (bs *BudgetState) GetUpdatedAt() (time.Time, error) {
	bs.mu.RLock()
	defer bs.mu.RUnlock()

	if !bs.initialized {
		return time.Time{}, fmt.Errorf("BudgetState 未初始化，请先调用 NewBudgetState")
	}
	return bs.updatedAt, nil
}

// validateBudgetAmount 校验预算金额是否在合法范围内。
//
// 参数：
//   - amount : 待校验的预算金额
//
// 返回：
//   - error : 金额 <= 0 或 > maxBudget 时返回错误
func validateBudgetAmount(amount Budget) error {
	if amount <= 0 {
		return fmt.Errorf("预算金额必须大于 0，当前值: %.2f", float64(amount))
	}
	if amount > maxBudget {
		return fmt.Errorf("预算金额不可超过 %.2f，当前值: %.2f",
			float64(maxBudget), float64(amount))
	}
	return nil
}

// isValidBudgetType 检查给定的预算类型是否为已定义的合法值。
//
// 参数：
//   - bt : 待校验的预算类型
//
// 返回：
//   - bool : 合法时返回 true
func isValidBudgetType(bt BudgetType) bool {
	switch bt {
	case BudgetTypeCost,
		BudgetTypeUsage,
		BudgetTypeRIUtilization,
		BudgetTypeRICoverage,
		BudgetTypeSPUtilization,
		BudgetTypeSPCoverage:
		return true
	default:
		return false
	}
}
