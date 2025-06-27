package main

import (
	"log"
	"net/http"
	"os"

	"github.com/go-chi/chi/v5"
	"github.com/joho/godotenv"
	"github.com/lordronz/news-feed-backend/db"
	"github.com/lordronz/news-feed-backend/handler"
	"github.com/lordronz/news-feed-backend/queue"
	"github.com/lordronz/news-feed-backend/sse"
	"github.com/rabbitmq/amqp091-go"
	"github.com/redis/go-redis/v9"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}
	if err = db.Connect(os.Getenv("DATABASE_URL")); err != nil {
		log.Fatal("DB connect error:", err)
	}

	redis := redis.NewClient(&redis.Options{
		Addr: os.Getenv("REDIS_URL"),
	})
	sse.Init(redis)
	queue.InitRedis()

	rmqConn, err := amqp091.Dial(os.Getenv("RABBITMQ_URL"))
	if err != nil {
		log.Fatalf("failed to connect to RabbitMQ: %v", err)
	}
	defer rmqConn.Close()

	if err := queue.InitProducer(rmqConn); err != nil {
		log.Fatalf("producer init error: %v", err)
	}
	if err := queue.StartConsumer(rmqConn); err != nil {
		log.Fatalf("consumer init error: %v", err)
	}

	r := chi.NewRouter()
	r.Get("/events", sse.HandleSSE)
	r.Get("/feeds", handler.GetFeeds)
	r.Post("/feeds", handler.CreateFeed)

	log.Println("Listening on :8080")
	http.ListenAndServe(":8080", r)
}
