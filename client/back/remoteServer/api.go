package remoteServer

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"messengerClient/back/crypto"
	"messengerClient/types"
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

func GetChatMessages(username, password, chatID string) ([]types.Message, error) {
	resp, err := makeSeveralRequests(username, password, fmt.Sprintf("getChatMessages_%s", chatID))
	if err != nil {
		return nil, err
	}

	messages := make([]types.Message, 0)

	for _, msg := range resp {
		message := types.Message{}
		err = json.Unmarshal(msg, &message)
		if err != nil {
			return nil, err
		}

		messages = append(messages, message)
	}

	return messages, nil
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

func CreateChat(user, password string, recipient string) (string, error) {
	resp, err := makeRequest(user, password, fmt.Sprintf("createRegularChat_%s", recipient), "")
	if err != nil {
		return "", err
	}

	if !strings.Contains(resp, "ok_") {
		return "", fmt.Errorf("failed to create chat: %s", resp)
	}

	parts := strings.Split(resp, "_")
	if len(parts) != 2 {
		return "", fmt.Errorf("failed to create chat: %s", resp)
	}

	return parts[1], nil
}

func CreateSecretChat(sender string, recipient string, cipherType string) error {
	// TODO
	return nil
}

// RabbitMQ user funcs

func UserExists(username string) (bool, error) {
	reqId := createRequestId()

	resp, err := makeRequest("guest", "guest", fmt.Sprintf("userExists_%s", username), reqId)
	if err != nil {
		return false, err
	}
	return resp == "true", nil
}

func UserLogin(username, password string) error {
	if username == "" || password == "" {
		return fmt.Errorf("username or password is empty")
	}

	conn, err := amqp.Dial(fmt.Sprintf("amqp://%s:%s@localhost:5672/", username, password))
	if err != nil {
		return fmt.Errorf("failed to login to RabbitMQ: %v", err)
	}
	conn.Close()
	return nil
}

func UserRegister(username string, password string) error {
	ok, err := UserExists(username)
	if err != nil {
		return fmt.Errorf("failed to check user existence: %v", err)
	}
	if ok {
		return fmt.Errorf("user %s already exists", username)
	}

	reqId := createRequestId()

	resp, err := makeRequest("guest", "guest", fmt.Sprintf("register_%s_%s", username, password), reqId)
	if err != nil {
		return err
	}

	if resp != "ok" {
		return fmt.Errorf("failed to register user: %s", resp)
	}

	return nil
}

func GetUserChats(username, password string) ([]string, error) {
	resp, err := makeRequest(username, password, "getUserChats", "")
	if err != nil {
		return make([]string, 0), err
	}

	if !strings.HasPrefix(resp, "ok_") {
		return make([]string, 0), fmt.Errorf("failed to get user chats: %s", resp)
	}

	chats := strings.Split(strings.TrimPrefix(resp, "ok_"), "_")

	if len(chats) == 1 && chats[0] == "" {
		return make([]string, 0), nil
	}

	return chats, nil
}

func GetUserSecretChats(username, password string) ([]string, error) {
	resp, err := makeRequest(username, password, "getUserSecretChats", "")
	if err != nil {
		return make([]string, 0), err
	}

	if !strings.HasPrefix(resp, "ok_") {
		return make([]string, 0), fmt.Errorf("failed to get user chats: %s", resp)
	}

	chats := strings.Split(strings.TrimPrefix(resp, "ok_"), "_")

	if len(chats) == 1 && chats[0] == "" {
		return make([]string, 0), nil
	}

	return chats, nil
}

// RabbitMQ helpers

func connectToRabbit(username, password string) (*amqp.Connection, error) {
	conn, err := amqp.Dial(fmt.Sprintf("amqp://%s:%s@localhost:5672/", username, password))
	if err != nil {
		return nil, fmt.Errorf("Failed to connect to RabbitMQ: %v", err)
	}

	return conn, nil
}

func makeRequest(username, password, request, requestId string) (string, error) {
	conn, err := connectToRabbit(username, password)
	if err != nil {
		return "", fmt.Errorf("failed to make request: %v", err)
	}
	defer conn.Close()

	ch, err := conn.Channel()
	if err != nil {
		return "", fmt.Errorf("failed to open a channel: %v", err)
	}
	defer ch.Close()

	log.Printf("[SERVER][RABBIT] Sending a request: %s from %s:%s", request, username, password)

	err = ch.Publish(
		"request",
		"request",
		false,
		false,
		amqp.Publishing{
			Headers: amqp.Table{
				"requestId": requestId,
				"username":  username,
				"password":  password,
			},
			ContentType: "text/plain",
			Body:        []byte(request),
		})

	if err != nil {
		return "", fmt.Errorf("failed to publish a message: %v", err)
	}

	responseChannel := ""
	if username == "guest" {
		responseChannel = fmt.Sprintf("response:%s:%s", username, requestId)
	} else {
		responseChannel = fmt.Sprintf("response:%s", username)
	}

	log.Printf("[SERVER][RABBIT] Waiting for a response: %s", responseChannel)

	ch.Close()

	time.Sleep(time.Second * 1)

	ch, err = conn.Channel()
	if err != nil {
		return "", fmt.Errorf("failed to open a consuming channel: %v", err)
	}
	defer ch.Close()
	resp, err := ch.Consume(
		responseChannel,
		"",
		true,
		false,
		false,
		false,
		nil,
	)

	if err != nil {
		return "", fmt.Errorf("failed to consume a queue: %v", err)
	}

	respCh := make(chan *amqp.Delivery)
	ctx, _ := context.WithDeadline(context.Background(), time.Now().Add(30*time.Second))
	log.Printf("[SERVER][RABBIT] Waiting for a response...")
	go func() {
		select {
		case <-ctx.Done():
			respCh <- nil
		case msg := <-resp:
			respCh <- &msg
		}
	}()

	response := <-respCh
	if response == nil {
		return "", fmt.Errorf("request timed out after 30 seconds")
	}

	return string(response.Body), nil
}

func makeSeveralRequests(username, password, request string) ([][]byte, error) {
	conn, err := connectToRabbit(username, password)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %v", err)
	}
	defer conn.Close()

	ch, err := conn.Channel()
	if err != nil {
		return nil, fmt.Errorf("failed to open a channel: %v", err)
	}
	defer ch.Close()

	log.Printf("[SERVER][RABBIT] Sending a request: %s from %s:%s", request, username, password)

	err = ch.Publish(
		"request",
		"request",
		false,
		false,
		amqp.Publishing{
			Headers: amqp.Table{
				"requestId": "",
				"username":  username,
				"password":  password,
			},
			ContentType: "text/plain",
			Body:        []byte(request),
		})

	if err != nil {
		return nil, fmt.Errorf("failed to publish a message: %v", err)
	}

	responseChannel := fmt.Sprintf("response:%s", username)

	log.Printf("[SERVER][RABBIT] Waiting for a response: %s", responseChannel)

	ch.Close()

	ch, err = conn.Channel()
	if err != nil {
		return nil, fmt.Errorf("failed to open a consuming channel: %v", err)
	}
	defer ch.Close()
	resp, err := ch.Consume(
		responseChannel,
		"",
		true,
		false,
		false,
		false,
		nil,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to consume a queue: %v", err)
	}

	respCh := make(chan *amqp.Delivery)
	log.Printf("[SERVER][RABBIT] Waiting for a response...")
	timer := time.NewTimer(10 * time.Second)
	defer timer.Stop()

	go func() {
		for {
			select {
			case <-timer.C:
				respCh <- nil
				return

			case msg := <-resp:
				respCh <- &msg
				timer.Reset(10 * time.Second)
			}
		}
	}()

	responses := make([][]byte, 0)
	for {
		select {
		case <-timer.C:
			return responses, nil
		case response := <-respCh:
			if response == nil {
				return responses, nil
			}
			responses = append(responses, response.Body)
		}
	}
}

func ListenToRabbit(ctx context.Context, username, password string) {
	conn, err := connectToRabbit(username, password)
	if err != nil {
		log.Printf("[SERVER][RABBIT LISTENER] Failed to connect to RabbitMQ: %v", err)
		return
	}

	ch, err := conn.Channel()
	if err != nil {
		log.Printf("[SERVER][RABBIT LISTENER] Failed to open a channel: %v", err)
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

func createRequestId() string {
	return crypto.Hash(fmt.Sprintf("%d", time.Now().UnixNano()+rand.Int63()))
}
