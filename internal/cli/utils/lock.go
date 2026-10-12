package utils

import (
	"errors"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"sync"
	"time"

	"github.com/ViitoJooj/devx/pkg/errorx"
)

const (
	lockName      = ".devx.lock"
	lockWait      = 15 * time.Second
	lockStaleTime = 5 * time.Minute
)

var (
	lockPath   string
	signalOnce sync.Once
)

// LockProject keeps two devx runs from editing the same project at once (main.go, openapi and the makefile are read and rewritten).
// A second run waits for the first; a lock older than lockStaleTime is from a crashed run and is taken over.
func LockProject(root string) error {

	if lockPath != "" {
		return nil
	}

	path := filepath.Join(root, lockName)
	deadline := time.Now().Add(lockWait)

	for {
		file, err := os.OpenFile(path, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0o644)
		if err == nil {
			fmt.Fprintf(file, "%d\n", os.Getpid())
			file.Close()

			lockPath = path
			signalOnce.Do(releaseOnInterrupt)
			return nil
		}

		if !os.IsExist(err) {
			return err
		}

		if info, statErr := os.Stat(path); statErr == nil && time.Since(info.ModTime()) > lockStaleTime {
			os.Remove(path)
			continue
		}

		if time.Now().After(deadline) {
			return errors.New("project is locked")
		}

		time.Sleep(100 * time.Millisecond)
	}
}

// UnlockProject releases the lock taken by LockProject; it is safe to call without one.
func UnlockProject() {
	if lockPath != "" {
		os.Remove(lockPath)
		lockPath = ""
	}
}

func releaseOnInterrupt() {

	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)

	go func() {
		<-interrupt
		errorx.Exit(130)
	}()
}
