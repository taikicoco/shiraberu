package main

import (
	"fmt"
	"os"

	"github.com/taikicoco/shiraberu/internal/config"
	"github.com/taikicoco/shiraberu/internal/github"
	"github.com/taikicoco/shiraberu/internal/html"
	"github.com/taikicoco/shiraberu/internal/pr"
	"github.com/taikicoco/shiraberu/internal/prompt"
	"github.com/taikicoco/shiraberu/internal/server"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func run() error {
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

	switch opts.Format {
	case "browser":
		return server.Serve(report)

	case "html":
		if opts.OutputPath == "" {
			return html.RenderHTML(os.Stdout, report)
		}
		f, err := os.Create(opts.OutputPath)
		if err != nil {
			return err
		}
		defer f.Close()
		if err := html.RenderHTML(f, report); err != nil {
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
