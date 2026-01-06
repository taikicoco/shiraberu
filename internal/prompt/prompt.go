package prompt

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/taikicoco/shiraberu/internal/config"
)

type Options struct {
	Org        string
	StartDate  time.Time
	EndDate    time.Time
	Format     string
	OutputPath string
}

func Run(cfg *config.Config) (*Options, error) {
	reader := bufio.NewReader(os.Stdin)
	opts := &Options{}

	// Organization
	opts.Org = promptText(reader, "Organization を入力してください", cfg.Org)
	if opts.Org == "" {
		return nil, fmt.Errorf("organization is required")
	}

	// Period selection
	startDate, endDate, err := promptPeriod(reader)
	if err != nil {
		return nil, err
	}
	opts.StartDate = startDate
	opts.EndDate = endDate

	// Format
	formats := []string{"Markdown", "HTML", "HTML (ブラウザで表示)"}
	formatValues := []string{"markdown", "html", "browser"}
	defaultIdx := 0
	for i, v := range formatValues {
		if v == cfg.Format {
			defaultIdx = i
			break
		}
	}
	idx := promptSelect(reader, "出力形式を選択してください", formats, defaultIdx)
	opts.Format = formatValues[idx]

	// Output path (auto-generate if output_dir is set)
	if opts.Format != "browser" && cfg.OutputDir != "" {
		ext := ".md"
		if opts.Format == "html" {
			ext = ".html"
		}
		filename := generateFilename(opts.StartDate, opts.EndDate, ext)
		opts.OutputPath = filepath.Join(cfg.OutputDir, filename)
	}

	return opts, nil
}

func generateFilename(start, end time.Time, ext string) string {
	if start.Equal(end) {
		return start.Format("2006-01-02") + ext
	}
	return start.Format("2006-01-02") + "_" + end.Format("2006-01-02") + ext
}

func promptPeriod(reader *bufio.Reader) (time.Time, time.Time, error) {
	// 1日 or 範囲
	modes := []string{"1日だけ", "範囲で指定"}
	modeIdx := promptSelect(reader, "期間の指定方法", modes, 0)

	now := time.Now()
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())

	if modeIdx == 0 {
		// 1日だけ
		return promptSingleDay(reader, today)
	}

	// 範囲で指定
	return promptDateRange(reader, today)
}

func promptSingleDay(reader *bufio.Reader, today time.Time) (time.Time, time.Time, error) {
	yesterday := today.AddDate(0, 0, -1)

	options := []string{
		fmt.Sprintf("今日 (%s)", today.Format("2006-01-02")),
		fmt.Sprintf("昨日 (%s)", yesterday.Format("2006-01-02")),
		"日付を入力",
	}
	idx := promptSelect(reader, "日付を選択", options, 0)

	var date time.Time
	switch idx {
	case 0:
		date = today
	case 1:
		date = yesterday
	case 2:
		date = promptDate(reader, "日付 (YYYY-MM-DD)", today)
	}

	return date, date, nil
}

