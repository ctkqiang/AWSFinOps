package test

import (
	"sync"
	"testing"

	"aws_fin_ops/internal/billing"
)

// TestNewBudgetState_Success 验证合法参数能正确创建 BudgetState。
func TestNewBudgetState_Success(t *testing.T) {
	bs, err := billing.NewBudgetState(1000, "USD", billing.BudgetTypeCost)
	if err != nil {
		t.Fatalf("期望无错误，实际得到 %v", err)
	}
	amount, _ := bs.GetAmount()
	if amount != 1000 {
		t.Fatalf("期望金额 1000，实际得到 %.2f", float64(amount))
	}
	currency, _ := bs.GetCurrency()
	if currency != "USD" {
		t.Fatalf("期望货币 USD，实际得到 %s", currency)
	}
	bt, _ := bs.GetBudgetType()
	if bt != billing.BudgetTypeCost {
		t.Fatalf("期望类型 COST，实际得到 %s", bt)
	}
}

// TestNewBudgetState_ZeroAmount 验证金额为 0 时返回错误。
func TestNewBudgetState_ZeroAmount(t *testing.T) {
	_, err := billing.NewBudgetState(0, "USD", billing.BudgetTypeCost)
	if err == nil {
		t.Fatal("期望金额为 0 时返回错误")
	}
}

// TestNewBudgetState_NegativeAmount 验证负数金额时返回错误。
func TestNewBudgetState_NegativeAmount(t *testing.T) {
	_, err := billing.NewBudgetState(-500, "USD", billing.BudgetTypeCost)
	if err == nil {
		t.Fatal("期望负数金额时返回错误")
	}
}

// TestNewBudgetState_ExceedsMax 验证超过最大限额（1,000,000）时返回错误。
func TestNewBudgetState_ExceedsMax(t *testing.T) {
	_, err := billing.NewBudgetState(1_000_001, "USD", billing.BudgetTypeCost)
	if err == nil {
		t.Fatal("期望超过最大限额时返回错误")
	}
}

// TestNewBudgetState_EmptyCurrency 验证空货币代码时返回错误。
func TestNewBudgetState_EmptyCurrency(t *testing.T) {
	_, err := billing.NewBudgetState(1000, "", billing.BudgetTypeCost)
	if err == nil {
		t.Fatal("期望空货币代码时返回错误")
	}
}

// TestNewBudgetState_InvalidType 验证无效预算类型时返回错误。
func TestNewBudgetState_InvalidType(t *testing.T) {
	_, err := billing.NewBudgetState(1000, "USD", "INVALID")
	if err == nil {
		t.Fatal("期望无效预算类型时返回错误")
	}
}

// TestNewBudgetState_AllTypes 验证所有合法预算类型均可成功创建。
func TestNewBudgetState_AllTypes(t *testing.T) {
	types := []billing.BudgetType{
		billing.BudgetTypeCost,
		billing.BudgetTypeUsage,
		billing.BudgetTypeRIUtilization,
		billing.BudgetTypeRICoverage,
		billing.BudgetTypeSPUtilization,
		billing.BudgetTypeSPCoverage,
	}
	for _, bt := range types {
		_, err := billing.NewBudgetState(100, "USD", bt)
		if err != nil {
			t.Fatalf("预算类型 %s 应当合法，实际得到错误: %v", bt, err)
		}
	}
}

// TestSetAmount_Success 验证正常更新金额。
func TestSetAmount_Success(t *testing.T) {
	bs, _ := billing.NewBudgetState(1000, "USD", billing.BudgetTypeCost)

	if err := bs.SetAmount(2000); err != nil {
		t.Fatalf("期望无错误，实际得到 %v", err)
	}
	amount, _ := bs.GetAmount()
	if amount != 2000 {
		t.Fatalf("期望金额 2000，实际得到 %.2f", float64(amount))
	}
}

// TestSetAmount_InvalidAmount 验证设置非法金额时返回错误且原值不变。
func TestSetAmount_InvalidAmount(t *testing.T) {
	bs, _ := billing.NewBudgetState(1000, "USD", billing.BudgetTypeCost)

	if err := bs.SetAmount(-100); err == nil {
		t.Fatal("期望负数金额时返回错误")
	}
	amount, _ := bs.GetAmount()
	if amount != 1000 {
		t.Fatalf("期望原值 1000 不变，实际得到 %.2f", float64(amount))
	}
}

// TestSetAmount_ExceedsMax 验证设置超过最大限额时返回错误且原值不变。
func TestSetAmount_ExceedsMax(t *testing.T) {
	bs, _ := billing.NewBudgetState(1000, "USD", billing.BudgetTypeCost)

	if err := bs.SetAmount(1_000_001); err == nil {
		t.Fatal("期望超过最大限额时返回错误")
	}
	amount, _ := bs.GetAmount()
	if amount != 1000 {
		t.Fatalf("期望原值 1000 不变，实际得到 %.2f", float64(amount))
	}
}

// TestGetUpdatedAt 验证更新金额后时间戳发生变化。
func TestGetUpdatedAt(t *testing.T) {
	bs, _ := billing.NewBudgetState(1000, "USD", billing.BudgetTypeCost)
	t1, _ := bs.GetUpdatedAt()

	bs.SetAmount(2000)
	t2, _ := bs.GetUpdatedAt()

	if !t2.After(t1) && t2 != t1 {
		t.Fatal("期望 SetAmount 后 updatedAt 时间戳更新")
	}
}

// TestConcurrentSetAmount 验证并发写入时数据一致性（无 data race）。
func TestConcurrentSetAmount(t *testing.T) {
	bs, _ := billing.NewBudgetState(1000, "USD", billing.BudgetTypeCost)

	var wg sync.WaitGroup
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func(val billing.Budget) {
			defer wg.Done()
			bs.SetAmount(val)
		}(billing.Budget(i + 1))
	}
	wg.Wait()

	amount, err := bs.GetAmount()
	if err != nil {
		t.Fatalf("期望无错误，实际得到 %v", err)
	}
	if amount <= 0 || amount > 100 {
		t.Fatalf("期望金额在 1-100 之间，实际得到 %.2f", float64(amount))
	}
}

// TestMainConstants 验证 main.go 中定义的预算常量能正确创建 BudgetState。
func TestMainConstants(t *testing.T) {
	const (
		budget   billing.Budget         = 500
		currency billing.BudgetCurrency = "USD"
	)
	budgetType := billing.BudgetTypeCost

	bs, err := billing.NewBudgetState(budget, currency, budgetType)
	if err != nil {
		t.Fatalf("main.go 常量应能正确初始化 BudgetState，实际得到 %v", err)
	}

	amount, _ := bs.GetAmount()
	if amount != 500 {
		t.Fatalf("期望金额 500，实际得到 %.2f", float64(amount))
	}
	cur, _ := bs.GetCurrency()
	if cur != "USD" {
		t.Fatalf("期望货币 USD，实际得到 %s", cur)
	}
	bt, _ := bs.GetBudgetType()
	if bt != billing.BudgetTypeCost {
		t.Fatalf("期望类型 COST，实际得到 %s", bt)
	}
}
