package raft

import (
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/rpc"
	"time"
)

/*
Starts the raft server at the given port.
Also handles heartbeats, elections, etc.
Blocker function, does not stop.
*/
func (r *Raft) StartServer() {
	rpcServer := rpc.NewServer()
	rpcServer.Register(r)

	listener, err := net.Listen("tcp", r.Port)
	if err != nil {
		panic(err)
	}

	mux := http.NewServeMux()
	mux.Handle(rpc.DefaultRPCPath, rpcServer)

	mux.HandleFunc("/get", r.HandleGet)
	mux.HandleFunc("/set", r.HandleGet)

	r.Server = &http.Server{Handler: mux}
	go r.Server.Serve(listener)

	// append entries for leader - function runs no matter what
	// if not leader, it just returns and goroutine goes to next loop iteration
	go func() {
		for {
			r.CreateAndSendAppendEntry()

			// heartbeat interval: 100 ms
			time.Sleep(100 * time.Millisecond)
		}
	}()
}

func (r *Raft) HandleGet(w http.ResponseWriter, req *http.Request) {
	key := req.URL.Query().Get("key")

	fmt.Println(key)
	if r.State != LEADER {
		// TODO: forward request. for now we will just return "not leader"
		http.Error(w, "Not Leader", http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write(r.PersistentState.Database[key].([]byte))
}

func (r *Raft) HandleSet(w http.ResponseWriter, req *http.Request) {
	r.Mu.Lock()
	defer r.Mu.Unlock()

	if r.State != LEADER {
		// TODO: forward request. for now we will just return "not leader"
		http.Error(w, "Not Leader", http.StatusBadRequest)
		return
	}

	var logEntry Log

	body, err := io.ReadAll(req.Body)
    if err != nil {
        http.Error(w, "Unable to read request body", http.StatusBadRequest)
        return
    }

    err = json.Unmarshal(body, &logEntry)
    if err != nil {
        http.Error(w, "Invalid request body", http.StatusBadRequest)
        return
    }

	logEntry.Operation = PUT
	logEntry.Term = r.PersistentState.CurrentTerm

	r.PersistentState.Logs = append(r.PersistentState.Logs, logEntry)

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Success"))
}