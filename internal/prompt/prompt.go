package prompt

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/manifoldco/promptui"
	"github.com/taikicoco/shiraberu/internal/config"
	apperrors "github.com/taikicoco/shiraberu/internal/errors"
	"github.com/taikicoco/shiraberu/internal/period"
)

const (
	backOption   = "← Back"
	monthsToShow = 12
)

// IO はユーザー入力を抽象化するインターフェース
type IO interface {
	// ReadLine はラベルとデフォルト値を表示し、ユーザーの入力を返す
	ReadLine(label string, defaultVal string) (string, error)
	// Select は選択肢を表示し、選択されたインデックスを返す
	Select(label string, options []string, defaultIdx int) (int, error)
}

// DefaultIO は標準入出力を使用するIO実装
type DefaultIO struct {
	reader *bufio.Reader
}

// NewDefaultIO は標準入力を使用するDefaultIOを作成する
func NewDefaultIO() *DefaultIO {
	return &DefaultIO{reader: bufio.NewReader(os.Stdin)}
}

// ReadLine はラベルとデフォルト値を表示し、ユーザーの入力を返す
func (d *DefaultIO) ReadLine(label string, defaultVal string) (string, error) {
	if defaultVal != "" {
		fmt.Printf("? %s [%s]: ", label, defaultVal)
	} else {
		fmt.Printf("? %s: ", label)
	}

	input, err := d.reader.ReadString('\n')
	if err != nil && err != io.EOF {
		return defaultVal, err
	}
	input = strings.TrimSpace(input)

	if input == "" {
		return defaultVal, nil
	}
	return input, nil
}

// Select は選択肢を表示し、選択されたインデックスを返す
func (d *DefaultIO) Select(label string, options []string, defaultIdx int) (int, error) {
	prompt := promptui.Select{
		Label:     label,
		Items:     options,
		CursorPos: defaultIdx,
		Templates: &promptui.SelectTemplates{
			Label:    "? {{ . }}",
			Active:   "\U0001F449 {{ . | cyan }}",
			Inactive: "   {{ . }}",
			Selected: "\U00002705 {{ . | green }}",
		},
	}

	idx, _, err := prompt.Run()
	if err != nil {
		return defaultIdx, nil
	}
	return idx, nil
}

// Runner はプロンプトの実行を管理する
type Runner struct {
	io IO
}

// NewRunner は指定されたIOを使用するRunnerを作成する
func NewRunner(io IO) *Runner {
	return &Runner{io: io}
}

type Options struct {
	Org        string
	Username   string
	StartDate  time.Time
	EndDate    time.Time
	PeriodType period.Type
	Format     string
	OutputPath string
}

type step int

const (
	stepOrg step = iota
	stepUsername
	stepPeriodMode
	stepPeriodDetail
	stepFormat
	stepDone
)

// Run は標準入出力を使用してプロンプトを実行する（後方互換性のため）
func Run(cfg *config.Config, defaultUsername string) (*Options, error) {
	r := NewRunner(NewDefaultIO())
	return r.Run(cfg, defaultUsername)
}

// Run はインタラクティブプロンプトを実行してオプションを収集する
func (r *Runner) Run(cfg *config.Config, defaultUsername string) (*Options, error) {
	opts := &Options{}

	currentStep := stepOrg

	for currentStep != stepDone {
		switch currentStep {
		case stepOrg:
			opts.Org = r.promptText("Organization", cfg.Org)
			if opts.Org == "" {
				return nil, apperrors.ErrOrgRequired
			}
			currentStep = stepUsername

		case stepUsername:
			opts.Username = r.promptText("GitHub username", defaultUsername)
			if opts.Username == "" {
				opts.Username = defaultUsername
			}
			currentStep = stepPeriodMode

		case stepPeriodMode:
			modes := []string{"Single day", "Date range", backOption}
			idx := r.promptSelect("Period type", modes, 0)
			if idx == 2 { // Back
				currentStep = stepUsername
				continue
			}
			var goBack bool
			if idx == 0 {
				opts.StartDate, opts.EndDate, goBack = r.promptSingleDay()
				opts.PeriodType = period.TypeCustom
			} else {
				opts.StartDate, opts.EndDate, opts.PeriodType, goBack = r.promptDateRange()
			}
			if goBack {
				continue // Stay at stepPeriodMode
			}
			currentStep = stepFormat

		case stepFormat:
			formats := []string{"HTML (open in browser)", "HTML", "Markdown", backOption}
			formatValues := []string{"browser", "html", "markdown"}
			defaultIdx := 0
			for i, v := range formatValues {
				if v == cfg.Format {
					defaultIdx = i
					break
				}
			}
			idx := r.promptSelect("Output format", formats, defaultIdx)
			if idx == 3 { // Back
				currentStep = stepPeriodMode
				continue
			}
			opts.Format = formatValues[idx]
			currentStep = stepDone
		}
	}

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
		return start.Format("20060102") + ext
	}
	return start.Format("20060102") + "-" + end.Format("20060102") + ext
}

