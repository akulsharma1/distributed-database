package raft

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/mux"
	rpc "github.com/gorilla/rpc"
	gjson "github.com/gorilla/rpc/json"
)

/*
Starts the raft server at the given port.
Also handles heartbeats, elections, etc.
Blocker function, does not stop.
*/
func (r *Raft) StartServer() {
	router := mux.NewRouter()

	router.HandleFunc("/get", r.HandleGet)
	router.HandleFunc("/set", r.HandleSet)

	rpcServer := rpc.NewServer()
	rpcServer.RegisterCodec(gjson.NewCodec(), "application/json")
	err := rpcServer.RegisterService(r, "")

	if err != nil {
		panic(err)
	}

	router.Handle("/rpc", rpcServer)
	
	var wg sync.WaitGroup
	wg.Add(1)

	go http.ListenAndServe(r.Port, router)

	r.Printf(fmt.Sprintf("Starting server at %v", r.Port))
	
	wg.Add(1)
	// append entries for leader - function runs no matter what
	// if not leader, it just returns and goroutine goes to next loop iteration
	go func() {
		for {
			r.CreateAndSendAppendEntry()

			// heartbeat interval: 100 ms
			time.Sleep(100 * time.Millisecond)
		}
	}()
	
	wg.Add(1)
	go r.CheckIfElectionRequired()

	wg.Add(1)
	go func() {
		for {
			StartElection := <-r.ElectionChan
			if StartElection && r.State != CANDIDATE {
				r.StartElection()
			}
		}
	}()
	
	wg.Wait()
}

func (r *Raft) HandleGet(w http.ResponseWriter, req *http.Request) {

	key := req.URL.Query().Get("key")

	if r.State != LEADER {
		// TODO: forward request. for now we will just return "not leader"
		resp := &HttpServerResponse{
			Success: false,
			Message: "Not Leader",
		}
		
		data, _ := json.Marshal(resp)

		w.WriteHeader(http.StatusBadRequest)
		w.Write(data)

		return
	}

	val, ok := r.PersistentState.Database[key]

	if !ok {
		resp := &HttpServerResponse{
			Success: false,
			Message: "Data not found",
		}

		data, _ := json.Marshal(resp)

		w.WriteHeader(http.StatusNotFound)
		w.Write(data)

		return
	}
	
	resp := &HttpServerResponse{
		Success: true,
		Message: "Found Data",
		Value: val,
	}

	data, _ := json.Marshal(resp)

	w.WriteHeader(http.StatusOK)
	w.Write(data)
}

func (r *Raft) HandleSet(w http.ResponseWriter, req *http.Request) {
	r.Mu.Lock()
	defer r.Mu.Unlock()

	if r.State != LEADER {
		// TODO: forward request. for now we will just return "not leader"
		resp := &HttpServerResponse{
			Success: false,
			Message: "Not Leader",
		}
		
		data, _ := json.Marshal(resp)

		w.WriteHeader(http.StatusBadRequest)
		w.Write(data)

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

	if logEntry.Key == "" || logEntry.Value == nil {
		http.Error(w, "key/value pair required", http.StatusBadRequest)
		return
	}

	logEntry.Operation = PUT
	logEntry.Term = r.PersistentState.CurrentTerm

	r.PersistentState.Logs = append(r.PersistentState.Logs, logEntry)

	r.PersistentState.Database[logEntry.Key] = logEntry.Value

	resp := &HttpServerResponse{
		Success: true,
		Message: "Set key/value pair",
	}

	data, _ := json.Marshal(resp)

	w.WriteHeader(http.StatusOK)
	w.Write(data)
}