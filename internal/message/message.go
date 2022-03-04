package message

import (
	"encoding/json"
	"node/api"
	"node/internal/peers"
)

type Message struct {
	Identifier string
	Sequence   int
	Type       int
	Sender     string
	Data       []byte
	Confirmed  bool
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

func Proposal2Message(identifier string, seq int) Message {
	tp := MSG_PROPOSE
	data, err := json.Marshal(seq)
	if err != nil {
		panic(err)
	}
	node_name := peers.GetCurNode().Name
	message := Message{identifier, seq, tp, node_name, data, false, 0}
	return message
}

func AgreedSeq2Message(identifier string, seq int) Message {
	tp := MSG_AGREE
	data, err := json.Marshal(seq)
	if err != nil {
		panic(err)
	}
	node_name := peers.GetCurNode().Name
	message := Message{identifier, seq, tp, node_name, data, true, 0}
	return message
}

func Proto2Message(proto api.Message) Message {
	message := Message{
		Identifier: proto.Identifier,
		Sequence:   0,
		Type:       proto.Type,
		Sender:     proto.Sender,
		Data:       proto.Data,
		Confirmed:  false,
		TimeStamp:  proto.TimeStamp,
	}
	return message
}

func Message2Proto(message Message) api.Message {
	proto := api.Message{
		Identifier: message.Identifier,
		Type:       message.Type,
		Sender:     message.Sender,
		Data:       message.Data,
		Length:     len(message.Data),
		TimeStamp:  message.TimeStamp,
	}
	return proto
}
