package raft

import (
	"fmt"
	"net"
	"net/http"
	"net/rpc"
)

func (r *Raft) StartServer() {
	rpcServer := rpc.NewServer()
	rpcServer.Register(r)

	listener, err := net.Listen("tcp", r.Port)
	if err != nil {
		panic(err)
	}

	mux := http.NewServeMux()
	mux.Handle(rpc.DefaultRPCPath, rpcServer)

	// TODO: add endpoints to handle functions
	mux.HandleFunc("/get", r.HandleGet)

	r.Server = &http.Server{Handler: mux}
	go r.Server.Serve(listener)
}

func (r *Raft) HandleGet(w http.ResponseWriter, req *http.Request) {
	key := req.URL.Query().Get("key")

	fmt.Println(key)
	// if r.State != LEADER {
	// 	// forward
	// }

	// TODO: HANDLE
}