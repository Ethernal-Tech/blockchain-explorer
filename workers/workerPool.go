package workers

import (
	"context"
	"sync"

	"github.com/sirupsen/logrus"
)

func worker(ctx context.Context, wg *sync.WaitGroup, jobs <-chan Job, results chan<- Result) {
	logrus.Debug("New worker is created")

	defer wg.Done()
	for {
		select {
		case job, ok := <-jobs:
			if !ok {
				return
			}
			// fan-in job execution multiplexing results into the results channel
			results <- job.execute(ctx)
		case <-ctx.Done():
			logrus.Error("Cancelled worker, err: ", ctx.Err())
			results <- Result{
				Err: ctx.Err(),
			}
			return
		}
	}
}

type WorkerPool struct {
	workersCount uint
	jobs         chan Job
	results      chan Result
	Done         chan struct{}
}

func New(wcount uint) WorkerPool {
	return WorkerPool{
		workersCount: wcount,
		jobs:         make(chan Job, wcount),
		results:      make(chan Result, wcount),
		Done:         make(chan struct{}),
	}
}

// Run starts worker goroutines for fetching data from blockchain. Workers read from jobs channel, execute Job function and Result write into results channel.
func (wp WorkerPool) Run(ctx context.Context, wg *sync.WaitGroup) {
	wg.Add(1)

	var i uint
	for i = 0; i < wp.workersCount; i++ {
		wg.Add(1)
		// fan out worker goroutines
		// reading from jobs channel and pushing calcs into results channel
		go worker(ctx, wg, wp.jobs, wp.results)
	}

	//wait until the goroutines have finished and all data from result channel has been read
	wg.Wait()
	close(wp.Done)
	close(wp.results)
}

// Results returns WorkerPool results channel.
func (wp WorkerPool) Results() <-chan Result {
	return wp.results
}

// GenerateFrom adds Jobs to WorkerPool jobs channel and closes it after adding all of them.
func (wp WorkerPool) GenerateFrom(jobsBulk []Job) {
	for i := range jobsBulk {
		wp.jobs <- jobsBulk[i]
	}
	close(wp.jobs)
}
