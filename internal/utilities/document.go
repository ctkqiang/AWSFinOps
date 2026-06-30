package utilities

import (
	_ "embed"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/go-pdf/fpdf"
)

//go:embed fonts/NotoSansSC-Regular.ttf
var notoSansSCFont []byte

const (
	// documentComponent 是本文件在日志中使用的组件名称标识。
	documentComponent = "DocumentExporter"

	// fontFamily 是用于 PDF 的字体族名称（Noto Sans SC，完整支持中文）。
	fontFamily = "NotoSansSC"

	// pdfMarginLeft 是 PDF 页面左边距（毫米）。
	pdfMarginLeft = 20.0

	// pdfMarginTop 是 PDF 页面上边距（毫米）。
	pdfMarginTop = 20.0

	// pdfMarginRight 是 PDF 页面右边距（毫米）。
	pdfMarginRight = 20.0

	// pdfMarginBottom 是 PDF 页面下边距（毫米）。
	pdfMarginBottom = 20.0

	// pdfPageWidth 是 A4 纸宽度（毫米）。
	pdfPageWidth = 210.0

	// pdfPageHeight 是 A4 纸高度（毫米）。
	pdfPageHeight = 297.0

	// pdfContentWidth 是 PDF 有效内容区宽度（毫米）。
	pdfContentWidth = pdfPageWidth - pdfMarginLeft - pdfMarginRight

	// watermarkAngle 是水印文本的旋转角度（度）。
	watermarkAngle = 45.0

	// watermarkFontSize 是水印文本的字号。
	watermarkFontSize = 60.0

	// watermarkAlpha 是水印文本的透明度（0.0 完全透明 ~ 1.0 完全不透明）。
	watermarkAlpha = 0.08
)

// ReportData 是文档导出的统一数据输入结构。
// 由 Worker 引擎在执行完毕后填充，传递给三种导出函数。
type ReportData struct {
	RunID        string           `json:"run_id"`
	AWSAccountID string           `json:"aws_account_id"`
	StartTime    time.Time        `json:"start_time"`
	EndTime      time.Time        `json:"end_time"`
	Duration     time.Duration    `json:"duration"`
	Status       string           `json:"status"`
	Steps        []ReportStepData `json:"steps"`
	Summary      string           `json:"summary"`
}

// ReportStepData 是单个步骤的导出数据。
type ReportStepData struct {
	Index    int           `json:"index"`
	Step     string        `json:"step"`
	Status   string        `json:"status"`
	Duration time.Duration `json:"duration"`
	Message  string        `json:"message,omitempty"`
	Error    string        `json:"error,omitempty"`
}

// ExportConfig 控制文档导出行为。
type ExportConfig struct {
	OutputDir    string `json:"output_dir"`
	EnablePDF    bool   `json:"enable_pdf"`
	EnableJSON   bool   `json:"enable_json"`
	EnableCSV    bool   `json:"enable_csv"`
	AWSAccountID string `json:"aws_account_id"`
}

// DefaultExportConfig 返回默认导出配置：三种格式全部启用，输出到 ./reports/ 目录。
//
// 返回：
//   - ExportConfig : 默认导出配置
func DefaultExportConfig() ExportConfig {
	return ExportConfig{
		OutputDir:  "./reports",
		EnablePDF:  true,
		EnableJSON: true,
		EnableCSV:  true,
	}
}

