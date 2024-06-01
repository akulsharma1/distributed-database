package raft

import "log"

func (r *Raft) Printf(message string) {
	log.Printf("[Node %v]: "+message+"\n", r.ID)
}