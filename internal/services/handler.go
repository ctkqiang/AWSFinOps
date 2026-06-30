package services

import (
	"aws_fin_ops/internal/billing"
	"aws_fin_ops/internal/utilities"
	"encoding/json"
	"fmt"
	"time"
)

// FinOpsRequest 是所有入口（HTTP / Lambda）共用的请求结构体。
//
// 字段：
//   - Action : 操作类型，目前支持 "get_budget"、"set_budget"、"health"
//   - Budget : 当 Action 为 "set_budget" 时，携带新的预算金额
type FinOpsRequest struct {
	Action string  `json:"action"`
	Budget float64 `json:"budget,omitempty"`
}

// FinOpsResponse 是所有入口共用的响应结构体。
//
// 字段：
//   - Status  : "ok" 或 "error"
//   - Message : 人类可读的描述信息
//   - Data    : 可选的结构化数据（如预算详情）
type FinOpsResponse struct {
	Status  string      `json:"status"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

// BudgetData 是 get_budget 响应中 Data 字段的具体结构。
//
// 字段：
//   - Amount   : 当前预算金额
//   - Currency : 货币代码
//   - Type     : 预算类型
type BudgetData struct {
	Amount   float64 `json:"amount"`
	Currency string  `json:"currency"`
	Type     string  `json:"type"`
}

const (
	handlerComponent = "Handler"
)

// HandleFinOpsRequest 是核心业务处理函数，由 HTTP server 和 Lambda handler 共同调用。
// 根据 FinOpsRequest.Action 分发到对应的业务逻辑。
//
// 参数：
//   - req         : 解析后的请求
//   - budgetState : 当前预算状态实例（由 main 初始化后注入）
//
// 返回：
//   - FinOpsResponse : 统一格式的响应
func HandleFinOpsRequest(req FinOpsRequest, budgetState *billing.BudgetState) FinOpsResponse {
	start := time.Now()
	utilities.LogStart(handlerComponent, req.Action)

	var resp FinOpsResponse

	switch req.Action {
	case "health":
		resp = handleHealth()
	case "get_budget":
		resp = handleGetBudget(budgetState)
	case "set_budget":
		resp = handleSetBudget(req, budgetState)
	default:
		resp = FinOpsResponse{
			Status:  "error",
			Message: fmt.Sprintf("未知操作: %q，支持的操作: health, get_budget, set_budget", req.Action),
		}
	}

	utilities.LogSuccess(handlerComponent, req.Action, time.Since(start),
		fmt.Sprintf("status=%s", resp.Status),
	)
	return resp
}

// ParseRequest 将 JSON 字节流解析为 FinOpsRequest。
//
// 参数：
//   - body : JSON 格式的请求体
//
// 返回：
//   - FinOpsRequest : 解析后的请求
//   - error         : JSON 格式错误时返回错误
func ParseRequest(body []byte) (FinOpsRequest, error) {
	var req FinOpsRequest
	if err := json.Unmarshal(body, &req); err != nil {
		return req, fmt.Errorf("请求体 JSON 解析失败: %w", err)
	}
	return req, nil
}

// handleHealth 返回服务健康状态。
//
// 返回：
//   - FinOpsResponse : status=ok 的健康检查响应
func handleHealth() FinOpsResponse {
	return FinOpsResponse{
		Status:  "ok",
		Message: "AWSFinOps service is running",
	}
}

// handleGetBudget 读取当前预算状态并返回。
//
// 参数：
//   - budgetState : 预算状态实例
//
// 返回：
//   - FinOpsResponse : 包含预算详情的响应
func handleGetBudget(budgetState *billing.BudgetState) FinOpsResponse {
	amount, err := budgetState.GetAmount()
	if err != nil {
		return FinOpsResponse{Status: "error", Message: err.Error()}
	}
	currency, _ := budgetState.GetCurrency()
	bt, _ := budgetState.GetBudgetType()

	return FinOpsResponse{
		Status:  "ok",
		Message: fmt.Sprintf("当前预算: %.2f %s", float64(amount), currency),
		Data: BudgetData{
			Amount:   float64(amount),
			Currency: string(currency),
			Type:     string(bt),
		},
	}
}

// handleSetBudget 更新预算金额。
//
// 参数：
//   - req         : 请求，Budget 字段为新金额
//   - budgetState : 预算状态实例
//
// 返回：
//   - FinOpsResponse : 更新结果
func handleSetBudget(req FinOpsRequest, budgetState *billing.BudgetState) FinOpsResponse {
	if err := budgetState.SetAmount(billing.Budget(req.Budget)); err != nil {
		return FinOpsResponse{Status: "error", Message: err.Error()}
	}

	amount, _ := budgetState.GetAmount()
	currency, _ := budgetState.GetCurrency()

	return FinOpsResponse{
		Status:  "ok",
		Message: fmt.Sprintf("预算已更新为 %.2f %s", float64(amount), currency),
		Data: BudgetData{
			Amount:   float64(amount),
			Currency: string(currency),
		},
	}
}