// cleanOldReports 清理输出目录中旧的 FinOps 报告文件（PDF/JSON/CSV）。
// 在生成新报告之前调用，确保目录中只保留最新一次运行的报告。
//
// 参数：
//   - outputDir : 报告输出目录
func cleanOldReports(outputDir string) {
	entries, err := os.ReadDir(outputDir)
	if err != nil {
		LogWarn(documentComponent, "cleanOldReports",
			fmt.Sprintf("读取目录失败，跳过清理: %v", err),
			0,
		)
		return
	}

	cleaned := 0
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		if strings.HasPrefix(name, "finops_report_") &&
			(strings.HasSuffix(name, ".pdf") ||
				strings.HasSuffix(name, ".json") ||
				strings.HasSuffix(name, ".csv")) {
			if err := os.Remove(filepath.Join(outputDir, name)); err != nil {
				LogWarn(documentComponent, "cleanOldReports",
					fmt.Sprintf("删除旧文件失败 %s: %v", name, err),
					0,
				)
				continue
			}
			cleaned++
		}
	}

	if cleaned > 0 {
		LogProgress(documentComponent, "cleanOldReports",
			fmt.Sprintf("已清理 %d 个旧报告文件", cleaned),
		)
	}
}

// ExportAll 根据配置依次导出 PDF / JSON / CSV 三种格式的报告。
// 任意一种格式导出失败不会影响其余格式的导出。
//
// 参数：
//   - data : 统一的报告数据
//   - cfg  : 导出配置，传 nil 则使用默认配置
//
// 返回：
//   - []string : 成功生成的文件路径列表
//   - error    : 全部失败时返回聚合错误，部分成功时返回 nil
func ExportAll(data *ReportData, cfg *ExportConfig) ([]string, error) {
	start := time.Now()
	LogStart(documentComponent, "ExportAll")

	c := DefaultExportConfig()
	if cfg != nil {
		if cfg.OutputDir != "" {
			c.OutputDir = cfg.OutputDir
		}
		c.EnablePDF = cfg.EnablePDF
		c.EnableJSON = cfg.EnableJSON
		c.EnableCSV = cfg.EnableCSV
		if cfg.AWSAccountID != "" {
			c.AWSAccountID = cfg.AWSAccountID
		}
	}

	if c.AWSAccountID != "" {
		data.AWSAccountID = c.AWSAccountID
	}

	if err := os.MkdirAll(c.OutputDir, 0755); err != nil {
		LogError(documentComponent, "ExportAll", err, time.Since(start), "step=mkdir")
		return nil, fmt.Errorf("创建输出目录失败: %w", err)
	}

	cleanOldReports(c.OutputDir)

	var paths []string
	var errs []string
	ts := data.StartTime.Format("20060102_150405")

	if c.EnablePDF {
		p := filepath.Join(c.OutputDir, fmt.Sprintf("finops_report_%s.pdf", ts))
		if err := ExportAsPDF(data, p); err != nil {
			errs = append(errs, fmt.Sprintf("PDF: %v", err))
		} else {
			paths = append(paths, p)
		}
	}

	if c.EnableJSON {
		p := filepath.Join(c.OutputDir, fmt.Sprintf("finops_report_%s.json", ts))
		if err := ExportAsJSON(data, p); err != nil {
			errs = append(errs, fmt.Sprintf("JSON: %v", err))
		} else {
			paths = append(paths, p)
		}
	}

	if c.EnableCSV {
		p := filepath.Join(c.OutputDir, fmt.Sprintf("finops_report_%s.csv", ts))
		if err := ExportAsCSV(data, p); err != nil {
			errs = append(errs, fmt.Sprintf("CSV: %v", err))
		} else {
			paths = append(paths, p)
		}
	}

	if len(paths) == 0 && len(errs) > 0 {
		err := fmt.Errorf("所有导出格式均失败: %s", strings.Join(errs, "; "))
		LogError(documentComponent, "ExportAll", err, time.Since(start))
		return nil, err
	}

	LogSuccess(documentComponent, "ExportAll", time.Since(start),
		fmt.Sprintf("files=%d", len(paths)),
		fmt.Sprintf("errors=%d", len(errs)),
	)
	return paths, nil
}

