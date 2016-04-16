package task

import (
	"errors"
	"fmt"
	"os"
	"os/signal"
	"time"
)

var (
	errTimeout   = errors.New("Received timeout")
	errInterrupt = errors.New("Received interrupt")
)

const defaultRepeatTaskTmo = 100 * time.Millisecond

// Runner is generic task runner for executing dependent tasks.
// Tasks can be any datatype that implements task/Task interface
// Each task is executed in its own goroutine returning result via channel
// Runner can end its execution in following ways:
// 1. All tasks finished their execution and returned
// 2. There was interrupt sent on interrupt channel
// 3. There was timeout
// 4. A task with behavior TaskFailedAbort failed during execution
type Runner struct {
	interrupt  chan os.Signal   // OS interrupt signals
	complete   chan error       // To report when processing is done
	timeout    <-chan time.Time // To report timeout for whole runner
	resultChan chan Result      // channel for task results
	tasks      map[ID]Task
	tasksNext  map[ID][]Task // Tasks that depend on task TaskID
	tasksPrev  map[ID][]Task // Tasks that are prerequisites for task TaskID
}

// NewRunner creates a new Runner instance
func NewRunner(d time.Duration) *Runner {
	return &Runner{
		interrupt:  make(chan os.Signal, 1), // Must use buffered channel
		complete:   make(chan error),
		resultChan: make(chan Result),
		timeout:    time.After(d),
		tasks:      make(map[ID]Task),
		tasksNext:  make(map[ID][]Task),
		tasksPrev:  make(map[ID][]Task),
	}
}

// Add adds one or more tasks to the runner
func (r *Runner) Add(tasks ...Task) {
	for _, t := range tasks {
		if t != nil {
			r.tasks[t.GetID()] = t
			r.tasksNext[t.GetID()] = []Task{}
			r.tasksPrev[t.GetID()] = []Task{}
		}
	}
}

// AddDependency specifies t1->t2 dependency.
// This means that t2 can be started only if t1 finished
func (r *Runner) AddDependency(tasks ...Task) error {
	if err := r.verifyTasksInRunner(tasks...); err != nil {
		return err
	}

	for i := 0; i < len(tasks)-1; i++ {
		t1, t2 := tasks[i], tasks[i+1]
		r.tasksNext[t1.GetID()] = append(r.tasksNext[t1.GetID()], t2)
		r.tasksPrev[t2.GetID()] = append(r.tasksPrev[t2.GetID()], t1)
	}

	return nil
}

// RemoveDependency removes dependency t1->t2
func (r *Runner) RemoveDependency(tasks ...Task) error {

	if err := r.verifyTasksInRunner(tasks...); err != nil {
		return err
	}

	for i := 0; i < len(tasks)-1; i++ {
		t1, t2 := tasks[i], tasks[i+1]
		r.tasksNext[t1.GetID()] = removeTaskFromSlice(r.tasksNext[t1.GetID()], t2)
		r.tasksPrev[t2.GetID()] = removeTaskFromSlice(r.tasksPrev[t2.GetID()], t1)
	}

	return nil
}

// Start starts execution of tasks in Runner
// goroutine that called Start will remain blocked untill runner finishes execution
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
		return errTimeout
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
			return errInterrupt
		}
		for _, t := range waitingTasks {
			t.SetState(TaskStateRunning)

			runningTasks = append(runningTasks, t)
			// Execute the registered task
			go t.Run(r.resultChan)
		}
		waitingTasks = []Task{}

		select {
		case result := <-r.resultChan:
			doneTask := r.tasks[result.ID]
			runningTasks = removeTaskFromSlice(runningTasks, doneTask)
			// Handle error
			if result.Err != nil {
				doneTask.SetState(TaskStateFailed)
				switch doneTask.GetFailedBehavior() {
				case TaskFailedAbort:
					r.waitForRunningTasks(runningTasks)
					return result.Err
				case TaskFailedRepeat:
					doneTask.SetState(TaskStateRunning)
					runningTasks = append(runningTasks, doneTask)
					// Want to wait for some time before retrying
					time.AfterFunc(defaultRepeatTaskTmo, func() {
						doneTask.Run(r.resultChan)
					})
				}
			} else {
				doneTask.SetState(TaskStateDone)
				waitingTasks = append(waitingTasks, r.getNextTasks(doneTask)...)
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

func (r *Runner) getNextTasks(doneTask Task) []Task {
	retTasks := []Task{}
	for _, nextTask := range r.tasksNext[doneTask.GetID()] {
		r.RemoveDependency(doneTask, nextTask)
		// if nextTask does not have more dependencies add it to the retTasks
		if len(r.tasksPrev[nextTask.GetID()]) == 0 {
			retTasks = append(retTasks, nextTask)
		}
	}
	return retTasks
}

func (r *Runner) verifyTasksInRunner(tasks ...Task) error {
	for _, t := range tasks {
		if t == nil {
			return fmt.Errorf("Task is nil! tasks: %v", tasks)
		}
		if foundTask := r.tasks[t.GetID()]; foundTask == nil {
			return fmt.Errorf("%v not found", t)
		}
	}
	return nil
}

func (r *Runner) waitForRunningTasks(runningTasks []Task) {
	for len(runningTasks) > 0 {
		select {
		case result := <-r.resultChan:
			doneTask := r.tasks[result.ID]
			runningTasks = removeTaskFromSlice(runningTasks, doneTask)
			if result.Err != nil {
				doneTask.SetState(TaskStateFailed)
			} else {
				doneTask.SetState(TaskStateDone)
			}
		}
	}
}

func removeTaskFromSlice(tasks []Task, task Task) []Task {
	for i, t := range tasks {
		if t.GetID() == task.GetID() {
			tasks = append(tasks[:i], tasks[i+1:]...)
		}
	}
	return tasks
}
