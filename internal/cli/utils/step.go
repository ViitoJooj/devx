package utils

import (
	"fmt"
	"os"
	"time"

	"golang.org/x/term"
)

var spinnerFrames = []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}

const clearLine = "\r\x1b[K"

// Step runs fn while showing a spinner with label, and leaves a ✓ or ✗ line behind.
// Without a terminal (CI, pipes) it prints plain lines instead of the animation.
// The spinner only writes to stdout: stdin and the terminal mode are not touched.
func Step(label string, fn func() error) error {

	if !term.IsTerminal(int(os.Stdout.Fd())) {
		err := fn()
		if err != nil {
			fmt.Println("failed: " + label)
			return err
		}
		fmt.Println("done: " + label)
		return nil
	}

	done := make(chan error, 1)
	go func() { done <- fn() }()

	ticker := time.NewTicker(80 * time.Millisecond)
	defer ticker.Stop()

	for frame := 0; ; frame++ {
		fmt.Print(clearLine + spinnerFrames[frame%len(spinnerFrames)] + " " + label)

		select {
		case err := <-done:
			mark := "✓"
			if err != nil {
				mark = "✗"
			}
			fmt.Println(clearLine + mark + " " + label)
			return err
		case <-ticker.C:
		}
	}
}
