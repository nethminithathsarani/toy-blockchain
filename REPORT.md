# Networked Toy Blockchain

## Golang Developer Assessment 2 Report

## 1. Introduction

This project is a networked toy blockchain developed using Go 1.22+ as part of a Software Engineering Internship assessment.

Assessment 2 extends the original single-process blockchain and ledger simulator into a network of independently running HTTP nodes. The system supports Ed25519-signed transactions, transaction and block gossip, message de-duplication, blockchain synchronization, cumulative Proof-of-Work fork selection, chain reorganization, orphaned-transaction recovery and race-free shared state.

The main objective was to understand how separate blockchain nodes exchange, validate and agree on distributed state while continuing to treat all received network data as untrusted.

## 2. Changes from Assessment 1

The original project already supported:

- SHA-256 block hashing
- Deterministic genesis creation
- Proof-of-Work mining
- Concurrent mining
- Merkle roots
- Pending transactions
- Ledger and balance validation
- Double-spending prevention
- Difficulty adjustment
- Local fork resolution
- JSON persistence
- CLI commands and automated tests

Assessment 2 introduced the following major changes:

- ECDSA was replaced with Ed25519.
- Sender addresses are derived from public keys.
- Important transaction fields are signed using a deterministic JSON payload.
- Per-sender transaction nonces prevent replay attacks.
- Deterministic transaction IDs support network de-duplication.
- The blockchain runs as an HTTP node with configurable peers.
- Transactions and blocks are gossiped between nodes.
- Nodes synchronize missing blocks from peers.
- Competing branches are compared using cumulative Proof-of-Work.
- Reorganization rebuilds confirmed balances and the pending pool.
- Valid transactions from orphaned blocks return to the pending pool.
- Shared node state is protected using `sync.RWMutex`.
- A PowerShell launcher starts a three-node local cluster.
- A multi-stage Docker image and Docker Compose launcher start the three-node network with one command.
- Separate listen and advertised addresses support communication through Docker service names.

## 3. System Design

### 3.1 Node architecture

Each running node contains:

```text
Node configuration
Blockchain and pending pool
Known transaction IDs
Known block hashes
Read/write mutex
HTTP router and handlers
```

The configuration contains the node’s listening address and initial peer list.
The initial list acts as a seed list. Nodes request peer lists from their seeds, recursively add new unique peers and register their own address with every contacted peer.

For example:

```text
Node A: localhost:8001
Node B: localhost:8002
Node C: localhost:8003
```

Every node keeps its own blockchain and pending pool in memory. Nodes communicate using JSON over HTTP.

### 3.2 HTTP wire format and endpoints

Transactions, blocks and API responses are encoded as JSON using Go’s standard `encoding/json` package.

The main read endpoints are:

| Method | Endpoint | Purpose |
|---|---|---|
| GET | `/status` | Return height and head hash |
| GET | `/peers` | Return known peers |
| GET | `/mempool` | Return pending-pool size |
| GET | `/balances` | Return confirmed balances |
| GET | `/chain` | Return the complete chain |
| GET | `/blocks?from=n` | Return blocks beginning at index `n` |
| GET | `/nonce?address=x` | Return the sender’s next nonce |

The state-changing endpoints are:

| Method | Endpoint | Purpose |
|---|---|---|
| POST | `/transactions` | Receive a signed transaction |
| POST | `/blocks` | Receive an already-mined block |
| POST | `/mine` | Mine the running node’s pending transactions |
| GET | `/merkle-proof?block=<index>&transaction=<id>` | Return a Merkle inclusion proof for one transaction |
| POST | `/peers` | Register a node as a peer |

The custom `X-Node-Address` header identifies the peer that forwarded a transaction or block. The receiver excludes that peer when forwarding, preventing an immediate return to the sender.

## 4. Transaction Security

### 4.1 Ed25519 wallets

Each wallet contains an Ed25519 public/private key pair.

- The private key signs transactions and must remain secret.
- The public key verifies signatures and may be shared.
- The sender address is the SHA-256 hash of the public key.
- Wallet files use restrictive file permissions where supported.

