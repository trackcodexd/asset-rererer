package taskqueue

import (
	"container/list"
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
	isSchedulerRunning bool
	tasks              *list.List
	interval           time.Duration
	mutex              sync.Mutex
}

func New[R any](interval time.Duration) *Queue[R] {
	return &Queue[R]{
		tasks:    list.New(),
		interval: interval,
	}
}

func (q *Queue[R]) QueueTask(f func() (R, error)) chan TaskResult[R] {
	taskChan := make(chan TaskResult[R])

	q.mutex.Lock()
	defer q.mutex.Unlock()

	q.tasks.PushBack(task[R]{
		Func: f,
		Chan: taskChan,
	})

	if !q.isSchedulerRunning {
		go q.startScheduler()
	}

	return taskChan
}

func (q *Queue[R]) startScheduler() {
	q.isSchedulerRunning = true

	for {
		q.mutex.Lock()
		if q.tasks.Len() == 0 {
			q.isSchedulerRunning = false
			q.mutex.Unlock()
			return
		}

		element := q.tasks.Front()
		task := element.Value.(task[R])

		q.tasks.Remove(element)
		q.mutex.Unlock()

		go func() {
			res, err := task.Func()
			task.Chan <- TaskResult[R]{
				Result: res,
				Error:  err,
			}
		}()

		time.Sleep(q.interval)
	}
}
