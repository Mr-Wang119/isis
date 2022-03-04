package message

import (
	"encoding/json"
	"node/internal/peers"
	"sort"
	"sync"
)

// the largest aggreed sequence number
var seq_aggree int = 0

// the largest proposed sequence number
var seq_propse int = 0

// hold-back queue
var holdBackQueue []*Message = make([]*Message, 0)

// delivery queue
var deliveryQueue []*Message = make([]*Message, 0)

var lock = &sync.Mutex{}

// sequences of messages that need to be confirmed
type msgNeedToConfirm struct {
	confirmed_nodes map[string]bool
	largest_seq     int
}

var sequenceForUnconfirmed = make(map[string]*msgNeedToConfirm)

func AfterSendMessage(message Message) {
	msg_record := msgNeedToConfirm{make(map[string]bool), 0}
	lock.Lock()
	defer lock.Unlock()
	// println("add to unconfirmed:", message.Identifier)
	sequenceForUnconfirmed[message.Identifier] = &msg_record
	holdBackQueue = append(holdBackQueue, &message)
}

func HandleMessage(message Message) {
	// store the message in the hold-back queue
	seq_propse = max(seq_aggree, seq_propse) + 1
	message.Sequence = seq_propse
	lock.Lock()
	defer lock.Unlock()
	holdBackQueue = append(holdBackQueue, &message)

	// send the proposal to the node
	msg_proposal := Proposal2Message(message.Identifier, seq_propse)
	SendProposal(message.Sender, msg_proposal)
}

func HandleProposal(message Message) {
	lock.Lock()
	defer lock.Unlock()
	// update the largest agreed sequence number
	msg_record, ok := sequenceForUnconfirmed[message.Identifier]
	if !ok {
		// println("error: no record for message:", message.Identifier, message.Type)
		// panic("no such message")
		return
	}
	var msg_seq int
	json.Unmarshal(message.Data, &msg_seq)
	msg_record.confirmed_nodes[message.Sender] = true
	msg_record.largest_seq = max(msg_record.largest_seq, msg_seq)
	if len(msg_record.confirmed_nodes) >= peers.GetConnectedNum() {
		// delete the message from the sequenceForUnconfirmed
		delete(sequenceForUnconfirmed, message.Identifier)
		// send confirmed message to all nodes
		msg_confirmed := AgreedSeq2Message(message.Identifier, msg_record.largest_seq)
		SendAgreedSeq(msg_confirmed)
		// update the sequence of the message in the current node
		updateSeqForMsg(message.Identifier, msg_record.largest_seq)
	}
}

func HandleAgreedSeq(message Message) {
	// update the largest agreed sequence number
	var msg_seq int
	json.Unmarshal(message.Data, &msg_seq)
	seq_aggree = max(seq_aggree, msg_seq)
	// update the sequence of the message in the current node
	lock.Lock()
	defer lock.Unlock()
	updateSeqForMsg(message.Identifier, msg_seq)
}

func FilterUnconnected(node_name string) {
	lock.Lock()
	defer lock.Unlock()
	for i, msg := range holdBackQueue {
		if msg.Sender == node_name {
			holdBackQueue = holdBackQueue[:i+copy(holdBackQueue[i:], holdBackQueue[i+1:])]
		}
	}
}

func CheckUnconfirmed(node_name string) {
	lock.Lock()
	defer lock.Unlock()
	for msg_identifier, msg_record := range sequenceForUnconfirmed {
		if _, ok := msg_record.confirmed_nodes[node_name]; !ok {
			msg_record.confirmed_nodes[node_name] = true
			sequenceForUnconfirmed[msg_identifier] = msg_record
		}
		checkMsgRecord(msg_identifier)
	}
}

func checkMsgRecord(msg_identifier string) {
	msg_record, ok := sequenceForUnconfirmed[msg_identifier]
	if !ok {
		return
	}
	if len(msg_record.confirmed_nodes) >= peers.GetConnectedNum() {
		// delete the message from the sequenceForUnconfirmed
		delete(sequenceForUnconfirmed, msg_identifier)
		// send confirmed message to all nodes
		msg_confirmed := AgreedSeq2Message(msg_identifier, msg_record.largest_seq)
		SendAgreedSeq(msg_confirmed)
		// update the sequence of the message in the current node
		updateSeqForMsg(msg_identifier, msg_record.largest_seq)
	}

}

func updateSeqForMsg(identifier string, seq int) {
	changed := false
	for i, msg := range holdBackQueue {
		if msg.Identifier == identifier {
			// println("update message:", identifier, "to", seq)
			changed = true
			holdBackQueue[i].Sequence = seq
			holdBackQueue[i].Confirmed = true
			break
		}
	}
	if !changed {
		return
	}
	sort.Slice(holdBackQueue, func(i, j int) bool {
		return holdBackQueue[i].Sequence < holdBackQueue[j].Sequence
	})

	// println("------------------------")
	// println("hold-back queue: ", len(holdBackQueue))
	// for _, msg := range holdBackQueue {
	// 	println(msg.Identifier, " ", msg.Sequence, " ", msg.Confirmed)
	// }
	// println("------------------------")
	// if len(holdBackQueue) > 0 && !holdBackQueue[0].Confirmed {
	// 	println("stucked by:", holdBackQueue[0].Identifier, holdBackQueue[0].Sequence, holdBackQueue[0].Sender)
	// }
	for len(holdBackQueue) > 0 && holdBackQueue[0].Confirmed {
		deliveryQueue = append(deliveryQueue, holdBackQueue[0])
		holdBackQueue = holdBackQueue[1:]
		deliverMsg()
	}
}

func deliverMsg() {
	for len(deliveryQueue) > 0 {
		// println("deliver message:", deliveryQueue[0].Identifier)
		msg := deliveryQueue[0]
		deliveryQueue = deliveryQueue[1:]
		DeliverMsg(msg)
	}
}

func min(x, y int) int {
	if x < y {
		return x
	}
	return y
}

func max(x, y int) int {
	if x > y {
		return x
	}
	return y
}
