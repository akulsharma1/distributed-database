package raft

import (
	"context"
	"fmt"
	"time"

	"github.com/ybbus/jsonrpc/v3"
)

func (r *Raft) CheckIfElectionRequired() {
	for {
		r.Mu.Lock()
		timeDiff := time.Since(*r.LastHeartbeat)
		r.Mu.Unlock()

		if timeDiff > *r.ElectionTimer {
			r.ElectionChan <- true
		}
	}
}

func (r *Raft) StartElection() {
	if r.State == CANDIDATE || r.State == LEADER {
		return
	}

	r.Mu.Lock()

	r.State = CANDIDATE

	r.PersistentState.CurrentTerm++

	numOfVotes := 1

	*r.LastHeartbeat = time.Now()

	r.Mu.Unlock()

	for _, peer := range r.Peers {
		if peer.ID == r.ID {
			continue
		}
		r.Mu.Lock()

		voteRequest := &RequestVote{
			Term: r.PersistentState.CurrentTerm,
			CandidateID: r.ID,
			LastLogIndex: len(r.PersistentState.Logs) - 1,
			LastLogTerm: r.PersistentState.Logs[len(r.PersistentState.Logs) - 1].Term,
		}

		r.Mu.Unlock()

		rpcClient := jsonrpc.NewClient(fmt.Sprintf("http://localhost:%v/rpc", peer.Port))
		resp, err := rpcClient.Call(context.Background(), "Raft.VoteRequestReply", voteRequest)

		if (err != nil) {
			continue
		}

		var voteRequestReply RequestVoteResp
		err = resp.GetObject(&voteRequestReply)

		if err != nil {
			continue
		}

		r.Mu.Lock()
		if voteRequestReply.Term > r.PersistentState.CurrentTerm {
			r.State = FOLLOWER
			r.PersistentState.CurrentTerm = voteRequest.Term

			r.Mu.Unlock()

			return
		}

		if voteRequestReply.VoteGranted {
			numOfVotes++
		}
		r.Mu.Unlock()
	}

	r.Mu.Lock()
	defer r.Mu.Unlock()
	
	if r.State != CANDIDATE {
		return
	}

	if numOfVotes > len(r.Peers) / 2 {
		r.State = LEADER
	}
}