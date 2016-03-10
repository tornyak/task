package main
import (
	"math/rand"
	"time"
	"fmt"
	"github.com/tornyak/goalg/queue"
	"tasks/counter"

)

type Result struct {
	ID uint64
	Status string
}

type Task interface {
	GetId() uint64
	Run(c chan<- Result)
	GetNextTask() []Task
	GetDependsOn() []uint64
	AddNextTask(Task)
	AddDependsOn(uint64)
	NotifyDone(uint64)

}

type DefaultTask struct {
	id uint64
	nextTask []Task
	dependsOn []uint64
}

func (dt *DefaultTask) GetId() uint64 {
	return dt.id
}

func (dt *DefaultTask) Run(resultChan chan<- Result) {
	time.Sleep(time.Duration(rand.Intn(1e3)) * time.Millisecond)
	resultChan <- Result{dt.id, fmt.Sprintf("Task %v done!", dt.id)}
}

func (dt *DefaultTask) AddNextTask(t Task) {
	dt.nextTask = append(dt.nextTask, t)
}

func (dt *DefaultTask) AddDependsOn(taskId uint64) {
	dt.dependsOn = append(dt.dependsOn, taskId)
}

func (dt *DefaultTask) GetDependsOn()[]uint64 {
	return dt.dependsOn
}

func (dt *DefaultTask) NotifyDone(taskId uint64) {
	for i, t := range dt.dependsOn {
		if t == taskId {
			dt.dependsOn = append(dt.dependsOn[:i], dt.dependsOn[i+1:]...)
		}
	}
}

func (dt *DefaultTask) GetNextTask() []Task {
	return dt.nextTask
}



type DefaultTaskFactory struct {
	nextID counter.Counter
}



func (tf *DefaultTaskFactory) NewTask() Task {
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

	t1.AddDependsOn(dummy.GetId())
	t2.AddDependsOn(dummy.GetId())
	t3.AddDependsOn(dummy.GetId())
	t4.AddDependsOn(t1.GetId())
	t4.AddDependsOn(t2.GetId())
	t5.AddDependsOn(t2.GetId())
	t6.AddDependsOn(t3.GetId())

	// For indexing tasks by ID
	taskMap := map[uint64]Task{
		dummy.GetId(): dummy,
		t1.GetId(): t1,
		t2.GetId(): t2,
		t3.GetId(): t3,
		t4.GetId(): t4,
		t5.GetId(): t5,
		t6.GetId(): t6,
	}

	// Tasks that are ready to be executed
	taskQueue := queue.NewLinkedQueue()

	fmt.Printf("Enqueue %v\n", dummy.GetId())
	taskQueue.Enqueue(dummy)
	resultChan := make(chan Result)

	cnt := 7

	for cnt > 0 {
		for !taskQueue.IsEmpty() {
			t := taskQueue.Dequeue().(Task)
			fmt.Printf("Run %v\n", t.GetId())
			go t.Run(resultChan)
		}
		select {
		case c := <- resultChan:
			cnt--
			fmt.Println(c.Status)
			finishedTask := taskMap[c.ID]
			nextTasks := finishedTask.GetNextTask()
			var nt Task
			for _, nt = range nextTasks {
				nt.NotifyDone(c.ID)
				fmt.Printf("NotifyDone %v\n", nt.GetId())
				fmt.Printf("%v DependsOn %v\n", nt.GetId(), nt.GetDependsOn())
				if len(nt.GetDependsOn()) == 0 {
					fmt.Printf("Enqueue %v\n", nt.GetId())
					taskQueue.Enqueue(nt)
				}
			}
		}
	}
}