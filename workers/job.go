package workers

import (
	"context"
)

type ExecutionFn func(ctx context.Context, args interface{}) interface{}

type Result struct {
	Value interface{}
	Err   error
}

type Job struct {
	ExecFn ExecutionFn
	Args   interface{}
}

func (j Job) execute(ctx context.Context) Result {
	value := j.ExecFn(ctx, j.Args)

	return Result{
		Value: value,
	}
}
