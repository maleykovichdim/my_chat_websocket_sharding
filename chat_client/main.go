package main

import (
	"bytes"
	"encoding/json"
	"log"
	"math/rand"
	"os"
	"os/signal"
	"strconv"
	"time"

	"github.com/sacOO7/gowebsocket"
)

const NumberSockets = 1000

var sockets [NumberSockets]gowebsocket.Socket
var isSocketOpen [NumberSockets]bool
var rand_ rand.Source

type MessageStruct struct {
	To   int64  `json:"to"`
	From int64  `json:"from"`
	Text string `json:"text"`
}

type RegistrationStruct struct {
	From int64 `json:"from"`
}

func forSocket(socket *gowebsocket.Socket) {
	socket.OnConnectError = func(err error, socket gowebsocket.Socket) {
		log.Fatal("Received connect error - ", err)
	}

	socket.OnConnected = func(socket gowebsocket.Socket) {
		log.Println("Connected to server")
	}

	socket.OnTextMessage = func(message string, socket gowebsocket.Socket) {
		log.Println("Received message - " + message)
	}

	socket.OnPingReceived = func(data string, socket gowebsocket.Socket) {
		log.Println("Received ping - " + data)
	}

	socket.OnPongReceived = func(data string, socket gowebsocket.Socket) {
		log.Println("Received pong - " + data)
	}

	socket.OnDisconnected = func(err error, socket gowebsocket.Socket) {
		log.Println("Disconnected from server ")
		return
	}
}

func createSocket(id int64, socket *gowebsocket.Socket) error {
	*socket = gowebsocket.New("ws://localhost:8086/ws")

	forSocket(socket)
	socket.Connect()
	reg := RegistrationStruct{From: id}

	reqBodyBytes := new(bytes.Buffer)
	json.NewEncoder(reqBodyBytes).Encode(reg)
	//println("first")
	socket.SendBinary(reqBodyBytes.Bytes())
	//println("second")
	return nil
}

func sendMes(message string, recipient int64, author int64, socket *gowebsocket.Socket) error {
	mes := MessageStruct{From: author, To: recipient, Text: message}
	reqBodyBytes := new(bytes.Buffer)
	json.NewEncoder(reqBodyBytes).Encode(mes)
	socket.SendBinary(reqBodyBytes.Bytes())
	//println("third")
	return nil
}

var letterRunes = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

func RandStringRunes(n int) string {
	b := make([]rune, n)
	for i := range b {
		b[i] = letterRunes[rand.Intn(len(letterRunes))]
	}
	return string(b)
}

func main() {
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)

	for j := 0; j < NumberSockets; j++ {
		isSocketOpen[j] = false
	}

	rand_ = rand.NewSource(time.Now().UnixNano())

	r1 := rand.New(rand_)
	for j := 0; j <= 100000; j++ {

		auth := r1.Intn(NumberSockets-2) + 1
		recip := r1.Intn(NumberSockets-2) + 1
		if auth != recip {

			if (!isSocketOpen[auth]) || (!sockets[auth].IsConnected) {
				println("try to create connection")
				e := createSocket(int64(auth), &sockets[auth])
				if e != nil {
					isSocketOpen[auth] = false
					continue
				} else {
					isSocketOpen[auth] = true
				}
			}

			text := RandStringRunes(12)
			e := sendMes(text, int64(recip), int64(auth), &sockets[auth])
			if e != nil {
				println("error message sending")
			} else {
				println("try to send connection")
				info := "Send: " + text + " from " + strconv.FormatInt(int64(auth), 10) +
					"  to: " + strconv.FormatInt(int64(recip), 10)
				println(info)
			}

		}
	}

	println("Sleeping..... ")
	time.Sleep(60 * time.Second)

	for j := 0; j < NumberSockets; j++ {
		if sockets[j].IsConnected {
			sockets[j].Close()
		}
	}

	//for {
	//	select {
	//	case <-interrupt:
	//		log.Println("interrupt")
	//		sockets[0].Close()
	//		return
	//	}
	//}
}
