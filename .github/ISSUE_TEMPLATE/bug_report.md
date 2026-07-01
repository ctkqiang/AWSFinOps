---
name: Bug Report / 缺陷报告
about: Report a bug in AWSFinOps Worker
title: '[BUG] '
labels: bug
assignees: ctkqiang
---

## Bug Description / 缺陷描述

<!-- A clear and concise description of what the bug is -->
<!-- 请清晰简洁地描述缺陷内容 -->

## Runtime Environment / 运行时环境

<!-- Mark the mode where the bug occurs / 标记缺陷发生的运行模式 -->

- [ ] Local loop mode / 本地循环模式 (`go run .`)
- [ ] Local once mode / 本地单次模式 (`WORKER_ONCE=true go run .`)
- [ ] AWS Lambda mode / Lambda 模式 (EventBridge trigger)

**Go version / Go 版本:** <!-- e.g. go1.26.1 -->
**OS / 操作系统:** <!-- e.g. macOS 14, Ubuntu 22.04, Amazon Linux 2023 -->
**AWS Region / AWS 区域:** <!-- e.g. ap-east-1 -->

## Affected Component / 受影响组件

<!-- Mark all that apply / 勾选所有适用项 -->

- [ ] Worker engine / Worker 引擎 (`internal/worker/engine.go`)
- [ ] Config loader / 配置加载器 (`internal/config.go`)
- [ ] Budget check / 预算检查 (`internal/billing/budget.go`)
- [ ] Cost analysis / 成本分析 (`internal/billing/analysis.go`)
- [ ] Rightsizing / 规格优化 (`internal/aws/rightsizing.go`)
- [ ] Cost Explorer / 成本查询 (`internal/aws/costexplorer.go`)
- [ ] Report export / 报告导出 (`internal/utilities/document.go`)
- [ ] S3 Glacier archive / S3 归档 (`internal/archive/s3.go`)
- [ ] Broadcast / 广播通知 (`internal/services/broadcast.go`)
- [ ] Lambda handler / Lambda 处理器 (`internal/aws/lambda.go`)
- [ ] Logging / 日志 (`internal/utilities/logger.go`)
- [ ] Other / 其他: <!-- describe / 描述 -->

## Affected Pipeline Step / 受影响的流水线步骤

<!-- If the bug is in a specific pipeline step / 若缺陷在特定流水线步骤中 -->

- [ ] Step 1: Budget check / 预算检查
- [ ] Step 2: Cost query / 成本查询
- [ ] Step 3: Cost forecast / 成本预测
- [ ] Step 4: Usage comparison / 用量对比
- [ ] Step 5: Rightsizing / 规格优化
- [ ] Step 6: Analysis / 综合分析
- [ ] Step 7: Broadcast / 广播通知

## Steps to Reproduce / 复现步骤

1. Configure `.env` with / 配置 `.env`: `...`
2. Run / 运行: `...`
3. Observe / 观察到: `...`

## Expected Behavior / 预期行为

<!-- What should have happened / 描述应该发生什么 -->

## Actual Behavior / 实际行为

<!-- What actually happened / 描述实际发生了什么 -->

## Error Output / Logs / 错误输出与日志

<!-- Paste relevant log output. Sensitive values (credentials, tokens) must be redacted. -->
<!-- 粘贴相关日志输出。敏感值（凭证、令牌）必须脱敏处理。 -->

```
paste log output here / 在此粘贴日志输出
```

## Environment Variables (redacted) / 环境变量（已脱敏）

<!-- List relevant env vars with values replaced by xxxxxxxxxx -->
<!-- 列出相关环境变量，值替换为 xxxxxxxxxx -->

```dotenv
AWS_REGION=xxxxxxxxxx
AWS_ACCESS_KEY_ID=xxxxxxxxxx
AWS_SECRET_ACCESS_KEY=xxxxxxxxxx
WORKER_INTERVAL=xxxxxxxxxx
BROADCAST_ENABLED=xxxxxxxxxx
EXPORT_REPORT=xxxxxxxxxx
```

## Additional Context / 补充说明

<!-- Any other context, screenshots, or related issues -->
<!-- 其他上下文、截图或关联 Issue -->
