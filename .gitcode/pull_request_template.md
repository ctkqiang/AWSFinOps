## Description / 描述

<!-- Provide a clear and concise summary of your changes -->
<!-- 请简洁清晰地描述本次变更内容 -->

**Problem / 问题：**
Describe the issue or requirement this PR addresses.
描述本 PR 解决的问题或需求。

**Solution / 解决方案：**
Explain how this PR solves the problem.
说明本 PR 如何解决该问题。

---

## Type of Change / 变更类型

<!-- Mark the relevant option with an "x" / 在相关选项前填写 "x" -->

- [ ] Bug fix / 缺陷修复 (non-breaking change that fixes an issue)
- [ ] Feature / 新功能 (non-breaking change that adds functionality)
- [ ] Refactor / 重构 (code improvement without changing behavior)
- [ ] Documentation / 文档 (documentation updates only)
- [ ] Performance / 性能优化 (performance improvement)
- [ ] Security / 安全 (security improvement or vulnerability fix)
- [ ] Chore / 杂项 (dependency update, linting, etc.)
- [ ] Breaking change / 破坏性变更 (breaking API change)

---

## Related Issues / 关联 Issue

Closes / 关闭: #<!-- issue number -->
Related to / 关联: #<!-- issue number -->

---

## Testing / 测试

### Test Plan / 测试计划

1. Set environment / 配置环境: `cp .env.example .env && vim .env`
2. Run once / 单次运行: `WORKER_ONCE=true go run .`
3. Run tests / 运行测试: `go test -v ./test/...`

### Test Coverage / 测试覆盖

- [ ] Unit tests added/updated / 单元测试已添加或更新 (`go test ./test/...`)
- [ ] Manual testing completed / 已完成手动测试

### Test Results / 测试结果

```
$ go test -v ./test/...
ok    aws_fin_ops/test    x.xxxs
PASS
```

---

## Changes Overview / 变更概览

### Files Changed / 变更文件

<!-- List key files modified / 列出主要变更文件 -->

- `internal/worker/engine.go` -- Worker 7-step pipeline / Worker 7 步流水线
- `internal/config.go` -- Config loader / 三层配置加载器
- `internal/billing/*.go` -- Budget & cost analysis / 预算与成本分析
- `internal/aws/*.go` -- AWS SDK wrappers / AWS SDK 封装
- `internal/services/broadcast.go` -- Multi-platform broadcast / 多平台广播
- `internal/utilities/document.go` -- PDF/JSON/CSV report export / 报告导出
- `internal/archive/s3.go` -- S3 Glacier archive / S3 归档

---

## Security & Compliance / 安全与合规

- [ ] No security impact / 无安全影响
- [ ] Handles sensitive data (AWS credentials, bot tokens) / 涉及敏感数据
- [ ] Affects AWS IAM permissions / 影响 IAM 权限

**Security Considerations / 安全说明：**

- All credentials loaded from environment variables only / 所有凭证仅从环境变量读取 [OK]
- Sensitive values masked via `utilities.Mask()` before logging / 敏感值已脱敏 [OK]
- No hardcoded secrets / 无硬编码密钥 [OK]

---

## Deployment Notes / 部署说明

### Runtime Mode Affected / 受影响的运行时模式

- [ ] Local loop mode / 本地循环模式
- [ ] Local once mode / 本地单次模式 (`WORKER_ONCE=true`)
- [ ] AWS Lambda mode / Lambda 模式 (EventBridge)

### Breaking Changes / 破坏性变更

- [ ] No breaking changes / 无破坏性变更
- [ ] Environment variable added or renamed / 新增或重命名了环境变量

---

## Deployment Checklist / 部署检查清单

- [ ] `go build ./...` compiles / 编译通过
- [ ] `go vet ./...` no warnings / 无警告
- [ ] `go test ./test/...` all passing / 测试通过
- [ ] Log messages in Simplified Chinese / 日志消息为简体中文
- [ ] Error handling uses `fmt.Errorf("operation: %w", err)` / 错误包装规范
- [ ] Logging uses `utilities.LogStart/LogSuccess/LogError/LogWarn` / 日志函数规范
- [ ] No hardcoded credentials or secrets / 无硬编码凭证
- [ ] Commit message follows `git-commit-message.md` format / 提交信息符合规范

---

**Reviewer / 审查人:** @ctkqiang_sr
**License / 许可证:** MIT -- Copyright (c) 2026 ctkqiang
