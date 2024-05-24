package raft

import (
	"context"
	"fmt"
	"log"

	"github.com/ybbus/jsonrpc/v3"
)

func (r *Raft) CreateAndSendAppendEntry() {
	if (r.State != LEADER) {
		return
	}

	for _, follower := range r.Peers {
		if (follower.ID == r.ID) {
			continue
		}

		appendEntry := AppendEntry{
			Term: r.PersistentState.CurrentTerm,
			LeaderID: r.ID,
			PrevLogIndex: r.VolatileState.LeaderVolatileState.NextIndex[follower.ID],
			PrevLogTerm: r.PersistentState.Logs[r.VolatileState.LeaderVolatileState.NextIndex[follower.ID] - 1].Term,
			Entries: r.PersistentState.Logs[r.VolatileState.LeaderVolatileState.NextIndex[follower.ID]:],
			LeaderCommitIndex: r.VolatileState.CommitIndex,
			LeaderPort: r.Port,
		}

		rpcClient := jsonrpc.NewClient(fmt.Sprintf("http://localhost:%v/rpc", follower.Port))
		resp, err := rpcClient.Call(context.Background(), "Raft.AppendEntryFollower", appendEntry)

		if (err != nil) {
			continue
		}

		var appendEntryResp AppendEntryResp
		err = resp.GetObject(&appendEntryResp)

		if err != nil {
			continue
		}

		// TODO: add handling for response
		
	}
}

func (r *Raft) AppendEntryFollower(req AppendEntry, resp *AppendEntryResp) {
	r.Mu.Lock()
	defer r.Mu.Unlock()

	if (r.State == LEADER) {
		return
	}

	r.LeaderAddr = req.LeaderPort
	r.LeaderID = req.LeaderID

	log.Printf("[Node %v]: Received Append Entry Request/Heartbeat from Leader (Node %v)\n", r.ID, req.LeaderID)
	
	resp.ReplyNodeID = r.ID

	if req.Term < r.PersistentState.CurrentTerm {
		resp.Success = false
		resp.Term = r.PersistentState.CurrentTerm
		return
	}

	if req.PrevLogIndex > 0 {
		if req.PrevLogIndex > len(r.PersistentState.Logs) || r.PersistentState.Logs[req.PrevLogIndex - 1].Term != req.PrevLogTerm {
			resp.Term = r.PersistentState.CurrentTerm
			resp.Success = false
			return
		}

		r.PersistentState.Logs = r.PersistentState.Logs[:req.PrevLogIndex]
	}

	r.PersistentState.Logs = append(r.PersistentState.Logs, req.Entries...)

	if req.LeaderCommitIndex > r.VolatileState.CommitIndex {
		r.VolatileState.CommitIndex = min(req.LeaderCommitIndex, len(r.PersistentState.Logs))
	}

	resp.Term = r.PersistentState.CurrentTerm
	resp.Success = true
}

// func (r *Raft) SendAppendEntryResponse() {

// }