func promptDateRange(reader *bufio.Reader, today time.Time) (time.Time, time.Time, error) {
	weekdays := []string{"日", "月", "火", "水", "木", "金", "土"}

	// 今週の月曜日を計算
	weekday := int(today.Weekday())
	if weekday == 0 {
		weekday = 7
	}
	thisMonday := today.AddDate(0, 0, -(weekday - 1))
	lastMonday := thisMonday.AddDate(0, 0, -7)
	lastSunday := thisMonday.AddDate(0, 0, -1)

	// 今月の1日
	thisMonthStart := time.Date(today.Year(), today.Month(), 1, 0, 0, 0, 0, today.Location())
	// 先月
	lastMonthStart := thisMonthStart.AddDate(0, -1, 0)
	lastMonthEnd := thisMonthStart.AddDate(0, 0, -1)

	options := []string{
		fmt.Sprintf("今週 (%s %s 〜 %s %s)", thisMonday.Format("1/2"), weekdays[thisMonday.Weekday()], today.Format("1/2"), weekdays[today.Weekday()]),
		fmt.Sprintf("先週 (%s %s 〜 %s %s)", lastMonday.Format("1/2"), weekdays[lastMonday.Weekday()], lastSunday.Format("1/2"), weekdays[lastSunday.Weekday()]),
		fmt.Sprintf("今月 (%s 〜 %s)", thisMonthStart.Format("1/2"), today.Format("1/2")),
		fmt.Sprintf("先月 (%s 〜 %s)", lastMonthStart.Format("1/2"), lastMonthEnd.Format("1/2")),
		"過去N日",
		"日付を入力",
	}
	idx := promptSelect(reader, "範囲を選択", options, 0)

	var start, end time.Time
	switch idx {
	case 0: // 今週
		start, end = thisMonday, today
	case 1: // 先週
		start, end = lastMonday, lastSunday
	case 2: // 今月
		start, end = thisMonthStart, today
	case 3: // 先月
		start, end = lastMonthStart, lastMonthEnd
	case 4: // 過去N日
		n := promptNumber(reader, "何日前から？", 7)
		start = today.AddDate(0, 0, -n+1)
		end = today
	case 5: // 日付を入力
		start = promptDate(reader, "開始日 (YYYY-MM-DD)", today.AddDate(0, 0, -7))
		end = promptDate(reader, "終了日 (YYYY-MM-DD)", today)
	}

	// 確認・修正
	return confirmDateRange(reader, start, end)
}

func confirmDateRange(reader *bufio.Reader, start, end time.Time) (time.Time, time.Time, error) {
	fmt.Printf("\n期間: %s 〜 %s\n", start.Format("2006-01-02"), end.Format("2006-01-02"))
	fmt.Print("? この期間でよいですか？ [Enter: OK / s: 開始日修正 / e: 終了日修正]: ")

	input, _ := reader.ReadString('\n')
	input = strings.TrimSpace(strings.ToLower(input))

	switch input {
	case "s":
		start = promptDate(reader, "開始日 (YYYY-MM-DD)", start)
		return confirmDateRange(reader, start, end)
	case "e":
		end = promptDate(reader, "終了日 (YYYY-MM-DD)", end)
		return confirmDateRange(reader, start, end)
	default:
		return start, end, nil
	}
}

func promptDate(reader *bufio.Reader, label string, defaultDate time.Time) time.Time {
	defaultStr := defaultDate.Format("2006-01-02")
	input := promptText(reader, label, defaultStr)

	parsed, err := time.Parse("2006-01-02", input)
	if err != nil {
		fmt.Println("  無効な日付形式です。デフォルトを使用します。")
		return defaultDate
	}
	return parsed
}

func promptNumber(reader *bufio.Reader, label string, defaultVal int) int {
	input := promptText(reader, label, strconv.Itoa(defaultVal))
	n, err := strconv.Atoi(input)
	if err != nil || n < 1 {
		return defaultVal
	}
	return n
}

func promptText(reader *bufio.Reader, label string, defaultVal string) string {
	if defaultVal != "" {
		fmt.Printf("? %s [%s]: ", label, defaultVal)
	} else {
		fmt.Printf("? %s: ", label)
	}

	input, _ := reader.ReadString('\n')
	input = strings.TrimSpace(input)

	if input == "" {
		return defaultVal
	}
	return input
}

func promptSelect(reader *bufio.Reader, label string, options []string, defaultIdx int) int {
	fmt.Printf("? %s:\n", label)
	for i, opt := range options {
		marker := "  "
		if i == defaultIdx {
			marker = "> "
		}
		fmt.Printf("  %s%d. %s\n", marker, i+1, opt)
	}
	fmt.Printf("番号を入力 [%d]: ", defaultIdx+1)

	input, _ := reader.ReadString('\n')
	input = strings.TrimSpace(input)

	if input == "" {
		return defaultIdx
	}

	idx, err := strconv.Atoi(input)
	if err != nil || idx < 1 || idx > len(options) {
		return defaultIdx
	}
	return idx - 1
}
