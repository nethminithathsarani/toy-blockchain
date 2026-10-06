# Networked Toy Blockchain

A networked toy blockchain developed in **Go 1.22+** for the Golang Developer Assessment 2.

The project extends the original single-process blockchain simulator into a multi-node HTTP network supporting Ed25519-signed transactions, transaction and block gossip, de-duplication, synchronization, fork resolution, chain reorganization, and race-free shared state.

## Main Features

### Blockchain

- Deterministic genesis block
- SHA-256 block hashing
- Previous-hash block linking
- Proof-of-Work mining
- Concurrent mining using goroutines
- Merkle roots and transaction inclusion proofs
- Difficulty adjustment
- Full-chain validation
- Tamper detection
- Cumulative Proof-of-Work chain selection

### Transactions and Wallets

- Ed25519 public/private key pairs
- Public-key-derived sender addresses
- Signed transactions
- Deterministic transaction IDs
- Per-sender transaction nonces
- Replay-attack protection
- Pending-pool validation
- Double-spending prevention
- Integer-based account balances

### Networking

- Independent HTTP blockchain nodes
- Configurable node addresses and peers
- Transaction gossip
- Block gossip
- Transaction and block de-duplication
- Peer forwarding with timeouts
- Three-node local cluster launcher
- Peer discovery from a single seed
- Two-way peer registration
- Duplicate and self-peer prevention
- Docker Compose three-node cluster
- Separate listen and advertised node addresses
- Container health checks


### Synchronization and Reorganization

- New-node synchronization
- Missing-block downloads
- Behind-node catch-up
- Competing-chain detection
- Cumulative-work fork selection
- Full candidate-chain validation
- Ledger rebuilding after reorganization
- Valid orphaned transactions returned to the pending pool

### Concurrency Safety

- `sync.RWMutex` protects shared node state
- Read locks for introspection endpoints
- Write locks for transactions, mining, synchronization and reorganization
- Network communication occurs after releasing blockchain locks
- Verified using Go’s race detector

## Requirements

- Go 1.22 or later
- PowerShell for the provided Windows cluster launcher
- WSL/Linux recommended for running the Go race detector on Windows

Check Go:

```bash
go version
```

## Installation

Clone the repository:

```bash
git clone <repository-url>
cd toy-blockchain
```

Verify the project:

```bash
go mod tidy
go test ./...
```

## Creating Wallets

Create wallets before submitting transactions:

```bash
go run . wallet Alice
go run . wallet Bob
```

Wallet files contain private keys and must not be committed or shared.

## Running One Node

Start a node on port `8001`:

```bash
go run . node -address localhost:8001
```

Start another node with Node 8001 as a peer:

```bash
go run . node -address localhost:8002 -peers localhost:8001
```

A node attempts initial synchronization with its configured peers before starting its HTTP server. If peers are unavailable, it logs the error and still starts.

## Running the Three-Node Cluster

The project includes:

```text
start-cluster.ps1
```

Run:

```powershell
.\start-cluster.ps1
```

If PowerShell blocks the script:

```powershell
powershell -ExecutionPolicy Bypass -File .\start-cluster.ps1
```
## Running the Cluster with Docker Compose

Requirements:

* Docker Desktop
* Docker Compose v2

Build and start all three nodes:

```bash
docker compose up --build -d
```

Check container health:

```bash
docker compose ps
```

View network logs:

```bash
docker compose logs
```

The services are available from the host at:

| Node   | Host API                | Docker advertised address |
| ------ | ----------------------- | ------------------------- |
| Node A | `http://localhost:8001` | `node-a:8001`             |
| Node B | `http://localhost:8002` | `node-b:8002`             |
| Node C | `http://localhost:8003` | `node-c:8003`             |

Node C is configured with only Node A as its seed. It discovers Node B and registers itself with both existing nodes.

Stop and remove the cluster containers and temporary network:

```bash
docker compose down
```

The Docker image is built using a multi-stage `Dockerfile` and runs the final application as a non-root user. Private wallet files, generated chain data and Git history are excluded through `.dockerignore`.

## Peer Discovery

A new node needs only one seed address. It requests the seed node's peer list, safely adds unknown addresses and continues discovery through newly learned peers.

