package raft

import (
	"context"
	"errors"
	"fmt"
	"time"

	"net/http"

	"github.com/ybbus/jsonrpc/v3"
)

func (r *Raft) CheckIfElectionRequired() {
	for {
		if r.State == LEADER {
			continue
		}
		r.Mu.Lock()
		timeDiff := time.Since(*r.LastHeartbeat)
		r.Mu.Unlock()

		if timeDiff > *r.ElectionTimer {
			r.Printf("Found time since last heartbeat > election timer. Sending message to start election.")
			r.ElectionChan <- true
		}
	}
}

func (r *Raft) StartElection() {
	if r.State == CANDIDATE || r.State == LEADER {
		return
	}
	// r.Printf("------Starting election.-------")

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

		go func(peer *Peer) {
			r.Printf(fmt.Sprintf("Sending vote request to peer %v", peer.ID))

			r.Mu.Lock()

			var voteRequest *RequestVote
			if len(r.PersistentState.Logs) == 0 {
				voteRequest = &RequestVote{
					Term: r.PersistentState.CurrentTerm,
					CandidateID: r.ID,
					LastLogIndex: 0,
					LastLogTerm: 0,
				}
			} else {
				voteRequest = &RequestVote{
					Term: r.PersistentState.CurrentTerm,
					CandidateID: r.ID,
					LastLogIndex: len(r.PersistentState.Logs) - 1,
					LastLogTerm: r.PersistentState.Logs[len(r.PersistentState.Logs) - 1].Term,
				}
			}
			

			r.Mu.Unlock()

			rpcClient := jsonrpc.NewClient(fmt.Sprintf("http://%v/rpc", peer.Address))
			resp, err := rpcClient.Call(context.Background(), "Raft.VoteRequestReply", &voteRequest)

			// log.Println("****VOTE REQUEST REPLY******")
			// log.Println(resp, err)

			if (err != nil) {
				return
			}

			var voteRequestReply *RequestVoteResp
			err = resp.GetObject(&voteRequestReply)

			if err != nil {
				return
			}

			r.Mu.Lock()
			if voteRequestReply.Term > r.PersistentState.CurrentTerm {
				r.State = FOLLOWER
				r.PersistentState.CurrentTerm = voteRequest.Term

				r.Mu.Unlock()

				return
			}

			if voteRequestReply.VoteGranted {
				r.Printf(fmt.Sprintf("Received vote from node %v", peer.ID))
				numOfVotes++
			}
			r.Mu.Unlock()
		}(peer)
	}

	r.Mu.Lock()
	defer r.Mu.Unlock()
	
	if r.State != CANDIDATE {
		return
	}

	r.Printf(fmt.Sprintf("Received %v total votes out of %v total peers during candidacy", numOfVotes, len(r.Peers)))
	if numOfVotes > len(r.Peers) / 2 {
		r.Printf("Becoming leader.")
		r.State = LEADER

		for _, peer := range r.Peers {
			if peer.ID == r.ID {
				continue
			}
			r.VolatileState.LeaderVolatileState.NextIndex[peer.ID] = len(r.PersistentState.Logs) + 1
			r.VolatileState.LeaderVolatileState.MatchIndex[peer.ID] = 0
		}
	} else {
		r.State = FOLLOWER
	}
}

func (r *Raft) VoteRequestReply(httpreq *http.Request, req *RequestVote, resp *RequestVoteResp) error {
	r.Mu.Lock()
	defer r.Mu.Unlock()

	resp.Term = r.PersistentState.CurrentTerm

	if req.Term < r.PersistentState.CurrentTerm {
		r.Printf("Received vote request. Voting no, invalid term.")
		resp.VoteGranted = false
		return errors.New("invalid term")
	}

	if req.LastLogIndex >= len(r.PersistentState.Logs) - 1 && req.LastLogTerm >= r.PersistentState.Logs[len(r.PersistentState.Logs) - 1].Term {
		if r.PersistentState.VotedFor == -1 {
			r.Printf("Received vote request. Voting yes.")
			resp.VoteGranted = true
			r.PersistentState.VotedFor = req.CandidateID
		} else {
			r.Printf("Denying vote request, already voted")
		}
	}

	r.Printf("Received vote request. Voting no.")

	resp.VoteGranted = false
	return nil
}