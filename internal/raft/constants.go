package raft

type State int

const (
	FOLLOWER State = iota
	CANDIDATE State = iota
	LEADER State = iota
)

type Operation string

const (
	// GET Operation = "GET"

	PUT Operation = "PUT"
	// DELETE Operation = "DELETE"
)