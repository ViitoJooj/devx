package errorx

import "os"

var exitHooks []func()

// OnExit registers a function that runs before the process ends through Exit, so cleanups (like the project lock) are not skipped.
func OnExit(hook func()) {
	exitHooks = append(exitHooks, hook)
}

// Exit runs the registered hooks and ends the process.
func Exit(code int) {
	for _, hook := range exitHooks {
		hook()
	}
	os.Exit(code)
}
