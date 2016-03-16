package task
import (
	"testing"
	"time"
)

func TestIndependentOk(t *testing.T){
	tmo := 10 * time.Millisecond

	r := NewRunner(tmo)
	dtf := DefaultTaskFactory{}
	t1 := dtf.NewTask(sleepFunc, 1 * time.Millisecond, TaskFailedAbort)
	t2 := dtf.NewTask(sleepFunc, 1 * time.Millisecond, TaskFailedAbort)
	t3 := dtf.NewTask(sleepFunc, 1 * time.Millisecond, TaskFailedAbort)
	r.Add(t1, t2, t3)

	if err := r.Start(); err != nil {
		t.Fatal("Task execution error")
	}
}

func TestTmo(t *testing.T) {
	tmo := 1 * time.Millisecond

	r := NewRunner(tmo)
	dtf := DefaultTaskFactory{}
	t1 := dtf.NewTask(sleepFunc, 3 * time.Millisecond, TaskFailedAbort)
	t2 := dtf.NewTask(sleepFunc, 1 * time.Millisecond, TaskFailedAbort)
	t3 := dtf.NewTask(sleepFunc, 1 * time.Millisecond, TaskFailedAbort)
	r.Add(t1, t2, t3)

	if err := r.Start(); err != nil {
		switch err {
		case ErrTimeout:
			t.Log("Timeout as expected")
			return
		default:
			t.Fatal("Unexpected error")
		}

	}
	t.Fatal("No error returned")
}

// Dependency t1->t2->t3
func TestSimpleDependentOk(t *testing.T){
	tmo := 10 * time.Millisecond

	r := NewRunner(tmo)
	dtf := DefaultTaskFactory{}
	t1 := dtf.NewTask(sleepFunc, 1 * time.Millisecond, TaskFailedAbort)
	t2 := dtf.NewTask(sleepFunc, 1 * time.Millisecond, TaskFailedAbort)
	t3 := dtf.NewTask(sleepFunc, 1 * time.Millisecond, TaskFailedAbort)
	r.Add(t1, t2, t3)

	r.AddDependency(t1,t2)
	r.AddDependency(t2,t3)

	if err := r.Start(); err != nil {
		t.Fatal("Task execution error")
	}

	checkDependency(t1, t2, t)
	checkDependency(t2, t3, t)
}


func TestComplexDependentOk(t *testing.T){
	tmo := 10 * time.Millisecond

	r := NewRunner(tmo)
	dtf := DefaultTaskFactory{}
	t1 := dtf.NewTask(sleepFunc, 1 * time.Millisecond, TaskFailedAbort)
	t2 := dtf.NewTask(sleepFunc, 1 * time.Millisecond, TaskFailedAbort)
	t3 := dtf.NewTask(sleepFunc, 1 * time.Millisecond, TaskFailedAbort)
	t4 := dtf.NewTask(sleepFunc, 1 * time.Millisecond, TaskFailedAbort)
	t5 := dtf.NewTask(sleepFunc, 1 * time.Millisecond, TaskFailedAbort)
	t6 := dtf.NewTask(sleepFunc, 1 * time.Millisecond, TaskFailedAbort)
	t7 := dtf.NewTask(sleepFunc, 1 * time.Millisecond, TaskFailedAbort)
	t8 := dtf.NewTask(sleepFunc, 1 * time.Millisecond, TaskFailedAbort)
	t9 := dtf.NewTask(sleepFunc, 1 * time.Millisecond, TaskFailedAbort)
	r.Add(t1, t2, t3, t4, t5, t6, t7, t8, t9)

	r.AddDependency(t1,t4)
	r.AddDependency(t2,t4)
	r.AddDependency(t3,t5)
	r.AddDependency(t1,t6)
	r.AddDependency(t4,t6)
	r.AddDependency(t5,t6)
	r.AddDependency(t5,t7)
	r.AddDependency(t6,t8)
	r.AddDependency(t6,t9)
	r.AddDependency(t7,t9)

	if err := r.Start(); err != nil {
		t.Fatal("Task execution error")
	}

	checkDependency(t1,t4,t)
	checkDependency(t2,t4,t)
	checkDependency(t3,t5,t)
	checkDependency(t1,t6,t)
	checkDependency(t4,t6,t)
	checkDependency(t5,t6,t)
	checkDependency(t5,t7,t)
	checkDependency(t6,t8,t)
	checkDependency(t6,t9,t)
	checkDependency(t7,t9,t)
}

func checkDependency(t1, t2 Task, t *testing.T) {
	if t2.GetStartTime().Before(t1.GetEndTime()) {
		t.Fatalf("%v started before %v finished", t2, t1)
	}
}

func sleepFunc(sleepTmo interface{})(interface{}, error) {
	time.Sleep(sleepTmo.(time.Duration))
	return nil, nil
}
