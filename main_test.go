package main

import (
	"os"
	"testing"
	"time"
)

func TestResolveWorkerInterval_Default(t *testing.T) {
	os.Unsetenv("WORKER_INTERVAL")

	got := resolveWorkerInterval()
	if got != defaultWorkerInterval {
		t.Fatalf("期望默认值 %v，实际得到 %v", defaultWorkerInterval, got)
	}
}

func TestResolveWorkerInterval_DurationFormat(t *testing.T) {
	cases := []struct {
		input    string
		expected time.Duration
	}{
		{"30s", 30 * time.Second},
		{"5m", 5 * time.Minute},
		{"2h", 2 * time.Hour},
		{"1h30m", 90 * time.Minute},
	}

	for _, tc := range cases {
		t.Run(tc.input, func(t *testing.T) {
			os.Setenv("WORKER_INTERVAL", tc.input)
			defer os.Unsetenv("WORKER_INTERVAL")

			got := resolveWorkerInterval()
			if got != tc.expected {
				t.Fatalf("WORKER_INTERVAL=%s: 期望 %v，实际得到 %v", tc.input, tc.expected, got)
			}
		})
	}
}

func TestResolveWorkerInterval_PureSeconds(t *testing.T) {
	os.Setenv("WORKER_INTERVAL", "120")
	defer os.Unsetenv("WORKER_INTERVAL")

	got := resolveWorkerInterval()
	if got != 120*time.Second {
		t.Fatalf("期望 120s，实际得到 %v", got)
	}
}

func TestResolveWorkerInterval_TooShort(t *testing.T) {
	cases := []string{"5s", "1s", "3"}

	for _, input := range cases {
		t.Run(input, func(t *testing.T) {
			os.Setenv("WORKER_INTERVAL", input)
			defer os.Unsetenv("WORKER_INTERVAL")

			got := resolveWorkerInterval()
			if got != defaultWorkerInterval {
				t.Fatalf("WORKER_INTERVAL=%s 过短，期望回退到 %v，实际得到 %v", input, defaultWorkerInterval, got)
			}
		})
	}
}

func TestResolveWorkerInterval_Invalid(t *testing.T) {
	cases := []string{"abc", "never", "3x", ""}

	for _, input := range cases {
		t.Run("input_"+input, func(t *testing.T) {
			if input == "" {
				os.Unsetenv("WORKER_INTERVAL")
			} else {
				os.Setenv("WORKER_INTERVAL", input)
				defer os.Unsetenv("WORKER_INTERVAL")
			}

			got := resolveWorkerInterval()
			if got != defaultWorkerInterval {
				t.Fatalf("WORKER_INTERVAL=%q 无效，期望回退到 %v，实际得到 %v", input, defaultWorkerInterval, got)
			}
		})
	}
}

func TestIsWorkerOnce_True(t *testing.T) {
	cases := []string{"true", "TRUE", "True", " true ", " TRUE "}

	for _, input := range cases {
		t.Run(input, func(t *testing.T) {
			os.Setenv("WORKER_ONCE", input)
			defer os.Unsetenv("WORKER_ONCE")

			if !isWorkerOnce() {
				t.Fatalf("WORKER_ONCE=%q 应返回 true", input)
			}
		})
	}
}

func TestIsWorkerOnce_False(t *testing.T) {
	cases := []string{"false", "FALSE", "0", "yes", ""}

	for _, input := range cases {
		t.Run("input_"+input, func(t *testing.T) {
			if input == "" {
				os.Unsetenv("WORKER_ONCE")
			} else {
				os.Setenv("WORKER_ONCE", input)
				defer os.Unsetenv("WORKER_ONCE")
			}

			if isWorkerOnce() {
				t.Fatalf("WORKER_ONCE=%q 应返回 false", input)
			}
		})
	}
}

func TestResolveWorkerInterval_WithSpaces(t *testing.T) {
	os.Setenv("WORKER_INTERVAL", "  1h  ")
	defer os.Unsetenv("WORKER_INTERVAL")

	got := resolveWorkerInterval()
	if got != time.Hour {
		t.Fatalf("期望 1h，实际得到 %v", got)
	}
}