The new node also registers its own address with every peer it contacts. This allows existing nodes to learn about the joining node and creates two-way connectivity.

Example:

```powershell
go run . node `
    -address localhost:8003 `
    -peers localhost:8001
```

If Node A already knows Node B, Node C forms the following peer list automatically:

```text
Node C: [localhost:8001 localhost:8002]
```

The live three-node experiment produced:

```text
Node A: [localhost:8002 localhost:8003]
Node B: [localhost:8001 localhost:8003]
Node C: [localhost:8001 localhost:8002]
```

Peer-list access uses the node's read/write mutex. Gossip and synchronization operate on safe snapshots so network requests do not hold the mutex.

The script starts:

| Node | Address | Initial peers |
|---|---|---|
| Node A | `localhost:8001` | `8002`, `8003` |
| Node B | `localhost:8002` | `8001`, `8003` |
| Node C | `localhost:8003` | `8001`, `8002` |

Stop each node using `Ctrl + C`.

## Submitting a Network Transaction

Submit an Ed25519-signed transaction to Node A:

```bash
go run . submit localhost:8001 Alice Bob 20
```

The command:

1. Loads Alice’s wallet.
2. Requests Alice’s next nonce from Node A.
3. Creates and signs the transaction.
4. Calculates its deterministic transaction ID.
5. Sends it to `POST /transactions`.
6. Allows the receiving node to validate and gossip it.

Check all pending pools:

```bash
curl.exe http://localhost:8001/mempool
curl.exe http://localhost:8002/mempool
curl.exe http://localhost:8003/mempool
```

Expected before mining:

```json
{"size":1}
```

## Mining Through the Network

Mine Node A’s pending transactions:

```bash
curl.exe -X POST http://localhost:8001/mine
```

The mined block is validated and forwarded to peers.

Check node status:

```bash
curl.exe http://localhost:8001/status
curl.exe http://localhost:8002/status
curl.exe http://localhost:8003/status
```

All synchronized nodes should show the same height and head hash.

After block acceptance, every pending pool should return:

```json
{"size":0}
```

## HTTP API

### Read endpoints

| Method | Endpoint | Purpose |
|---|---|---|
| `GET` | `/status` | Return chain height and head hash |
| `GET` | `/peers` | Return configured peers |
| `GET` | `/mempool` | Return pending-pool size |
| `GET` | `/balances` | Return confirmed balances |
| `GET` | `/chain` | Return the complete blockchain |
| `GET` | `/blocks?from=<index>` | Return blocks from an index |
| `GET` | `/nonce?address=<address>` | Return a sender’s next nonce |
| `GET` | `/merkle-proof?block=<index>&transaction=<id>` | Return a Merkle inclusion proof for one transaction |
| `GET` | `/peers` | Return known peers |
| `POST` | `/peers` | Register a node as a peer |

## Merkle Inclusion Proofs

Request proof that one transaction is included in a block:

```bash
curl.exe "http://localhost:8001/merkle-proof?block=1&transaction=<transaction-id>"
```

The response contains:

* The block and transaction indexes
* The complete transaction
* The block's Merkle root
* The neighboring hashes and their left/right positions
* A `verified` result

The proof allows the transaction to be checked against the block's Merkle root without downloading every transaction in that block. Changing any signed transaction field, including the amount or nonce, causes verification to fail.


### State-changing endpoints

| Method | Endpoint | Purpose |
|---|---|---|
| `POST` | `/transactions` | Receive and gossip a signed transaction |
| `POST` | `/blocks` | Receive and gossip an already-mined block |
| `POST` | `/mine` | Mine pending transactions and gossip the block |

HTTP data is encoded as JSON.

## Transaction Validation Flow

```text
Receive transaction
→ Recalculate transaction ID
→ Check de-duplication map
→ Verify public-key-derived address
→ Verify Ed25519 signature
→ Validate sender nonce
→ Validate available balance
→ Add to pending pool
→ Forward to peers
```

The transaction ID prevents repeated network forwarding. The transaction nonce prevents replaying an already-authorized payment.

## Block Validation Flow

