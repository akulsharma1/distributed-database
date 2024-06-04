package raft

import (
	"net/http"
	"sync"
	"time"
)

type Raft struct {
	Mu sync.Mutex
	
	Peers []*Peer // a list of all peers - TODO: change
	Port string // port for raft node server
	Server *http.Server

	State State // current state of node (leader, candidate, or follower)
	ID int // node ID

	LeaderAddr string
	LeaderID int

	ElectionTimer *time.Duration // election timer (randomly assigned between 150-300ms)
	LastHeartbeat *time.Time // last heartbeat/append entry rpc the node received
	ElectionChan chan bool

	PersistentState PersistentState
	VolatileState VolatileState
}

type PersistentState struct {
	CurrentTerm int // latest term server has seen
	VotedFor int // candidateID they voted for, -1 if hasn't voted
	Logs []Log // first index is 1

	Database map[string]interface{} // database key/value store
}

type VolatileState struct {
	CommitIndex int // index of highest log known to be committed
	LastApplied int // index of highest log applied to state machine

	LeaderVolatileState LeaderVolatileState
}

type LeaderVolatileState struct {
	NextIndex map[int]int // index of the next log to send to each following server
	MatchIndex map[int]int // index of highest log known to be replicated on following servers
}

type Log struct {
	Operation Operation

	Key string
	Value interface{}

	Term int
}

type RequestVote struct {
	Term int
	CandidateID int
	LastLogIndex int // index of last log candidate has
	LastLogTerm int // term of last log candidate has
}

type RequestVoteResp struct {
	Term int // term of peer for updating
	VoteGranted bool // whether vote was granted from peer
}

type AppendEntry struct {
	Term int
	LeaderID int
	PrevLogIndex int
	PrevLogTerm int
	Entries []Log
	LeaderCommitIndex int

	LeaderPort string
}

type AppendEntryResp struct {
	ReplyNodeID int // id of raft node sending reply
	Term int // current term for leader to update itself
	Success bool // true if follower contained entry matching PrevLogIndex and PrevLogTerm
}

type Peer struct {
	ID int `json:"id"`
	Address string `json:"address"`
}

type HttpServerResponse struct {
	Success bool `json:"success"`
	Message string `json:"message"`
	Value interface{} `json:"value"`
}