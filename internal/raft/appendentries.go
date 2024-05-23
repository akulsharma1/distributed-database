package raft

// func (r *Raft) createAppendEntry() *AppendEntry {
// 	return &AppendEntry{
// 		Term: r.PersistentState.CurrentTerm,
// 		LeaderID: r.ID,
// 		PrevLogIndex: r.,
// 	}
// }

func (r *Raft) AppendEntryFollower(req AppendEntry) *AppendEntryResp {
	r.Mu.Lock()
	defer r.Mu.Unlock()
	
	resp := &AppendEntryResp{ReplyNodeID: r.ID}

	if req.Term < r.PersistentState.CurrentTerm {
		resp.Success = false
		resp.Term = r.PersistentState.CurrentTerm
		return resp
	}

	if req.PrevLogIndex > 0 {
		if req.PrevLogIndex > len(r.PersistentState.Logs) || r.PersistentState.Logs[req.PrevLogIndex - 1].Term != req.PrevLogTerm {
			resp.Term = r.PersistentState.CurrentTerm
			resp.Success = false
			return resp
		}

		r.PersistentState.Logs = r.PersistentState.Logs[:req.PrevLogIndex]
	}

	r.PersistentState.Logs = append(r.PersistentState.Logs, req.Entries...)

	if req.LeaderCommitIndex > r.VolatileState.CommitIndex {
		r.VolatileState.CommitIndex = min(req.LeaderCommitIndex, len(r.PersistentState.Logs))
	}

	resp.Term = r.PersistentState.CurrentTerm
	resp.Success = true

	return resp
}