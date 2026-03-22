<div align="center">
  <h1>🚀 My Redis Clone (Go)</h1>
  <p><strong>A high-performance, concurrent, in-memory key-value store built from scratch in Go.</strong></p>

  <p>
    <img src="https://img.shields.io/badge/Language-Go%201.21+-00ADD8?style=for-the-badge&logo=go" alt="Go Version" />
    <img src="https://img.shields.io/badge/Protocol-RESP-DC382D?style=for-the-badge&logo=redis" alt="Protocol" />
    <img src="https://img.shields.io/badge/Persistence-AOF-brightgreen?style=for-the-badge" alt="Storage" />
  </p>
</div>

---

## 📖 Overview

**My Redis Clone** is a lightweight, custom implementation of a Redis-like server written entirely in Go. It handles raw TCP connections and implements the **Redis Serialization Protocol (RESP)**, making it compatible with standard Redis clients like `redis-cli`. 

Designed for concurrency and safety, it uses Go's powerful concurrency model (goroutines and channels) along with `sync.RWMutex` to efficiently manage multiple client connections and thread-safe data access. It also guarantees data durability through **Append-Only File (AOF)** persistence with background `fsync`.

---

## ✨ Key Features

- **⚡ High Concurrency:** Handles numerous simultaneous client connections effortlessly using goroutines and safe channel-based pipelines.
- **🔄 RESP Compatibility:** Fully parses and serializes data using RESP, functioning seamlessly with standard Redis tools and clients.
- **🔒 Thread-Safe Operations:** Utilizes `sync.RWMutex` to ensure safe concurrent reads and exclusive writes to the in-memory data store.
- **💾 Data Persistence (AOF):** 
  - Writes every mutating command to an Append-Only File (`dump.aof`).
  - Employs a background goroutine to `fsync` data to disk every second, minimizing the risk of data loss.
  - Automatically reconstructs the database state from the AOF on startup.
- **⏱️ Key Expiration (TTL):** Supports setting expiration times on keys (using the `EX` argument), with automatic background cleanup using `time.AfterFunc`.

---

## 🛠️ Supported Commands

Currently, the server supports the following core Redis commands:

| Command | Usage | Description |
| :--- | :--- | :--- |
| `SET` | `SET key value [EX seconds]` | Stores the `key` and `value`. Optional `EX` sets a time-to-live in seconds. |
| `GET` | `GET key` | Retrieves the value associated with the `key`. Returns `nil` if not found. |
| `DEL` | `DEL key` | Deletes the specified `key` from the store. |

---

## 🏗️ Architecture Architecture

1. **Server (`server.go`)**: Acts as the central hub. It listens for incoming TCP connections and spawns a new `Peer` goroutine for each client. Channel pipelines (`msgCh`, `addPeerCh`, `delPeerCh`) safely transport messages to the main event loop, eliminating race conditions.
2. **Peers (`peer.go`)**: Manages individual client lifecycles. Reads byte streams, decodes RESP arrays, and forwards structured `Command` objects to the main server loop.
3. **KV Store (`kv.go`)**: The core in-memory map protected by an `RWMutex`, allowing multiple simultaneous reads but isolated writes.
4. **AOF Engine (`aof.go`)**: Manages the `dump.aof` file. It sequentially writes RESP-encoded commands and forces a disk sync (`fsync`) every second for persistence.
5. **Protocol Parser (`proto.go`)**: Translates RESP structures sent by clients into Go `Command` structs and vice versa.

---

## 🚦 Getting Started

### Prerequisites
- [Go](https://go.dev/dl/) installed on your machine (version 1.21+ recommended).
- A standard Redis client like `redis-cli` (optional, for testing).

### Installation & Execution

1. **Clone the repository:**
   ```bash
   git clone <your-repo-url>
   cd my-redis-clone
   ```
2. **Download Dependencies:**
   ```bash
   go mod tidy
   ```
3. **Run the server:**
   ```bash
   go run .
   ```
   *The server will start listening on `0.0.0.0:5001` or `localhost:5001` by default.*

---

## 💻 Usage Examples

Once the server is running, you can connect to it using `redis-cli` or `netcat`/`telnet`.

### Using `redis-cli`

Connect to the server on port `5001`:

```bash
redis-cli -p 5001
```

Execute commands:

```redis
127.0.0.1:5001> SET mykey "Hello, World!"
+Ok
127.0.0.1:5001> GET mykey
"Hello, World!"
127.0.0.1:5001> SET temp_key "This will vanish" EX 5
+Ok
127.0.0.1:5001> DEL mykey
:1
```

### Using `nc` (Netcat)

If you understand RESP, you can communicate directly via TCP:

```bash
nc localhost 5001
*2
$3
GET
$5
mykey
$-1
```

---

## 🤔 Future Improvements

- Implement the `Passive` & `Active` background key expiration strategies for massive datasets to preserve CPU/RAM.
- Provide support for more complex data structures like Hashes, Lists, and Sets.
- Add Master-Slave replication.
- Create more robust AOF rewriting to shrink file size over time.

---

<p align="center">
  <i>Built with ❤️ in Go</i>
</p>
