package utilities

import (
	"encoding/json"
	"fmt"
	"os"
	"runtime"
	"strings"
	"sync"
	"time"
)

const (
	APP_NAME = "AWSFinOps"
	VERSION  = "1.0.0"
	TZ       = "CST"
)

// LogLevel controls the minimum severity emitted.
type LogLevel int

const (
	DEBUG   LogLevel = iota // verbose diagnostics (local dev only)
	INFO                    // general operational messages (default)
	WARN                    // recoverable issues, degraded mode
	ERROR                   // failures requiring attention
	VERBOSE                 // per-request metrics, audit trails
)

func (l LogLevel) String() string {
	switch l {
	case DEBUG:
		return "DEBUG"
	case INFO:
		return "INFO"
	case WARN:
		return "WARN"
	case ERROR:
		return "ERROR"
	case VERBOSE:
		return "VERBOSE"
	default:
		return "UNKNOWN"
	}
}

// cloudWatchLevel maps internal levels to CloudWatch / EMF standard severity strings.
// CloudWatch Logs Insights recognises these values for metric filters.
func (l LogLevel) cloudWatchLevel() string {
	switch l {
	case DEBUG:
		return "DEBUG"
	case INFO:
		return "INFO"
	case WARN:
		return "WARNING"
	case ERROR:
		return "ERROR"
	case VERBOSE:
		return "INFO"
	default:
		return "INFO"
	}
}

const (
	colorReset  = "\033[0m"
	colorRed    = "\033[31m"
	colorYellow = "\033[33m"
	colorBlue   = "\033[34m"
	colorPink   = "\033[35m"
	colorGreen  = "\033[32m"
	colorBold   = "\033[1m"
	colorNoBold = "\033[22m"
)

var (
	CurrentLevel  = INFO
	startTime     = time.Now()
	errorCallback func(string)
	goroutineSeq  int
	goroutineMu   sync.Mutex

	// cloudWatchMode is true when the CLOUDWATCH_MODE env var is set to "1" or "true".
	// In this mode all output is newline-delimited JSON (no ANSI escapes) so that the
	// CloudWatch Logs agent / Lambda runtime can parse and index every field.
	cloudWatchMode bool
)

// cloudWatchEntry is the canonical JSON schema emitted in CloudWatch mode.
// Field names follow the CloudWatch Embedded Metric Format (EMF) conventions
// so that Logs Insights can filter on them without a custom parser.
type cloudWatchEntry struct {
	Timestamp string `json:"timestamp"`          // RFC3339Nano, UTC — required by CloudWatch
	Level     string `json:"level"`              // CloudWatch standard severity
	App       string `json:"app"`                // application / service name
	Version   string `json:"version"`            // semantic version
	Component string `json:"component"`          // subsystem (e.g. "BillingService")
	Operation string `json:"operation"`          // action being performed
	Status    string `json:"status,omitempty"`   // START | OK | FAIL | IN_PROGRESS | WARN
	TaskID    string `json:"taskId,omitempty"`   // monotonic task sequence identifier
	Caller    string `json:"caller,omitempty"`   // function name at call site
	MemoryMB  string `json:"memoryMB,omitempty"` // heap allocated at log time
	Elapsed   string `json:"elapsed,omitempty"`  // human-readable duration
	Message   string `json:"message"`            // free-form log message
	// Extra holds all caller-supplied key=value detail pairs.
	Extra map[string]string `json:"extra,omitempty"`
}

// SetLogLevel parses a string level name and sets the global threshold.
// Defaults to INFO on unrecognised input. Called automatically from
// init() via the LOG_LEVEL env var.
func SetLogLevel(s string) {
	switch strings.ToUpper(s) {
	case "DEBUG":
		CurrentLevel = DEBUG
	case "INFO":
		CurrentLevel = INFO
	case "WARN":
		CurrentLevel = WARN
	case "ERROR":
		CurrentLevel = ERROR
	case "VERBOSE":
		CurrentLevel = VERBOSE
	default:
		CurrentLevel = INFO
	}
}

// RegisterErrorCallback sets a callback invoked on every ERROR log line.
func RegisterErrorCallback(cb func(string)) { errorCallback = cb }

// Bold wraps text with ANSI bold escapes for emphasis inside log lines.
// In CloudWatch mode the escapes are stripped so the raw text is returned.
func Bold(text string) string {
	if cloudWatchMode {
		return text
	}
	return colorBold + text + colorNoBold
}

// Error emits an ERROR-level log.
func Error(format string, a ...interface{}) { Log(ERROR, format, a...) }

// Info emits an INFO-level log.
func Info(format string, a ...interface{}) { Log(INFO, format, a...) }

// Debug emits a DEBUG-level log. Only visible when LOG_LEVEL=DEBUG.
func Debug(format string, a ...interface{}) { Log(DEBUG, format, a...) }

