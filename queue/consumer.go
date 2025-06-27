package queue

import (
	"context"
	"encoding/json"
	"log"
	"os"
	"time"

	"github.com/lordronz/news-feed-backend/db"
	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/redis/go-redis/v9"
)

var redisClient *redis.Client

func InitRedis() {
	redisClient = redis.NewClient(&redis.Options{
		Addr: os.Getenv("REDIS_URL"),
	})
}

func StartConsumer(conn *amqp.Connection) error {
	ch, err := conn.Channel()
	if err != nil {
		return err
	}

	err = ch.ExchangeDeclare("feed", "fanout", true, false, false, false, nil)
	if err != nil {
		return err
	}

	q, err := ch.QueueDeclare("", false, true, true, false, nil)
	if err != nil {
		return err
	}

	err = ch.QueueBind(q.Name, "", "feed", false, nil)
	if err != nil {
		return err
	}

	msgs, err := ch.Consume(q.Name, "", true, true, false, false, nil)
	if err != nil {
		return err
	}

	go func() {
		ctx := context.Background()
		for msg := range msgs {
			var event FeedEvent
			if err := json.Unmarshal(msg.Body, &event); err != nil {
				log.Printf("❌ Failed to parse event: %v", err)
				continue
			}

			log.Printf("📥 Received feed.created event: %s", event.ID)

			err := db.InsertFeedEvent(ctx, event.FeedID(), event.Author, time.Unix(event.Time, 0))
			if err != nil {
				log.Printf("❌ Failed to log feed event: %v", err)
			}

			redisPayload, _ := json.Marshal(event)
			redisClient.Publish(ctx, "feeds:new", redisPayload)
		}
	}()

	return nil
}

func (f FeedEvent) FeedID() string {
	return f.ID
}
