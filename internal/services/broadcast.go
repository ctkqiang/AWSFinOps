package services

import "aws_fin_ops/internal/utilities"

type (
	FileType         string
	BroadcastService int
)

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

type File struct {
	Name string
	Type FileType
}

func Broadcast(message string, file File) {
	utilities.LogProgress(
		"broadcast",
		"broadcast",
		message,
		file.Name,
	)
}

func TelegramGetBotFather() string {
	return ""
}