### 4.2 Signing payload

The deterministic signing payload contains:

```text
Sender
Sender address
Receiver
Amount
Nonce
```

The payload is encoded using a fixed Go structure and JSON field order. The private key signs those bytes.

If the receiver, amount, address or nonce changes afterward, the recreated payload differs and Ed25519 verification fails.

### 4.3 Transaction nonce

The transaction nonce increases independently for every sender:

```text
Alice transaction 1 → nonce 1
Alice transaction 2 → nonce 2
Bob transaction 1   → nonce 1
```

The nonce prevents an exact signed transaction from being processed repeatedly. Because it is included in the signed payload, an attacker cannot change the nonce without invalidating the signature.

The transaction nonce is different from the mining nonce. A mining nonce changes repeatedly until the block hash satisfies Proof of Work.

### 4.4 Transaction ID

A transaction ID is calculated using SHA-256 over the complete signed transaction data:

```text
Sender
Sender address
Receiver
Amount
Nonce
Public key
Signature
```

The ID is represented as 64 hexadecimal characters.

Every receiving node recalculates the ID instead of trusting the received `ID` field. The same signed transaction produces the same ID on every node.

The nonce prevents replayed spending, while the transaction ID supports network de-duplication.

## 5. Gossip and De-duplication

### 5.1 Transaction gossip

Transaction handling follows this process:

```text
Receive transaction JSON
→ Recalculate transaction ID
→ Check whether ID is already known
→ Verify sender address and Ed25519 signature
→ Validate nonce and balance
→ Add transaction to pending pool
→ Remember transaction ID
→ Forward to peers except the sender
```

Invalid and duplicate transactions are not forwarded.

### 5.2 Block gossip

A block is identified using its block hash.

A received block is already mined, so the receiving node does not repeat Proof of Work. Instead, it:

```text
Checks duplicate block hash
→ Builds a temporary candidate chain
→ Validates the complete chain
→ Appends the valid block
→ Removes its confirmed transactions from pending
→ Remembers block and transaction IDs
→ Forwards it to peers
```

Blocks that do not extend the current head trigger synchronization or fork-resolution logic instead of being blindly appended.

### 5.3 Gossip message cost

The gossip-cost experiment used three fully connected nodes. Each node knew the other two nodes as peers. Before the experiment, all nodes had the same genesis block and an empty pending pool.

One signed transaction was submitted to Node A:

```text
Transaction ID:
e9313e9f7687d56d74e635491f0ff5e6313f2151b86829f1242a86c0b747b3ff
```

The logs showed the following propagation:

```text
Client → Node A: initial submission
Node A → Node B: accepted
Node B → Node C: accepted
Node C → Node A: duplicate ignored
Node A → Node C: duplicate ignored
```

The measured message counts were:

| Measurement                     | Count |
| ------------------------------- | ----: |
| Initial client submission       |     1 |
| Peer-to-peer gossip POSTs       |     4 |
| Total transaction POST requests |     5 |
| Nodes accepting the transaction |     3 |
| Duplicate deliveries detected   |     2 |

Each node added the transaction exactly once. Although two redundant deliveries occurred, neither duplicate was added or forwarded again.

Every node recalculated the deterministic transaction ID and checked its `seenTransactions` map before adding or forwarding the transaction. Once an ID was already known, the node returned a duplicate response and stopped processing it. This prevented the transaction from circulating endlessly.

The experiment demonstrates that the simple full-mesh approach creates some redundant traffic, but de-duplication bounds the propagation and prevents repeated state changes. A larger network could reduce redundant messages using time-to-live values, randomized peer selection or more complete message-origin tracking.


## 6. Blockchain Synchronization

A new or behind node first requests a peer’s `/status` endpoint.

If the peer is ahead on the same branch, the local node requests:

```http
GET /blocks?from=<local-height+1>
```

Each missing block is validated and appended in order. This is necessary because every block’s `PreviousHash` depends on the block before it.

The synchronization flow is:

