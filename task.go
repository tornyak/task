package tasks
import (
	"math/rand"
	"time"
	"fmt"
)

type Result string

type Task interface {
	GetId() uint64
	Run() Result
	GetNextTask() []*Task
	GetDependsOn() []*Task
	AddNextTask(*Task)
}

type DefaultTask struct {
	id uint64
	nextTask []*Task
}

func (dt *DefaultTask) GetId() uint64 {
	return dt.id
}

func (dt *DefaultTask) Run() Result {
	time.Sleep(time.Duration(rand.Intn(1e3)) * time.Millisecond)
	return fmt.Sprintf("Task %v done!", dt.id)
}

func (dt *DefaultTask) AddNextTask(t *Task) {
	dt.nextTask = append(dt.nextTask, t)
}

func (dt *DefaultTask) GetNextTask() []*Task {
	return dt.nextTask
}



type DefaultTaskFactory struct {
	nextID Counter
}



func (tf *DefaultTaskFactory) NewTask() *DefaultTask {
	return &DefaultTask{
		id: tf.nextID.GetAndInc(),
	}
}

// TaskManager must know when task is done
// Then it will try to start next task if possible
// if not possible it waits for next task to end
func main() {
	dtf := DefaultTaskFactory{}
	dummy := dtf.NewTask()
	t1 := dtf.NewTask()
	t2 := dtf.NewTask()
	t3 := dtf.NewTask()
	t4 := dtf.NewTask()
	t5 := dtf.NewTask()
	t6 := dtf.NewTask()

	dummy.AddNextTask(t1)
	dummy.AddNextTask(t2)
	dummy.AddNextTask(t3)
	t1.AddNextTask(t4)
	t2.AddNextTask(t4)
	t2.AddNextTask(t5)
	t3.AddNextTask(t6)




}