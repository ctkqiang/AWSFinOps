package main

import (
	"aws_fin_ops/internal"
	"aws_fin_ops/internal/utilities"
	"context"
	"os"
)

func main() {
	initialContext := context.Background()

	if err := internal.LoadConfig(initialContext); err != nil {
		utilities.Error("配置加载失败: %v", err)
		os.Exit(1)
	}
}
