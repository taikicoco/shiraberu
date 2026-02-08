package server

import (
	"bytes"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"runtime"
	"time"

	"github.com/taikicoco/shiraberu/internal/pr"
	"github.com/taikicoco/shiraberu/internal/render"
)

const (
	DefaultPort      = "7777"
	browserOpenDelay = 500 * time.Millisecond
)

// BrowserOpener はブラウザを開く機能を抽象化するインターフェース
type BrowserOpener interface {
	Open(url string) error
}

// DefaultBrowserOpener はOSのデフォルトブラウザを開く実装
type DefaultBrowserOpener struct{}

// Open はURLをデフォルトブラウザで開く
func (o *DefaultBrowserOpener) Open(url string) error {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "darwin":
		cmd = exec.Command("open", url)
	case "linux":
		cmd = exec.Command("xdg-open", url)
	case "windows":
		cmd = exec.Command("rundll32", "url.dll,FileProtocolHandler", url)
	}
	if cmd != nil {
		return cmd.Start()
	}
	return nil
}

// Server はHTTPサーバーの設定を保持する
type Server struct {
	browserOpener BrowserOpener
}

// ServerOption はServerの設定オプション
type ServerOption func(*Server)

// WithBrowserOpener はBrowserOpenerを設定するオプション
func WithBrowserOpener(opener BrowserOpener) ServerOption {
	return func(s *Server) {
		s.browserOpener = opener
	}
}

// NewServer は新しいServerインスタンスを作成する
func NewServer(opts ...ServerOption) *Server {
	s := &Server{
		browserOpener: &DefaultBrowserOpener{},
	}
	for _, opt := range opts {
		opt(s)
	}
	return s
}

// Serve はデフォルト設定でサーバーを起動する（後方互換性のため）
func Serve(report *pr.Report, previousReport *pr.Report) error {
	return NewServer().ServeReport(report, previousReport)
}

// ServeReport はレポートをHTTPサーバーで提供する
func (s *Server) ServeReport(report *pr.Report, previousReport *pr.Report) error {
	port := os.Getenv("SHIRABERU_PORT")
	if port == "" {
		port = DefaultPort
	}
	return s.ServeWithAddr(report, previousReport, ":"+port)
}

// ServeWithAddr は指定アドレスでサーバーを起動する（後方互換性のため）
func ServeWithAddr(report *pr.Report, previousReport *pr.Report, addr string) error {
	return NewServer().ServeWithAddr(report, previousReport, addr)
}

// ServeWithAddr は指定アドレスでサーバーを起動する
func (s *Server) ServeWithAddr(report *pr.Report, previousReport *pr.Report, addr string) error {
	var buf bytes.Buffer
	if err := render.RenderHTML(&buf, report, previousReport); err != nil {
		return err
	}
	content := buf.Bytes()

	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		_, _ = w.Write(content)
	})

	server := &http.Server{
		Addr:    addr,
		Handler: mux,
	}

	go func() {
		time.Sleep(browserOpenDelay)
		_ = s.browserOpener.Open("http://localhost" + addr)
	}()

	fmt.Println("✓ Opening in browser: http://localhost" + addr)
	fmt.Println("Press Ctrl+C to stop the server")

	return server.ListenAndServe()
}
