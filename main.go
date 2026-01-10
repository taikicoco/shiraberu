package main

import (
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/taikicoco/shiraberu/internal/config"
	"github.com/taikicoco/shiraberu/internal/demo"
	"github.com/taikicoco/shiraberu/internal/github"
	"github.com/taikicoco/shiraberu/internal/html"
	"github.com/taikicoco/shiraberu/internal/pr"
	"github.com/taikicoco/shiraberu/internal/prompt"
	"github.com/taikicoco/shiraberu/internal/server"
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

	fmt.Println("\nレポートを生成しています...")

	client, err := github.NewClient()
	if err != nil {
		return fmt.Errorf("failed to create GitHub client: %w", err)
	}

	fetcher := pr.NewFetcher(client)
	report, err := fetcher.Fetch(opts.Org, opts.StartDate, opts.EndDate)
	if err != nil {
		return fmt.Errorf("failed to fetch PRs: %w", err)
	}

	// 前期間のデータを取得（同じ期間長さで、直前の期間）
	duration := opts.EndDate.Sub(opts.StartDate)
	prevEndDate := opts.StartDate.AddDate(0, 0, -1)
	prevStartDate := prevEndDate.Add(-duration)
	previousReport, err := fetcher.Fetch(opts.Org, prevStartDate, prevEndDate)
	if err != nil {
		// 前期間の取得に失敗しても続行（差分表示なしで）
		previousReport = nil
	}

	switch opts.Format {
	case "browser":
		return server.Serve(report, previousReport)

	case "html":
		if opts.OutputPath == "" {
			return html.RenderHTML(os.Stdout, report, previousReport)
		}
		f, err := os.Create(opts.OutputPath)
		if err != nil {
			return err
		}
		defer f.Close()
		if err := html.RenderHTML(f, report, previousReport); err != nil {
			return err
		}
		fmt.Printf("✓ %s に出力しました\n", opts.OutputPath)

	default: // markdown
		if opts.OutputPath == "" {
			return html.RenderMarkdown(os.Stdout, report)
		}
		f, err := os.Create(opts.OutputPath)
		if err != nil {
			return err
		}
		defer f.Close()
		if err := html.RenderMarkdown(f, report); err != nil {
			return err
		}
		fmt.Printf("✓ %s に出力しました\n", opts.OutputPath)
	}

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
