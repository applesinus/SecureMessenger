package remoteServer

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/streadway/amqp"
)

func SendMessage(chatID string, message string) chan int {
	// TODO
	progress := make(chan int)

	percentage := 1000
	go func() {
		defer close(progress)

		for {
			time.Sleep(time.Second * 1)

			if percentage == 1100 {
				percentage = 2000
			} else if percentage == 2100 {
				percentage = 0
			}

			progress <- percentage
			if percentage == 0 {
				return
			}
			percentage += 5
		}
	}()

	return progress
}

func SendFile(chatID string, file *os.File) chan int {
	// TODO
	progress := make(chan int)

	percentage := 1000
	go func() {
		defer close(progress)

		for {
			time.Sleep(time.Second * 1)

			if percentage == 1100 {
				percentage = 2000
			} else if percentage == 2100 {
				percentage = 0
			}

			progress <- percentage
			percentage += 5

			if percentage == 0 {
				return
			}
		}
	}()

	return progress
}

func CreateChat(user, password string, recipient string) error {
	resp, err := makeRequest(user, password, fmt.Sprintf("createRegularChat_%s", recipient))
	if err != nil {
		return err
	}

	if !strings.Contains(resp, "ok_") {
		return fmt.Errorf("failed to create chat: %s", resp)
	}

	return nil
}

func CreateSecretChat(sender string, recipient string, cipherType string) error {
	// TODO
	return nil
}

func UserRegister(name string, password string) error {
	resp, err := makeRequest("guest", "guest", fmt.Sprintf("register_%s_%s", name, password))
	if err != nil {
		return err
	}

	if resp != "ok" {
		return fmt.Errorf("failed to register user: %s", resp)
	}

	return nil
}

// RabbitMQ user funcs

func UserLogin(username, password string) error {
	conn, err := amqp.Dial(fmt.Sprintf("amqp://%s:%s@localhost:5672/", username, password))
	if err != nil {
		return fmt.Errorf("Failed to login to RabbitMQ: %v", err)
	}
	conn.Close()
	return nil
}

func UserExists(name string) (bool, error) {
	resp, err := makeRequest("guest", "guest", "userExists_"+name)
	if err != nil {
		return false, err
	}
	return resp == "true", nil
}

func GetUserChats(name string) []string {
	return nil
}

func GetUserSecretChats(name string) []string {
	return nil
}

// RabbitMQ helpers

func connectToRabbit(username, password string) (*amqp.Channel, error) {
	conn, err := amqp.Dial(fmt.Sprintf("amqp://%s:%s@localhost:5672/", username, password))
	if err != nil {
		return nil, fmt.Errorf("Failed to connect to RabbitMQ: %v", err)
	}
	defer conn.Close()

	ch, err := conn.Channel()
	if err != nil {
		return nil, fmt.Errorf("failed to open a channel: %v", err)
	}

	return ch, nil
}

func makeRequest(username, password, request string) (string, error) {
	ch, err := connectToRabbit(username, password)
	if err != nil {
		return "", fmt.Errorf("failed to make request: %v", err)
	}

	err = ch.Publish(
		"request",
		"request",
		false,
		false,
		amqp.Publishing{
			ContentType: "text/plain",
			Body:        []byte(request),
		})

	if err != nil {
		return "", fmt.Errorf("failed to publish a message: %v", err)
	}

	return "", nil
}

func ListenToRabbit(ctx context.Context, username, password string) {
	ch, err := connectToRabbit(username, password)
	if err != nil {
		log.Printf("[SERVER][RABBIT LISTENER] Failed to connect to RabbitMQ: %v", err)
		return
	}

	channelName := fmt.Sprintf("response:%s", username)

	messages, err := ch.Consume(
		channelName, // queue
		"",          // consumer
		true,        // auto-ack
		false,       // exclusive
		false,       // no-local
		false,       // no-wait
		nil,         // args
	)
	if err != nil {
		log.Printf("[SERVER][RABBIT LISTENER] Failed to register a consumer: %v", err)
		return
	}

	select {
	case <-ctx.Done():
		return

	case msg := <-messages:
		log.Printf("[SERVER][RABBIT LISTENER] Received a response: %s", msg.Body)
	}
}

func RabbitIsConnected() bool {
	conn, err := amqp.Dial("amqp://guest:guest@localhost:5672/")
	if err != nil {
		return false
	}
	conn.Close()
	return true
}
