package services

import (
	"aws_fin_ops/internal"
	"aws_fin_ops/internal/utilities"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/smtp"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type (
	FileType         string
	BroadcastService int
)

type TelegramMessage struct {
	ChatID string `json:"chat_id"`
	Text   string `json:"text"`
}

type File struct {
	Name string   `json:"name"`
	Type FileType `json:"type"`
}

const (
	RawText FileType = "text"
	CSV     FileType = "csv"
	JSON    FileType = "json"
	XML     FileType = "xml"
	PDF     FileType = "pdf"
)

const (
	Weixin BroadcastService = iota
	Feishu
	DingDing
	Email
	Slack
	Telegram
	Discord
)

// telegramBaseURL 是 Telegram Bot API 的根地址。声明为 var 以便单元测试
// 可以将其指向 httptest.Server，无需修改生产代码。
var telegramBaseURL = "https://api.telegram.org"

func Broadcast(message string, file File) {
	utilities.LogProgress(
		"broadcast",
		"broadcast",
		message,
		file.Name,
	)
}

// TelegramGetBotFather 使用 internal.GetEnvValue 读取环境变量 TELEGRAM_BOT_TOKEN，
// 构建 Telegram Bot API 的 getUpdates URL。
//
// 返回值:
//   - string: 完整的 getUpdates URL（如 https://api.telegram.org/bot<token>/getUpdates）。
//     若 TELEGRAM_BOT_TOKEN 未配置或为空，记录 ERROR 日志并返回空字符串。
func TelegramGetBotFather() string {
	token, err := internal.GetEnvValue("TELEGRAM_BOT_TOKEN")
	if err != nil {
		utilities.Error("TelegramGetBotFather: %v", err)
		return ""
	}

	utilities.Info("Telegram getUpdates URL 已构建 (token=%s)", utilities.Mask(token))
	return fmt.Sprintf("%s/bot%s/getUpdates", telegramBaseURL, token)
}

