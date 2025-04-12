package taskqueue

import (
	"slices"
	"sync"
	"time"
)

type TaskResult[R any] struct {
	Result R
	Error  error
}

type task[R any] struct {
	Func func() (R, error)
	Chan chan TaskResult[R]
}

type Queue[R any] struct {
	tasks    []task[R]
	interval time.Duration
	mutex    sync.Mutex
}

func New[R any](interval time.Duration) *Queue[R] {
	return &Queue[R]{
		tasks:    make([]task[R], 0),
		interval: interval,
	}
}

func (q *Queue[R]) QueueTask(f func() (R, error)) chan TaskResult[R] {
	taskChan := make(chan TaskResult[R])

	q.mutex.Lock()

	q.tasks = append(q.tasks, task[R]{
		Func: f,
		Chan: taskChan,
	})

	if len(q.tasks) == 1 {
		go q.startScheduler()
	}

	q.mutex.Unlock()

	return taskChan
}

func (q *Queue[R]) startScheduler() {
	for {
		if len(q.tasks) == 0 {
			return
		}

		task := q.tasks[0]
		go func() {
			res, err := task.Func()
			task.Chan <- TaskResult[R]{
				Result: res,
				Error:  err,
			}
		}()

		q.mutex.Lock()
		q.tasks = slices.Delete(q.tasks, 0, 1)
		q.mutex.Unlock()

		time.Sleep(q.interval)
	}
}
