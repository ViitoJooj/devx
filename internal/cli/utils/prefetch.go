package utils

import (
	"os"
	"os/exec"
	"sync"
)

// Fetch is a module to download into the module cache before the project asks for it.
type Fetch struct {
	Path string
	Tool bool
}

// Prefetch downloads every module at the same time, each in its own scratch module, so the go get and the go mod tidy
// of the project only read the cache. It is best effort: failures are ignored, because the project commands retry on their own.
// The returned function waits for the downloads and removes the scratch folders.
func Prefetch(fetches ...Fetch) func() {

	var wg sync.WaitGroup

	for _, fetch := range fetches {
		wg.Go(func() { prefetchOne(fetch) })
	}

	return wg.Wait
}

func prefetchOne(fetch Fetch) {

	dir, err := os.MkdirTemp("", "devx-prefetch-")
	if err != nil {
		return
	}
	defer os.RemoveAll(dir)

	env := append(os.Environ(), "GOWORK=off", "GOFLAGS=-mod=mod")

	init := exec.Command("go", "mod", "init", "prefetch")
	init.Dir, init.Env = dir, env
	if init.Run() != nil {
		return
	}

	args := []string{"get"}
	if fetch.Tool {
		args = append(args, "-tool")
	}

	get := exec.Command("go", append(args, fetch.Path)...)
	get.Dir, get.Env = dir, env
	get.Run()
}
