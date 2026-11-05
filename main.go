// Shiraberu is a CLI tool that generates pull request activity reports
// from GitHub. It fetches merged PRs for a specified user and organization,
// then renders the results as HTML or Markdown reports.
//
// Usage:
//
//	shiraberu [flags]
//
// Flags:
//
//	-demo    Run with demo data (no GitHub API calls)
//	-since   Start date (YYYY-MM-DD, JST). Runs without prompts and writes JSON
//	-until   End date (YYYY-MM-DD, JST). Defaults to -since
//	-org     GitHub organization. Defaults to SHIRABERU_ORG
//	-user    GitHub username. Defaults to the authenticated user
//	-format  Output format without prompts. Only json is supported
//	-output  Output file path. Defaults to stdout
package main

import (
	"flag"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/taikicoco/shiraberu/internal/config"
	"github.com/taikicoco/shiraberu/internal/demo"
	apperrors "github.com/taikicoco/shiraberu/internal/errors"
	"github.com/taikicoco/shiraberu/internal/github"
	"github.com/taikicoco/shiraberu/internal/period"
	"github.com/taikicoco/shiraberu/internal/pr"
	"github.com/taikicoco/shiraberu/internal/prompt"
	"github.com/taikicoco/shiraberu/internal/render"
	"github.com/taikicoco/shiraberu/internal/server"
	"github.com/taikicoco/shiraberu/internal/spinner"
	"github.com/taikicoco/shiraberu/internal/timezone"
)

const dateLayout = "2006-01-02"

var (
	demoMode   = flag.Bool("demo", false, "Run with demo data (no GitHub API calls)")
	sinceFlag  = flag.String("since", "", "Start date (YYYY-MM-DD, JST). Runs without prompts and writes JSON")
	untilFlag  = flag.String("until", "", "End date (YYYY-MM-DD, JST). Defaults to -since")
	orgFlag    = flag.String("org", "", "GitHub organization. Defaults to SHIRABERU_ORG")
	userFlag   = flag.String("user", "", "GitHub username. Defaults to the authenticated user")
	formatFlag = flag.String("format", "json", "Output format without prompts. Only json is supported")
	outputFlag = flag.String("output", "", "Output file path. Defaults to stdout")
)

func main() {
	flag.Parse()

	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func run() error {
	// Demo mode
	if *demoMode {
		return runDemo()
	}

	if flag.NFlag() > 0 {
		return runWithoutPrompts()
	}

	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	client, err := github.NewClient()
	if err != nil {
		return fmt.Errorf("failed to create GitHub client: %w", err)
	}

	defaultUsername := client.Username()
	opts, err := prompt.Run(cfg, defaultUsername)
	if err != nil {
		return err
	}

	fetcher := pr.NewFetcher(client)

	// Fetch current period with spinner
	spin := spinner.New("Fetching PRs...")
	spin.Start()
	report, err := fetcher.Fetch(opts.Org, opts.Username, opts.StartDate, opts.EndDate)
	if err != nil {
		spin.Fail("Failed to fetch PRs")
		return fmt.Errorf("failed to fetch PRs: %w", err)
	}
	spin.Success("Fetched PRs")

	// Fetch previous period for comparison
	var previousReport *pr.Report
	spin = spinner.New("Fetching previous period...")
	spin.Start()
	prevStartDate, prevEndDate := period.CalcPrevious(opts.StartDate, opts.EndDate, opts.PeriodType)
	previousReport, err = fetcher.Fetch(opts.Org, opts.Username, prevStartDate, prevEndDate)
	if err != nil {
		spin.Fail("Previous period unavailable")
		previousReport = nil
	} else {
		spin.Success("Fetched previous period")
	}

	switch opts.Format {
	case "browser":
		return server.Serve(report, previousReport)
	case "html":
		return writeOutput(opts.OutputPath, func(w io.Writer) error {
			return render.RenderHTML(w, report, previousReport)
		})
	default: // markdown
		return writeOutput(opts.OutputPath, func(w io.Writer) error {
			return render.RenderMarkdown(w, report)
		})
	}
}

// writeOutput はレンダリング結果をファイルまたは標準出力に書き込む
func writeOutput(path string, renderer func(io.Writer) error) error {
	if path == "" {
		return renderer(os.Stdout)
	}

	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()

	if err := renderer(f); err != nil {
		return err
	}

	fmt.Printf("✓ Saved to %s\n", path)
	return nil
}

// batchOptions は対話なしで実行するときのオプション
type batchOptions struct {
	Org       string
	StartDate time.Time
	EndDate   time.Time
	Output    string
}

// parseBatchOptions はフラグの値を検証して batchOptions を作る
func parseBatchOptions(since, until, org, format, output string) (*batchOptions, error) {
	if since == "" {
		return nil, apperrors.ErrSinceRequired
	}
	if org == "" {
		return nil, apperrors.ErrOrgRequired
	}
	if format != "json" {
		return nil, fmt.Errorf("%w: %s", apperrors.ErrUnsupportedFormat, format)
	}
	if until == "" {
		until = since
	}

	start, err := time.ParseInLocation(dateLayout, since, timezone.JST)
	if err != nil {
		return nil, fmt.Errorf("%w: -since %s", apperrors.ErrInvalidDate, since)
	}
	end, err := time.ParseInLocation(dateLayout, until, timezone.JST)
	if err != nil {
		return nil, fmt.Errorf("%w: -until %s", apperrors.ErrInvalidDate, until)
	}
	if start.After(end) {
		return nil, apperrors.ErrInvalidDateRange
	}

	return &batchOptions{Org: org, StartDate: start, EndDate: end, Output: output}, nil
}

// runWithoutPrompts は対話なしで活動を取得し、JSON を書き出す。
// 標準出力は JSON だけにするため、スピナーは使わない
func runWithoutPrompts() error {
	org := *orgFlag
	if org == "" {
		org = config.LoadEnv().Org
	}

	opts, err := parseBatchOptions(*sinceFlag, *untilFlag, org, *formatFlag, *outputFlag)
	if err != nil {
		return err
	}

	client, err := github.NewClient()
	if err != nil {
		return fmt.Errorf("failed to create GitHub client: %w", err)
	}

	username := *userFlag
	if username == "" {
		username = client.Username()
	}

	report, err := pr.NewActivityFetcher(client).Fetch(opts.Org, username, opts.StartDate, opts.EndDate)
	if err != nil {
		return fmt.Errorf("failed to fetch activities: %w", err)
	}

	return writeOutput(opts.Output, func(w io.Writer) error {
		return render.RenderJSON(w, report)
	})
}

func runDemo() error {
	fmt.Println("Demo mode: Generating sample data...")

	// Generate demo data for last 30 days
	endDate := time.Now()
	startDate := endDate.AddDate(0, 0, -30)

	report, previousReport := demo.GenerateReport(startDate, endDate)

	fmt.Printf("Generated %d days of demo data\n", len(report.Days))

	return server.Serve(report, previousReport)
}
