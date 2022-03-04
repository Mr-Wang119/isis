# Distributed Deposit and Transfer System based on ISIS Algorithm

![截屏2022-03-04 下午3.49.52](https://s3.amazonaws.com/images.masen.com/2022/03/caa45a1b8cc767d06f530d85ec101295.png)

Distributed Deposit and Transfer System is a distributed system of multiple processes that maintain accounts and transactions. We use [the ISIS algorithm](https://en.wikipedia.org/wiki/IS-IS) to satisfy the requirement of total ordering and build a reliable multicast through implementing [the reliable multicast algorithm](https://en.wikipedia.org/wiki/Reliable_multicast). TCP errors are captured to detect failures.

## Makefile:

We do not have a makefile since we are using Golang, which is not a compiled language. We do, however, use go build to compile executables.

## Libraries:

    go 1.15  google/uuid v1.3.0

## Project Structure:

```
-cmd/
	-node/
		-node.go     // main function
-test/
	-logs/         // save the logs for the evaluation matrix
	-node          // compiled binary
	-gentx.py      // generate stimulated transactions, including "DEPOSIT" and "TRANSFER"
	-config.txt    // record the address of each node in the distributed system
-api/
	-proto.go      // transmitted message protocal
-internal/
	-peers/
		-peers.go    // store the information of the current node and peers
	-accounts/
		-accounts.go // store and manage the accounts information
	-message/
		-messageService.go // send and handle the receival of messages
		-ISIS.go     // ISIS algorithm implementation
		-message.go  // internal message structure
	-deposit/
		-depositAndTransfer.go // process the stimulated transactions
-logs/
-graph.ipynb     // generate the evaluation graph
-go.sum
-go.mod
```

## Running:

1. Go into the folder “mp1/test/”

2. Executable files “mp1_node” should be present

3. Run with the same commands as specified on CS 425’s website with any valid corresponding node name and config filename

​		```python3 -u gentx.py 0.5 | ./node node1 config.txt```

4. If somehow those executables did not work, please try our building instructions.

## Building:

1. Go into the folder “mp1/test/”, if it is not presented, please create a new folder with the name “test” under “mp1”, and put “gentx.py” and config file into the “test” folder

2. Run the following commands in the “test” folder

​		```go build ../api/node/node.go```

## Implementation Details:

We use the ISIS algorithm to satisfy the requirement of total ordering, and build a reliable multicast through implementing the reliable multicast algorithm. TCP errors are captured to detect failures.

- Total ordering:

​       We implement the ISIS algorithm for total ordering. After every node receives the message, it would reply to the proposed sequence to the sender. After the sender collected all the proposed sequence numbers and selects the largest one, as the next agreed sequence number. Then, it would R-multicasts to other nodes.

![B_KufC9U4AEZMfC](https://s3.amazonaws.com/images.masen.com/2022/03/1d2ad0ad0db6499070666c3f2c3e3510.png)

- Reliable delivery under failures:

​       To ensure reliable delivery, we implement the reliable multicast algorithm. After a process B-multicasts the message to the processes in the destination group, the recipient checks if it has received the message, if not, it would then B-multicasts the message to the group. In this way, the failure of delivering messages between two nodes would be more reliable.

![image-20220304152727261](https://s3.amazonaws.com/images.masen.com/2022/03/84be41fecb8a644b0b79bb82c66b2e0f.png)

​       To detect failures, each node would check the state of the connections with other nodes regularly. If some connection is broken, the node would delete the message sent by the failed node from the hold-back queue. Then, it would check the messages sent by the current node which have not received all responses from other nodes. If the failed node has not responded, we would “pretend” the broken node has responded to the proposed sequence number to the message, and check if the process collects all the proposed sequence numbers.

## Evaluation:

The performance of the system is evaluated using the “transaction processing time”, measuring the time difference between the message generation and the message delivery. The time difference distribution is represented as the CDF (cumulative distribution function).

- 3 nodes, 0.5 Hz each, running for 100 seconds

![image-20220304153158728](https://s3.amazonaws.com/images.masen.com/2022/03/901453fa031574ac3564a67d5edb78b1.png)

- 8 nodes, 5 Hz each, running for 100 seconds

![image-20220304153205023](https://s3.amazonaws.com/images.masen.com/2022/03/0583d3b05f387f6ceed7eef0c538be24.png)

- 3 nodes, 0.5 Hz each, running for 100 seconds, then one node fails, and the rest continue to run for 100 seconds

![image-20220304153214748](https://s3.amazonaws.com/images.masen.com/2022/03/e17d84b043ca7a2f77ee278e7ddf22d4.png)

- 8 nodes, 5 Hz each, running for 100 seconds, then 3 nodes fail simultaneously, and the rest continue to run for 100 seconds

![image-20220304153225283](https://s3.amazonaws.com/images.masen.com/2022/03/61141d3bb197b1b1adb34959d303c944.png)
