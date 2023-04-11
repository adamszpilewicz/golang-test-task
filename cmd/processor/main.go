package main

import (
	"context"
	"encoding/json"
	"github.com/go-redis/redis/v8"
	amqp "github.com/rabbitmq/amqp091-go"
	"log"
)

const (
	amqpURL          = "amqp://user:password@localhost:7001/"
	redisAddr        = "localhost:6379"
	redisPassword    = ""
	redisDB          = 0
	messageQueueName = "message_queue"
)

type Message struct {
	Sender   string `json:"sender"`
	Receiver string `json:"receiver"`
	Content  string `json:"message"`
}

type MessageProcessor struct {
	conn *amqp.Connection
	ch   *amqp.Channel
	q    amqp.Queue
	rdb  *redis.Client
}

func (mp *MessageProcessor) Connect() error {
	var err error
	mp.conn, err = amqp.Dial(amqpURL)
	if err != nil {
		return err
	}

	mp.ch, err = mp.conn.Channel()
	if err != nil {
		return err
	}

	mp.q, err = mp.ch.QueueDeclare(
		messageQueueName,
		false,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return err
	}

	mp.rdb = redis.NewClient(&redis.Options{
		Addr:     redisAddr,
		Password: redisPassword,
		DB:       redisDB,
	})

	return nil
}

func (mp *MessageProcessor) Close() {
	mp.ch.Close()
	mp.conn.Close()
	mp.rdb.Close()
}

func (mp *MessageProcessor) ProcessMessages() (<-chan amqp.Delivery, error) {
	return mp.ch.Consume(
		mp.q.Name,
		"",
		true,
		false,
		false,
		false,
		nil,
	)
}

func (mp *MessageProcessor) SaveMessageToRedis(msg Message) error {
	key := msg.Sender + "_" + msg.Receiver
	value, _ := json.Marshal(msg)

	log.Printf("Saving message to Redis: %s", value)

	return mp.rdb.LPush(context.Background(), key, value).Err()
}

func main() {
	mp := MessageProcessor{}
	err := mp.Connect()
	if err != nil {
		panic(err)
	}
	defer mp.Close()

	msgs, err := mp.ProcessMessages()
	if err != nil {
		panic(err)
	}

	go func() {
		for d := range msgs {
			var msg Message
			err := json.Unmarshal(d.Body, &msg)
			if err != nil {
				log.Printf("Error decoding message: %s", err)
				continue
			}

			err = mp.SaveMessageToRedis(msg)
			if err != nil {
				log.Printf("Error saving message to Redis: %s", err)
				continue
			}
		}
	}()

	log.Printf("MessageProcessor is running. Press CTRL+C to exit.")
	forever := make(chan bool)
	<-forever
}