// ExportAsPDF 生成高级排版的 PDF 执行报告（面向高管 / 决策层）。
//
// PDF 特性：
//   - A4 纵向布局，专业的标题页 + 执行摘要 + 步骤详情表格
//   - AWS Account ID（大写）作为对角线水印印在每一页
//   - 颜色编码的状态标签（OK=绿色, WARN=橙色, FAIL=红色, STUB=灰色）
//   - 页眉含应用名 + 运行 ID，页脚含页码和生成时间
//
// 参数：
//   - data     : 统一的报告数据
//   - filePath : 输出文件路径
//
// 返回：
//   - error : 生成或写入失败时返回错误
func ExportAsPDF(data *ReportData, filePath string) error {
	start := time.Now()
	LogStart(documentComponent, "ExportAsPDF")

	pdf := fpdf.New("P", "mm", "A4", "")
	pdf.SetAutoPageBreak(true, pdfMarginBottom)

	// 注册 Noto Sans SC 中文字体，确保所有 Unicode 字符正确渲染。
	// 风格参数必须为空字符串 ""，与 SetFont 的 lookup key 逻辑一致。
	pdf.AddUTF8FontFromBytes(fontFamily, "", notoSansSCFont)

	accountWatermark := strings.ToUpper(strings.TrimSpace(data.AWSAccountID))
	if accountWatermark == "" {
		accountWatermark = "CONFIDENTIAL"
	}

	pdf.SetHeaderFuncMode(func() {
		pdf.SetFont(fontFamily, "", 8)
		pdf.SetTextColor(150, 150, 150)
		pdf.CellFormat(pdfContentWidth/2, 6, "AWS FinOps Report", "", 0, "L", false, 0, "")
		pdf.CellFormat(pdfContentWidth/2, 6, data.RunID, "", 0, "R", false, 0, "")
		pdf.Ln(8)
		pdf.SetDrawColor(0, 120, 215)
		pdf.SetLineWidth(0.5)
		pdf.Line(pdfMarginLeft, pdf.GetY(), pdfPageWidth-pdfMarginRight, pdf.GetY())
		pdf.Ln(4)
	}, true)

	pdf.SetFooterFunc(func() {
		pdf.SetY(-15)
		pdf.SetFont(fontFamily, "", 7)
		pdf.SetTextColor(150, 150, 150)
		pdf.CellFormat(pdfContentWidth/2, 8,
			fmt.Sprintf("Generated: %s", time.Now().Format("2006-01-02 15:04:05 MST")),
			"", 0, "L", false, 0, "")
		pdf.CellFormat(pdfContentWidth/2, 8,
			fmt.Sprintf("Page %d / {nb}", pdf.PageNo()),
			"", 0, "R", false, 0, "")
	})

	pdf.AliasNbPages("")
	pdf.AddPage()

	drawWatermark(pdf, accountWatermark)
	drawTitleSection(pdf, data)
	drawExecutiveSummary(pdf, data)
	drawStepsTable(pdf, data, accountWatermark)
	drawFooterNote(pdf, data)

	if err := pdf.OutputFileAndClose(filePath); err != nil {
		LogError(documentComponent, "ExportAsPDF", err, time.Since(start))
		return fmt.Errorf("PDF 文件写入失败: %w", err)
	}

	LogSuccess(documentComponent, "ExportAsPDF", time.Since(start),
		fmt.Sprintf("path=%s", filePath),
	)
	return nil
}

// drawWatermark 在当前页面绘制对角线大写水印。
func drawWatermark(pdf *fpdf.Fpdf, text string) {
	pdf.SetFont(fontFamily, "", watermarkFontSize)
	pdf.SetAlpha(watermarkAlpha, "Normal")
	pdf.SetTextColor(180, 180, 180)

	centerX := pdfPageWidth / 2
	centerY := pdfPageHeight / 2

	pdf.TransformBegin()
	pdf.TransformRotate(watermarkAngle, centerX, centerY)
	textW := pdf.GetStringWidth(text)
	pdf.SetXY(centerX-textW/2, centerY-watermarkFontSize/2)
	pdf.CellFormat(textW, watermarkFontSize, text, "", 0, "C", false, 0, "")
	pdf.TransformEnd()

	pdf.SetAlpha(1.0, "Normal")
}

