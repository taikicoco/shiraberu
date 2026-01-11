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
package main

import (
	"flag"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/taikicoco/shiraberu/internal/config"
	"github.com/taikicoco/shiraberu/internal/demo"
	"github.com/taikicoco/shiraberu/internal/github"
	"github.com/taikicoco/shiraberu/internal/period"
	"github.com/taikicoco/shiraberu/internal/pr"
	"github.com/taikicoco/shiraberu/internal/prompt"
	"github.com/taikicoco/shiraberu/internal/render"
	"github.com/taikicoco/shiraberu/internal/server"
	"github.com/taikicoco/shiraberu/internal/spinner"
)

var demoMode = flag.Bool("demo", false, "Run with demo data (no GitHub API calls)")

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

func runDemo() error {
	fmt.Println("Demo mode: Generating sample data...")

	// Generate demo data for last 30 days
	endDate := time.Now()
	startDate := endDate.AddDate(0, 0, -30)

	report, previousReport := demo.GenerateReport(startDate, endDate)

	fmt.Printf("Generated %d days of demo data\n", len(report.Days))

	return server.Serve(report, previousReport)
}
