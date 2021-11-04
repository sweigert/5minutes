package main

import (
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"strconv"
	"time"
)

type taskItem struct {
	MsgType string `json:"MsgType"`

	Tid        int    `json:"tid"`
	TaskStr    string `json:"TaskStr"`
	CorrectCnt int    `json:"CorrectCnt"`
	FailCnt    int    `json:"FailCnt"`

	a int
	b int
}

type Task struct {
	hub      *Hub
	start    chan *Client
	taskList []taskItem
}

func (t *Task) waitForClients() {
	for {
		client := <-t.start
		go t.serveTask(client)
	}
}

func (t *Task) serveTask(client *Client) {
	tid := 0
	correctCnt := 0
	failCnt := 0

	for _, nextTask := range t.taskList {
		nextTask.CorrectCnt = correctCnt
		nextTask.FailCnt = failCnt
		bytes, err := json.Marshal(nextTask)
		if err != nil {
			log.Printf("Error marshalling task %v: %v", nextTask, err)
			continue
		}

		client.send <- bytes

		select {
		case msg := <-client.receive:
			res, err := strconv.ParseInt(string(msg), 10, 64)
			if err != nil {
				log.Printf("Error reading answer from client %v: %v", client, err)
			}
			if res == int64(nextTask.a+nextTask.b) {
				correctCnt += 1
			} else {
				failCnt += 1
			}

			tid += 1
		case <-time.After(60 * time.Second):
			log.Printf("timeout after 60s for client %v", client)
			client.done <- true
			return
		}
	}
}

func newTask(hub *Hub) *Task {
	task := &Task{
		hub:      hub,
		start:    make(chan *Client),
		taskList: make([]taskItem, 100),
	}

	for i := 0; i < 100; i++ {
		a := rand.Intn(10)
		b := rand.Intn(10)

		task.taskList[i] = taskItem{
			MsgType:    "Task",
			Tid:        i,
			TaskStr:    fmt.Sprintf("%d + %d = ", a, b),
			CorrectCnt: 0,
			FailCnt:    0,

			a: a,
			b: b,
		}
	}

	go task.waitForClients()

	return task
}