// drawTitleSection 绘制报告标题区域。
func drawTitleSection(pdf *fpdf.Fpdf, data *ReportData) {
	pdf.Ln(10)

	pdf.SetFont(fontFamily, "", 24)
	pdf.SetTextColor(0, 51, 102)
	pdf.CellFormat(pdfContentWidth, 14, "AWS FinOps", "", 0, "C", false, 0, "")
	pdf.Ln(16)

	pdf.SetFont(fontFamily, "", 14)
	pdf.SetTextColor(0, 102, 153)
	pdf.CellFormat(pdfContentWidth, 8, "Cost Optimization Execution Report", "", 0, "C", false, 0, "")
	pdf.Ln(14)

	pdf.SetDrawColor(0, 120, 215)
	pdf.SetLineWidth(1.0)
	lineStart := (pdfPageWidth - 80) / 2
	pdf.Line(lineStart, pdf.GetY(), lineStart+80, pdf.GetY())
	pdf.Ln(10)

	pdf.SetFont(fontFamily, "", 10)
	pdf.SetTextColor(100, 100, 100)
	if data.AWSAccountID != "" {
		pdf.CellFormat(pdfContentWidth, 6,
			fmt.Sprintf("AWS Account: %s", strings.ToUpper(data.AWSAccountID)),
			"", 0, "C", false, 0, "")
		pdf.Ln(7)
	}
	pdf.CellFormat(pdfContentWidth, 6,
		fmt.Sprintf("Run ID: %s", data.RunID),
		"", 0, "C", false, 0, "")
	pdf.Ln(7)
	pdf.CellFormat(pdfContentWidth, 6,
		fmt.Sprintf("Date: %s", data.StartTime.Format("2006-01-02 15:04:05 MST")),
		"", 0, "C", false, 0, "")
	pdf.Ln(14)
}

// drawExecutiveSummary 绘制执行摘要面板。
func drawExecutiveSummary(pdf *fpdf.Fpdf, data *ReportData) {
	pdf.SetFont(fontFamily, "", 13)
	pdf.SetTextColor(0, 51, 102)
	pdf.CellFormat(pdfContentWidth, 8, "EXECUTIVE SUMMARY", "", 0, "L", false, 0, "")
	pdf.Ln(10)

	pdf.SetDrawColor(230, 230, 230)
	pdf.SetFillColor(245, 248, 252)
	pdf.SetLineWidth(0.3)

	panelY := pdf.GetY()
	panelH := 52.0
	pdf.RoundedRect(pdfMarginLeft, panelY, pdfContentWidth, panelH, 3, "1234", "FD")

	pdf.SetXY(pdfMarginLeft+6, panelY+5)

	statusR, statusG, statusB := statusColor(data.Status)
	pdf.SetFont(fontFamily, "", 11)
	pdf.SetTextColor(60, 60, 60)
	pdf.CellFormat(35, 7, "Status:", "", 0, "L", false, 0, "")
	pdf.SetTextColor(statusR, statusG, statusB)
	pdf.SetFont(fontFamily, "", 11)
	pdf.CellFormat(60, 7, strings.ToUpper(data.Status), "", 0, "L", false, 0, "")
	pdf.Ln(9)

	pdf.SetX(pdfMarginLeft + 6)
	pdf.SetTextColor(60, 60, 60)
	pdf.SetFont(fontFamily, "", 10)
	pdf.CellFormat(35, 7, "Duration:", "", 0, "L", false, 0, "")
	pdf.SetFont(fontFamily, "", 10)
	pdf.CellFormat(60, 7, data.Duration.Round(time.Millisecond).String(), "", 0, "L", false, 0, "")

	pdf.SetFont(fontFamily, "", 10)
	pdf.CellFormat(25, 7, "Steps:", "", 0, "L", false, 0, "")
	pdf.SetFont(fontFamily, "", 10)
	pdf.CellFormat(30, 7, fmt.Sprintf("%d", len(data.Steps)), "", 0, "L", false, 0, "")
	pdf.Ln(9)

	ok, pending, warned, failed := CountStepStatuses(data.Steps)

	pdf.SetX(pdfMarginLeft + 6)
	pdf.SetFont(fontFamily, "", 10)
	pdf.CellFormat(35, 7, "Breakdown:", "", 0, "L", false, 0, "")
	pdf.SetFont(fontFamily, "", 10)
	pdf.SetTextColor(0, 150, 0)
	pdf.CellFormat(0, 7,
		fmt.Sprintf("OK: %d  |  Pending: %d  |  Warn: %d  |  Fail: %d", ok, pending, warned, failed),
		"", 0, "L", false, 0, "")
	pdf.Ln(9)

	pdf.SetX(pdfMarginLeft + 6)
	pdf.SetTextColor(60, 60, 60)
	pdf.SetFont(fontFamily, "", 10)
	pdf.CellFormat(35, 7, "Period:", "", 0, "L", false, 0, "")
	pdf.SetFont(fontFamily, "", 10)
	pdf.CellFormat(0, 7,
		fmt.Sprintf("%s  ~  %s",
			data.StartTime.Format("2006-01-02 15:04:05"),
			data.EndTime.Format("2006-01-02 15:04:05")),
		"", 0, "L", false, 0, "")

	pdf.SetY(panelY + panelH + 8)
}

