package main

import (
	"fmt"
	"net"
	"node/internal/deposit"
	"node/internal/message"
	"node/internal/peers"
	"os"
	"strconv"
)

func main() {
	// parse command line arguments
	if len(os.Args) != 3 {
		fmt.Println("Usage: ./mp1_node <node_name> <config_file>")
		return
	}
	node_name := os.Args[1]
	config_file := os.Args[2]

	// clear logs
	clearlogs()

	// parse config file
	peers.ParseConfigFile(config_file)

	// set current node
	peers.SetCurNode(node_name)

	// start node
	go startNode(peers.GetCurNode())

	// connect to peers

	peers.ConnectToPeersUntilSucc(message.HandleConnection)

	// start deposit and transfer
	deposit.ReadInstructions()
}

func startNode(currentNode peers.Node) {
	address := currentNode.Ip + ":" + strconv.Itoa(currentNode.Port)
	listener, err := net.Listen("tcp", address)
	if err != nil {
		fmt.Println(err)
		return
	}
	for {
		conn, err := listener.Accept()
		if err != nil {
			fmt.Println(err)
			continue
		}
		go message.HandleConnection(conn)
	}
}

func clearlogs() {
	err := os.RemoveAll("logs/")
	if err != nil {
		fmt.Println(err)
		return
	}
	os.MkdirAll("logs/", 0777)
}
