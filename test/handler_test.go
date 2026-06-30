package test

import (
	"aws_fin_ops/internal/billing"
	"aws_fin_ops/internal/services"
	"testing"
)

// TestParseRequest_Valid 验证合法 JSON 能正确解析为 FinOpsRequest。
func TestParseRequest_Valid(t *testing.T) {
	body := []byte(`{"action":"health"}`)
	req, err := services.ParseRequest(body)
	if err != nil {
		t.Fatalf("期望无错误，实际得到 %v", err)
	}
	if req.Action != "health" {
		t.Fatalf("期望 action=health，实际得到 %s", req.Action)
	}
}

// TestParseRequest_InvalidJSON 验证非法 JSON 返回错误。
func TestParseRequest_InvalidJSON(t *testing.T) {
	body := []byte(`{invalid}`)
	_, err := services.ParseRequest(body)
	if err == nil {
		t.Fatal("期望非法 JSON 时返回错误")
	}
}

// TestParseRequest_WithBudget 验证带 budget 字段的请求能正确解析。
func TestParseRequest_WithBudget(t *testing.T) {
	body := []byte(`{"action":"set_budget","budget":750.50}`)
	req, err := services.ParseRequest(body)
	if err != nil {
		t.Fatalf("期望无错误，实际得到 %v", err)
	}
	if req.Action != "set_budget" {
		t.Fatalf("期望 action=set_budget，实际得到 %s", req.Action)
	}
	if req.Budget != 750.50 {
		t.Fatalf("期望 budget=750.50，实际得到 %f", req.Budget)
	}
}

// TestHandleFinOpsRequest_Health 验证 health 操作返回 ok 状态。
func TestHandleFinOpsRequest_Health(t *testing.T) {
	bs, _ := billing.NewBudgetState(500, "USD", billing.BudgetTypeCost)
	req := services.FinOpsRequest{Action: "health"}

	resp := services.HandleFinOpsRequest(req, bs)
	if resp.Status != "ok" {
		t.Fatalf("期望 status=ok，实际得到 %s", resp.Status)
	}
}

// TestHandleFinOpsRequest_GetBudget 验证 get_budget 返回正确的预算数据。
func TestHandleFinOpsRequest_GetBudget(t *testing.T) {
	bs, _ := billing.NewBudgetState(500, "USD", billing.BudgetTypeCost)
	req := services.FinOpsRequest{Action: "get_budget"}

	resp := services.HandleFinOpsRequest(req, bs)
	if resp.Status != "ok" {
		t.Fatalf("期望 status=ok，实际得到 %s", resp.Status)
	}
	data, ok := resp.Data.(services.BudgetData)
	if !ok {
		t.Fatal("期望 Data 为 BudgetData 类型")
	}
	if data.Amount != 500 {
		t.Fatalf("期望 amount=500，实际得到 %f", data.Amount)
	}
	if data.Currency != "USD" {
		t.Fatalf("期望 currency=USD，实际得到 %s", data.Currency)
	}
}

// TestHandleFinOpsRequest_SetBudget 验证 set_budget 能正确更新预算。
func TestHandleFinOpsRequest_SetBudget(t *testing.T) {
	bs, _ := billing.NewBudgetState(500, "USD", billing.BudgetTypeCost)
	req := services.FinOpsRequest{Action: "set_budget", Budget: 800}

	resp := services.HandleFinOpsRequest(req, bs)
	if resp.Status != "ok" {
		t.Fatalf("期望 status=ok，实际得到 %s", resp.Status)
	}

	amount, _ := bs.GetAmount()
	if amount != 800 {
		t.Fatalf("期望金额更新为 800，实际得到 %.2f", float64(amount))
	}
}

// TestHandleFinOpsRequest_SetBudget_Invalid 验证设置非法金额时返回错误。
func TestHandleFinOpsRequest_SetBudget_Invalid(t *testing.T) {
	bs, _ := billing.NewBudgetState(500, "USD", billing.BudgetTypeCost)
	req := services.FinOpsRequest{Action: "set_budget", Budget: -100}

	resp := services.HandleFinOpsRequest(req, bs)
	if resp.Status != "error" {
		t.Fatalf("期望 status=error，实际得到 %s", resp.Status)
	}

	amount, _ := bs.GetAmount()
	if amount != 500 {
		t.Fatalf("期望原值 500 不变，实际得到 %.2f", float64(amount))
	}
}

// TestHandleFinOpsRequest_UnknownAction 验证未知操作返回错误。
func TestHandleFinOpsRequest_UnknownAction(t *testing.T) {
	bs, _ := billing.NewBudgetState(500, "USD", billing.BudgetTypeCost)
	req := services.FinOpsRequest{Action: "delete_everything"}

	resp := services.HandleFinOpsRequest(req, bs)
	if resp.Status != "error" {
		t.Fatalf("期望 status=error，实际得到 %s", resp.Status)
	}
}