// drawStepsTable 绘制步骤详情表格。
func drawStepsTable(pdf *fpdf.Fpdf, data *ReportData, watermark string) {
	pdf.SetFont(fontFamily, "", 13)
	pdf.SetTextColor(0, 51, 102)
	pdf.CellFormat(pdfContentWidth, 8, "STEP DETAILS", "", 0, "L", false, 0, "")
	pdf.Ln(10)

	colWidths := []float64{12, 40, 22, 22, pdfContentWidth - 96}
	headers := []string{"#", "Step", "Status", "Duration", "Message"}

	pdf.SetFont(fontFamily, "", 9)
	pdf.SetFillColor(0, 51, 102)
	pdf.SetTextColor(255, 255, 255)
	pdf.SetDrawColor(0, 51, 102)
	pdf.SetLineWidth(0.3)

	for i, h := range headers {
		pdf.CellFormat(colWidths[i], 8, h, "1", 0, "C", true, 0, "")
	}
	pdf.Ln(-1)

	pdf.SetFont(fontFamily, "", 8)
	pdf.SetDrawColor(220, 220, 220)

	for idx, step := range data.Steps {
		if pdf.GetY() > pdfPageHeight-50 {
			pdf.AddPage()
			drawWatermark(pdf, watermark)

			pdf.SetFont(fontFamily, "", 9)
			pdf.SetFillColor(0, 51, 102)
			pdf.SetTextColor(255, 255, 255)
			pdf.SetDrawColor(0, 51, 102)
			for i, h := range headers {
				pdf.CellFormat(colWidths[i], 8, h, "1", 0, "C", true, 0, "")
			}
			pdf.Ln(-1)
			pdf.SetFont(fontFamily, "", 8)
			pdf.SetDrawColor(220, 220, 220)
		}

		if idx%2 == 0 {
			pdf.SetFillColor(250, 250, 255)
		} else {
			pdf.SetFillColor(255, 255, 255)
		}

		r, g, b := statusColor(step.Status)
		pdf.SetTextColor(60, 60, 60)

		msg := step.Message
		if step.Error != "" {
			msg = step.Error
		}

		lineHeight := 6.5
		textWidth := pdf.GetStringWidth(msg)
		numLines := 1
		if textWidth > colWidths[4] {
			numLines = int(math.Ceil(textWidth / colWidths[4]))
		}
		rowHeight := lineHeight * float64(numLines)
		if rowHeight < 7 {
			rowHeight = 7
		}

		pdf.CellFormat(colWidths[0], rowHeight, fmt.Sprintf("%d", idx+1), "1", 0, "C", true, 0, "")

		pdf.SetFont(fontFamily, "", 8)
		pdf.CellFormat(colWidths[1], rowHeight, step.Step, "1", 0, "L", true, 0, "")

		pdf.SetFont(fontFamily, "", 8)
		pdf.SetTextColor(r, g, b)
		statusLabel := stepStatusLabel(step.Status)
		pdf.CellFormat(colWidths[2], rowHeight, statusLabel, "1", 0, "C", true, 0, "")

		pdf.SetTextColor(60, 60, 60)
		pdf.SetFont(fontFamily, "", 8)
		pdf.CellFormat(colWidths[3], rowHeight, formatDurationShort(step.Duration), "1", 0, "C", true, 0, "")

		pdf.SetTextColor(60, 60, 60)
		pdf.SetFont(fontFamily, "", 8)
		pdf.MultiCell(colWidths[4], lineHeight, msg, "1", "L", true)

		pdf.Ln(-1)
	}
}

