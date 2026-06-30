package services

import (
	"aws_fin_ops/internal/billing"
	"aws_fin_ops/internal/utilities"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"
)

const (
	serverComponent = "LocalServer"
)

// StartLocalServer 启动本地 HTTP 服务器，将所有 /api 请求路由到共享的 HandleFinOpsRequest。
// 仅在本地开发模式下由 main.go 调用。
//
// 参数：
//   - budgetState : 由 main 初始化的预算状态实例，注入到每个请求的处理链中
//
// 返回：
//   - error : 服务器启动失败时返回错误（如端口被占用）
func StartLocalServer(budgetState *billing.BudgetState) error {
	start := time.Now()
	utilities.LogStart(serverComponent, "StartLocalServer")

	host := envOrDefault("APP_HOST", "0.0.0.0")
	port := envOrDefault("APP_PORT", "8080")
	addr := fmt.Sprintf("%s:%s", host, port)

	mux := http.NewServeMux()

	mux.HandleFunc("/api", APIHandler(budgetState))
	mux.HandleFunc("/health", HealthHandler())

	utilities.LogSuccess(serverComponent, "StartLocalServer", time.Since(start),
		fmt.Sprintf("addr=%s", addr),
	)
	utilities.Info("Local HTTP server listening on %s", addr)

	return http.ListenAndServe(addr, mux)
}

// APIHandler 返回处理 /api 端点的 http.HandlerFunc。
// 仅接受 POST 方法，将请求体解析为 FinOpsRequest 后委托给 HandleFinOpsRequest。
//
// 参数：
//   - budgetState : 预算状态实例
//
// 返回：
//   - http.HandlerFunc : 标准 HTTP 处理函数
func APIHandler(budgetState *billing.BudgetState) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		if r.Method != http.MethodPost {
			writeJSON(w, http.StatusMethodNotAllowed, FinOpsResponse{
				Status:  "error",
				Message: fmt.Sprintf("不支持 %s 方法，请使用 POST", r.Method),
			})
			return
		}

		body, err := io.ReadAll(r.Body)
		if err != nil {
			writeJSON(w, http.StatusBadRequest, FinOpsResponse{
				Status:  "error",
				Message: "读取请求体失败",
			})
			return
		}
		defer r.Body.Close()

		req, err := ParseRequest(body)
		if err != nil {
			writeJSON(w, http.StatusBadRequest, FinOpsResponse{
				Status:  "error",
				Message: err.Error(),
			})
			return
		}

		resp := HandleFinOpsRequest(req, budgetState)

		status := http.StatusOK
		if resp.Status == "error" {
			status = http.StatusBadRequest
		}

		utilities.LogProgress(serverComponent, "apiHandler",
			fmt.Sprintf("action=%s elapsed=%v", req.Action, time.Since(start)),
		)

		writeJSON(w, status, resp)
	}
}

// HealthHandler 返回处理 /health 端点的 http.HandlerFunc。
// 接受任意 HTTP 方法，直接返回健康状态。
//
// 返回：
//   - http.HandlerFunc : 标准 HTTP 处理函数
func HealthHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, FinOpsResponse{
			Status:  "ok",
			Message: "AWSFinOps service is running",
		})
	}
}

// writeJSON 将响应序列化为 JSON 并写入 http.ResponseWriter。
//
// 参数：
//   - w      : HTTP 响应写入器
//   - status : HTTP 状态码
//   - v      : 待序列化的响应对象
func writeJSON(w http.ResponseWriter, status int, v interface{}) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

// envOrDefault 读取环境变量，未设置时返回默认值。
//
// 参数：
//   - key          : 环境变量名称
//   - defaultValue : 未设置时的默认值
//
// 返回：
//   - string : 环境变量值或默认值
func envOrDefault(key, defaultValue string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return defaultValue
}
