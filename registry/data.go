package registry

import (
	"encoding/json"
	"errors"
	"io"
	"log"
	"os"

	"github.com/akulsharma1/distributed-database/internal/raft"
)

func GetNodes() ([]*raft.Peer, error) {
	file, err := os.Open("./nodes.json")
	if err != nil {
		log.Printf("Failed to open file: %v", err)
		return []*raft.Peer{}, errors.New("error opening registry file")
	}
	defer file.Close()

	// Read the contents of the file
	data, err := io.ReadAll(file)
	if err != nil {
		log.Printf("Failed to read file: %v", err)
		return []*raft.Peer{}, errors.New("error reading registry file")
	}

	// Unmarshal the data into a Registry instance
	var registry Registry
	err = json.Unmarshal(data, &registry)
	if err != nil {
		log.Printf("Failed to unmarshal JSON: %v", err)
		return []*raft.Peer{}, errors.New("error unmarshaling registry file")
	}

	return registry.Nodes, nil
}

func AddNode(node *raft.Peer) error {
	file, err := os.Open("./nodes.json")
	if err != nil {
		log.Printf("Failed to open file: %v", err)
		return errors.New("error opening registry file")
	}
	defer file.Close()

	// Read the contents of the file
	data, err := io.ReadAll(file)
	if err != nil {
		log.Printf("Failed to read file: %v", err)
		return errors.New("error reading registry file")
	}

	// Unmarshal the data into a Registry instance
	var registry Registry
	err = json.Unmarshal(data, &registry)
	if err != nil {
		log.Printf("Failed to unmarshal JSON: %v", err)
		return errors.New("error unmarshaling registry file to struct")
	}

	registry.Nodes = append(registry.Nodes, node)

	updatedData, err := json.Marshal(registry)

	if err != nil {
		log.Printf("Failed to marshal JSON: %v", err)
		return errors.New("error marshaling struct to json for registry file write")
	}
	
	_, err = file.Write(updatedData)

	if err != nil {
		log.Printf("Failed to write to registry file: %v", err)
		return errors.New("error writing to registry file")
	}

	return nil
}