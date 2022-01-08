// Copyright 2013 The Gorilla WebSocket Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	m "chat/common"
	"database/sql"
	"encoding/json"
	"fmt"
	"strconv"
	"sync"

	_ "github.com/go-sql-driver/mysql"
)

var (
	ring *m.Ring
)

// Hub maintains the set of active clients and broadcasts messages to the
// clients.
type Hub struct {
	// Registered clients.
	clients map[*Client]int64
	mu      sync.Mutex

	// Registered clients.
	iclients map[int64]*Client

	// Inbound messages from the clients.
	broadcast chan []byte

	// Register requests from the clients.
	register chan *Client

	// Unregister requests from clients.
	unregister chan *Client

	// Inbound messages from the clients.
	messageProcess chan []byte
}

func newHub() *Hub {
	return &Hub{
		broadcast:      make(chan []byte),
		messageProcess: make(chan []byte),
		register:       make(chan *Client),
		unregister:     make(chan *Client),
		clients:        make(map[*Client]int64),
		iclients:       make(map[int64]*Client),
	}
}

func preRun() {
	ring = m.NewRing()
	m.InitDB()
	ring.AddNode("shard1")
	ring.AddNode("shard2")
	ring.AddNode("shard3")
	ring.AddNode("shard4")
}

//var sh1 = 1
//var sh2 = 1
//var sh3 = 1
//var sh4 = 1

func GetShardDbBody(from int64, to int64) *sql.DB {
	var idChat string
	if from < to {
		idChat = strconv.FormatInt(from, 10) + "_" + strconv.FormatInt(to, 10)
	} else {
		idChat = strconv.FormatInt(to, 10) + "_" + strconv.FormatInt(from, 10)
	}

	shard := ring.Get(idChat)
	//println(shard)
	//println("shard1: ", sh1, "shard2:  ", sh2, "shard3: ", sh3, "shard4: ", sh4, " ")
	switch shard {
	case "shard1":
		//sh1++
		return m.Db1
	case "shard2":
		//sh2++
		return m.Db2
	case "shard3":
		//sh3++
		return m.Db3
	case "shard4":
		//sh4++
		return m.Db4
	default:
		fmt.Println("case error")
		return m.Db1
	}

}

func GetShardDb(mes *m.MessageStruct) *sql.DB {
	return GetShardDbBody(mes.From, mes.To)
}

func (h *Hub) run() {

	preRun()

	defer m.Db1.Close()
	defer m.Db2.Close()
	defer m.Db3.Close()
	defer m.Db4.Close()

	for {
		select {
		case client := <-h.register:
			h.clients[client] = -1
			//println("register")
		case client := <-h.unregister:
			v, ok := h.clients[client]
			if ok {
				h.mu.Lock()
				delete(h.clients, client)
				if v > 0 {
					delete(h.iclients, v)
				}

				close(client.send)
				h.mu.Unlock()
			}

		case message := <-h.broadcast:
			for client := range h.clients {
				select {
				case client.send <- message:
				default:
					close(client.send)
					delete(h.clients, client)
				}
			}
		case message := <-h.messageProcess:
			var mes m.MessageStruct
			if err := json.Unmarshal(message, &mes); err != nil {
				println("ERROR: " + string(message))
				continue
			}

			db := GetShardDb(&mes)

			query := "INSERT INTO message (id_author, id_recipient, text) VALUES (?, ?, ?)"
			_, err := db.Exec(query, mes.From, mes.To, mes.Text)
			if err != nil {
				fmt.Println("DB Error: " + query)
				continue
			}

			println(">>> from: " + strconv.FormatInt(mes.From, 10) + " To:" + strconv.FormatInt(mes.To, 10) + " Text: " + mes.Text + "<<<<")
			recipient, ok := h.iclients[mes.To]
			if !ok {
				println("recipient inactive: " + strconv.FormatInt(mes.To, 10))
				continue
			}
			recipient.send <- message
		}

	}
}
