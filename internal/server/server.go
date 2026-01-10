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

const DefaultPort = "7777"

// openBrowserFunc is a variable for mocking openBrowser in tests
var openBrowserFunc = openBrowserDefault

func Serve(report *pr.Report, previousReport *pr.Report) error {
	port := os.Getenv("SHIRABERU_PORT")
	if port == "" {
		port = DefaultPort
	}
	return ServeWithAddr(report, previousReport, ":"+port)
}

func ServeWithAddr(report *pr.Report, previousReport *pr.Report, addr string) error {
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
		time.Sleep(500 * time.Millisecond)
		openBrowserFunc("http://localhost" + addr)
	}()

	fmt.Println("✓ Opening in browser: http://localhost" + addr)
	fmt.Println("Press Ctrl+C to stop the server")

	return server.ListenAndServe()
}

func openBrowserDefault(url string) {
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
		_ = cmd.Start()
	}
}