// Warn emits a WARN-level log.
func Warn(format string, a ...interface{}) { Log(WARN, format, a...) }

// Log is the single-line logger. In CloudWatch mode it emits a minimal JSON
// object; in local mode it emits a coloured human-readable line.
func Log(level LogLevel, format string, a ...interface{}) {
	if level < CurrentLevel {
		return
	}
	msg := fmt.Sprintf(format, a...)

	if cloudWatchMode {
		entry := cloudWatchEntry{
			Timestamp: utcNow(),
			Level:     level.cloudWatchLevel(),
			App:       APP_NAME,
			Version:   VERSION,
			Component: "General",
			Operation: "Log",
			Message:   msg,
		}
		emitJSON(entry)
	} else {
		line := fmt.Sprintf("[%s] [%s] [%s] %s",
			APP_NAME, time.Now().Format("2006-01-02 15:04:05"), level.String(), msg)
		c := levelColor(level)
		if c != "" {
			fmt.Fprintf(os.Stderr, "%s%s%s\n", c, line, colorReset)
		} else {
			fmt.Fprintln(os.Stderr, line)
		}
	}

	if level == ERROR && errorCallback != nil {
		errorCallback(msg)
	}
}

// Logf emits a structured log entry. In CloudWatch mode every field is a
// top-level JSON key so Logs Insights can filter/aggregate without parsing
// nested strings. In local mode the original coloured block format is kept.
func Logf(component, operation string, level LogLevel, status string, elapsed time.Duration, details ...string) {
	if level < CurrentLevel {
		return
	}
	id := nextTaskID()
	funcName := callerName(3)
	heapMB := heapAllocMB()

	if cloudWatchMode {
		extra := make(map[string]string, len(details))
		for _, d := range details {
			k, v, ok := strings.Cut(d, "=")
			if ok {
				extra[strings.TrimSpace(k)] = strings.TrimSpace(v)
			} else if d != "" {
				extra[d] = ""
			}
		}
		entry := cloudWatchEntry{
			Timestamp: utcNow(),
			Level:     level.cloudWatchLevel(),
			App:       APP_NAME,
			Version:   VERSION,
			Component: component,
			Operation: operation,
			Status:    status,
			TaskID:    id,
			Caller:    funcName,
			MemoryMB:  fmt.Sprintf("%.2f", heapMB),
			Elapsed:   fmtElapsed(elapsed),
			Message:   fmt.Sprintf("%s::%s [%s]", component, operation, status),
			Extra:     extra,
		}
		emitJSON(entry)
	} else {
		header := fmt.Sprintf("[%s@%s]::%s:: (%s:%s>>%s::%s)",
			APP_NAME, nowCompact(), level.String(), component, operation, id, funcName)

		rows := [][]string{
			{"Status", status},
			{"Type", "ACTION"},
			{"Memory", fmt.Sprintf("%.2fMB", heapMB)},
			{"Routine", id},
			{"Elapsed", fmtElapsed(elapsed)},
		}
		for _, d := range details {
			k, v, ok := strings.Cut(d, "=")
			if ok {
				rows = append(rows, []string{strings.TrimSpace(k), strings.TrimSpace(v)})
			} else {
				rows = append(rows, []string{d, ""})
			}
		}
		c := levelColor(level)
		fmt.Fprint(os.Stderr, buildBlock(header, c, rows))
	}

	if level == ERROR && errorCallback != nil {
		errorCallback(fmt.Sprintf("%s::%s [%s]", component, operation, status))
	}
}

// LogProgress emits an INFO-level IN_PROGRESS log for intermediate checkpoints.
func LogProgress(component, operation, msg string, details ...string) {
	resolved := msg
	if strings.Contains(msg, "%") && len(details) > 0 {
		verbCount := strings.Count(msg, "%s") + strings.Count(msg, "%d") + strings.Count(msg, "%v")
		if verbCount > 0 && verbCount <= len(details) {
			args := make([]interface{}, verbCount)
			for i := 0; i < verbCount; i++ {
				args[i] = details[i]
			}
			resolved = fmt.Sprintf(msg, args...)
			details = details[verbCount:]
		}
	}
	all := append([]string{"Progress=" + resolved}, details...)
	Logf(component, operation, INFO, "IN_PROGRESS", 0, all...)
}

// LogStart emits a START marker for an operation.
func LogStart(component, operation string) {
	Logf(component, operation, INFO, "START", 0)
}

// LogSuccess emits an OK marker with elapsed time.
func LogSuccess(component, operation string, elapsed time.Duration, details ...string) {
	Logf(component, operation, INFO, "OK", elapsed, details...)
}

