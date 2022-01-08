package common

import "time"

type MessageStruct struct {
	To   int64  `json:"to"`
	From int64  `json:"from"`
	Text string `json:"text"`
}

type MessageStruct_old struct {
	To   int64  `json:"to"`
	Text string `json:"text"`
}

type RegistrationStruct struct {
	From int64 `json:"from"`
}

type Message struct {
	Id   int64     `json:"id"`
	From int64     `json:"from"`
	To   int64     `json:"to"`
	Text string    `json:"text"`
	Time time.Time `json:"time"`
}
