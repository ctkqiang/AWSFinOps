package test

import (
	"os"
	"testing"

	"aws_fin_ops/internal"
)

// TestGetEnvValue_Set 验证环境变量已设置时能正确返回值。
func TestGetEnvValue_Set(t *testing.T) {
	const key = "TEST_GET_ENV_VALUE_SET"
	const want = "hello123"
	os.Setenv(key, want)
	defer os.Unsetenv(key)

	got, err := internal.GetEnvValue(key)
	if err != nil {
		t.Fatalf("期望无错误，实际得到 %v", err)
	}
	if got != want {
		t.Fatalf("期望 %q，实际得到 %q", want, got)
	}
}

// TestGetEnvValue_Empty 验证环境变量未设置时返回错误。
func TestGetEnvValue_Empty(t *testing.T) {
	const key = "TEST_GET_ENV_VALUE_EMPTY"
	os.Unsetenv(key)

	_, err := internal.GetEnvValue(key)
	if err == nil {
		t.Fatal("期望未设置的环境变量返回错误，实际得到 nil")
	}
}

// TestGetEnvValue_ExplicitlyEmpty 验证环境变量显式设为空字符串时返回错误。
func TestGetEnvValue_ExplicitlyEmpty(t *testing.T) {
	const key = "TEST_GET_ENV_VALUE_EXPLICIT_EMPTY"
	os.Setenv(key, "")
	defer os.Unsetenv(key)

	_, err := internal.GetEnvValue(key)
	if err == nil {
		t.Fatal("期望空字符串环境变量返回错误，实际得到 nil")
	}
}
