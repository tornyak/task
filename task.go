package task

import (
	"sync/atomic"
	"time"
	"fmt"
)

type TaskID uint64
type TaskState int
type TaskFailedBehavior int

const (
	TaskStateNew TaskState = iota
	TaskStateRunning
	TaskStateDone
	TaskStateFailed
)

const (
	TaskFailedContinue TaskFailedBehavior = iota
	TaskFailedRepeat
	TaskFailedAbort
)

type Task interface {
	GetId() TaskID
	Run(c chan<- Result)
	GetStartTime() time.Time
	GetEndTime() time.Time
	GetState() TaskState
	SetState(TaskState)
	GetFailedBehavior() TaskFailedBehavior
}

type Result struct {
	ID    TaskID
	Err   error
	Value interface{}
}

type DefaultTask struct {
	id             TaskID
	state          TaskState
	failedBehavior TaskFailedBehavior
	runFunc        func(interface{}) (interface{}, error)
	runArg         interface{}
	startTime      time.Time
	endTime        time.Time
}

func (dt *DefaultTask) GetId() TaskID {
	return dt.id
}

func (dt *DefaultTask) GetState() TaskState {
	return dt.state
}

func (dt *DefaultTask) SetState(state TaskState) {
	dt.state = state
}

func (dt *DefaultTask) GetStartTime() time.Time {
	return dt.startTime
}

func (dt *DefaultTask) GetEndTime() time.Time {
	return dt.endTime
}

func (dt *DefaultTask) GetFailedBehavior() TaskFailedBehavior {
	return dt.failedBehavior
}

func (dt *DefaultTask) Run(resultChan chan<- Result) {
	dt.startTime = time.Now()
	result, err := dt.runFunc(dt.runArg)
	dt.endTime = time.Now()
	resultChan <- Result{ID: dt.id, Err: err, Value: result}
}

func (dt *DefaultTask) String() string {
	return fmt.Sprintf("Task (id: %v, state: %v)", dt.id, dt.state)
}

type DefaultTaskFactory struct {
	nextID uint64
}

func (tf *DefaultTaskFactory) NewTask(f func(interface{}) (interface{}, error),
	arg interface{},
	behavior TaskFailedBehavior) Task {

	return &DefaultTask{
		id:    TaskID(atomic.AddUint64(&tf.nextID, 1)),
		state: TaskStateNew,
		failedBehavior: behavior,
		runFunc: f,
		runArg: arg,
	}
}
