// Copyright 2013 The Gorilla WebSocket Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"bytes"
	m "chat/common"
	"encoding/json"
	"fmt"
	"github.com/gorilla/websocket"
	"log"
	"net/http"
	"strconv"
	"time"
)

const (
	// Time allowed to write a message to the peer.
	writeWait = 10 * time.Second

	// Time allowed to read the next pong message from the peer.
	pongWait = 60 * time.Second

	// Send pings to peer with this period. Must be less than pongWait.
	pingPeriod = (pongWait * 9) / 10

	// Maximum message size allowed from peer.
	maxMessageSize = 512 * 12
)

var (
	newline = []byte{'\n'}
	space   = []byte{' '}
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

// Client is a middleman between the websocket connection and the hub.
type Client struct {
	hub *Hub

	// The websocket connection.
	conn *websocket.Conn

	// Buffered channel of outbound messages.
	send chan []byte

	id int64
}

// readPump pumps messages from the websocket connection to the hub.
//
// The application runs readPump in a per-connection goroutine. The application
// ensures that there is at most one reader on a connection by executing all
// reads from this goroutine.
func (c *Client) readPump() {

	defer func() {
		c.hub.unregister <- c

		c.conn.Close()
	}()
	c.conn.SetReadLimit(maxMessageSize)
	c.conn.SetReadDeadline(time.Now().Add(pongWait))
	c.conn.SetPongHandler(func(string) error { c.conn.SetReadDeadline(time.Now().Add(pongWait)); return nil })
	for {
		_, message, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("error: %v", err)
			}
			break
		}

		message2 := bytes.TrimSpace(bytes.Replace(message, newline, space, -1))
		println(message2)

		if c.id == -1 || c.id == 0 {
			var rmes m.RegistrationStruct
			if err2 := json.Unmarshal(message, &rmes); err2 == nil {
				//println("registration " + string(rmes.From))
				var id = rmes.From
				c.hub.mu.Lock()
				c.hub.clients[c] = id
				c.hub.iclients[id] = c
				c.id = id
				c.hub.mu.Unlock()
				//println("get registration data")
			} else {
				//println("parsing registration error")
			}
			continue
		}

		//var mes m.MessageStruct
		//if err := json.Unmarshal(message, &mes); err != nil {
		//	panic(err)
		//	//TODO continue
		//}
		//println(">>>" + string(mes.From) + " " + string(mes.To) + " " + mes.Text + "<<<<")

		message = bytes.TrimSpace(bytes.Replace(message, newline, space, -1))
		//c.hub.broadcast <- message //TODO remove
		c.hub.messageProcess <- message
	}
}

// writePump pumps messages from the hub to the websocket connection.
//
// A goroutine running writePump is started for each connection. The
// application ensures that there is at most one writer to a connection by
// executing all writes from this goroutine.
func (c *Client) writePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.conn.Close()
	}()
	for {
		select {
		case message, ok := <-c.send:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				// The hub closed the channel.
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			w, err := c.conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			w.Write(message)

			// Add queued chat messages to the current websocket message.
			n := len(c.send)
			for i := 0; i < n; i++ {
				w.Write(newline)
				w.Write(<-c.send)
			}

			if err := w.Close(); err != nil {
				return
			}
		case <-ticker.C:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

// serveWs handles websocket requests from the peer.
func serveWs(hub *Hub, w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return
	}
	client := &Client{hub: hub, conn: conn, send: make(chan []byte, 256)}
	client.hub.register <- client

	// Allow collection of memory referenced by the caller by doing all work in
	// new goroutines.
	go client.writePump()
	go client.readPump()
}

// chat history
func getLastMessagesFor(hub *Hub, w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	firstId := q.Get("first")
	secondId := q.Get("second")
	firstId_, err1 := strconv.ParseInt(firstId, 10, 64)
	if err1 != nil {
		respondErr(w, err1)
		return
	}
	secondId_, err2 := strconv.ParseInt(secondId, 10, 64)
	if err2 != nil {
		respondErr(w, err2)
		return
	}

	db := GetShardDbBody(firstId_, secondId_)

	query := "SELECT id, id_author, id_recipient, text, time FROM message " +
		" WHERE (id_author=" + firstId + " AND  id_recipient=" + secondId + ") " +
		"OR (id_recipient=" + secondId + " AND  id_author=" + firstId + ") " +
		" ORDER BY  time ASC"

	fmt.Println(query)
	rows, err := db.Query(query)

	if err != nil {
		fmt.Println("Query get chat error " + err.Error())
		respondErr(w, err)
		return
	}
	defer rows.Close()

	var out []m.Message
	//uu := make([](m.UserProfile), 0, first)
	for rows.Next() {
		var u m.Message
		//var avatar sql.NullString
		//var t time.Time
		dest := []interface{}{&u.Id, &u.From, &u.To, &u.Text, &u.Time}
		if err = rows.Scan(dest...); err != nil {
			//return nil, fmt.Errorf("could not scan user: %v", err)
			fmt.Println("error of scanning db response" + err.Error())
			respondErr(w, err)
			return
		}
		out = append(out, u)
	}
	respond(w, out, http.StatusOK)
}

func respond(w http.ResponseWriter, v interface{}, statusCode int) {
	b, err := json.Marshal(v)
	if err != nil {
		respondErr(w, fmt.Errorf("could not marshal response: %v", err))
	}
	w.Header().Set("Content-Type", "application/json: charset=utf-8")
	w.WriteHeader(statusCode)
	w.Write(b)
}

func respondErr(w http.ResponseWriter, err error) {
	log.Println(err)
	//http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
	http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
}
