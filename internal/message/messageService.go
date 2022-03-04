package message

import (
	"encoding/gob"
	"encoding/json"
	"fmt"
	"net"
	"node/api"
	"node/internal/accounts"
	"node/internal/peers"
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/google/uuid"
)

///////////////////////////// receive messages /////////////////////////////

func HandleConnection(conn net.Conn) {
	dec := gob.NewDecoder(conn)
	for {
		var message api.Message
		err := dec.Decode(&message)
		// println("ReceiveNewMessage: ", message.Identifier, " ", message.Sender, " ", message.Type)
		if err != nil {
			// fmt.Println("HandleConnection: ", err, conn.RemoteAddr().String())
			// some node disconnected
			peers.RemoveFromConnectedPeers(conn.RemoteAddr().String())
			FilterUnconnected(peers.GetNodeName(conn.RemoteAddr().String()))
			CheckUnconfirmed(peers.GetNodeName(conn.RemoteAddr().String()))
			return
		}
		msg_internal := Proto2Message(message)
		switch message.Type {
		case api.MSG_HELO:
			handleHello(message, conn)
		case api.MSG_DEPOSIT, api.MSG_TRANSFER:
			HandleMessage(msg_internal)
		case api.MSG_PROPOSE:
			if !checkIfReceived(message) {
				rMulticast(message)
			}
			HandleProposal(msg_internal)
		case api.MSG_AGREE:
			if !checkIfReceived(message) {
				rMulticast(message)
			}
			HandleAgreedSeq(msg_internal)
		default:
			fmt.Println("HandleConnection: Unknown message type: ", message.Type)
		}
	}
}

func handleHello(message api.Message, conn net.Conn) {
	peers.AddToConnectedPeers(message.Sender, conn)
}

func DeliverMsg(message *Message) {
	switch message.Type {
	case api.MSG_DEPOSIT:
		var deposit Deposit
		err := json.Unmarshal(message.Data, &deposit)
		if err != nil {
			panic(err)
		}
		accounts.DepositMoney(deposit.Account, deposit.Amount)
	case api.MSG_TRANSFER:
		var transfer Transfer
		err := json.Unmarshal(message.Data, &transfer)
		if err != nil {
			panic(err)
		}
		accounts.TransferMoney(transfer.From, transfer.To, transfer.Amount)
	default:
		fmt.Println("DeliverMsg: Unknown message type", message.Type)
	}
	accounts.ShowAccounts()
	logProcessingTime(*message)
}

func logProcessingTime(message Message) {
	time_now := time.Now().UnixNano() / int64(time.Microsecond)
	time_diff := time_now - message.TimeStamp
	file_name := "logs/" + peers.GetCurNode().Name + ".log"
	f, err := os.OpenFile(file_name, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		f, err = os.Create(file_name)
		if err != nil {
			panic(err)
		}
	}
	defer f.Close()
	f.WriteString(fmt.Sprintf("%s,%d\n", message.Identifier, time_diff))
}

///////////////////////////// send messages /////////////////////////////

func TMulticast(message Message) string {
	uid := uuid.New()
	msg_proto := Message2Proto(message)
	msg_proto.Identifier = uid.String()
	message.Identifier = uid.String()
	AfterSendMessage(message)
	putIntoReceived(msg_proto)
	sendMessageToPeers(msg_proto)
	return uid.String()
}

func SendProposal(destination string, message Message) {
	msg_proto := Message2Proto(message)
	sendMessage(msg_proto, destination)
}

func SendAgreedSeq(message Message) {
	msg_proto := Message2Proto(message)
	sendMessageToPeers(msg_proto)
}

func sendMessageToPeers(message api.Message) {
	for _, node := range peers.GetPeers() {
		if node == peers.GetCurNode().Name || message.Sender == node {
			continue
		}
		go sendMessage(message, node)
	}
}

func sendMessage(message api.Message, node_name string) {
	// println("sendMessageto: ", node_name, " ", message.Identifier, " ", message.Type)
	conn := peers.GetConnection(node_name)
	if conn == nil {
		fmt.Println("sendMessage: Connection to ", node_name, " is nil")
		return
	}
	enc := peers.GetEncoder(node_name)
	err := enc.Encode(message)
	if err != nil {
		fmt.Println(err)
		return
	}

}

///////////////////////////// R-Multicast /////////////////////////////
var recieved_messages sync.Map = sync.Map{} // received messages Identifer->true

func rMulticast(message api.Message) {
	sendMessageToPeers(message)
}

func checkIfReceived(message api.Message) bool {
	type_str := strconv.Itoa(message.Type)
	identifer := message.Identifier + message.Sender + type_str
	if _, ok := recieved_messages.Load(identifer); ok {
		return true
	}
	putIntoReceived(message)
	return false
}

func putIntoReceived(message api.Message) {
	type_str := strconv.Itoa(message.Type)
	identifer := message.Identifier + message.Sender + type_str
	recieved_messages.Store(identifer, true)
}
