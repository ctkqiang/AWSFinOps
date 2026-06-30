package test

import (
	"aws_fin_ops/internal/billing"
	"aws_fin_ops/internal/services"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// TestAPIHandler_Health 验证 /api 端点处理 health 请求返回 200 + ok。
func TestAPIHandler_Health(t *testing.T) {
	bs, _ := billing.NewBudgetState(500, "USD", billing.BudgetTypeCost)
	handler := services.APIHandler(bs)

	body := `{"action":"health"}`
	req := httptest.NewRequest(http.MethodPost, "/api", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("期望状态码 200，实际得到 %d", w.Code)
	}

	var resp services.FinOpsResponse
	json.NewDecoder(w.Body).Decode(&resp)
	if resp.Status != "ok" {
		t.Fatalf("期望 status=ok，实际得到 %s", resp.Status)
	}
}

// TestAPIHandler_GetBudget 验证 /api 端点处理 get_budget 请求。
func TestAPIHandler_GetBudget(t *testing.T) {
	bs, _ := billing.NewBudgetState(500, "USD", billing.BudgetTypeCost)
	handler := services.APIHandler(bs)

	body := `{"action":"get_budget"}`
	req := httptest.NewRequest(http.MethodPost, "/api", strings.NewReader(body))
	w := httptest.NewRecorder()

	handler(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("期望状态码 200，实际得到 %d", w.Code)
	}

	var resp services.FinOpsResponse
	json.NewDecoder(w.Body).Decode(&resp)
	if resp.Status != "ok" {
		t.Fatalf("期望 status=ok，实际得到 %s", resp.Status)
	}
}

// TestAPIHandler_SetBudget 验证 /api 端点处理 set_budget 请求并更新预算。
func TestAPIHandler_SetBudget(t *testing.T) {
	bs, _ := billing.NewBudgetState(500, "USD", billing.BudgetTypeCost)
	handler := services.APIHandler(bs)

	body := `{"action":"set_budget","budget":999}`
	req := httptest.NewRequest(http.MethodPost, "/api", strings.NewReader(body))
	w := httptest.NewRecorder()

	handler(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("期望状态码 200，实际得到 %d", w.Code)
	}

	amount, _ := bs.GetAmount()
	if amount != 999 {
		t.Fatalf("期望金额 999，实际得到 %.2f", float64(amount))
	}
}

// TestAPIHandler_MethodNotAllowed 验证非 POST 方法返回 405。
func TestAPIHandler_MethodNotAllowed(t *testing.T) {
	bs, _ := billing.NewBudgetState(500, "USD", billing.BudgetTypeCost)
	handler := services.APIHandler(bs)

	req := httptest.NewRequest(http.MethodGet, "/api", nil)
	w := httptest.NewRecorder()

	handler(w, req)

	if w.Code != http.StatusMethodNotAllowed {
		t.Fatalf("期望状态码 405，实际得到 %d", w.Code)
	}
}

// TestAPIHandler_InvalidJSON 验证非法 JSON 返回 400。
func TestAPIHandler_InvalidJSON(t *testing.T) {
	bs, _ := billing.NewBudgetState(500, "USD", billing.BudgetTypeCost)
	handler := services.APIHandler(bs)

	body := `{not json}`
	req := httptest.NewRequest(http.MethodPost, "/api", strings.NewReader(body))
	w := httptest.NewRecorder()

	handler(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("期望状态码 400，实际得到 %d", w.Code)
	}
}

// TestHealthHandler 验证 /health 端点返回 ok。
func TestHealthHandler(t *testing.T) {
	handler := services.HealthHandler()

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	w := httptest.NewRecorder()

	handler(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("期望状态码 200，实际得到 %d", w.Code)
	}

	var resp services.FinOpsResponse
	json.NewDecoder(w.Body).Decode(&resp)
	if resp.Status != "ok" {
		t.Fatalf("期望 status=ok，实际得到 %s", resp.Status)
	}
}
