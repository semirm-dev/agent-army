package cli

import (
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"runtime"
	"syscall"
	"time"

	"github.com/spf13/cobra"
)

// WebDir is set at build time via ldflags to the absolute path of army/web/be.
var WebDir string

func newServeCmd() *cobra.Command {
	var port int
	var noOpen bool

	cmd := &cobra.Command{
		Use:   "serve",
		Short: "Start the web management UI",
		RunE: func(cmd *cobra.Command, args []string) error {
			// 1. Find the web backend directory
			webDir := resolveWebDir()
			if webDir == "" {
				return fmt.Errorf("web UI not found. Set ARMY_WEB_DIR or run from the project root after 'make web-build'")
			}

			mainJS := filepath.Join(webDir, "dist", "main.js")

			// 2. Set up environment
			exe, _ := os.Executable()
			cwd, _ := os.Getwd()
			env := append(os.Environ(),
				fmt.Sprintf("ARMY_BIN=%s", exe),
				fmt.Sprintf("ARMY_CWD=%s", cwd),
				fmt.Sprintf("PORT=%d", port),
				"NODE_ENV=production",
			)

			// 3. Spawn node process
			nodeCmd := exec.Command("node", mainJS)
			nodeCmd.Env = env
			nodeCmd.Dir = webDir
			nodeCmd.Stdout = os.Stdout
			nodeCmd.Stderr = os.Stderr

			if err := nodeCmd.Start(); err != nil {
				return fmt.Errorf("starting web server: %w", err)
			}

			// 4. Wait for server to be ready (poll /api/health)
			url := fmt.Sprintf("http://localhost:%d", port)
			if err := waitForServer(url, 10*time.Second); err != nil {
				fmt.Fprintf(os.Stderr, "Warning: %v\n", err)
			}

			// 5. Open browser (unless --no-open)
			if !noOpen {
				openBrowser(url)
			}

			fmt.Printf("Web UI available at %s\n", url)
			fmt.Println("Press Ctrl+C to stop.")

			// 6. Handle signals and wait
			sigCh := make(chan os.Signal, 1)
			signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

			done := make(chan error, 1)
			go func() { done <- nodeCmd.Wait() }()

			select {
			case <-sigCh:
				nodeCmd.Process.Signal(syscall.SIGTERM)
				<-done
			case err := <-done:
				return err
			}

			return nil
		},
	}

	cmd.Flags().IntVar(&port, "port", 3141, "Port to serve on")
	cmd.Flags().BoolVar(&noOpen, "no-open", false, "Don't open browser automatically")
	return cmd
}

func waitForServer(url string, timeout time.Duration) error {
	deadline := time.Now().Add(timeout)
	client := &http.Client{Timeout: 500 * time.Millisecond}
	for time.Now().Before(deadline) {
		resp, err := client.Get(url + "/api/health")
		if err == nil {
			resp.Body.Close()
			if resp.StatusCode == http.StatusOK {
				return nil
			}
		}
		time.Sleep(250 * time.Millisecond)
	}
	return fmt.Errorf("server did not become ready within %s", timeout)
}

func resolveWebDir() string {
	// 1. Environment variable override (always wins)
	if dir := os.Getenv("ARMY_WEB_DIR"); dir != "" {
		if hasMainJS(dir) {
			return dir
		}
	}

	// 2. Build-time embedded path (set via ldflags)
	if WebDir != "" && hasMainJS(WebDir) {
		return WebDir
	}

	// 3. Relative to the executable
	exe, err := os.Executable()
	if err == nil {
		exe, _ = filepath.EvalSymlinks(exe)
		exeDir := filepath.Dir(exe)
		for _, rel := range []string{
			filepath.Join("..", "army", "web", "be"), // dev: .build/army
			filepath.Join("..", "web", "be"),         // installed alongside
		} {
			c := filepath.Join(exeDir, rel)
			if hasMainJS(c) {
				return c
			}
		}
	}

	// 4. Relative to CWD (running from project root)
	cwd, err := os.Getwd()
	if err == nil {
		c := filepath.Join(cwd, "army", "web", "be")
		if hasMainJS(c) {
			return c
		}
	}

	return ""
}

func hasMainJS(dir string) bool {
	_, err := os.Stat(filepath.Join(dir, "dist", "main.js"))
	return err == nil
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
	default:
		return
	}
	cmd.Start()
}