// drawFooterNote 绘制报告底部声明。
func drawFooterNote(pdf *fpdf.Fpdf, data *ReportData) {
	pdf.Ln(12)

	pdf.SetDrawColor(200, 200, 200)
	pdf.SetLineWidth(0.3)
	pdf.Line(pdfMarginLeft, pdf.GetY(), pdfPageWidth-pdfMarginRight, pdf.GetY())
	pdf.Ln(6)

	pdf.SetFont(fontFamily, "", 8)
	pdf.SetTextColor(130, 130, 130)
	pdf.MultiCell(pdfContentWidth, 5,
		fmt.Sprintf(
			"This report was automatically generated by AWSFinOps Worker (Run: %s). "+
				"Data reflects the state at execution time and may differ from real-time AWS billing. "+
				"For official billing data, please refer to the AWS Billing Console.",
			data.RunID,
		),
		"", "L", false)
}

// ExportAsJSON 生成格式化的 JSON 报告（面向开发者）。
//
// JSON 特性：
//   - 4 空格缩进，易于阅读和 diff 对比
//   - 包含完整的元数据、步骤详情和时间信息
//   - 时间字段使用 RFC3339 格式
//
// 参数：
//   - data     : 统一的报告数据
//   - filePath : 输出文件路径
//
// 返回：
//   - error : 序列化或写入失败时返回错误
func ExportAsJSON(data *ReportData, filePath string) error {
	start := time.Now()
	LogStart(documentComponent, "ExportAsJSON")

	type jsonStep struct {
		Index       int    `json:"index"`
		Step        string `json:"step"`
		Status      string `json:"status"`
		StatusLabel string `json:"status_label"`
		DurationMs  int64  `json:"duration_ms"`
		DurationStr string `json:"duration"`
		Message     string `json:"message,omitempty"`
		Error       string `json:"error,omitempty"`
	}

	type jsonReport struct {
		SchemaVersion string     `json:"schema_version"`
		App           string     `json:"app"`
		RunID         string     `json:"run_id"`
		AWSAccountID  string     `json:"aws_account_id,omitempty"`
		StartTime     string     `json:"start_time"`
		EndTime       string     `json:"end_time"`
		DurationMs    int64      `json:"duration_ms"`
		DurationStr   string     `json:"duration"`
		Status        string     `json:"status"`
		TotalSteps    int        `json:"total_steps"`
		StepCounts    stepCounts `json:"step_counts"`
		Steps         []jsonStep `json:"steps"`
		Summary       string     `json:"summary"`
		GeneratedAt   string     `json:"generated_at"`
	}

	ok, pending, warned, failed := CountStepStatuses(data.Steps)

	steps := make([]jsonStep, len(data.Steps))
	for i, s := range data.Steps {
		steps[i] = jsonStep{
			Index:       s.Index,
			Step:        s.Step,
			Status:      s.Status,
			StatusLabel: stepStatusLabel(s.Status),
			DurationMs:  s.Duration.Milliseconds(),
			DurationStr: formatDurationShort(s.Duration),
			Message:     s.Message,
			Error:       s.Error,
		}
	}

	report := jsonReport{
		SchemaVersion: "1.0",
		App:           APP_NAME,
		RunID:         data.RunID,
		AWSAccountID:  data.AWSAccountID,
		StartTime:     data.StartTime.Format(time.RFC3339),
		EndTime:       data.EndTime.Format(time.RFC3339),
		DurationMs:    data.Duration.Milliseconds(),
		DurationStr:   data.Duration.Round(time.Millisecond).String(),
		Status:        data.Status,
		TotalSteps:    len(data.Steps),
		StepCounts:    stepCounts{OK: ok, Pending: pending, Warn: warned, Fail: failed},
		Steps:         steps,
		Summary:       data.Summary,
		GeneratedAt:   time.Now().Format(time.RFC3339),
	}

	out, err := json.MarshalIndent(report, "", "    ")
	if err != nil {
		LogError(documentComponent, "ExportAsJSON", err, time.Since(start))
		return fmt.Errorf("JSON 序列化失败: %w", err)
	}

	if err := os.WriteFile(filePath, out, 0644); err != nil {
		LogError(documentComponent, "ExportAsJSON", err, time.Since(start))
		return fmt.Errorf("JSON 文件写入失败: %w", err)
	}

	LogSuccess(documentComponent, "ExportAsJSON", time.Since(start),
		fmt.Sprintf("path=%s", filePath),
		fmt.Sprintf("bytes=%d", len(out)),
	)
	return nil
}

