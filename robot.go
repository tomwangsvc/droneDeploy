package main

import (
	"fmt"
	"log"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
)

type Robot interface {
	EnqueueTask(commands string) (taskID string, state chan RobotState, err chan error)

	CancelTask(taskID string) error

	CurrentState() RobotState
}

type RobotState struct {
	X int
	Y int
}

func (r RobotState) Valid() bool {
	return r.X <= 10 && r.X >= 0 && r.Y <= 10 && r.Y >= 0
}

type robot struct {
	mu     sync.Mutex
	state  RobotState
	err    error
	tasks  []task
	done   chan struct{} // signal for every single command proccessed
	cancel chan struct{} // signal to cancel current task (i.e first element in robot.tasks)
}

type task struct {
	id       string
	commands []string
}

func (r *robot) removeFirstTask() {
	if len(r.tasks) == 0 {
		return
	}
	r.mu.Lock()
	r.tasks = r.tasks[1:]
	r.mu.Unlock()
}

func NewRobot() *robot {
	r := &robot{
		done:   make(chan struct{}),
		cancel: make(chan struct{}),
	}
	go r.runTasks()
	return r
}

func (r *robot) runTasks() {
	go func() {
		<-r.cancel
		// remove the running task (i.e first one) when cancelled
		r.removeFirstTask()
	}()

	for {
		if len(r.tasks) == 0 {
			time.Sleep(time.Second) // wait for one second and then check new tasks
			continue
		}

		runningTask := r.tasks[0]
		for i, command := range runningTask.commands {
			if len(r.tasks) == 0 || r.tasks[0].id != runningTask.id {
				log.Println("running task has been changed and it is possibly cancelled")
				break
			}

			if err := r.runSingleCommand(i, command, i+1 == len(runningTask.commands)); err != nil {
				r.err = err
			}
		}
	}
}

func (r *robot) runSingleCommand(i int, command string, taskComplete bool) (err error) {
	defer func() {
		r.done <- struct{}{}
		if taskComplete {
			// remove the first one task when complete
			r.removeFirstTask()
		}
	}()

	i++
	taskId := r.tasks[0].id

	switch command {
	case commandNorthwards:
		// perform some task that can consume time and move northwards
		log.Printf(fmt.Sprintf("%s: %d.command start to move northwards", taskId, i))
		if err = move(
			func() {
				r.state.Y++
			},
			r.state,
			func() {
				r.state.Y--
			}); err != nil {
			return
		}
		log.Printf("%s: %d.command finish to move northwards", taskId, i)

	case commandEastwards:
		// perform some task that can consume time and move eastwards
		log.Printf("%s: %d.command start to move eastwards", taskId, i)

		if err = move(
			func() {
				r.state.X++
			},
			r.state,
			func() {
				r.state.X--
			}); err != nil {
			return
		}
		log.Printf("%s: %d.command finish to move eastwards", taskId, i)

	case commandSouthwards:
		// perform some task that can consume time and move southwards
		log.Printf("%s: %d.command start to move southwards", taskId, i)
		if err = move(
			func() {
				r.state.Y--
			},
			r.state,
			func() {
				r.state.Y++
			}); err != nil {
			return
		}
		log.Printf("%s: %d.command finish to move southwards", taskId, i)

	case commandWestwards:
		// perform some task that can consume time and move westwards
		log.Printf("%s: %d.command start to move westwards", taskId, i)
		if err = move(
			func() {
				r.state.X--
			},
			r.state,
			func() {
				r.state.X++
			}); err != nil {
			return
		}
		log.Printf("%s: %d.command finish to move westwards", taskId, i)

	default:
		err = fmt.Errorf("command not recognized: %s", command)
		return
	}

	return
}

func move(act func(), state RobotState, rollBack func()) error {
	act()

	if !state.Valid() {
		// roll back to previous state when move out of boundary
		rollBack()
		return fmt.Errorf("robot move out of boundary with state %+v", state)
	}

	// assume action will take one sec
	time.Sleep(1 * time.Second)
	
	return nil
}

func (r *robot) EnqueueTask(commands string) (taskID string, chanState chan RobotState, chanError chan error) {
	log.Println("enqueue task", commands)

	taskID = uuid.New().String()
	commandsArr := strings.Split(commands, " ")
	if len(commandsArr) != 0 {
		r.mu.Lock()
		r.tasks = append(r.tasks, task{taskID, commandsArr})
		r.mu.Unlock()
	}

	chanState = make(chan RobotState)
	chanError = make(chan error)

	go func() {
		for {
			<-r.done
			chanState <- r.state
			chanError <- r.err
		}
	}()

	return
}

func (r *robot) CurrentState() RobotState {
	return r.state
}

func (r *robot) CancelTask(taskID string) error {
	log.Printf("cancelling task %s", taskID)
	found := false
	for i, task := range r.tasks {
		if task.id == taskID {
			// check if task is running
			if i == 0 {
				r.cancel <- struct{}{}
			}
			found = true
			r.tasks = append(r.tasks[:i], r.tasks[i+1:]...)
		}
	}

	if !found {
		return fmt.Errorf("task with id %s not found", taskID)
	}

	log.Printf("cancelled task %s", taskID)
	return nil
}
