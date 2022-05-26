package executor

import (
	"sync"
	"time"

	"go.uber.org/zap"
)

// E represents the executor that will accept tasks and run them on the
// first available worker.
type E struct {
	terminated chan bool

	queueSize int
	poolSize  int
	taskQueue chan *Task

	tmu   sync.Mutex
	tasks map[string]*Task

	log *zap.Logger
}

// Task represents a single unit of work executed by the Executor
type Task struct {
	// Task name
	Name string
	// Task duration in milliseconds
	Duration time.Duration

	rmu     sync.Mutex
	running bool
}

// New creates a new Executor ready to be started.
// poolSize is the number of workers executing tasks
// queueSize is the number of tasks that can be queued up
// waiting to be executed. If the queue is full, Submit will block.
func New(poolSize int, queueSize int, log *zap.Logger) *E {
	return &E{
		terminated: make(chan bool),
		queueSize:  queueSize,
		poolSize:   poolSize,
		tasks:      make(map[string]*Task),
		taskQueue:  make(chan *Task, queueSize),
        log: log,
	}
}

// Run starts up the exector
func (e *E) Run() {
	for i := 0; i < e.poolSize; i++ {
		go func(id int) {
			e.worker(id)
		}(i)
	}

	e.log.Sugar().Infof("Executor started: workers: %d, queue capacity: %d", e.poolSize, e.queueSize)
	<-e.terminated
}

// Submit submits a task to set of tasks waiting to be executed by a worker.
// Returns false if there is already a pending or running task with the same name,
// otherwise it returns true.
// Submit will block if the task queue gets full.
func (e *E) Submit(t *Task) bool {
	e.tmu.Lock()
	defer e.tmu.Unlock()

	if _, ok := e.tasks[t.Name]; ok {
		e.log.Sugar().Debugf("Duplicate task: %v", t)
		return false
	}

	e.tasks[t.Name] = t

	e.taskQueue <- t

	e.log.Sugar().Debugf("Submitied task: %v", t)
	return true
}

// GetRunningTasks return list of currently running tasks
func (e *E) GetRunningTasks() []*Task {
	var r []*Task

	e.tmu.Lock()
	defer e.tmu.Unlock()
	for _, v := range e.tasks {
		if v.running {
			r = append(r, v)
		}
	}

	return r
}

// GetPendingTasks return list of currently pending tasks, waiting to be executed
func (e *E) GetPendingTasks() []*Task {
	var r []*Task

	e.tmu.Lock()
	defer e.tmu.Unlock()
	for _, v := range e.tasks {
		if !v.running {
			r = append(r, v)
		}
	}

	return r
}

// Close stops the executor
func (e *E) Close() {
	e.log.Sugar().Info("Closing the exeutor...")
	e.terminated <- true
}

// worker takes tasks from the queue end "executes" them.
func (e *E) worker(id int) {
	for {
		t := <-e.taskQueue

		t.rmu.Lock()
		t.running = true
		t.rmu.Unlock()

		e.log.Sugar().Debugf("Worker %d - Executing task: %v", id, t)

		// processing the tasks
		time.Sleep(t.Duration)

		e.tmu.Lock()
		delete(e.tasks, t.Name)
		e.tmu.Unlock()
	}
}