// ExportAsCSV 生成标准 CSV 财务报告。
//
// CSV 特性：
//   - 首行为元数据头（run_id, account_id, status, start, end, duration_ms, summary）
//   - 第二行为元数据值
//   - 空行分隔
//   - 步骤数据表头 + 每行一步
//   - 兼容 Excel / Google Sheets / 财务系统直接导入
//
// 参数：
//   - data     : 统一的报告数据
//   - filePath : 输出文件路径
//
// 返回：
//   - error : 写入失败时返回错误
func ExportAsCSV(data *ReportData, filePath string) error {
	start := time.Now()
	LogStart(documentComponent, "ExportAsCSV")

	f, err := os.Create(filePath)
	if err != nil {
		LogError(documentComponent, "ExportAsCSV", err, time.Since(start))
		return fmt.Errorf("CSV 文件创建失败: %w", err)
	}
	defer f.Close()

	if _, err := f.WriteString("\xEF\xBB\xBF"); err != nil {
		LogError(documentComponent, "ExportAsCSV", err, time.Since(start))
		return fmt.Errorf("BOM 写入失败: %w", err)
	}

	w := csv.NewWriter(f)
	defer w.Flush()

	metaHeaders := []string{
		"report_run_id",
		"aws_account_id",
		"overall_status",
		"start_time",
		"end_time",
		"duration_ms",
		"total_steps",
		"steps_ok",
		"steps_pending",
		"steps_warn",
		"steps_fail",
		"summary",
	}
	if err := w.Write(metaHeaders); err != nil {
		LogError(documentComponent, "ExportAsCSV", err, time.Since(start))
		return fmt.Errorf("CSV 元数据头写入失败: %w", err)
	}

	ok, pending, warned, failed := CountStepStatuses(data.Steps)
	metaValues := []string{
		data.RunID,
		data.AWSAccountID,
		data.Status,
		data.StartTime.Format(time.RFC3339),
		data.EndTime.Format(time.RFC3339),
		fmt.Sprintf("%d", data.Duration.Milliseconds()),
		fmt.Sprintf("%d", len(data.Steps)),
		fmt.Sprintf("%d", ok),
		fmt.Sprintf("%d", pending),
		fmt.Sprintf("%d", warned),
		fmt.Sprintf("%d", failed),
		data.Summary,
	}
	if err := w.Write(metaValues); err != nil {
		LogError(documentComponent, "ExportAsCSV", err, time.Since(start))
		return fmt.Errorf("CSV 元数据值写入失败: %w", err)
	}

	if err := w.Write([]string{}); err != nil {
		LogError(documentComponent, "ExportAsCSV", err, time.Since(start))
		return fmt.Errorf("CSV 分隔行写入失败: %w", err)
	}

	stepHeaders := []string{
		"step_index",
		"step_name",
		"step_status",
		"step_status_label",
		"step_duration_ms",
		"step_duration",
		"step_message",
		"step_error",
	}
	if err := w.Write(stepHeaders); err != nil {
		LogError(documentComponent, "ExportAsCSV", err, time.Since(start))
		return fmt.Errorf("CSV 步骤表头写入失败: %w", err)
	}

	for _, step := range data.Steps {
		row := []string{
			fmt.Sprintf("%d", step.Index),
			step.Step,
			step.Status,
			stepStatusLabel(step.Status),
			fmt.Sprintf("%d", step.Duration.Milliseconds()),
			formatDurationShort(step.Duration),
			step.Message,
			step.Error,
		}
		if err := w.Write(row); err != nil {
			LogError(documentComponent, "ExportAsCSV", err, time.Since(start))
			return fmt.Errorf("CSV 步骤行写入失败: %w", err)
		}
	}

	w.Flush()
	if err := w.Error(); err != nil {
		LogError(documentComponent, "ExportAsCSV", err, time.Since(start))
		return fmt.Errorf("CSV flush 失败: %w", err)
	}

	LogSuccess(documentComponent, "ExportAsCSV", time.Since(start),
		fmt.Sprintf("path=%s", filePath),
		fmt.Sprintf("rows=%d", len(data.Steps)),
	)
	return nil
}

