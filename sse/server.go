package sse

import (
	"context"
	"fmt"
	"log"
	"net/http"

	"github.com/redis/go-redis/v9"
)

var redisClient *redis.Client

func Init(redis *redis.Client) {
	redisClient = redis
}

type Client struct {
	ctx     context.Context
	writer  http.ResponseWriter
	flusher http.Flusher
}

func HandleSSE(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "Streaming unsupported", http.StatusInternalServerError)
		return
	}

	pubsub := redisClient.Subscribe(ctx, "feeds:new")
	defer pubsub.Close()

	fmt.Fprintf(w, "event: ping\ndata: ready\n\n")
	flusher.Flush()

	log.Println("📡 SSE client connected")

	for {
		select {
		case <-ctx.Done():
			log.Println("❌ SSE client disconnected")
			return
		case msg := <-pubsub.Channel():
			fmt.Fprintf(w, "data: %s\n\n", msg.Payload)
			flusher.Flush()
		}
	}
}
