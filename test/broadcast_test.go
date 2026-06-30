package test

import (
	"os"
	"strings"
	"testing"

	"aws_fin_ops/internal/services"
)

// TestTelegramGetBotFather_WithToken 验证 token 已设置时能正确构建 getUpdates URL。
func TestTelegramGetBotFather_WithToken(t *testing.T) {
	const token = "123456:ABC-DEF1234ghIkl-zyx57W2v1u123ew11"
	os.Setenv("TELEGRAM_BOT_TOKEN", token)
	defer os.Unsetenv("TELEGRAM_BOT_TOKEN")

	got := services.TelegramGetBotFather()
	want := "https://api.telegram.org/bot" + token + "/getUpdates"
	if got != want {
		t.Fatalf("期望 %q，实际得到 %q", want, got)
	}
}

// TestTelegramGetBotFather_NoToken 验证 token 未设置时返回空字符串。
func TestTelegramGetBotFather_NoToken(t *testing.T) {
	os.Unsetenv("TELEGRAM_BOT_TOKEN")

	got := services.TelegramGetBotFather()
	if got != "" {
		t.Fatalf("期望 token 缺失时返回空字符串，实际得到 %q", got)
	}
}

// TestTelegramSendMessage_NoToken 验证 TELEGRAM_BOT_TOKEN 缺失时返回包含变量名的错误。
func TestTelegramSendMessage_NoToken(t *testing.T) {
	os.Unsetenv("TELEGRAM_BOT_TOKEN")
	os.Setenv("TELEGRAM_CHAT_ID", "123")
	defer os.Unsetenv("TELEGRAM_CHAT_ID")

	err := services.TelegramSendMessage("hello")
	if err == nil {
		t.Fatal("期望 TELEGRAM_BOT_TOKEN 缺失时返回错误")
	}
	if !strings.Contains(err.Error(), "TELEGRAM_BOT_TOKEN") {
		t.Fatalf("错误信息应包含 TELEGRAM_BOT_TOKEN，实际得到: %v", err)
	}
}

// TestTelegramSendMessage_NoChatID 验证 TELEGRAM_CHAT_ID 缺失时返回包含变量名的错误。
func TestTelegramSendMessage_NoChatID(t *testing.T) {
	os.Setenv("TELEGRAM_BOT_TOKEN", "fake-token")
	defer os.Unsetenv("TELEGRAM_BOT_TOKEN")
	os.Unsetenv("TELEGRAM_CHAT_ID")

	err := services.TelegramSendMessage("hello")
	if err == nil {
		t.Fatal("期望 TELEGRAM_CHAT_ID 缺失时返回错误")
	}
	if !strings.Contains(err.Error(), "TELEGRAM_CHAT_ID") {
		t.Fatalf("错误信息应包含 TELEGRAM_CHAT_ID，实际得到: %v", err)
	}
}
