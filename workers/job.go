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
	// if err != nil {
	// 	logrus.Error("Execute error ", err.Error())
	// 	return Result{
	// 		Err: err,
	// 	}
	// }

	return Result{
		Value: value,
	}
}
