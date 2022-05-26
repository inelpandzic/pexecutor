package executor

import (
	"fmt"
	"testing"
    "time"

	"go.uber.org/zap"
)

func TestExecutor(t *testing.T) {

	// create 10 longrunning tasks
	// have a queue of 5
	// check that 5 are running and 5 are pending
	// sleep so that all can be executed
	// check that 0 are running 0 are pending

	log, _ := zap.NewDevelopment()
    poolSize := 5
	ex := New(poolSize, 100, log)
    go ex.Run()
	defer ex.Close()


    // Ensure that the executor is up and runningn
    time.Sleep(100 * time.Millisecond)

    const totalTasks = 10

	for i := 0; i < totalTasks; i++ {
		t := &Task{
			Name:     fmt.Sprintf("T-%d", i),
			Duration: 3 * time.Second,
		}
		ex.Submit(t)
	}

    // Ensure that tasks are submitted
    time.Sleep(100 * time.Millisecond)

    runningTasks := len(ex.GetRunningTasks())
    if runningTasks != poolSize {
        t.Fatalf("Number of running tasks should be %d, but got %d", poolSize,runningTasks)
    }

    pendingTasks := len(ex.GetPendingTasks())
    if pendingTasks != totalTasks - poolSize {
        t.Fatalf("Number of pending tasks should be %d, but got %d", totalTasks - poolSize, pendingTasks)
    }


    // Ensure that all tasks are executed
    time.Sleep(6 * time.Second)

    runningTasks = len(ex.GetRunningTasks())
    if runningTasks != 0 {
        t.Fatalf("Number of running tasks should be %d, but got %d", 0, runningTasks)
    }

    pendingTasks = len(ex.GetPendingTasks())
    if pendingTasks != 0 {
        t.Fatalf("Number of pending tasks should be %d, but got %d", 0, pendingTasks)
    }
}