func (r *Runner) promptSingleDay() (time.Time, time.Time, bool) {
	now := time.Now()
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	yesterday := today.AddDate(0, 0, -1)

	options := []string{
		fmt.Sprintf("Today (%s)", today.Format("2006-01-02")),
		fmt.Sprintf("Yesterday (%s)", yesterday.Format("2006-01-02")),
		"Enter date",
		backOption,
	}
	idx := r.promptSelect("Select date", options, 0)

	if idx == 3 { // Back
		return time.Time{}, time.Time{}, true
	}

	var date time.Time
	switch idx {
	case 0:
		date = today
	case 1:
		date = yesterday
	case 2:
		date = r.promptDate("Date (YYYY-MM-DD)", today)
	}

	return date, date, false
}

func (r *Runner) promptDateRange() (time.Time, time.Time, period.Type, bool) {
	now := time.Now()
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	weekdays := []string{"Sun", "Mon", "Tue", "Wed", "Thu", "Fri", "Sat"}

	// Calculate this Monday
	weekday := int(today.Weekday())
	if weekday == 0 {
		weekday = 7
	}
	thisMonday := today.AddDate(0, 0, -(weekday - 1))
	lastMonday := thisMonday.AddDate(0, 0, -7)
	lastSunday := thisMonday.AddDate(0, 0, -1)

	// First day of this month
	thisMonthStart := time.Date(today.Year(), today.Month(), 1, 0, 0, 0, 0, today.Location())
	// Last month
	lastMonthStart := thisMonthStart.AddDate(0, -1, 0)
	lastMonthEnd := thisMonthStart.AddDate(0, 0, -1)

	options := []string{
		fmt.Sprintf("This week (%s %s - %s %s)", thisMonday.Format("1/2"), weekdays[thisMonday.Weekday()], today.Format("1/2"), weekdays[today.Weekday()]),
		fmt.Sprintf("Last week (%s %s - %s %s)", lastMonday.Format("1/2"), weekdays[lastMonday.Weekday()], lastSunday.Format("1/2"), weekdays[lastSunday.Weekday()]),
		fmt.Sprintf("This month (%s - %s)", thisMonthStart.Format("1/2"), today.Format("1/2")),
		fmt.Sprintf("Last month (%s - %s)", lastMonthStart.Format("1/2"), lastMonthEnd.Format("1/2")),
		"Select month",
		"Enter dates",
		backOption,
	}
	idx := r.promptSelect("Select range", options, 0)

	if idx == 6 { // Back
		return time.Time{}, time.Time{}, "", true
	}

	var start, end time.Time
	var periodType period.Type
	switch idx {
	case 0: // This week
		start, end = thisMonday, today
		periodType = period.TypeWeek
	case 1: // Last week
		start, end = lastMonday, lastSunday
		periodType = period.TypeWeek
	case 2: // This month
		start, end = thisMonthStart, today
		periodType = period.TypeMonth
	case 3: // Last month
		start, end = lastMonthStart, lastMonthEnd
		periodType = period.TypeMonth
	case 4: // Select month
		start, end = r.promptSelectMonth(today)
		periodType = period.TypeMonth
	case 5: // Enter dates
		start = r.promptDate("Start date (YYYY-MM-DD)", today.AddDate(0, 0, -7))
		end = r.promptDate("End date (YYYY-MM-DD)", today)
		periodType = period.TypeCustom
	}

	// Confirm
	start, end = r.confirmDateRange(start, end)
	return start, end, periodType, false
}

func (r *Runner) promptSelectMonth(today time.Time) (time.Time, time.Time) {
	// Generate past months
	options := make([]string, monthsToShow)
	months := make([]time.Time, monthsToShow)

	for i := 0; i < monthsToShow; i++ {
		monthStart := time.Date(today.Year(), today.Month(), 1, 0, 0, 0, 0, today.Location()).AddDate(0, -i, 0)
		months[i] = monthStart
		options[i] = monthStart.Format("2006-01")
	}

	idx := r.promptSelect("Select month", options, 0)

	start := months[idx]
	end := start.AddDate(0, 1, -1) // Last day of the month

	return start, end
}

func (r *Runner) confirmDateRange(start, end time.Time) (time.Time, time.Time) {
	label := fmt.Sprintf("Period: %s - %s. Confirm? [Enter: OK / s: change start / e: change end]",
		start.Format("2006-01-02"), end.Format("2006-01-02"))

	input, err := r.io.ReadLine(label, "")
	if err != nil {
		return start, end
	}
	input = strings.TrimSpace(strings.ToLower(input))

	switch input {
	case "s":
		start = r.promptDate("Start date (YYYY-MM-DD)", start)
		return r.confirmDateRange(start, end)
	case "e":
		end = r.promptDate("End date (YYYY-MM-DD)", end)
		return r.confirmDateRange(start, end)
	default:
		return start, end
	}
}

func (r *Runner) promptDate(label string, defaultDate time.Time) time.Time {
	defaultStr := defaultDate.Format("2006-01-02")
	input := r.promptText(label, defaultStr)

	parsed, err := time.Parse("2006-01-02", input)
	if err != nil {
		return defaultDate
	}
	return parsed
}

func (r *Runner) promptText(label string, defaultVal string) string {
	input, err := r.io.ReadLine(label, defaultVal)
	if err != nil {
		return defaultVal
	}
	return input
}

func (r *Runner) promptSelect(label string, options []string, defaultIdx int) int {
	idx, err := r.io.Select(label, options, defaultIdx)
	if err != nil {
		return defaultIdx
	}
	return idx
}
