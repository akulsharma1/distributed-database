package raft

import (
	"net/http"
	"sync"
)

type Raft struct {
	Mu sync.Mutex
	
	Peers []string
	Port string
	Server *http.Server
	State State
	ID int

	PersistentState struct {
		CurrentTerm int // latest term server has seen
		VotedFor int // candidateID they voted for, -1 if hasn't voted
		Logs []Log // first index is 1
	}
	
	VolatileState struct {
		CommitIndex int
		LastApplied int

		LeaderVolatileState struct {

		}
	}

}

type Log struct {
	Data interface{}
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
}

type AppendEntryResp struct {
	ReplyNodeID int // id of raft node sending reply
	Term int // current term for leader to update itself
	Success bool // true if follower contained entry matching PrevLogIndex and PrevLogTerm
}