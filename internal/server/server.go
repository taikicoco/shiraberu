package server

import (
	"bytes"
	"fmt"
	"net/http"
	"os/exec"
	"runtime"
	"time"

	"github.com/taikicoco/shiraberu/internal/html"
	"github.com/taikicoco/shiraberu/internal/pr"
)

func Serve(report *pr.Report) error {
	var buf bytes.Buffer
	if err := html.RenderHTML(&buf, report); err != nil {
		return err
	}
	content := buf.Bytes()

	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		_, _ = w.Write(content)
	})

	server := &http.Server{
		Addr:    ":7777",
		Handler: mux,
	}

	go func() {
		time.Sleep(500 * time.Millisecond)
		openBrowser("http://localhost:7777")
	}()

	fmt.Println("✓ ブラウザで開きます: http://localhost:7777")
	fmt.Println("サーバーを停止するには Ctrl+C を押してください")

	return server.ListenAndServe()
}

func openBrowser(url string) {
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
