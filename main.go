package main

import (
	"aws_fin_ops/internal"
	internalaws "aws_fin_ops/internal/aws"
	"aws_fin_ops/internal/billing"
	"aws_fin_ops/internal/services"
	"aws_fin_ops/internal/utilities"
	"context"
	"os"
)

var (
	// Budget 是初始预算金额（500 美元）。
	// 这是一个示例值，实际使用时应根据业务需求调整。
	Budget billing.Budget = 500

	// BudgetCurrency 是预算使用的货币代码。
	BudgetCurrency billing.BudgetCurrency = "USD"

	// BudgetType 是预算类型，使用基于费用的预算。
	BudgetType = billing.BudgetTypeCost
)

func main() {
	initialContext := context.Background()

	if err := internal.LoadConfig(initialContext); err != nil {
		utilities.Error("配置加载失败: %v", err)
		os.Exit(1)
	}

	budgetState, err := billing.NewBudgetState(Budget, BudgetCurrency, BudgetType)
	if err != nil {
		utilities.Error("预算初始化失败: %v", err)
		os.Exit(1)
	}

	amount, _ := budgetState.GetAmount()
	currency, _ := budgetState.GetCurrency()
	utilities.Info("预算已就绪: %.2f %s", float64(amount), currency)

	if internal.IsLocalMode() {
		utilities.LogProgress("FinOps", "main", "启动本地 HTTP 服务器")
		if err := services.StartLocalServer(budgetState); err != nil {
			utilities.Error("本地服务器启动失败: %v", err)
			os.Exit(1)
		}
	} else {
		utilities.LogProgress("FinOps", "main", "启动 Lambda handler")
		internalaws.StartLambda(budgetState)
	}
}
