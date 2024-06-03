# Distributed Database

An implementation of a key/value database in Golang. Uses the [Raft consensus algorithm](https://raft.github.io/) for server consensus.

**How to start a node**

`go run . --nodeID <node_id_here> --port <port here (e.g. localhost:5555)>`
In order to run multiple nodes, open multiple terminals and run the same command with different NodeIDs and ports.

**General Info**
- The implementation currently uses a central JSON file which all nodes write to when they join the cluster. On the future todo list is adding consul support as a central registry.

**Future Todo (in no particular order)**
- Add consul support as a central registry.
- Add sharding to the database.
- Add crash recovery.
- Use in built memory instead of just maps for storage.
