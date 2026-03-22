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

- **🧠 AI-Powered CLI:** Includes a custom CLI powered by Google's Gemini (`gemini-2.5-flash`), allowing you to interact with your database using pure natural language (e.g., *"create a vector for car with values 0.2 0.8 0.1"*).
- **📉 Vector Database Capabilities:** First-class support for storing high-dimensional vectors and discovering nearest neighbors via Cosine Similarity (`VSET`, `VSEARCH`).
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
| `VSET` | `VSET key float1 float2 ...` | Stores a multi-dimensional floating point vector associated with `key`. |
| `VSEARCH` | `VSEARCH float1 float2 ... [LIMIT n]`| Performs a Cosine Similarity search to find the top `n` most similar vectors stored in the DB. |

---

## 🏗️ Architecture

1. **Server (`server.go`)**: Acts as the central hub. It listens for incoming TCP connections and spawns a new `Peer` goroutine for each client. Channel pipelines (`msgCh`, `addPeerCh`, `delPeerCh`) safely transport messages to the main event loop, eliminating race conditions.
2. **Peers (`peer.go`)**: Manages individual client lifecycles. Reads byte streams, decodes RESP arrays, and forwards structured `Command` objects to the main server loop.
3. **KV Store (`kv.go`)**: The core in-memory map protected by an `RWMutex`. Now includes a secondary `vdata` map optimized for storing arrays of floats and executing fast Cosine Similarity math for vector queries.
4. **AOF Engine (`aof.go`)**: Manages the `dump.aof` file. It sequentially writes RESP-encoded commands and forces a disk sync (`fsync`) every second for persistence.
5. **Protocol Parser (`proto.go`)**: Translates RESP structures sent by clients into Go `Command` structs and vice versa.
6. **AI CLI (`client/ai_cli/main.go`)**: A unique, decoupled client application utilizing the Google GenAI SDK to transform natural English prompts directly into raw byte-stream RESP queries.

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

### Using the AI-Powered CLI (Gemini)

Don't want to type raw commands? You can talk to the database in plain English! 
Make sure you have exported your `Gemini_API_KEY`.

```bash
export Gemini_API_KEY="your-api-key-here"
cd client/ai_cli
go run main.go
```

```text
Welcome to AI powered Redis CLI 
Type 'exit' to quit.

Ai-CLI > create a vector for car with values 0.2 0.8 0.1
[Gemini Translated to RESP] -> "*5\r\n$4\r\nVSET\r\n$3\r\ncar\r\n$3\r\n0.2\r\n$3\r\n0.8\r\n$3\r\n0.1\r\n"
Server: "+OK\r\n"

Ai-CLI > find 1 vector most similar to 0.2 0.9 0.1
[Gemini Translated to RESP] -> "*6\r\n$7\r\nVSEARCH\r\n$3\r\n0.2\r\n$3\r\n0.9\r\n$3\r\n0.1\r\n$5\r\nLIMIT\r\n$1\r\n1\r\n"
Server: "*2\r\n$3\r\ncar\r\n$6\r\n0.9996\r\n"
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
