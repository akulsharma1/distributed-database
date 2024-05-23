package raft

type State int

const (
	FOLLOWER State = iota
	CANDIDATE State = iota
	LEADER State = iota
)