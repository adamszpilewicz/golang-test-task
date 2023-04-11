package main

import (
	"context"
	"encoding/json"
	"github.com/gin-gonic/gin"
	amqp "github.com/rabbitmq/amqp091-go"
	"log"
)

const (
	apiPort  = "localhost:8088"
	ampqpUrl = "amqp://user:password@localhost:7001/"
)

// Message is the message object that will be sent between users
type Message struct {
	Sender   string `json:"sender"`
	Receiver string `json:"receiver"`
	Content  string `json:"message"`
}

// RabbitMQConnection is the connection to RabbitMQ
type RabbitMQConnection struct {
	conn *amqp.Connection
	ch   *amqp.Channel
	q    amqp.Queue
}

// Connect connects to RabbitMQ
func (r *RabbitMQConnection) Connect() error {
	var err error
	r.conn, err = amqp.Dial(ampqpUrl)
	if err != nil {
		return err
	}

	r.ch, err = r.conn.Channel()
	if err != nil {
		return err
	}

	r.q, err = r.ch.QueueDeclare(
		"message_queue",
		false,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return err
	}

	return nil
}

// Close closes the connection to RabbitMQ
func (r *RabbitMQConnection) Close() {
	r.ch.Close()
	r.conn.Close()
}

// Publish publishes a message to RabbitMQ
func (r *RabbitMQConnection) Publish(msg Message) error {
	body, err := json.Marshal(msg)
	if err != nil {
		return err
	}

	ctx := context.Background()
	err = r.ch.PublishWithContext(
		ctx,
		"",
		r.q.Name,
		false,
		false,
		amqp.Publishing{
			ContentType: "application/json",
			Body:        body,
		},
	)

	return err
}

func main() {
	rmq := RabbitMQConnection{}
	err := rmq.Connect()
	if err != nil {
		panic(err)
	}
	defer rmq.Close()

	r := gin.Default()
	r.POST("/message", func(c *gin.Context) {
		var msg Message
		if err := c.ShouldBindJSON(&msg); err != nil {
			c.JSON(400, gin.H{"error": err.Error()})
			return
		}

		err = rmq.Publish(msg)
		if err != nil {
			c.JSON(400, gin.H{"error": "Error publishing message"})
			return
		}

		c.JSON(200, gin.H{"status": "OK"})
	})

	log.Fatal(r.Run(apiPort))
}
