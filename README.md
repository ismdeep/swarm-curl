# swarm-curl

Distributed collaborative download tool

# 1 Architecture

- **swarm-curl**: Client that mimics curl for file downloads
- **swarm-curl-daemon**: Server that receives download requests and returns file chunks

# 2 Configuration

Create a configuration file at `~/.swarm-curl/config.yaml`:

```
endpoints:
  - address: http://localhost:8001
    token: token1
  - address: http://localhost:8002
    token: token2
```

# 3 Build

```
make build
```

# 4 Install

```
make install
```

# 5 Usage

Start the daemon:

```bash
./swarm-curl-daemon --addr :8001 --token token1
./swarm-curl-daemon --addr :8002 --token token2
```

Download a file:

```bash
./swarm-curl https://example.com/file.zip -o file.zip
```

