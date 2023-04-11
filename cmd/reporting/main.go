package main

import (
	"context"
	"encoding/json"
	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	"log"
)

const (
	apiPort       = "localhost:8089"
	redisAddr     = "localhost:6379"
	redisPassword = ""
	redisDB       = 0
)

// Message is the message object that will be sent between users
type Message struct {
	Sender   string `json:"sender"`
	Receiver string `json:"receiver"`
	Content  string `json:"message"`
}

// ReportingAPI is the API that will be used to retrieve messages
type ReportingAPI struct {
	rdb *redis.Client
}

// Connect connects to Redis
func (ra *ReportingAPI) Connect() {
	ra.rdb = redis.NewClient(&redis.Options{
		Addr:     redisAddr,
		Password: redisPassword,
		DB:       redisDB,
	})
}

// Close closes the connection to Redis
func (ra *ReportingAPI) Close() {
	ra.rdb.Close()
}

// GetMessageList returns a list of messages between two users
func (ra *ReportingAPI) GetMessageList(sender, receiver string) ([]Message, error) {
	key := sender + "_" + receiver
	messages, err := ra.rdb.LRange(context.Background(), key, 0, -1).Result()
	if err != nil {
		return nil, err
	}

	var response []Message
	for _, msg := range messages {
		var messageObj Message
		err = json.Unmarshal([]byte(msg), &messageObj)
		if err == nil {
			response = append(response, messageObj)
		}
	}

	return response, nil
}

func main() {
	ra := ReportingAPI{}
	ra.Connect()
	defer ra.Close()

	r := gin.Default()
	r.GET("/message/list", func(c *gin.Context) {
		sender := c.Query("sender")
		receiver := c.Query("receiver")

		if sender == "" || receiver == "" {
			c.JSON(400, gin.H{"error": "Both sender and receiver parameters are required"})
			return
		}

		response, err := ra.GetMessageList(sender, receiver)
		if err != nil {
			c.JSON(500, gin.H{"error": "Error fetching messages from Redis"})
			return
		}

		c.JSON(200, response)
	})

	log.Fatal(r.Run(apiPort))
}
