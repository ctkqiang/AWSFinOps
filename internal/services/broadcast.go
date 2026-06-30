package services

import (
	"aws_fin_ops/internal"
	"aws_fin_ops/internal/utilities"
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
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

	utilities.Info("Telegram getUpdates URL built (token=%s)", utilities.Mask(token))
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
		utilities.Error("TelegramSendMessage: failed to marshal payload: %v", err)
		return fmt.Errorf("TelegramSendMessage: marshal: %w", err)
	}

	url := fmt.Sprintf("%s/bot%s/sendMessage", telegramBaseURL, token)

	resp, err := http.Post(url, "application/json", bytes.NewBuffer(data))
	if err != nil {
		utilities.LogError("Telegram", "SendMessage", err, time.Since(start))
		return fmt.Errorf("TelegramSendMessage: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		err := fmt.Errorf("telegram API returned %s", resp.Status)
		utilities.LogError("Telegram", "SendMessage", err, time.Since(start))
		return err
	}

	utilities.LogSuccess("Telegram", "SendMessage", time.Since(start),
		fmt.Sprintf("ChatID=%s", utilities.Mask(chatID)),
	)
	return nil
}