```text
Request peer height
→ Compare local height and head
→ Download missing blocks
→ Validate each block
→ Append each valid block
→ Confirm equal final height
```

If missing blocks cannot connect, the peer may be on a competing branch. The local node then downloads the complete peer chain through `/chain` and starts fork resolution.

At startup, a node tries its configured peers. If one peer is offline, it tries the next. If synchronization fails completely, the server still starts because peers may only be temporarily unavailable.

## 7. Fork Resolution and Reorganization

### 7.1 Cumulative Proof-of-Work

The original project used only chain length. Assessment 2 compares cumulative Proof-of-Work.

For hexadecimal difficulty `d`, estimated block work is:

\[
16^d
\]

Total chain work is the sum of all block-work values.

A candidate is considered only when:

```text
Candidate cumulative work > Local cumulative work
```

Equal-work and weaker chains are rejected. This avoids unnecessary switching between equally strong branches.

The candidate must still pass:

- Shared genesis validation
- Merkle-root verification
- Block-hash verification
- Previous-hash linkage
- Index and timestamp checks
- Required difficulty checks
- Proof-of-Work checks
- Transaction ID and signature checks
- Nonce validation
- Balance validation

A chain cannot win merely by claiming a high difficulty because validation checks whether that difficulty was actually required and whether its hash satisfies it.

### 7.2 Finding the fork point

The fork point is the last block shared by both branches.

Example:

```text
Local:     Genesis → Block 1 → Block 2A
Candidate: Genesis → Block 1 → Block 2B → Block 3B
```

The fork point is Block 1. Block 2A becomes orphaned when the candidate is selected.

### 7.3 Ledger rebuilding

Balances are not copied from the old branch. They are recalculated using:

```text
Initial balances
→ Replay transactions from the selected chain
→ Produce confirmed balances
```

Transactions in orphaned blocks no longer affect confirmed balances.

The node then rebuilds its pending pool:

1. Save existing pending transactions.
2. Collect transactions from orphaned blocks.
3. Remove transactions already confirmed in the selected chain.
4. Remove duplicates.
5. Revalidate signatures, IDs, nonces and balances.
6. Return only valid transactions to pending.

This prevents valid orphaned transactions from being silently lost while also preventing invalid or already-confirmed transactions from being processed twice.

## 8. Race-Free Shared State

Go’s HTTP server can handle several requests concurrently. For example, one request may read balances while another receives a transaction or mines a block.

The node protects shared state using `sync.RWMutex`.

Read locks are used for:

- Status
- Mempool size
- Balances
- Chain dump
- Missing-block responses
- Nonce lookup

Write locks are used for:

- Adding transactions
- Accepting blocks
- Mining
- Synchronization
- Fork resolution
- Rebuilding de-duplication maps

Network forwarding is performed after releasing the write lock. Otherwise, an offline peer could keep the blockchain locked during an HTTP timeout.

The project was tested using:

```bash
go test -race ./...
```

No data races were reported.

## 9. Experiments and Results

### 9.1 Three-node transaction gossip

The cluster launcher started:

```text
Node A → localhost:8001
Node B → localhost:8002
Node C → localhost:8003
```

A signed transaction was submitted to Node A:

```bash
go run . submit localhost:8001 Alice Bob 20
```

Before mining:

| Node | Pending transactions |
|---|---:|
| Node 8001 | 1 |
| Node 8002 | 1 |
| Node 8003 | 1 |

This demonstrated that a transaction submitted to one node propagated to every node.

### 9.2 Block propagation

Mining was requested from Node A:

```bash
curl.exe -X POST http://localhost:8001/mine
```

The mined block had:

```text
Index: 1
Difficulty: 4
Nonce: 131135
Hash: 0000f6b4ae122cfa7c8701d01a177e5b8b309d3641792316ea2b7c221ad76293
```

All nodes reported:

```text
Height: 1
Head hash: 0000f6b4ae122cfa7c8701d01a177e5b8b309d3641792316ea2b7c221ad76293
```

After propagation:

