## Webhookr.go
An implementation of the original [Webhookr](http://webhookr.com/) by [Matt W.](https://github.com/mattwilliamson/webhookr) completely re-written in Go and powered by the [Gorilla WebSocket library](https://github.com/gorilla/websocket).
Test HTTP/S callbacks quickly by creating [new Webhookr](/new) and pointing your callback to it.
Webhookr.go collects no data, ever.

## How to run it
- `CGO_ENABLED=0 GOOS=linux go build -a -o webhookr .`
- `docker build -t webhookr .`
- `docker run -e HOST="localhost" -e PORT="8080" -p 8080:8080 webhookr`

