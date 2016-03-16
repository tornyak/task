package task

import (
	"errors"
	"fmt"
	"os"
	"os/signal"
	"time"
)

var (
	ErrTimeout   = errors.New("Received timeout")
	ErrInterrupt = errors.New("Received interrupt")
)

type Runner struct {
	interrupt  chan os.Signal   // OS interrupt signals
	complete   chan error       // To report when processing is done
	timeout    <-chan time.Time // To report timeout
	resultChan chan Result
	tasks      map[TaskID]Task
	tasksNext  map[TaskID][]Task
	tasksPrev  map[TaskID][]Task
}

func NewRunner(d time.Duration) *Runner {
	return &Runner{
		interrupt:  make(chan os.Signal, 1), // must use buffered channel, otherwise signal might be lost
		complete:   make(chan error),        // channel for runner result
		resultChan: make(chan Result),       // channel for task results
		timeout:    time.After(d),           // Timeout for whole runner
		tasks:      make(map[TaskID]Task),
		tasksNext:  make(map[TaskID][]Task), // Tasks that depend on task TaskID
		tasksPrev:  make(map[TaskID][]Task), // Tasks that are prerequisites for task TaskID
	}
}

func (r *Runner) Add(tasks ...Task) {
	for _, t := range tasks {
		r.tasks[t.GetId()] = t
		r.tasksNext[t.GetId()] = []Task{}
		r.tasksPrev[t.GetId()] = []Task{}
	}
}

// AddDependency specifies t1->t2 dependency.
// This means that t2 can be started only if t1 finished
func (r *Runner) AddDependency(t1, t2 Task) {
	r.tasksNext[t1.GetId()] = append(r.tasksNext[t1.GetId()], t2)
	r.tasksPrev[t2.GetId()] = append(r.tasksPrev[t2.GetId()], t1)
}

// RemoveDependency removes dependency t1->t2
func (r *Runner) RemoveDependency(t1, t2 Task) {
	r.tasksNext[t1.GetId()] = removeTaskFromSlice(r.tasksNext[t1.GetId()], t2)
	r.tasksPrev[t2.GetId()] = removeTaskFromSlice(r.tasksPrev[t2.GetId()], t1)
}

func (r *Runner) Start() error {
	// receive all interrupt based signals
	signal.Notify(r.interrupt, os.Interrupt)

	go func() {
		r.complete <- r.run()
	}()

	select {
	case err := <-r.complete:
		return err
	case <-r.timeout:
		return ErrTimeout
	}
}

func (r *Runner) run() error {

	waitingTasks := []Task{}
	runningTasks := []Task{}

	for id, prev := range r.tasksPrev {
		if len(prev) == 0 {
			waitingTasks = append(waitingTasks, r.tasks[id])
		}
	}

	for len(waitingTasks) > 0 || len(runningTasks) > 0 {
		// Check for interrupt from OS
		if r.gotInterrupt() {
			return ErrInterrupt
		}
		for _, t := range waitingTasks {
			t.SetState(TaskStateRunning)

			fmt.Printf("Run: %v\n", t)
			// Execute the registered task
			go t.Run(r.resultChan)
		}
		runningTasks = append(runningTasks, waitingTasks...)
		waitingTasks = []Task{}

		select {
		case result := <-r.resultChan:
			doneTask := r.tasks[result.ID]
			fmt.Printf("Task Done: %v\n", doneTask)
			runningTasks = removeTaskFromSlice(runningTasks, doneTask)
			// Handle error
			if result.Err != nil {
				doneTask.SetState(TaskStateFailed)
				switch doneTask.GetFailedBehavior() {
				case TaskFailedAbort:
					return result.Err
				case TaskFailedRepeat:
					doneTask.SetState(TaskStateNew)
					waitingTasks = append(waitingTasks, doneTask)
				}
			} else {
				waitingTasks = append(waitingTasks, r.getNextTasks(doneTask)...)
				fmt.Printf("WaitingTasks: %v\n", waitingTasks)
			}
		}

	}

	return nil
}

func (r *Runner) gotInterrupt() bool {
	select {
	case <-r.interrupt:
		signal.Stop(r.interrupt)
		return true
	default:
		return false
	}
}

func (r *Runner) getNextTasks(doneTask Task)[]Task {
	retTasks := []Task{}
	for _, nextTask := range r.tasksNext[doneTask.GetId()] {
		r.RemoveDependency(doneTask, nextTask)
		// if nextTask does not have more dependencies add it to the retTasks
		if len(r.tasksPrev[nextTask.GetId()]) == 0 {
			retTasks = append(retTasks, nextTask)
		}
	}
	return retTasks
}

func removeTaskFromSlice(tasks []Task, task Task) []Task {
	for i, t := range tasks {
		if t.GetId() == task.GetId() {
			tasks = append(tasks[:i], tasks[i+1:]...)
		}
	}
	return tasks
}