```text
Receive block
→ Check block-hash de-duplication
→ Confirm index and previous hash
→ Recalculate Merkle root and block hash
→ Validate difficulty and Proof of Work
→ Verify transaction IDs, signatures and nonces
→ Validate balances
→ Append block
→ Remove confirmed pending transactions
→ Forward to peers
```

Received blocks are validated but are not mined again.

## Chain Synchronization

When a node starts, it asks a peer for its height.

If the peer is ahead on the same branch:

```text
Request missing blocks
→ Validate each block in order
→ Append each valid block
```

If the peer has a competing stronger branch:

```text
Download complete peer chain
→ Validate candidate chain
→ Compare cumulative Proof of Work
→ Find fork point
→ Adopt stronger valid chain
→ Rebuild balances and pending transactions
```

## Fork Resolution and Reorganization

The project selects a candidate chain only when it has more cumulative Proof of Work.

Estimated work per block is:

```text
16 ^ difficulty
```

During reorganization:

1. The last common block is identified.
2. Removed local blocks become orphaned.
3. The stronger candidate chain is validated and adopted.
4. Confirmed balances are recalculated from the selected chain.
5. Existing pending transactions are revalidated.
6. Valid orphaned transactions not confirmed in the new chain return to the pending pool.

Equal-work, weaker and invalid candidate chains are rejected.

## CLI Commands

```text
wallet <name>
add <sender> <receiver> <amount>
submit <node-address> <sender> <receiver> <amount>
mine
print
validate
balance
node -address <address> -peers <peer1,peer2>
```

The `add` and `mine` commands support the original local workflow.

The `submit`, HTTP `/mine`, and `node` commands support the networked workflow.

## Testing

Run normal tests:

```bash
go test ./...
```

Run static analysis:

```bash
go vet ./...
```

Run race detection:

```bash
go test -race ./...
```

On Windows, the race detector can be run through WSL:

```bash
cd /mnt/c/Users/<username>/toy-blockchain
go test -race ./...
```

Tests cover:

- Deterministic hashing
- Ed25519 signing and verification
- Address generation
- Transaction ID generation
- Nonce and replay protection
- Mining difficulty
- Merkle roots and inclusion proofs
- Tamper detection
- Pending double-spending
- Transaction de-duplication
- Transaction gossip
- Block gossip
- Missing-block synchronization
- Cumulative-work calculation
- Fork resolution
- Orphaned-transaction recovery
- HTTP introspection endpoints
- Race-free shared state

## Three-Node Experiment Result

A transaction was submitted to Node `8001`.

Before mining:

```text
Node 8001 mempool: 1
Node 8002 mempool: 1
Node 8003 mempool: 1
```

The transaction was mined through Node `8001`. All nodes reached height `1` and shared the same head hash:

```text
0000f6b4ae122cfa7c8701d01a177e5b8b309d3641792316ea2b7c221ad76293
```

After block propagation:

```text
Node 8001 mempool: 0
Node 8002 mempool: 0
Node 8003 mempool: 0
```

This demonstrates transaction gossip, block gossip, Proof-of-Work validation and network convergence.

## Changes from Assessment 1

Assessment 2 extends the original blockchain with:

- ECDSA replaced by Ed25519
- Public-key-derived addresses
- Deterministic transaction IDs
- Transaction nonces and replay protection
- HTTP node service
- Configurable peers and ports
- Transaction and block gossip
- De-duplication
- Race-free shared state
- New-node synchronization
- Missing-block downloads
- Cumulative-work chain selection
- Reorganization and orphan recovery
- Three-node cluster launcher

## Persistence

The original local CLI can save and load blockchain data using:

```text
chain.json
```

Generated chain and wallet files are excluded from version control.

The running network nodes primarily maintain independent in-memory state. Complete per-node restart persistence and separate per-node data directories are not implemented.

## Known Limitations

This is an educational blockchain, not a production cryptocurrency.

Current limitations include:


- No automatic peer-health management
- No encrypted peer communication
- No authentication for administrative endpoints such as `/mine`
- No transaction fees or mining rewards
- No smart contracts
- No production-grade private-key protection
- No Byzantine-fault-tolerant consensus
- Network nodes do not yet use separate persistent data directories

## Author

Developed as part of a Software Engineering Internship Golang assessment.