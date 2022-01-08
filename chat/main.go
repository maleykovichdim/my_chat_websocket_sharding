// Copyright 2013 The Gorilla WebSocket Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"database/sql"
	"flag"
	"github.com/joho/godotenv"
	"log"
	"net/http"
	"os"
	"time"
)

const (
	MAX_NUM_CONN  = 1000
	MAX_IDLE_CONN = 100
	MAX_LIFETIME  = 3
)

var (
	// databaseURL1 = env("JAWSDB_URL", "root:toor@tcp(shard1:3306)/socialnet?parseTime=true")
	// databaseURL2 = env("JAWSDB_URL", "root:toor@tcp(shard2:3306)/socialnet?parseTime=true")
	// databaseURL3 = env("JAWSDB_URL", "root:toor@tcp(shard3:3306)/socialnet?parseTime=true")
	// databaseURL4 = env("JAWSDB_URL", "root:toor@tcp(shard4:3306)/socialnet?parseTime=true")

	addr = flag.String("addr", ":8086", "http service address")
)

func serveHome(w http.ResponseWriter, r *http.Request) {
	log.Println(r.URL)
	if r.URL.Path != "/" {
		http.Error(w, "Not found", http.StatusNotFound)
		return
	}
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	http.ServeFile(w, r, "home.html")
}

func serveHomeJs(w http.ResponseWriter, r *http.Request) {
	log.Println(r.URL)
	if r.URL.Path != "BigInteger.min.js" {
		http.Error(w, "Not found", http.StatusNotFound)
		return
	}
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	http.ServeFile(w, r, "BigInteger.min.js")
}

func SetDB(pathDB string) *sql.DB {
	db, err1 := sql.Open("mysql", pathDB)
	if err1 != nil {
		log.Println(" ---------------------  ERROR database opening " + pathDB)
		panic(err1)
	}
	db.SetConnMaxLifetime(time.Minute * MAX_LIFETIME)
	db.SetMaxOpenConns(MAX_NUM_CONN)
	db.SetMaxIdleConns(MAX_IDLE_CONN)
	return db
}

func main() {
	godotenv.Load()
	flag.Parse()
	hub := newHub()

	//db part
	time.Sleep(2 * time.Second)

	go hub.run()
	http.HandleFunc("/", serveHome)
	http.HandleFunc("/js", serveHomeJs)
	http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		serveWs(hub, w, r)
	})

	http.HandleFunc("/chat", func(w http.ResponseWriter, r *http.Request) {
		getLastMessagesFor(hub, w, r)
	})

	err := http.ListenAndServe(*addr, nil)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}

func env(key, fallbackValue string) string {
	s, ok := os.LookupEnv(key)
	if !ok {
		return fallbackValue
	}
	return s
}
