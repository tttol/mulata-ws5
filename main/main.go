package main

import (
	"fmt"
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/rs/cors"
	"github.com/tttol/mulata-ws5/aws"
)

func main() {
	runHttp()      // :8000
	runWebSocket() // :3001
	runTranslate() // :3002

	http.ListenAndServe(":8080", nil)
}

func runWebSocket() {
	http.HandleFunc("/ws", aws.HandleConnection)

	// Run WebSocket server in a goroutine
	go func() {
		slog.Info("WebSocket server started on :3001...")
		if err := http.ListenAndServe(":3001", nil); err != nil {
			slog.Error("ListenAndServe error on port 3001:", err)
		}
	}()
}

func runHttp() {
	ginEngine := gin.Default()
	ginEngine.StaticFile("/", "./html/index.html")
	ginEngine.StaticFS("/static", http.Dir("./static"))
	if err := ginEngine.Run(":8000"); err != nil {
		slog.Error("Failed to start Gin server:", err)
	}
	slog.Info("HTTP server started on :8000...")
}

func runTranslate() {
	mux := http.NewServeMux()
	mux.HandleFunc("/get/translate", func(w http.ResponseWriter, r *http.Request) {
		result, err := aws.GetResult()
		if err == nil {
			slog.Info("Success to get result")
			fmt.Fprint(w, result)
		} else {
			slog.Error("Failed to get result:", err)
			fmt.Fprint(w, err)
		}
	})

	handler := cors.Default().Handler(mux)

	http.ListenAndServe(":3002", handler)
}
