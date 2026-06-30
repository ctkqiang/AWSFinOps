package aws

import (
	"aws_fin_ops/internal/billing"
	"aws_fin_ops/internal/services"
	"aws_fin_ops/internal/utilities"
	"context"
	"encoding/json"
	"time"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
)

const (
	lambdaComponent = "LambdaHandler"
)

// StartLambda 注册 Lambda handler 并启动 Lambda 运行时循环。
// 仅在 AWS Lambda 环境下由 main.go 调用。
//
// 参数：
//   - budgetState : 由 main 初始化的预算状态实例，在 Lambda 冷启动时创建，
//     后续 warm invocation 复用同一实例
func StartLambda(budgetState *billing.BudgetState) {
	utilities.LogStart(lambdaComponent, "StartLambda")
	lambda.Start(newHandler(budgetState))
}

// newHandler 返回一个符合 API Gateway proxy 集成签名的 Lambda handler 函数。
// 将 API Gateway 请求体解析为 FinOpsRequest，委托给共享的 HandleFinOpsRequest，
// 再将 FinOpsResponse 序列化为 API Gateway proxy 响应。
//
// 参数：
//   - budgetState : 预算状态实例
//
// 返回：
//   - func(ctx, request) (response, error) : Lambda handler 函数
func newHandler(budgetState *billing.BudgetState) func(context.Context, events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	return func(ctx context.Context, event events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
		start := time.Now()
		utilities.LogStart(lambdaComponent, "Invoke")

		req, err := services.ParseRequest([]byte(event.Body))
		if err != nil {
			utilities.LogError(lambdaComponent, "Invoke", err, time.Since(start))
			return gatewayResponse(400, services.FinOpsResponse{
				Status:  "error",
				Message: err.Error(),
			})
		}

		resp := services.HandleFinOpsRequest(req, budgetState)

		statusCode := 200
		if resp.Status == "error" {
			statusCode = 400
		}

		utilities.LogSuccess(lambdaComponent, "Invoke", time.Since(start),
			"action="+req.Action,
			"status="+resp.Status,
		)

		return gatewayResponse(statusCode, resp)
	}
}

// gatewayResponse 将 FinOpsResponse 序列化为 API Gateway proxy 响应。
//
// 参数：
//   - statusCode : HTTP 状态码
//   - resp       : 业务响应
//
// 返回：
//   - events.APIGatewayProxyResponse : API Gateway 响应
//   - error                          : JSON 序列化失败时返回错误
func gatewayResponse(statusCode int, resp services.FinOpsResponse) (events.APIGatewayProxyResponse, error) {
	body, err := json.Marshal(resp)
	if err != nil {
		return events.APIGatewayProxyResponse{
			StatusCode: 500,
			Body:       `{"status":"error","message":"响应序列化失败"}`,
			Headers:    map[string]string{"Content-Type": "application/json"},
		}, nil
	}

	return events.APIGatewayProxyResponse{
		StatusCode: statusCode,
		Body:       string(body),
		Headers:    map[string]string{"Content-Type": "application/json"},
	}, nil
}
