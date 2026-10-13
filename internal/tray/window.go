package tray

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sync"
)

var windowProcess struct {
	sync.Mutex
	cmd *exec.Cmd
}

// openWindow keeps a single window process alive at a time.
func openWindow() error {
	windowProcess.Lock()
	defer windowProcess.Unlock()
	if windowProcess.cmd != nil {
		return nil
	}
	path, err := os.Executable()
	if err != nil {
		return err
	}
	sibling := filepath.Join(filepath.Dir(path), "monctl-window")
	var cmd *exec.Cmd
	if info, err := os.Stat(sibling); err == nil && info.Mode().IsRegular() && info.Mode()&0111 != 0 {
		cmd = exec.Command(sibling)
	} else if path, err := exec.LookPath("monctl-window"); err == nil {
		cmd = exec.Command(path)
	} else if _, err := os.Stat("go.mod"); err == nil {
		// Allow `go run ./cmd/monctl-tray` while developing in the repository.
		cmd = exec.Command("go", "run", "-tags=nucular_shiny", "./cmd/monctl-window")
	} else {
		return fmt.Errorf("monctl-window not installed; install the package or build it with -tags=nucular_shiny")
	}
	cmd.Stdout, cmd.Stderr = os.Stdout, os.Stderr
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("start slider window: %w", err)
	}
	windowProcess.cmd = cmd
	go func() {
		_ = cmd.Wait()
		windowProcess.Lock()
		windowProcess.cmd = nil
		windowProcess.Unlock()
	}()
	return nil
}
