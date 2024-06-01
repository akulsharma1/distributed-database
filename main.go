package main

import (
	"flag"
	"fmt"
	"math/rand"
	"time"

	"github.com/akulsharma1/distributed-database/internal/raft"
)

var (
	flagNodeID = flag.Int("nodeID", -1, "node id")
	flagPort = flag.String("port", "", "port to start node")
)

func main() {

	if *flagNodeID == -1 {
		fmt.Println("Invalid nodeID flag")
		return
	}
	if *flagPort == "" {
		fmt.Println("Invalid port flag")
		return
	}

	t := time.Now()

	r := &raft.Raft{
		Peers: []*raft.Peer{},
		Port: *flagPort,
		State: raft.FOLLOWER,
		ID: *flagNodeID,
		ElectionTimer: generateElectionTime(),
		LastHeartbeat: &t,
		ElectionChan: make(chan bool),
		PersistentState: raft.PersistentState{
			CurrentTerm: 0,
			VotedFor: -1,
			Logs: []raft.Log{},
			Database: make(map[string]interface{}),
		},
		VolatileState: raft.VolatileState{
			CommitIndex: 0,
			LastApplied: 0,
			LeaderVolatileState: raft.LeaderVolatileState{
				NextIndex: make(map[int]int),
				MatchIndex: make(map[int]int),
			},
		},
	}

	r.StartServer()
}

func generateElectionTime() *time.Duration {
	n := 150 + rand.Intn(300-150+1)

	t, _ := time.ParseDuration(fmt.Sprintf("%vms", n))

	return &t
}