type stepCounts struct {
	OK      int `json:"ok"`
	Pending int `json:"pending"`
	Warn    int `json:"warn"`
	Fail    int `json:"fail"`
}

// CountStepStatuses 统计各状态的步骤数量。
func CountStepStatuses(steps []ReportStepData) (ok, pending, warned, failed int) {
	for _, s := range steps {
		switch s.Status {
		case "ok":
			ok++
		case "pending_implementation":
			pending++
		case "warn":
			warned++
		case "error":
			failed++
		}
	}
	return
}

// statusColor 返回状态对应的 RGB 颜色值（用于 PDF）。
func statusColor(status string) (int, int, int) {
	switch status {
	case "ok", "completed":
		return 0, 150, 0
	case "warn", "completed_with_warnings":
		return 230, 140, 0
	case "error", "partial_failure":
		return 200, 0, 0
	case "pending_implementation":
		return 140, 140, 140
	default:
		return 80, 80, 80
	}
}

// stepStatusLabel 返回状态的可读标签。
func stepStatusLabel(status string) string {
	switch status {
	case "ok":
		return "OK"
	case "warn":
		return "WARN"
	case "error":
		return "FAIL"
	case "pending_implementation":
		return "PENDING"
	default:
		return strings.ToUpper(status)
	}
}

// formatDurationShort 将 time.Duration 格式化为简短可读字符串。
func formatDurationShort(d time.Duration) string {
	if d == 0 {
		return "0s"
	}
	us := float64(d.Microseconds())
	if math.Abs(us) < 1000 {
		return fmt.Sprintf("%.0fus", us)
	}
	ms := float64(d.Milliseconds())
	if math.Abs(ms) < 1000 {
		return fmt.Sprintf("%.1fms", float64(d.Microseconds())/1000.0)
	}
	return fmt.Sprintf("%.2fs", d.Seconds())
}
