package raft

import (
	"context"
	"fmt"

	"github.com/ybbus/jsonrpc/v3"
)

/*
Append Entry function - leader only function.
Sends Append Entry RPC to every follower
*/
func (r *Raft) CreateAndSendAppendEntry() {
	if (r.State != LEADER) {
		return
	}

	successCount := 0

	for _, follower := range r.Peers {
		if (follower.ID == r.ID) {
			continue
		}

		r.Printf(fmt.Sprintf("Sending Append Entry to follower Node %v at port %v\n", follower.ID, follower.Address))

		r.Mu.Lock()

		appendEntry := AppendEntry{
			Term: r.PersistentState.CurrentTerm,
			LeaderID: r.ID,
			PrevLogIndex: r.VolatileState.LeaderVolatileState.NextIndex[follower.ID],
			PrevLogTerm: r.PersistentState.Logs[r.VolatileState.LeaderVolatileState.NextIndex[follower.ID] - 1].Term,
			Entries: r.PersistentState.Logs[r.VolatileState.LeaderVolatileState.NextIndex[follower.ID]:],
			LeaderCommitIndex: r.VolatileState.CommitIndex,
			LeaderPort: r.Port,
		}

		r.Mu.Unlock()

		rpcClient := jsonrpc.NewClient(fmt.Sprintf("http://localhost:%v/rpc", follower.Address))
		resp, err := rpcClient.Call(context.Background(), "Raft.AppendEntryFollower", appendEntry)

		if (err != nil) {
			continue
		}

		var appendEntryResp AppendEntryResp
		err = resp.GetObject(&appendEntryResp)

		if err != nil {
			continue
		}

		r.Mu.Lock()

		if !appendEntryResp.Success {
			if appendEntryResp.Term < r.PersistentState.CurrentTerm {
				r.PersistentState.CurrentTerm = appendEntryResp.Term
				r.State = FOLLOWER
				r.Mu.Unlock()
				return
			}

			r.VolatileState.LeaderVolatileState.NextIndex[follower.ID]--
		} else {
			r.VolatileState.LeaderVolatileState.MatchIndex[follower.ID] = r.PersistentState.CurrentTerm
			r.VolatileState.LeaderVolatileState.NextIndex[follower.ID] = len(r.PersistentState.Logs) - 1
			successCount++
		}

		r.Mu.Unlock()

		
	}

	// at least half of the followers say they have added it to their logs
	if successCount > len(r.Peers) / 2 {
		// update commit index
		r.VolatileState.CommitIndex = len(r.PersistentState.Logs) - 1
	}
}

/*
Follower handling for appending entry RPC. Used in RPC Call.
*/
func (r *Raft) AppendEntryFollower(req AppendEntry, resp *AppendEntryResp) {
	r.Mu.Lock()
	defer r.Mu.Unlock()

	if (r.State == LEADER) {
		return
	}

	r.LeaderAddr = req.LeaderPort
	r.LeaderID = req.LeaderID

	r.Printf(fmt.Sprintf("Received Append Entry Request/Heartbeat from Leader (Node %v)\n", req.LeaderID))
	
	resp.ReplyNodeID = r.ID

	if req.Term < r.PersistentState.CurrentTerm {
		resp.Success = false
		resp.Term = r.PersistentState.CurrentTerm
		return
	} else if req.Term > r.PersistentState.CurrentTerm {
		r.PersistentState.CurrentTerm = req.Term
		r.State = FOLLOWER
	}

	if req.PrevLogIndex > 0 {
		if req.PrevLogIndex > len(r.PersistentState.Logs) || r.PersistentState.Logs[req.PrevLogIndex].Term != req.PrevLogTerm {
			resp.Term = r.PersistentState.CurrentTerm
			resp.Success = false
			r.PersistentState.Logs = r.PersistentState.Logs[:req.PrevLogIndex]
			return
		}
	}

	/*
	Per https://thesquareplanet.com/blog/students-guide-to-raft/

	The if here is crucial. If the follower has all the entries the leader sent, the follower MUST NOT truncate its log. 
	Any elements following the entries sent by the leader MUST be kept. This is because we could be receiving an outdated 
	AppendEntries RPC from the leader, and truncating the log would mean “taking back” entries 
	that we may have already told the leader that we have in our log.
	*/

	// NOTE: this code may have a bug where if logs are deleted then the database might be stale - maybe have to undo any deleted logs?
	for i, log := range r.PersistentState.Logs[req.PrevLogIndex:] {
		if i >= len(req.Entries) {
			break
		}
		if log.Term != req.Entries[i].Term {
			r.PersistentState.Logs = r.PersistentState.Logs[:i]
			break
		}
	}


	for _, log := range req.Entries {
		r.HandleLog(log)
	}

	r.PersistentState.Logs = append(r.PersistentState.Logs, req.Entries...)

	if req.LeaderCommitIndex > r.VolatileState.CommitIndex {
		r.VolatileState.CommitIndex = min(req.LeaderCommitIndex, len(r.PersistentState.Logs) - 1)
	}

	resp.Term = r.PersistentState.CurrentTerm
	resp.Success = true
}

/*
Database handler for logs which are appended.
*/
func (r *Raft) HandleLog(log Log) {
	if log.Operation == PUT {
		r.PersistentState.Database[log.Key] = log.Value
	} 
	// else if log.Operation == DELETE {
	// 	delete(r.PersistentState.Database, log.Key)
	// }
}