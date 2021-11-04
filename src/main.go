// Copyright 2013 The Gorilla WebSocket Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"flag"
	"log"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/gorilla/sessions"
)

var addr = flag.String("addr", ":8080", "http service address")

var (
	// key must be 16, 24 or 32 bytes long (AES-128, AES-192 or AES-256)
	key   = []byte("super-secret-key")
	store = sessions.NewCookieStore(key)
)

func serveLogin(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "html/login.html")
}

func serveHome(w http.ResponseWriter, r *http.Request) {
	session, err := store.Get(r, "session")
	if err != nil {
		log.Printf("Error reading cookie: %v", err)
	}

	// Check if user is authenticated
	userName, ok := session.Values["name"].(string)
	if !ok {
		log.Printf("Error reading username from cookie: %v", session.Values["name"])
		return
	}

	log.Printf("Username: %v", userName)

	log.Println(r.URL)
	if r.URL.Path != "/internal" {
		http.Error(w, "Not found", http.StatusNotFound)
		return
	}
	if r.Method != "GET" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	http.ServeFile(w, r, "html/home.html")
}

func login(w http.ResponseWriter, r *http.Request) {
	name := r.FormValue("login")
	log.Printf("Username: %v", name)
	session, err := store.Get(r, "session")
	if err != nil {
		log.Printf("Error reading cookie: %v", err)
	}

	session.Values["name"] = name
	session.Save(r, w)

	http.Redirect(w, r, "/internal", http.StatusTemporaryRedirect)
}

func main() {
	flag.Parse()
	hub := newHub()
	go hub.run()

	router := mux.NewRouter()
	router.HandleFunc("/login", login)
	router.HandleFunc("/", serveLogin)
	router.HandleFunc("/internal", serveHome)
	router.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		serveWs(hub, w, r)
	})

	http.Handle("/", router)
	err := http.ListenAndServe(*addr, nil)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}
