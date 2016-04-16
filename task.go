package task

import (
	"fmt"
	"sync/atomic"
	"time"
)

// ID is a unique Task identifier
type ID uint64

// State represents current state of the Task
type State int

// FailedBehavior defines behavior in case that Task execution fails
type FailedBehavior int

// Task state
const (
	TaskStateNew State = iota
	TaskStateRunning
	TaskStateDone
	TaskStateFailed
)

// What to do if task fails
const (
	TaskFailedContinue FailedBehavior = iota
	TaskFailedRepeat
	TaskFailedAbort
)

// Task is interface that tasks must implement to be supported by task/Runner
type Task interface {
	GetID() ID
	GetDescription() string
	Run(c chan<- Result)
	GetStartTime() time.Time
	GetEndTime() time.Time
	GetState() State
	SetState(State)
	GetFailedBehavior() FailedBehavior
}

// Result represents result that Tasks return to the Runner
type Result struct {
	ID    ID
	Err   error
	Value interface{}
}

// DefaultTask is simple implementation of Task interface
type DefaultTask struct {
	id             ID
	description    string
	state          State
	failedBehavior FailedBehavior
	runFunc        func(...interface{}) (interface{}, error)
	runArgs        []interface{}
	startTime      time.Time
	endTime        time.Time
}

// GetID returns Task ID
func (dt *DefaultTask) GetID() ID {
	return dt.id
}

// GetDescription return text description of the Task
func (dt *DefaultTask) GetDescription() string {
	return dt.description
}

// GetState returns Task state
func (dt *DefaultTask) GetState() State {
	return dt.state
}

// SetState sets Task state
func (dt *DefaultTask) SetState(state State) {
	dt.state = state
}

// GetStartTime returns time when task execution started
func (dt *DefaultTask) GetStartTime() time.Time {
	return dt.startTime
}

// GetEndTime returns time when task execution ended
func (dt *DefaultTask) GetEndTime() time.Time {
	return dt.endTime
}

// GetFailedBehavior returns behavior for failed tasks
func (dt *DefaultTask) GetFailedBehavior() FailedBehavior {
	return dt.failedBehavior
}

// Run executes the Task and returns result on resultChan
func (dt *DefaultTask) Run(resultChan chan<- Result) {
	var result interface{}
	var err error

	dt.startTime = time.Now()
	// Need to handle possible panic in dt.runFunc in a safe way
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("Panic in %v: %v", dt, r)
		}
		dt.endTime = time.Now()
		resultChan <- Result{ID: dt.id, Err: err, Value: result}
	}()
	result, err = dt.runFunc(dt.runArgs...)
}

func (dt *DefaultTask) String() string {
	return fmt.Sprintf("Task (id: %v, state: %v, %v)", dt.id, dt.state, dt.description)
}

// DefaultTaskFactory is simple factory for creating instances of DefaultTask
type DefaultTaskFactory struct {
	nextID ID
}

// NewTask creates a new DefaultTask instance
func (tf *DefaultTaskFactory) NewTask(f func(...interface{}) (interface{}, error),
	description string,
	behavior FailedBehavior,
	arg ...interface{}) Task {

	return &DefaultTask{
		id:             ID(atomic.AddUint64((*uint64)(&tf.nextID), 1)),
		description:    description,
		state:          TaskStateNew,
		failedBehavior: behavior,
		runFunc:        f,
		runArgs:        append([]interface{}{}, arg...),
	}
}
