package task

import (
	"math"
	"os"
	"syscall"
	"testing"
	"time"
)

func TestIndependentOk(t *testing.T) {

	r := NewRunner(1 * time.Second)
	dtf := DefaultTaskFactory{}

	tasks := make([]Task, 3, 3)
	for i := range tasks {
		tasks[i] = dtf.NewTask(sleepFunc, "sleep", TaskFailedAbort, 1*time.Millisecond)
	}
	r.Add(tasks...)

	if err := r.Start(); err != nil {
		t.Fatal("Task execution error")
	}

	verifyTaskState(t, TaskStateDone, tasks...)
}

func TestTmo(t *testing.T) {
	tmo := 1 * time.Millisecond

	r := NewRunner(tmo)
	dtf := DefaultTaskFactory{}

	tasks := make([]Task, 3, 3)
	for i := range tasks {
		tasks[i] = dtf.NewTask(sleepFunc, "sleep", TaskFailedAbort, 1*time.Millisecond)
	}
	r.Add(tasks...)

	if err := r.Start(); err != nil {
		switch err {
		case errTimeout:
			t.Log("Timeout as expected")
			return
		default:
			t.Fatal("Unexpected error")
		}

	}
	t.Fatal("No error returned")
}

func TestSignalInterrupt(t *testing.T) {

	r := NewRunner(10 * time.Second)
	dtf := DefaultTaskFactory{}

	tasks := make([]Task, 3, 3)
	for i := range tasks {
		tasks[i] = dtf.NewTask(sleepFunc, "sleep", TaskFailedAbort, 1*time.Second)
	}
	r.Add(tasks...)

	go syscall.Kill(syscall.Getpid(), syscall.SIGINT)

	if err := r.Start(); err != nil {
		switch err {
		case errInterrupt:
			t.Log("Interrupt as expected")
			return
		default:
			t.Fatal("Unexpected error")
		}

	}
	t.Fatal("No error returned")
}

func TestSimulateInterrupt(t *testing.T) {

	r := NewRunner(10 * time.Second)
	dtf := DefaultTaskFactory{}

	tasks := make([]Task, 3, 3)
	for i := range tasks {
		tasks[i] = dtf.NewTask(sleepFunc, "sleep", TaskFailedAbort, 1*time.Second)
	}
	r.Add(tasks...)

	go func() {
		r.interrupt <- os.Interrupt
	}()

	if err := r.Start(); err != nil {
		switch err {
		case errInterrupt:
			t.Log("Interrupt as expected")
			return
		default:
			t.Fatal("Unexpected error")
		}

	}
	t.Fatal("No error returned")
}

func TestDependentSimpleOk(t *testing.T) {

	r := NewRunner(1 * time.Second)
	dtf := DefaultTaskFactory{}

	tasks := make([]Task, 3, 3)
	for i := range tasks {
		tasks[i] = dtf.NewTask(sleepFunc, "sleep", TaskFailedAbort, 1*time.Millisecond)
	}
	r.Add(tasks...)

	r.AddDependency(tasks[0], tasks[1])
	r.AddDependency(tasks[1], tasks[2])

	if err := r.Start(); err != nil {
		t.Fatal("Task execution error")
	}

	checkDependency(t, tasks...)
}

func TestAddDependencyNonExistant(t *testing.T) {

	r := NewRunner(1 * time.Second)
	dtf := DefaultTaskFactory{}

	tasks1 := dtf.NewTask(sleepFunc, "sleep", TaskFailedAbort, 1*time.Millisecond)
	tasks2 := dtf.NewTask(sleepFunc, "sleep", TaskFailedAbort, 1*time.Millisecond)

	if err := r.AddDependency(tasks1, tasks2); err == nil {
		t.Fatal("AddDependency should return err")
	}
}

func TestDependantComplexOk(t *testing.T) {

	r := NewRunner(1 * time.Second)
	dtf := DefaultTaskFactory{}
	t1 := dtf.NewTask(sleepFunc, "sleep", TaskFailedAbort, 1*time.Millisecond)
	t2 := dtf.NewTask(sleepFunc, "sleep", TaskFailedAbort, 2*time.Millisecond)
	t3 := dtf.NewTask(sleepFunc, "sleep", TaskFailedAbort, 1*time.Millisecond)
	t4 := dtf.NewTask(sleepFunc, "sleep", TaskFailedAbort, 3*time.Millisecond)
	t5 := dtf.NewTask(sleepFunc, "sleep", TaskFailedAbort, 1*time.Millisecond)
	t6 := dtf.NewTask(sleepFunc, "sleep", TaskFailedAbort, 1*time.Millisecond)
	t7 := dtf.NewTask(sleepFunc, "sleep", TaskFailedAbort, 1*time.Millisecond)
	t8 := dtf.NewTask(sleepFunc, "sleep", TaskFailedAbort, 1*time.Millisecond)
	t9 := dtf.NewTask(sleepFunc, "sleep", TaskFailedAbort, 1*time.Millisecond)
	r.Add(t1, t2, t3, t4, t5, t6, t7, t8, t9)

	r.AddDependency(t1, t4)
	r.AddDependency(t2, t4)
	r.AddDependency(t3, t5)
	r.AddDependency(t1, t6)
	r.AddDependency(t4, t6)
	r.AddDependency(t5, t6)
	r.AddDependency(t5, t7)
	r.AddDependency(t6, t8)
	r.AddDependency(t6, t9)
	r.AddDependency(t7, t9)

	if err := r.Start(); err != nil {
		t.Fatal("Task execution error")
	}

	checkDependency(t, t1, t4)
	checkDependency(t, t2, t4)
	checkDependency(t, t3, t5)
	checkDependency(t, t1, t6)
	checkDependency(t, t4, t6)
	checkDependency(t, t5, t6)
	checkDependency(t, t5, t7)
	checkDependency(t, t6, t8)
	checkDependency(t, t6, t9)
	checkDependency(t, t7, t9)
}

