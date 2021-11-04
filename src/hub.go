// Copyright 2013 The Gorilla WebSocket Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"encoding/json"
	"log"
)

// Hub maintains the set of active clients and broadcasts messages to the
// clients.
type Hub struct {
	task *Task

	// Registered clients.
	clients map[*Client]bool

	// Inbound messages from the clients.
	broadcast chan []byte

	// Register requests from the clients.
	register chan *Client

	// Unregister requests from clients.
	unregister chan *Client
}

type socialUpdate struct {
	MsgType string `json:"MsgType"`

	PlayerCnt int `json:"PlayerCnt"`
}

func newHub() *Hub {
	h := &Hub{
		broadcast:  make(chan []byte),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		clients:    make(map[*Client]bool),
	}
	h.task = newTask(h)

	return h
}

func (h *Hub) updatePlayerCnt() {
	su := socialUpdate{
		MsgType:   "SocialUpdate",
		PlayerCnt: len(h.clients),
	}

	bytes, err := json.Marshal(su)
	if err != nil {
		log.Printf("Error marshalling socialUpdate %v: %v", su, err)
	}

	h.broadcast <- bytes
}

func (h *Hub) run() {
	for {
		select {
		case client := <-h.register:
			h.clients[client] = true
			h.task.start <- client

			go h.updatePlayerCnt()
		case client := <-h.unregister:
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				close(client.send)
			}

			go h.updatePlayerCnt()
		case message := <-h.broadcast:
			for client := range h.clients {
				client.send <- message
			}
		}
	}
}
