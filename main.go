package main

import (
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/taikicoco/shiraberu/internal/config"
	"github.com/taikicoco/shiraberu/internal/demo"
	"github.com/taikicoco/shiraberu/internal/github"
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

	opts, err := prompt.Run(cfg)
	if err != nil {
		return err
	}

	client, err := github.NewClient()
	if err != nil {
		return fmt.Errorf("failed to create GitHub client: %w", err)
	}

	fetcher := pr.NewFetcher(client)

	// Fetch current period with spinner
	spin := spinner.New("Fetching PRs...")
	spin.Start()
	report, err := fetcher.Fetch(opts.Org, opts.StartDate, opts.EndDate)
	if err != nil {
		spin.Fail("Failed to fetch PRs")
		return fmt.Errorf("failed to fetch PRs: %w", err)
	}
	spin.Success("Fetched PRs")

	// Fetch previous period for comparison
	var previousReport *pr.Report
	spin = spinner.New("Fetching previous period...")
	spin.Start()
	prevStartDate, prevEndDate := calcPreviousPeriod(opts.StartDate, opts.EndDate, opts.PeriodType)
	previousReport, err = fetcher.Fetch(opts.Org, prevStartDate, prevEndDate)
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
		if opts.OutputPath == "" {
			return render.RenderHTML(os.Stdout, report, previousReport)
		}
		f, err := os.Create(opts.OutputPath)
		if err != nil {
			return err
		}
		defer f.Close()
		if err := render.RenderHTML(f, report, previousReport); err != nil {
			return err
		}
		fmt.Printf("✓ Saved to %s\n", opts.OutputPath)

	default: // markdown
		if opts.OutputPath == "" {
			return render.RenderMarkdown(os.Stdout, report)
		}
		f, err := os.Create(opts.OutputPath)
		if err != nil {
			return err
		}
		defer f.Close()
		if err := render.RenderMarkdown(f, report); err != nil {
			return err
		}
		fmt.Printf("✓ Saved to %s\n", opts.OutputPath)
	}

	return nil
}

func calcPreviousPeriod(startDate, endDate time.Time, periodType prompt.PeriodType) (time.Time, time.Time) {
	switch periodType {
	case prompt.PeriodTypeWeek:
		// Previous week (Monday to Sunday)
		prevEndDate := startDate.AddDate(0, 0, -1)
		prevStartDate := prevEndDate.AddDate(0, 0, -6)
		return prevStartDate, prevEndDate

	case prompt.PeriodTypeMonth:
		// Previous month (1st to last day)
		prevEndDate := startDate.AddDate(0, 0, -1)
		prevStartDate := time.Date(prevEndDate.Year(), prevEndDate.Month(), 1, 0, 0, 0, 0, prevEndDate.Location())
		return prevStartDate, prevEndDate

	default: // PeriodTypeCustom
		// Same duration before
		duration := endDate.Sub(startDate) + 24*time.Hour
		prevEndDate := startDate.AddDate(0, 0, -1)
		prevStartDate := prevEndDate.Add(-duration + 24*time.Hour)
		return prevStartDate, prevEndDate
	}
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