| Node | Pending transactions |
|---|---:|
| Node 8001 | 0 |
| Node 8002 | 0 |
| Node 8003 | 0 |

This demonstrated mining, block gossip, independent validation and network convergence.

### 9.3 New-node synchronization

The synchronization test created:

```text
Peer node:  Genesis → Block 1
Local node: Genesis
```

The local node requested the peer’s height and downloaded Block 1. After validation, both nodes had the same height and head hash.

### 9.4 Live fork and reorganization experiment

Two independent nodes were started without configured peers:

```text
Node A: localhost:8001
Node B: localhost:8002
```

Because the nodes were disconnected, a different transaction was submitted and mined on each node:

```text
Node A Block 1:
Alice → Bob, amount 20
Hash: 0000a4c43efe3a796c99787fa7713cee43f8847ac99f79f95cc7270a758d9b22

Node B Block 1:
Charlie → Bob, amount 10
Hash: 0000fa6a20a3602eacde188774b4b9df9b027bbab6867e5fb094f22b3bbf0dab
```

Both nodes were at height 1, but their different head hashes confirmed that a fork had formed.

Another transaction was then submitted and mined on Node B:

```text
Charlie → Alice, amount 5

Node B Block 2 hash:
0000a147dcf5120002ece47b63be19874fa915a09691b17fffe4c97e498db123
```

The chains were then:

```text
Node A: Genesis → Block 1A
Node B: Genesis → Block 1B → Block 2B
```

Node B therefore had greater cumulative Proof of Work. Its latest block was sent to Node A with Node B’s address in the `X-Node-Address` header, simulating reconnection.

Node A produced the following relevant logs:

```text
Competing block received; synchronizing with: localhost:8002
Synchronization started with: localhost:8002
Missing blocks downloaded: 1 from localhost:8002
Competing chain detected from: localhost:8002
Chain reorganization completed: fork point 0 orphaned transactions returned 1
Stronger chain adopted from: localhost:8002
```

Node A could not append Block 2B directly because it did not contain Block 1B. It therefore downloaded Node B’s complete chain, validated it, compared cumulative Proof of Work, found the fork point at genesis and adopted Node B’s stronger chain.

Block 1A was orphaned. Its Alice-to-Bob transaction was still valid on the selected chain, so it was returned to Node A’s pending pool:

```text
Node A mempool size: 1
```

After reorganization, both nodes reported height 2 and the same head hash:

```text
Node A height: 2
Node B height: 2

Shared head:
0000a147dcf5120002ece47b63be19874fa915a09691b17fffe4c97e498db123

Head-hash equality check: true
```

This experiment demonstrated automatic fork detection, synchronization, cumulative-work chain selection, chain reorganization, orphaned-transaction recovery and final network convergence.


### 9.5 Test results

The following checks completed successfully:

```bash
gofmt -w .
go vet ./...
go test ./...
go test -race ./...
```

Tests cover:

- Ed25519 signing and tamper rejection
- Deterministic addresses and transaction IDs
- Nonce replay protection
- Pending double-spending
- Transaction gossip and de-duplication
- Block gossip
- HTTP introspection
- Missing-block synchronization
- Cumulative-work calculation
- Fork-point detection
- Reorganization
- Orphaned-transaction recovery
- Race-free operation
### 9.6 Peer-discovery experiment

Three nodes were started with the following initial configuration:

```text
Node A: seed Node B
Node B: seed Node A
Node C: seed Node A only
```

Node C requested Node A's peer list and produced:

```text
Peer discovered: localhost:8002
```

Node C then registered its own address with Node A and the newly discovered Node B. The final `/peers` responses were:

```text
Node A: [localhost:8002 localhost:8003]
Node B: [localhost:8001 localhost:8003]
Node C: [localhost:8001 localhost:8002]
```

This demonstrated that a new node can join through one seed address and that two-way registration allows every node to learn about the new participant. Duplicate peers and each node's own address are rejected. Peer-list snapshots and updates are protected by `sync.RWMutex`, and the complete suite passed the Go race detector.
### 9.7 Docker Compose cluster experiment

