package contracts

import (
	"context"
	"time"
)

type IWorker interface {
	Name() string
	Interval() time.Duration
	Run(ctx context.Context) error
}

type WorkerOptions struct {
	Interval time.Duration
}

type WorkerOption func(*WorkerOptions)

func WithInterval(d time.Duration) WorkerOption {
	return func(o *WorkerOptions) { o.Interval = d }
}

func NewWorkerOptions(defaultInterval time.Duration, opts ...WorkerOption) WorkerOptions {
	o := WorkerOptions{Interval: defaultInterval}
	for _, opt := range opts {
		opt(&o)
	}
	return o
}
