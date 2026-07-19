package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"
)

func clipboardWrite(text string) error {
	var cmd *exec.Cmd

	switch runtime.GOOS {
	case "darwin":
		cmd = exec.Command("pbcopy")
	case "linux":
		if isCommandAvailable("xclip") {
			cmd = exec.Command("xclip", "-selection", "clipboard")
		} else if isCommandAvailable("xsel") {
			cmd = exec.Command("xsel", "--clipboard", "--input")
		} else {
			return fmt.Errorf("no clipboard tool found. Install xclip or xsel")
		}
	case "windows":
		cmd = exec.Command("clip")
	default:
		return fmt.Errorf("unsupported platform: %s", runtime.GOOS)
	}

	cmd.Stdin = strings.NewReader(text)
	cmd.Stderr = os.Stderr

	return cmd.Run()
}

func isCommandAvailable(name string) bool {
	_, err := exec.LookPath(name)
	return err == nil
}
