FROM docker.1ms.run/library/golang:1.26.0-bookworm AS builder
WORKDIR /app
COPY . .
RUN make build

FROM docker.1ms.run/library/debian:13
ENV SWARM_CURL_ADDR=0.0.0.0:8080
RUN apt update && apt install -y ca-certificates
WORKDIR /app
COPY --from=builder /app/build/swarm-curl-daemon /app/swarm-curl-daemon
EXPOSE 8080
ENTRYPOINT ["/app/swarm-curl-daemon"]
