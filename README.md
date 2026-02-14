# swarm-curl

Distributed collaborative download tool

# 1 Architecture

- **swarm-curl**: Client that mimics curl for file downloads
- **swarm-curl-daemon**: Server that receives download requests and returns file chunks

# 2 Configuration

Create a configuration file at `~/.swarm-curl/config.json`:

```json
{
  "endpoints": [
    {
      "address": "http://host1:8001",
      "token": "your-token-1"
    },
    {
      "address": "http://host2:8002",
      "token": "your-token-2"
    }
  ]
}
```

# 3 Build

```bash
go build -o swarm-curl ./cmd/swarm-curl
go build -o swarm-curl-daemon ./cmd/swarm-curl-daemon
```

# 4 Usage

Start the daemon:

```bash
./swarm-curl-daemon :8001 token1
./swarm-curl-daemon :8002 token2
```

Download a file:

```bash
./swarm-curl https://example.com/file.zip -o file.zip
```

