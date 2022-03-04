package api

import (
	"encoding/json"
	"time"
)

type Message struct {
	Identifier string
	Type       int
	Sender     string
	Length     int
	Data       []byte
	TimeStamp  int64
}

// types of messages
const (
	MSG_PING = iota
	MSG_PONG
	MSG_HELO
	MSG_DEPOSIT
	MSG_TRANSFER
	MSG_PROPOSE
	MSG_AGREE
)

type Deposit struct {
	Account string `json:"account"`
	Amount  int    `json:"amount"`
}

type Transfer struct {
	From   string `json:"from"`
	To     string `json:"to"`
	Amount int    `json:"amount"`
}

func Helo2Message(msg string) Message {
	tp := MSG_HELO
	data, err := json.Marshal(msg)
	if err != nil {
		panic(err)
	}
	length := len(data)
	message := Message{"", tp, msg, length, data, time.Now().UnixNano() / int64(time.Microsecond)}
	return message
}
