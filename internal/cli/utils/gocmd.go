package utils

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

const goAttempts = 3

// networkErrors are transient failures when go talks to the module proxy.
var networkErrors = []string{"dial tcp", "server misbehaving", "i/o timeout", "connection reset", "TLS handshake timeout"}

// RunGo runs a go command quietly, retrying network failures, and returns its output only when it fails.
func RunGo(dir string, args ...string) error {

	var err error
	for attempt := 1; attempt <= goAttempts; attempt++ {
		err = Run(dir, "go", args...)
		if err == nil || !isNetworkError(err) {
			return err
		}

		time.Sleep(time.Duration(attempt) * time.Second)
	}

	if offlineErr := runGoOffline(dir, args...); offlineErr == nil {
		return nil
	}

	return err
}

// runGoOffline uses the local module cache as proxy, so modules downloaded before still work without network.
func runGoOffline(dir string, args ...string) error {

	cache, err := exec.Command("go", "env", "GOMODCACHE").Output()
	if err != nil {
		return err
	}

	cmd := exec.Command("go", args...)
	cmd.Dir = dir
	cmd.Env = append(os.Environ(),
		"GOPROXY=file://"+filepath.Join(strings.TrimSpace(string(cache)), "cache", "download"),
		"GOSUMDB=off",
		"GOFLAGS=-mod=mod",
	)

	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("%s: %w", output, err)
	}

	return nil
}

// Run executes a command without stdin, returning its output only when it fails.
func Run(dir, name string, args ...string) error {

	cmd := exec.Command(name, args...)
	cmd.Dir = dir

	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("%s %s: %w\n%s", name, strings.Join(args, " "), err, output)
	}

	return nil
}

func isNetworkError(err error) bool {
	for _, text := range networkErrors {
		if strings.Contains(err.Error(), text) {
			return true
		}
	}
	return false
}
