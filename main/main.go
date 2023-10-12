package main

import (
	"log/slog"
	"net/http"

	"github.com/tttol/mulata-ws5/aws"
)

func main() {
	http.HandleFunc("/ws", aws.HandleConnection)

	slog.Info("WebSocket server started on :3001...")
	err := http.ListenAndServe(":3001", nil)
	if err != nil {
		slog.Error("ListenAndServe error on port=3001:", err)
	}
}
