package registry

import "github.com/akulsharma1/distributed-database/internal/raft"

type Registry struct {
	Nodes []*raft.Peer `json:"nodes"`
}