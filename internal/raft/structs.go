package raft

import (
	"net/http"
	"sync"
)

type Raft struct {
	Mu sync.Mutex
	
	Peers []*Raft // a list of all peers - TODO: change
	Port string // port for raft node server
	Server *http.Server

	State State // current state of node (leader, candidate, or follower)
	ID int // node ID

	// LeaderAddr string
	// LeaderID int

	PersistentState struct {
		CurrentTerm int // latest term server has seen
		VotedFor int // candidateID they voted for, -1 if hasn't voted
		Logs []Log // first index is 1

		Database map[string]interface{} // database key/value store
	}
	
	VolatileState struct {
		CommitIndex int // index of highest log known to be committed
		LastApplied int // index of highest log applied to state machine

		LeaderVolatileState struct {
			NextIndex map[int]int // index of the next log to send to each following server
			MatchIndex map[int]int // index of highest log known to be replicated on following servers
		}
	}
}

type Log struct {
	Operation Operation

	Key string
	Value interface{}

	Term int
}

type RequestVote struct {

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