A multi-stage Dockerfile was used to compile the Go application and copy only the final executable into a small Alpine runtime image. The application runs as a non-root user. A `.dockerignore` file excludes private wallets, generated chains and Git history from the build context.

The complete cluster was started with:

```bash
docker compose up --build -d
```

Docker Compose started three healthy containers and exposed their APIs through host ports 8001, 8002 and 8003. Each node listened on `0.0.0.0` inside its container while advertising its Docker DNS address, such as `node-a:8001`, to peers.

Node C started with only Node A as its seed and discovered Node B:

```text
Node A: [node-b:8002 node-c:8003]
Node B: [node-a:8001 node-c:8003]
Node C: [node-a:8001 node-b:8002]
```

A signed Alice-to-Bob transaction was submitted to Node A. Before mining, all three pending pools contained one transaction. Node A then mined the transaction, and block gossip caused all nodes to reach height 1 with the same head hash:

```text
0000de05a376990354bc99c03fc1c7299e5b0c82...
```

The equality check across all three head hashes returned `true`. This demonstrated that the one-command container cluster supports peer discovery, transaction gossip, mining, block gossip and network convergence.

## 10. Challenges and Solutions

### Challenge 1: Distinguishing the two nonces

The mining nonce and transaction nonce initially appeared to conflict.

The solution was to treat them separately:

- Block nonce: repeatedly changed during Proof of Work.
- Transaction nonce: incremented once per sender transaction to prevent replay.

### Challenge 2: Verifying transactions on other nodes

The original verification depended on a locally stored wallet. Other network nodes should not possess the sender’s private wallet.

Verification was changed to use only public transaction information:

- Public key
- Derived sender address
- Signature
- Signed payload

### Challenge 3: Preventing gossip loops

Nodes could repeatedly forward the same transaction or block.

The solution was:

- Deterministic transaction IDs
- Block-hash identifiers
- Per-node known-ID maps
- Sender-node headers
- Duplicate detection before forwarding

### Challenge 4: Safely receiving mined blocks

Calling the original `AddBlock()` would mine another block instead of accepting the received block.

A separate `AddReceivedBlock()` path was implemented. It validates an already-mined block using a temporary candidate chain before modifying local state.

### Challenge 5: Recovering from competing branches

Missing-block synchronization alone fails when a peer’s new block references a different parent.

The node now falls back to downloading the full peer chain, comparing cumulative work and performing a reorganization.

### Challenge 6: Avoiding data races and deadlocks

HTTP handlers access shared blockchain state concurrently. Incorrect lock placement caused risks such as double unlock or calling a locking function while already holding the same lock.

The solution was to use clear lock paths:

```text
Lock
→ Modify shared state
→ Unlock exactly once
→ Perform network calls
```

Race detection was used to verify the final implementation.
## 11. Discussion Questions

### 11.1 Why agreement is probabilistic and what a 51 percent attack is

The cumulative Proof-of-Work rule allows independent nodes to select the valid chain containing the most total mining work. Normally, honest miners continue adding blocks to the same chain, causing it to become increasingly difficult for a competing branch to catch up.

However, agreement is probabilistic rather than immediately permanent. Two miners may find different valid blocks at nearly the same time, temporarily creating two branches. Nodes may initially receive and follow different branches. The network converges when one branch accumulates more Proof of Work, but there remains a decreasing probability that a competing branch could later overtake it.

A 51 percent attack occurs when one miner or coordinated group controls most of the network’s total mining power. The attacker could build a private chain faster than the honest network and later publish it as the stronger chain. This could reorganize recent blocks and allow the attacker to reverse their own transactions, enabling double-spending. However, control of mining power does not allow the attacker to forge another user’s Ed25519 signature or spend funds without the corresponding private key.

In this small toy network, a 51 percent attack is easier because there are very few miners. The implementation defends against invalid chains by checking hashes, Proof of Work, transaction signatures, nonces and balances, but it cannot prevent a majority of mining power from producing a stronger valid chain.
### 11.2 Finality and confirmations

