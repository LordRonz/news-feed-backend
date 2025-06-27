package queue

import (
	"encoding/json"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

var channel *amqp.Channel

func InitProducer(conn *amqp.Connection) error {
	ch, err := conn.Channel()
	if err != nil {
		return err
	}
	channel = ch

	return channel.ExchangeDeclare(
		"feed",
		"fanout",
		true,
		false,
		false,
		false,
		nil,
	)
}

type FeedEvent struct {
	ID      string `json:"id"`
	Content string `json:"content"`
	Author  string `json:"author"`
	Time    int64  `json:"time"`
}

func PublishFeedCreated(event FeedEvent) error {
	body, err := json.Marshal(event)
	if err != nil {
		return err
	}

	return channel.Publish(
		"feed",
		"",
		false,
		false,
		amqp.Publishing{
			ContentType: "application/json",
			Body:        body,
			Timestamp:   time.Now(),
		},
	)
}