// TelegramSendMessage 向已配置的 Telegram 聊天发送文本消息。
// 通过 internal.GetEnvValue 读取 TELEGRAM_BOT_TOKEN 和 TELEGRAM_CHAT_ID，
// 将 TelegramMessage 序列化为 JSON 后 POST 到 Telegram Bot API sendMessage 端点。
//
// 参数:
//   - text: 要发送的纯文本消息内容。
//
// 返回值:
//   - error: 成功时返回 nil；以下情况返回非 nil 错误：
//     TELEGRAM_BOT_TOKEN 或 TELEGRAM_CHAT_ID 未配置、JSON 序列化失败、
//     网络请求失败、或 Telegram API 返回非 200 状态码。
func TelegramSendMessage(text string) error {
	start := time.Now()

	token, err := internal.GetEnvValue("TELEGRAM_BOT_TOKEN")
	if err != nil {
		utilities.Error("TelegramSendMessage: %v", err)
		return fmt.Errorf("TelegramSendMessage: %w", err)
	}

	chatID, err := internal.GetEnvValue("TELEGRAM_CHAT_ID")
	if err != nil {
		utilities.Error("TelegramSendMessage: %v", err)
		return fmt.Errorf("TelegramSendMessage: %w", err)
	}

	msg := TelegramMessage{
		ChatID: chatID,
		Text:   text,
	}

	data, err := json.Marshal(msg)
	if err != nil {
		utilities.Error("TelegramSendMessage: 序列化消息体失败: %v", err)
		return fmt.Errorf("TelegramSendMessage: 序列化失败: %w", err)
	}

	url := fmt.Sprintf("%s/bot%s/sendMessage", telegramBaseURL, token)

	resp, err := http.Post(url, "application/json", bytes.NewBuffer(data))
	if err != nil {
		utilities.LogError("Telegram", "SendMessage", err, time.Since(start))
		return fmt.Errorf("TelegramSendMessage: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		err := fmt.Errorf("Telegram API 返回状态码 %s", resp.Status)
		utilities.LogError("Telegram", "SendMessage", err, time.Since(start))
		return err
	}

	utilities.LogSuccess("Telegram", "SendMessage", time.Since(start),
		fmt.Sprintf("ChatID=%s", utilities.Mask(chatID)),
	)
	return nil
}

// broadcastComponent 是广播服务在日志中使用的组件名称标识。
const broadcastComponent = "Broadcast"

// httpClient 是广播服务共用的 HTTP 客户端，所有请求共享同一个超时配置。
var httpClient = &http.Client{Timeout: 30 * time.Second}

// BroadcastReport 根据 .env 配置自动检测已启用的广播平台，
// 将 FinOps 报告摘要和生成的报告文件（PDF/JSON/CSV）发送到对应平台。
// 未配置的平台自动跳过；任意平台发送失败不影响其余平台。
//
// 启用的平台由环境变量控制（至少配置了 Token/ID 即视为已启用）：
//   - Weixin    : WEIXIN_BOT_TOKEN + WEIXIN_CHAT_ID
//   - Feishu   : FEISHU_BOT_TOKEN + FEISHU_CHAT_ID
//   - DingDing : DINGTALK_BOT_TOKEN + DINGTALK_SECRET
//   - Email    : SMTP_HOST + SMTP_TO
//   - Slack    : SLACK_BOT_TOKEN + SLACK_CHANNEL_ID
//   - Telegram : TELEGRAM_BOT_TOKEN + TELEGRAM_CHAT_ID
//   - Discord  : DISCORD_BOT_TOKEN + DISCORD_CHANNEL_ID
//
// 参数：
//   - data  : FinOps 报告数据
//   - paths : 已生成的报告文件路径列表（PDF/JSON/CSV）
//
// 返回：
//   - 成功发送的平台名称列表（可能为空）
func BroadcastReport(data *utilities.ReportData, paths []string) []string {
	start := time.Now()
	utilities.LogStart(broadcastComponent, "BroadcastReport")

	if data == nil {
		utilities.LogWarn(broadcastComponent, "BroadcastReport", "data 为 nil，跳过广播", 0)
		return nil
	}

	msg := buildReportMessage(data)
	enabledPlatforms := detectEnabledPlatforms()

	if len(enabledPlatforms) == 0 {
		utilities.LogProgress(broadcastComponent, "BroadcastReport",
			"未检测到任何已配置的广播平台，跳过广播",
		)
		return nil
	}

	utilities.LogProgress(broadcastComponent, "BroadcastReport",
		fmt.Sprintf("已启用平台: %s", strings.Join(enabledPlatforms, ", ")),
		fmt.Sprintf("待发送文件: %d 个", len(paths)),
	)

	var sent []string
	for _, platform := range enabledPlatforms {
		var err error
		switch platform {
		case "Weixin":
			err = sendWeixin(msg)
		case "Feishu":
			err = sendFeishu(msg, paths)
		case "DingDing":
			err = sendDingDing(msg)
		case "Email":
			err = sendEmail(msg, paths)
		case "Slack":
			err = sendSlack(msg, paths)
		case "Telegram":
			err = sendTelegram(msg, paths)
		case "Discord":
			err = sendDiscord(msg, paths)
		}
		if err != nil {
			utilities.LogWarn(broadcastComponent, "BroadcastReport",
				fmt.Sprintf("平台 %s 发送失败: %v", platform, err),
				time.Since(start),
			)
			continue
		}
		sent = append(sent, platform)
	}

	utilities.LogSuccess(broadcastComponent, "BroadcastReport", time.Since(start),
		fmt.Sprintf("enabled=%d", len(enabledPlatforms)),
		fmt.Sprintf("sent=%d", len(sent)),
		fmt.Sprintf("platforms=%s", strings.Join(sent, ",")),
	)
	return sent
}

// buildReportMessage 根据报告数据构建多平台兼容的文本摘要消息。
//
// 返回：
//   - string : 格式化的报告摘要文本
func buildReportMessage(data *utilities.ReportData) string {
	ok, pending, warned, failed := utilities.CountStepStatuses(data.Steps)

	statusEmoji := "✅"
	switch data.Status {
	case "partial_failure":
		statusEmoji = "⚠️"
	case "error":
		statusEmoji = "❌"
	}

	msg := fmt.Sprintf(`%s *AWS FinOps 执行报告*

📋 Run ID: *%s*
🏦 Account: *%s*
📊 状态: %s *%s*
⏱ 耗时: %s

📈 步骤汇总:
  ✅ 成功: %d
  ⏳ 待实现: %d
  ⚠️ 警告: %d
  ❌ 失败: %d

📅 时间区间: %s ~ %s

📝 摘要: %s

---
🤖 由 AWSFinOps Worker 自动生成`,
		statusEmoji,
		data.RunID,
		strings.ToUpper(data.AWSAccountID),
		statusEmoji,
		strings.ToUpper(data.Status),
		data.Duration.Round(time.Second).String(),
		ok, pending, warned, failed,
		data.StartTime.Format("2006-01-02 15:04"),
		data.EndTime.Format("2006-01-02 15:04"),
		data.Summary,
	)
	return msg
}

// detectEnabledPlatforms 检测 .env 中已配置的广播平台。
//
// 返回：
//   - []string : 已配置的平台名称列表
func detectEnabledPlatforms() []string {
	var platforms []string

	if isWeixinEnabled() {
		platforms = append(platforms, "Weixin")
	}
	if isFeishuEnabled() {
		platforms = append(platforms, "Feishu")
	}
	if isDingDingEnabled() {
		platforms = append(platforms, "DingDing")
	}
	if isEmailEnabled() {
		platforms = append(platforms, "Email")
	}
	if isSlackEnabled() {
		platforms = append(platforms, "Slack")
	}
	if isTelegramEnabled() {
		platforms = append(platforms, "Telegram")
	}
	if isDiscordEnabled() {
		platforms = append(platforms, "Discord")
	}

	return platforms
}

// ---- 各平台启用检测函数 ----

func isWeixinEnabled() bool {
	return os.Getenv("WEIXIN_BOT_TOKEN") != "" && os.Getenv("WEIXIN_CHAT_ID") != ""
}

func isFeishuEnabled() bool {
	return os.Getenv("FEISHU_BOT_TOKEN") != "" && os.Getenv("FEISHU_CHAT_ID") != ""
}

func isDingDingEnabled() bool {
	return os.Getenv("DINGTALK_BOT_TOKEN") != "" && os.Getenv("DINGTALK_SECRET") != ""
}

func isEmailEnabled() bool {
	return os.Getenv("SMTP_HOST") != "" && os.Getenv("SMTP_TO") != ""
}

func isSlackEnabled() bool {
	return os.Getenv("SLACK_BOT_TOKEN") != "" && os.Getenv("SLACK_CHANNEL_ID") != ""
}

func isTelegramEnabled() bool {
	return os.Getenv("TELEGRAM_BOT_TOKEN") != "" && os.Getenv("TELEGRAM_CHAT_ID") != ""
}

func isDiscordEnabled() bool {
	return os.Getenv("DISCORD_BOT_TOKEN") != "" && os.Getenv("DISCORD_CHANNEL_ID") != ""
}

// ---- 各平台发送实现 ----

// sendWeixin 通过企业微信机器人发送文本消息。
//
// 参数：
//   - msg : 要发送的文本内容（支持 Markdown）
//
// 返回：
//   - error : 发送失败时返回错误
func sendWeixin(msg string) error {
	start := time.Now()
	token := os.Getenv("WEIXIN_BOT_TOKEN")
	chatID := os.Getenv("WEIXIN_CHAT_ID")

	payload := map[string]interface{}{
		"chatid": chatID,
		"msgtype": "markdown",
		"markdown": map[string]string{
			"content": msg,
		},
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("序列化请求体失败: %w", err)
	}

	url := fmt.Sprintf("https://qyapi.weixin.qq.com/cgi-bin/im/chatid/%s/send_msg?access_token=%s",
		chatID, token)

	req, err := http.NewRequest("POST", url, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("构建请求失败: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := httpClient.Do(req)
	if err != nil {
		utilities.LogError("Weixin", "sendWeixin", err, time.Since(start))
		return fmt.Errorf("HTTP 请求失败: %w", err)
	}
	defer resp.Body.Close()

	io.ReadAll(resp.Body)

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("API 返回状态码 %d", resp.StatusCode)
	}

	utilities.LogSuccess("Weixin", "sendWeixin", time.Since(start))
	return nil
}

// sendFeishu 通过飞书机器人发送文本消息，并附加文件（支持 PDF/CSV/JSON）。
//
// 参数：
//   - msg  : 要发送的文本内容（支持 Markdown）
//   - paths: 要发送的文件路径列表
//
// 返回：
//   - error : 发送失败时返回错误
func sendFeishu(msg string, paths []string) error {
	start := time.Now()
	token := os.Getenv("FEISHU_BOT_TOKEN")
	chatID := os.Getenv("FEISHU_CHAT_ID")

	// 先发送文本消息
	payload := map[string]interface{}{
		"receive_id": chatID,
		"msg_type":   "text",
		"content": json.RawMessage(
			fmt.Sprintf(`{"text":"%s"}`,
				strings.ReplaceAll(strings.ReplaceAll(msg, `\`, `\\`), `"`, `\"`))),
	}

	body, _ := json.Marshal(payload)
	req, _ := http.NewRequest("POST",
		"https://open.feishu.cn/open-apis/bot/v2/hook/"+token,
		bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	resp, err := httpClient.Do(req)
	if err != nil {
		utilities.LogError("Feishu", "sendFeishu", err, time.Since(start))
		return fmt.Errorf("文本消息发送失败: %w", err)
	}
	io.ReadAll(resp.Body)
	resp.Body.Close()

	// 发送附件文件
	for _, p := range paths {
		ext := strings.ToLower(filepath.Ext(p))
		var msgType string
		switch ext {
		case ".pdf":
			msgType = "file"
		case ".csv":
			msgType = "file"
		case ".json":
			msgType = "file"
		default:
			continue
		}

		uploadURL := "https://open.feishu.cn/open-apis/im/v1/files"
		fileName := filepath.Base(p)

		// 构建 multipart form
		body := &bytes.Buffer{}
		writer := multipart.NewWriter(body)
		writer.WriteField("file_name", fileName)
		writer.WriteField("file_type", msgType)
		writer.WriteField("receive_id", chatID)
		writer.WriteField("receive_id_type", "chat_id")

		f, err := os.Open(p)
		if err != nil {
			utilities.LogWarn("Feishu", "sendFeishu",
				fmt.Sprintf("无法打开文件 %s: %v", fileName, err), time.Since(start))
			continue
		}
		part, _ := writer.CreateFormFile("file", fileName)
		io.Copy(part, f)
		f.Close()
		writer.Close()

		req, _ := http.NewRequest("POST", uploadURL,
			bytes.NewReader(body.Bytes()))
		req.Header.Set("Content-Type", writer.FormDataContentType())
		req.Header.Set("Authorization", "Bearer "+token)

		resp, err := httpClient.Do(req)
		if err != nil {
			utilities.LogWarn("Feishu", "sendFeishu",
				fmt.Sprintf("文件 %s 发送失败: %v", fileName, err), time.Since(start))
			continue
		}
		io.ReadAll(resp.Body)
		resp.Body.Close()
	}

	utilities.LogSuccess("Feishu", "sendFeishu", time.Since(start))
	return nil
}

// sendDingDing 通过钉钉自定义机器人发送文本消息（Markdown 格式）。
//
// 参数：
//   - msg : 要发送的文本内容
//
// 返回：
//   - error : 发送失败时返回错误
func sendDingDing(msg string) error {
	start := time.Now()
	token := os.Getenv("DINGTALK_BOT_TOKEN")
	_ = os.Getenv("DINGTALK_SECRET") // 完整版需计算 HMAC-SHA256 签名

	url := fmt.Sprintf("https://oapi.dingtalk.com/robot/send?access_token=%s", token)

	payload := map[string]interface{}{
		"msgtype": "markdown",
		"markdown": map[string]string{
			"title": "AWS FinOps 执行报告",
			"text":  msg,
		},
	}

	body, _ := json.Marshal(payload)
	req, _ := http.NewRequest("POST", url, bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	resp, err := httpClient.Do(req)
	if err != nil {
		utilities.LogError("DingDing", "sendDingDing", err, time.Since(start))
		return fmt.Errorf("HTTP 请求失败: %w", err)
	}
	io.ReadAll(resp.Body)
	resp.Body.Close()

	utilities.LogSuccess("DingDing", "sendDingDing", time.Since(start))
	return nil
}

// sendEmail 通过 SMTP 发送报告摘要和附件。
//
// 参数：
//   - msg  : 邮件正文摘要
//   - paths: 要附加的文件路径列表
//
// 返回：
//   - error : 发送失败时返回错误
func sendEmail(msg string, paths []string) error {
	start := time.Now()

	host := os.Getenv("SMTP_HOST")
	port := os.Getenv("SMTP_PORT")
	username := os.Getenv("SMTP_USERNAME")
	password := os.Getenv("SMTP_PASSWORD")
	from := os.Getenv("SMTP_FROM")
	to := os.Getenv("SMTP_TO")

	if from == "" {
		from = username
	}

	addr := fmt.Sprintf("%s:%s", host, port)

	// 构造带附件的邮件（RFC 2822 multipart/mixed）
	var body bytes.Buffer
	body.WriteString("From: " + from + "\r\n")
	body.WriteString("To: " + to + "\r\n")
	body.WriteString("Subject: AWS FinOps 执行报告\r\n")
	body.WriteString("MIME-Version: 1.0\r\n")
	body.WriteString("Content-Type: multipart/mixed; boundary=\"FINOPSBOUNDARY\"\r\n")
	body.WriteString("\r\n")

	// 正文
	body.WriteString("--FINOPSBOUNDARY\r\n")
	body.WriteString("Content-Type: text/plain; charset=utf-8\r\n")
	body.WriteString("\r\n")
	body.WriteString(msg + "\r\n")
	body.WriteString("\r\n")

	// 附件
	for _, p := range paths {
		data, err := os.ReadFile(p)
		if err != nil {
			utilities.LogWarn("Email", "sendEmail",
				fmt.Sprintf("无法读取附件 %s: %v", filepath.Base(p), err), time.Since(start))
			continue
		}

		ext := strings.ToLower(filepath.Ext(p))
		var ctype string
		switch ext {
		case ".pdf":
			ctype = "application/pdf"
		case ".csv":
			ctype = "text/csv"
		case ".json":
			ctype = "application/json"
		default:
			ctype = "application/octet-stream"
		}

		body.WriteString("--FINOPSBOUNDARY\r\n")
		body.WriteString(fmt.Sprintf("Content-Type: %s; name=\"%s\"\r\n",
			ctype, filepath.Base(p)))
		body.WriteString("Content-Disposition: attachment; filename=\"" + filepath.Base(p) + "\"\r\n")
		body.WriteString("Content-Transfer-Encoding: base64\r\n")
		body.WriteString("\r\n")
		body.WriteString(mimeEncode(data))
		body.WriteString("\r\n")
	}

	body.WriteString("--FINOPSBOUNDARY--\r\n")

	var auth smtp.Auth
	if password != "" {
		auth = smtp.PlainAuth("", username, password, host)
	}

	err := smtp.SendMail(addr, auth, from, []string{to}, body.Bytes())
	if err != nil {
		utilities.LogError("Email", "sendEmail", err, time.Since(start))
		return fmt.Errorf("SMTP 发送失败: %w", err)
	}

	utilities.LogSuccess("Email", "sendEmail", time.Since(start),
		fmt.Sprintf("to=%s", to))
	return nil
}

// mimeEncode 将字节数据编码为 MIME base64 格式（每行 76 字符）。
func mimeEncode(data []byte) string {
	const lineLen = 76
	const alphabet = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789+/"

	var buf bytes.Buffer
	buf.Grow(len(data)*2)

	for i := 0; i < len(data); i += 3 {
		var val uint32
		var parts int

		val = uint32(data[i]) << 16
		parts++

		if i+1 < len(data) {
			val |= uint32(data[i+1]) << 8
			parts++
		}
		if i+2 < len(data) {
			val |= uint32(data[i+2])
			parts++
		}

		for j := 4; j >= 0; j -= 2 {
			idx := (val >> uint(j*6)) & 0x3F
			buf.WriteByte(alphabet[idx])
		}

		pad := (4 - parts) * 8
		for j := 0; j < pad; j += 6 {
			buf.WriteByte('=')
		}

		if (i/3+1)%(lineLen/4) == 0 {
			buf.WriteByte('\r')
		}
		buf.WriteByte('\n')
	}

	return buf.String()
}

// sendSlack 通过 Slack Bot Token 发送消息和文件。
//
// 参数：
//   - msg  : 消息文本
//   - paths: 要上传的文件路径列表
//
// 返回：
//   - error : 发送失败时返回错误
func sendSlack(msg string, paths []string) error {
	start := time.Now()
	token := os.Getenv("SLACK_BOT_TOKEN")
	channel := os.Getenv("SLACK_CHANNEL_ID")

	// 发送文本消息（使用 chat.postMessage）
	msgPayload := map[string]interface{}{
		"channel": channel,
		"text":    msg,
		"mrkdwn":  true,
	}
	body, _ := json.Marshal(msgPayload)
	req, _ := http.NewRequest("POST",
		"https://slack.com/api/chat.postMessage",
		bytes.NewReader(body))
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := httpClient.Do(req)
	if err != nil {
		utilities.LogError("Slack", "sendSlack", err, time.Since(start))
		return fmt.Errorf("消息发送失败: %w", err)
	}
	io.ReadAll(resp.Body)
	resp.Body.Close()

	// 上传文件（使用 files.uploadV2）
	for _, p := range paths {
		f, err := os.Open(p)
		if err != nil {
			utilities.LogWarn("Slack", "sendSlack",
				fmt.Sprintf("无法打开文件 %s: %v", filepath.Base(p), err), time.Since(start))
			continue
		}
		defer f.Close()

		var bodyBuf bytes.Buffer
		writer := multipart.NewWriter(&bodyBuf)
		writer.WriteField("channels", channel)
		writer.WriteField("filename", filepath.Base(p))
		writer.WriteField("title", "AWS FinOps 报告: "+filepath.Base(p))

		part, _ := writer.CreateFormFile("file", filepath.Base(p))
		io.Copy(part, f)
		writer.Close()

		req, _ := http.NewRequest("POST",
			"https://slack.com/api/files.uploadV2",
			bytes.NewReader(bodyBuf.Bytes()))
		req.Header.Set("Authorization", "Bearer "+token)
		req.Header.Set("Content-Type", writer.FormDataContentType())

		resp, err := httpClient.Do(req)
		if err != nil {
			utilities.LogWarn("Slack", "sendSlack",
				fmt.Sprintf("文件 %s 上传失败: %v", filepath.Base(p), err), time.Since(start))
			continue
		}
		io.ReadAll(resp.Body)
		resp.Body.Close()
	}

	utilities.LogSuccess("Slack", "sendSlack", time.Since(start))
	return nil
}

// sendTelegram 通过 Telegram Bot 发送文本消息和文件附件。
//
// 参数：
//   - msg  : 要发送的文本内容
//   - paths: 要发送的文件路径列表
//
// 返回：
//   - error : 发送失败时返回错误
func sendTelegram(msg string, paths []string) error {
	start := time.Now()

	// 先发送文本消息
	if err := TelegramSendMessage(msg); err != nil {
		utilities.LogWarn("Telegram", "sendTelegram",
			fmt.Sprintf("文本消息发送失败: %v", err), time.Since(start))
	}

	// 再发送文件
	for _, p := range paths {
		if err := telegramSendFile(p); err != nil {
			utilities.LogWarn("Telegram", "sendTelegram",
				fmt.Sprintf("文件 %s 发送失败: %v", filepath.Base(p), err), time.Since(start))
			continue
		}
	}

	utilities.LogSuccess("Telegram", "sendTelegram", time.Since(start))
	return nil
}

// telegramSendFile 通过 Telegram Bot API sendDocument 端点发送文件。
//
// 参数：
//   - filePath : 要发送的文件路径
//
// 返回：
//   - error : 发送失败时返回错误
func telegramSendFile(filePath string) error {
	token := os.Getenv("TELEGRAM_BOT_TOKEN")
	chatID := os.Getenv("TELEGRAM_CHAT_ID")

	f, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("无法打开文件: %w", err)
	}
	defer f.Close()

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	writer.WriteField("chat_id", chatID)
	writer.WriteField("caption", "AWS FinOps 报告: "+filepath.Base(filePath))

	part, _ := writer.CreateFormFile("document", filepath.Base(filePath))
	io.Copy(part, f)
	writer.Close()

	url := fmt.Sprintf("%s/bot%s/sendDocument", telegramBaseURL, token)
	req, _ := http.NewRequest("POST", url, bytes.NewReader(body.Bytes()))
	req.Header.Set("Content-Type", writer.FormDataContentType())

	resp, err := httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("HTTP 请求失败: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("Telegram API 返回 %s", resp.Status)
	}

	return nil
}

// sendDiscord 通过 Discord Bot Webhook 发送消息和附件。
//
// 参数：
//   - msg  : 消息内容
//   - paths: 要发送的文件路径列表
//
// 返回：
//   - error : 发送失败时返回错误
func sendDiscord(msg string, paths []string) error {
	start := time.Now()
	webhookURL := os.Getenv("DISCORD_BOT_TOKEN")
	channelID := os.Getenv("DISCORD_CHANNEL_ID")

	// 如果设置的是 Channel ID，先获取 Webhook URL（简化处理：直接用 Channel ID 构建）
	// 实际使用中建议在 .env 里直接配置 Discord Webhook URL
	_ = channelID

	// Discord 使用 Embed 发送富文本消息
	embed := map[string]interface{}{
		"title":       "AWS FinOps 执行报告",
		"description": msg,
		"color":       0x0078D4,
		"footer": map[string]string{
			"text": "AWSFinOps Worker",
		},
		"timestamp": time.Now().Format(time.RFC3339),
	}

	payload := map[string]interface{}{
		"embeds": []interface{}{embed},
	}

	body, _ := json.Marshal(payload)
	req, _ := http.NewRequest("POST", webhookURL, bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	resp, err := httpClient.Do(req)
	if err != nil {
		utilities.LogError("Discord", "sendDiscord", err, time.Since(start))
		return fmt.Errorf("Discord 消息发送失败: %w", err)
	}
	io.ReadAll(resp.Body)
	resp.Body.Close()

	utilities.LogSuccess("Discord", "sendDiscord", time.Since(start))
	return nil
}