Finality describes the level of certainty that an accepted transaction will remain permanently recorded in the blockchain. Hard finality means that once a transaction is finalized, it cannot later be reversed.

This small Proof-of-Work blockchain does not provide hard finality. A transaction may be included in the current chain but later removed if the network receives and adopts a competing valid chain with greater cumulative Proof of Work. The removed block becomes orphaned, and its transactions may return to the pending pool if they remain valid.

Proof-of-Work networks reduce this risk by waiting for confirmations. A transaction has its first confirmation when it is included in a mined block. Each additional block built on top of that block adds another confirmation. Replacing an older transaction would require rebuilding its block and all the work performed after it.

Therefore, more confirmations make a successful reorganization increasingly unlikely. However, confirmations provide probabilistic confidence rather than an absolute guarantee. In this toy network, the protection is much weaker than in a large real network because there are only a few miners and comparatively little total mining power.
### 11.3 Signature protection and malicious peers

Ed25519 signatures allow nodes to verify that a transaction was authorized by the holder of the corresponding private key. Because the sender address, receiver, amount and transaction nonce are included in the signed payload, changing any of these fields after signing causes verification to fail. Signatures therefore prevent transaction forgery and unauthorized modification.

However, signatures do not prove that a peer is honest. A malicious peer could repeatedly send an old valid transaction in an attempt to process the same payment more than once. The signature would still be valid because the transaction was originally authorized.

The node defends against this replay attempt using transaction nonces and deterministic transaction IDs. The nonce must be the next expected value for the sender, so an already-confirmed transaction has an old nonce and is rejected. The transaction ID and de-duplication map also prevent the same transaction from being repeatedly added or forwarded through the network.

A malicious peer could also send invalid blocks or a fabricated chain. Each node independently verifies the block hash, previous-hash link, Merkle root, required difficulty, Proof of Work, transaction signatures, nonces and resulting balances. A candidate chain is adopted only when it is fully valid and contains more cumulative Proof of Work than the local chain.

## 12. Known Limitations

The implementation is educational and is not suitable for production use.

Limitations include:


- Unreachable peers are not permanently removed.
- HTTP communication is not encrypted.
- Administrative endpoints such as `/mine` are unauthenticated.
- Wallet files are not password-encrypted.
- There are no transaction fees or mining rewards.
- There is no smart-contract execution.
- Network nodes primarily keep independent in-memory state.
- Separate persistent data directories and automatic restart restoration are not implemented.
- The simple full-mesh gossip strategy creates redundant messages.
- There is no finality rule or Byzantine-fault-tolerant consensus.
- The project is designed for a small local trusted network.

## 13. Conclusion

The project successfully extended a local blockchain simulator into a networked multi-node blockchain.

The final system supports Ed25519-secured transactions, public-key-derived addresses, transaction nonces, deterministic transaction IDs, HTTP nodes, gossip, de-duplication, network mining, missing-block synchronization, cumulative-work fork resolution, reorganization and orphaned-transaction recovery.

The three-node experiment demonstrated that transactions and blocks propagate across the network and that all nodes converge to the same blockchain head. Automated synchronization and fork tests demonstrated catch-up behavior and stronger-chain adoption. Go race detection confirmed that shared network state is accessed safely.

The project provided practical experience with Go networking, JSON APIs, concurrency control, cryptographic validation, distributed-state synchronization and blockchain reorganization.

## References

1. The Go Authors, “Package net/http”: https://pkg.go.dev/net/http

2. The Go Authors, “Package crypto/ed25519”: https://pkg.go.dev/crypto/ed25519

3. The Go Authors, “Data Race Detector”: https://go.dev/doc/articles/race_detector

4. Satoshi Nakamoto, “Bitcoin: A Peer-to-Peer Electronic Cash System”: https://bitcoin.org/bitcoin.pdf

5. RFC 8032, “Edwards-Curve Digital Signature Algorithm (EdDSA)”: https://www.rfc-editor.org/rfc/rfc8032