// Creates binary dependency tree where task dependencies look like:
// 0 -> 1,2
// 1 -> 3,4
// 2 -> 5,6
// 3 -> 7,8
// 4 -> 9,10
// .........
func TestDependantTreeOk(t *testing.T) {

	numTasks := 1023 // 2^10 - 1

	r := NewRunner(10 * time.Second)
	dtf := DefaultTaskFactory{}
	tasks := make([]Task, int(numTasks), int(numTasks))
	for i := range tasks {
		tasks[i] = dtf.NewTask(sleepFunc, "sleep", TaskFailedAbort, 1*time.Millisecond)
	}
	r.Add(tasks...)

	// Create a tree of tasks
	for i := 0; i < 9; i++ {
		// layer i contains 2^i tasks
		// need to connect each of them with two of 2^(i+1) tasks in layer i+1
		tasksInLayerI := int(math.Exp2(float64(i)))
		for j := 0; j < tasksInLayerI; j++ {
			taskID1 := tasksInLayerI - 1 + j
			taskID2 := 2*tasksInLayerI + 2*j - 1
			taskID3 := taskID2 + 1

			r.AddDependency(tasks[taskID1], tasks[taskID2])
			r.AddDependency(tasks[taskID1], tasks[taskID3])
		}
	}

	if err := r.Start(); err != nil {
		t.Fatal("Task execution error")
	}
}

func TestDependantPanicContinue(t *testing.T) {
	r := NewRunner(1 * time.Second)
	dtf := DefaultTaskFactory{}
	t1 := dtf.NewTask(sleepFunc, "sleep", TaskFailedAbort, 1*time.Millisecond)
	t2 := dtf.NewTask(panicFunc, "panic", TaskFailedContinue, 1*time.Millisecond)
	t3 := dtf.NewTask(sleepFunc, "sleep", TaskFailedAbort, 2*time.Millisecond)
	t4 := dtf.NewTask(sleepFunc, "sleep", TaskFailedAbort, 1*time.Millisecond)
	r.Add(t1, t2, t3, t4)

	r.AddDependency(t1, t2)
	r.AddDependency(t1, t3)
	r.AddDependency(t2, t4)

	if err := r.Start(); err != nil {
		t.Fatal("Task execution error")
	}

	checkDependency(t, t1, t2)
	checkDependency(t, t1, t3)

	verifyTaskState(t, TaskStateDone, t1, t3)
	verifyTaskState(t, TaskStateFailed, t2)
	verifyTaskState(t, TaskStateNew, t4)
}

func TestDependantPanicAbort(t *testing.T) {
	r := NewRunner(1 * time.Second)
	dtf := DefaultTaskFactory{}
	t1 := dtf.NewTask(sleepFunc, "sleep", TaskFailedAbort, 1*time.Millisecond)
	t2 := dtf.NewTask(panicFunc, "panic", TaskFailedAbort, 1*time.Millisecond)
	t3 := dtf.NewTask(sleepFunc, "sleep", TaskFailedAbort, 2*time.Millisecond)
	t4 := dtf.NewTask(sleepFunc, "sleep", TaskFailedAbort, 1*time.Millisecond)
	r.Add(t1, t2, t3, t4)

	r.AddDependency(t1, t2)
	r.AddDependency(t1, t3)
	r.AddDependency(t2, t4)

	var err error

	if err = r.Start(); err == nil {
		t.Fatal("No error returned")
	}

	if err.Error() != "Panic in Task (id: 2, state: 1, panic): panicFunc" {
		t.Fatalf("Unexpected error: %v", err)
	}

	verifyTaskState(t, TaskStateDone, t1, t3)
	verifyTaskState(t, TaskStateFailed, t2)
	verifyTaskState(t, TaskStateNew, t4)
}

func checkDependency(t *testing.T, tasks ...Task) {
	for i := 1; i < len(tasks); i++ {
		if tasks[i].GetStartTime().Before(tasks[i-1].GetEndTime()) {
			t.Fatalf("%v started before %v finished", tasks[i], tasks[i-1])
		}
	}
}

func verifyTaskState(t *testing.T, state State, tasks ...Task) {
	for _, task := range tasks {
		if task.GetState() != state {
			t.Fatalf("%v wrong state! Expected: %v Got %v", task, state, task.GetState())
		}
	}
}

func sleepFunc(sleepTmo ...interface{}) (interface{}, error) {
	time.Sleep(sleepTmo[0].(time.Duration))
	return nil, nil
}

func panicFunc(sleepTmo ...interface{}) (interface{}, error) {
	time.Sleep(sleepTmo[0].(time.Duration))
	panic("panicFunc")
}
