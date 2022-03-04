package deposit

import (
	"bufio"
	"encoding/json"
	"node/internal/message"
	"node/internal/peers"
	"os"
	"strconv"
	"strings"
	"time"
)

const MaxInt = int(^uint(0) >> 1)

// read instractions from terminal
func ReadInstructions() {
	// println("Wait for 10 seconds...")
	// time.Sleep(10 * time.Second)
	input := bufio.NewScanner(os.Stdin)
	for input.Scan() {
		line := input.Text()
		tokens := strings.Fields(line)
		if tokens[0] == "DEPOSIT" {
			// println("depoist from:", tokens[1], "amount:", tokens[2])
			amount, _ := strconv.Atoi(tokens[2])
			deposit := message.Deposit{
				Account: tokens[1],
				Amount:  amount,
			}
			msg := deposit2Message(deposit)
			message.TMulticast(msg)
		}
		if tokens[0] == "TRANSFER" {
			// println("transfer from:", tokens[1], "to:", tokens[3], "amount:", tokens[4])
			amount, _ := strconv.Atoi(tokens[4])
			transfer := message.Transfer{
				From:   tokens[1],
				To:     tokens[3],
				Amount: amount,
			}
			msg := transfer2Message(transfer)
			message.TMulticast(msg)
		}
	}
}

func deposit2Message(msg message.Deposit) message.Message {
	tp := message.MSG_DEPOSIT
	data, err := json.Marshal(msg)
	if err != nil {
		panic(err)
	}
	node_name := peers.GetCurNode().Name
	message := message.Message{"", MaxInt, tp, node_name, data, false, time.Now().UnixNano() / int64(time.Microsecond)}
	return message
}

func transfer2Message(msg message.Transfer) message.Message {
	tp := message.MSG_TRANSFER
	data, err := json.Marshal(msg)
	if err != nil {
		panic(err)
	}
	node_name := peers.GetCurNode().Name
	message := message.Message{"", MaxInt, tp, node_name, data, false, time.Now().UnixNano() / int64(time.Microsecond)}
	return message
}