// LogError emits a FAIL marker with the error message.
func LogError(component, operation string, err error, elapsed time.Duration, details ...string) {
	all := append([]string{"Error=" + err.Error()}, details...)
	Logf(component, operation, ERROR, "FAIL", elapsed, all...)
}

// LogWarn emits a WARN marker.
func LogWarn(component, operation, msg string, elapsed time.Duration, details ...string) {
	all := append([]string{"Warn=" + msg}, details...)
	Logf(component, operation, WARN, "WARN", elapsed, all...)
}

// Mask redacts a sensitive value, showing the first few characters followed
// by [REDACTED]. Useful for tokens, keys, and PII in logs.
func Mask(s string) string {
	runes := []rune(s)
	if len(runes) <= 4 {
		return "****"
	}
	n := 10
	if len(runes) <= n {
		n = len(runes) / 3
	}
	return string(runes[:n]) + "[REDACTED]"
}

// RetryWithBackoff executes an operation up to maxAttempts times with a
// fixed backoff between attempts. Returns the last error on exhaustion.
func RetryWithBackoff(name string, maxAttempts int, backoff time.Duration, fn func() error) error {
	var last error
	for i := 0; i < maxAttempts; i++ {
		if err := fn(); err == nil {
			return nil
		} else {
			last = err
			Warn("%s attempt %d/%d failed: %v — retrying in %v", name, i+1, maxAttempts, err, backoff)
			time.Sleep(backoff)
		}
	}
	return fmt.Errorf("%s: exhausted %d retries: %w", name, maxAttempts, last)
}

func init() {
	SetLogLevel(os.Getenv("LOG_LEVEL"))
	v := strings.ToLower(strings.TrimSpace(os.Getenv("CLOUDWATCH_MODE")))
	cloudWatchMode = v == "1" || v == "true"
}

// utcNow returns the current time as an RFC3339Nano string in UTC.
// CloudWatch requires UTC timestamps for correct log stream ordering.
func utcNow() string {
	return time.Now().UTC().Format(time.RFC3339Nano)
}

// emitJSON serialises entry to a single-line JSON object and writes it to
// stdout. CloudWatch agents and Lambda runtimes read from stdout; each
// newline-terminated JSON object becomes one log event.
func emitJSON(entry cloudWatchEntry) {
	b, err := json.Marshal(entry)
	if err != nil {
		fmt.Fprintf(os.Stderr, `{"level":"ERROR","message":"logger: json.Marshal failed: %s"}`+"\n", err.Error())
		return
	}
	fmt.Fprintf(os.Stdout, "%s\n", b)
}

func levelColor(l LogLevel) string {
	switch l {
	case DEBUG:
		return colorYellow
	case INFO:
		return colorBlue
	case WARN:
		return colorPink
	case ERROR:
		return colorRed
	case VERBOSE:
		return colorGreen
	default:
		return ""
	}
}

func nowCompact() string {
	t := time.Now()
	return fmt.Sprintf("%d%02d%02d:%02d:%02d:%02d%s",
		t.Year(), t.Month(), t.Day(), t.Hour(), t.Minute(), t.Second(), TZ)
}

func nextTaskID() string {
	goroutineMu.Lock()
	id := goroutineSeq
	goroutineSeq++
	goroutineMu.Unlock()
	return fmt.Sprintf("TASK-%03d", id)
}

func callerName(depth int) string {
	pc, _, _, ok := runtime.Caller(depth)
	if !ok {
		return "Unknown"
	}
	name := runtime.FuncForPC(pc).Name()
	if i := strings.LastIndexByte(name, '.'); i >= 0 {
		return name[i+1:]
	}
	return name
}

func heapAllocMB() float64 {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	return float64(m.Alloc) / 1024 / 1024
}

func fmtElapsed(d time.Duration) string {
	switch {
	case d == 0:
		return "0us"
	case d < time.Microsecond:
		return fmt.Sprintf("%.2fus", float64(d.Nanoseconds())/1000.0)
	case d < time.Millisecond:
		return fmt.Sprintf("%.2fus", float64(d.Microseconds()))
	case d < time.Second:
		return fmt.Sprintf("%.2fms", float64(d.Milliseconds()))
	default:
		return fmt.Sprintf("%.2fs", d.Seconds())
	}
}

func buildBlock(header, color string, rows [][]string) string {
	var sb strings.Builder
	keyW := 0
	for _, r := range rows {
		if len(r) == 2 && len(r[0]) > keyW {
			keyW = len(r[0])
		}
	}
	sb.WriteString(color)
	sb.WriteString(colorBold)
	sb.WriteString(header)
	sb.WriteString(colorReset)
	sb.WriteString("\n")
	for _, r := range rows {
		if len(r) == 2 {
			fmt.Fprintf(&sb, "%s  | %-*s : %s%s\n", color, keyW, r[0], r[1], colorReset)
		}
	}
	return sb.String()
}
