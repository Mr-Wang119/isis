package peers

import (
	"bufio"
	"encoding/gob"
	"fmt"
	"net"
	"node/api"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
)

type Node struct {
	Name string
	Ip   string
	Port int
}

var peers map[string]Node = make(map[string]Node)      // the information of peers
var connected_peers_encoder sync.Map = sync.Map{}      // connected peers encoder
var connected_peers_connection sync.Map = sync.Map{}   // connected peers connection
var connected_peers_address2name sync.Map = sync.Map{} // address to names
var num_connection int = 0                             // num of connected peers
var currentNode Node = Node{"", "", 0}                 // the information of the current node

///////////////////////////// parse config file /////////////////////////////

func ParseConfigFile(config_file string) {
	file, err := os.Open(config_file)
	if err != nil {
		fmt.Println("Error opening config file: ", err)
		os.Exit(1)
	}
	defer file.Close()
	scanner := bufio.NewScanner(file)
	scanner.Scan() // skip the first line
	for scanner.Scan() {
		line := scanner.Text()
		if line == "" {
			continue
		}
		tokens := strings.Split(line, " ")
		if len(tokens) != 3 {
			fmt.Println("Error parsing config file: ", line)
			os.Exit(1)
		}
		host, _ := strconv.Atoi(tokens[2])
		peer := Node{tokens[0], tokens[1], host}
		peers[peer.Name] = peer
	}
}

///////////////////////////// API /////////////////////////////

func GetCurNode() Node {
	return currentNode
}

func SetCurNode(node_name string) {
	currentNode = peers[node_name]
}

func AddToConnectedPeers(name string, conn net.Conn) {
	connected_peers_connection.Store(name, conn)
	connected_peers_encoder.Store(name, gob.NewEncoder(conn))
	connected_peers_address2name.Store(conn.RemoteAddr().String(), name)
	num_connection++
	fmt.Println("Connected to ", name)
}

func RemoveFromConnectedPeers(address string) {
	name, ok := connected_peers_address2name.Load(address)
	if !ok {
		println("Error: cannot find name of ", address)
		return
	}
	peer, ok := connected_peers_connection.Load(name)
	if !ok {
		println("Error: cannot find connection of ", name)
		return
	}
	conn := peer.(net.Conn)
	connected_peers_connection.Delete(name)
	connected_peers_encoder.Delete(name)
	connected_peers_address2name.Delete(address)
	num_connection--
	println("Disconnected from ", name.(string))
	conn.Close()
	logDisconnections(name.(string))
}

func GetConnection(name string) net.Conn {
	conn, ok := connected_peers_connection.Load(name)
	if !ok {
		return nil
	}
	return conn.(net.Conn)
}

func GetEncoder(name string) *gob.Encoder {
	encoder, ok := connected_peers_encoder.Load(name)
	if !ok {
		return nil
	}
	return encoder.(*gob.Encoder)
}

func GetPeers() []string {
	var result []string
	for _, peer := range peers {
		_, ok := connected_peers_connection.Load(peer.Name)
		if ok {
			result = append(result, peer.Name)
		}
	}
	return result
}

func GetConnectedNum() int {
	return num_connection
}

func GetNodeName(address string) string {
	name, ok := connected_peers_address2name.Load(address)
	if !ok {
		return ""
	}
	return name.(string)
}

///////////////////////////// Init Peer Connections /////////////////////////////

func ConnectToPeersUntilSucc(callback func(net.Conn)) {
	for {
		var all_connected bool = true
		for name, node := range peers {
			if name == currentNode.Name {
				continue
			}
			_, ok := connected_peers_connection.Load(name)
			if !ok {
				all_connected = false
				go tryToConnect(node, callback)
			}
		}
		if all_connected {
			break
		}
		time.Sleep(time.Second)
	}
	fmt.Println("All connected")
}

func tryToConnect(node Node, callback func(net.Conn)) {
	address := node.Ip + ":" + strconv.Itoa(node.Port)
	conn, err := net.Dial("tcp", address)
	// fmt.Println("Connecting to ", node.Name)
	if err != nil {
		return
	}
	AddToConnectedPeers(node.Name, conn)

	// hello message
	msg := api.Helo2Message(currentNode.Name)
	enc := GetEncoder(node.Name)
	err = enc.Encode(msg)
	if err != nil {
		fmt.Println(err)
		return
	}
	go callback(conn)
}

func logDisconnections(node_name string) {
	file_name := "logs/" + GetCurNode().Name + ".log"
	f, err := os.OpenFile(file_name, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		f, err = os.Create(file_name)
		if err != nil {
			panic(err)
		}
	}
	defer f.Close()
	f.WriteString(fmt.Sprintf("DISCONNECTION %s\n", node_name))
}
