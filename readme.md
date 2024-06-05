# Distributed Database

An implementation of a key/value database in Golang. Uses the [Raft consensus algorithm](https://raft.github.io/) for server consensus. This implementation is a proof of concept and may have bugs.

## How to start a node

`go run . --nodeID <node_id_here> --address <address here (e.g. localhost:5555)>`
In order to run multiple nodes, open multiple terminals and run the same command with different NodeIDs and ports.

## Requests
- `GET` `<address>/get?key=<key>`
```
GET http://localhost:5555/get?key=abc
```

- `POST` `<address>/set` with the following json body:
```
{
    "key": <key>,
    "value": <value>
}
```
Note that value can be of any type; it isn't restricted to strings.

**Example:**
```
POST http://localhost:5555/set
{
    "key": "abc",
    "value": {
        "name": "bob",
        "email": "bob@gmail.com"
    }
}
```

**General Info**
- The implementation currently uses a central JSON file which all nodes write to when they join the cluster. On the future todo list is adding consul support as a central registry.

**Future Todo (in no particular order)**
- Add consul support as a central registry.
- Add sharding to the database.
- Add crash recovery.
- Use in built memory instead of just maps for storage.
- Add request forwarding to leader
