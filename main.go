package main

import (
	"flag"
	"fmt"

	"github.com/akulsharma1/distributed-database/internal/raft"
)

var (
	flagNodeID = flag.Int("nodeID", -1, "node id")
	flagPort = flag.String("port", "", "port to start node")
)

func main() {
	if *flagNodeID == -1 {
		fmt.Println("Invalid nodeID flag")
	}
	if *flagPort == "" {
		fmt.Println("Invalid port flag")
	}

	r := &raft.Raft{
		Peers: []*raft.Raft{}, // TODO: add
		Port: *flagPort,
		State: raft.FOLLOWER,
		ID: *flagNodeID,
	}

	r.StartServer()
}