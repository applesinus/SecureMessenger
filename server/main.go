// main.go
package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/streadway/amqp"

	"messengerServer/api"
	"messengerServer/consts"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	conn, err := amqp.Dial(fmt.Sprintf("amqp://%s:%s@localhost:5672/", consts.RabbitmqUser, consts.RabbitmqPassword))
	if consts.LogIfError(err, "Failed to connect to RabbitMQ") {
		return
	}
	defer conn.Close()

	ch, err := conn.Channel()
	if err != nil {
		log.Fatalf("Failed to open a channel: %v", err)
	}
	defer ch.Close()

	err = createRequestsExchange(ch)
	if consts.LogIfError(err, "Failed to create exchange") {
		return
	}
	go listenRequests(ctx, conn)

	api.CreateGuestUser()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	for {
		select {
		case <-sigChan:
			cancel()
		case <-ctx.Done():
			return
		}
	}
}

func listenRequests(ctx context.Context, conn *amqp.Connection) {
	ch, err := conn.Channel()
	if consts.LogIfError(err, "[REQUEST LISTENER] Failed to open a channel") {
		return
	}
	defer ch.Close()

	messages, err := ch.Consume(
		"request", // queue
		"",        // consumer
		true,      // auto-ack
		false,     // exclusive
		false,     // no-local
		false,     // no-wait
		nil,       // args
	)
	if consts.LogIfError(err, "[REQUEST LISTENER] Failed to register a consumer") {
		return
	}

	select {
	case <-ctx.Done():
		return

	case msg := <-messages:
		log.Printf("Received a request: %s", msg.Body)
		body := string(msg.Body)

		// Check if user exists
		if strings.Contains(body, "userExists_") {
			parts := strings.Split(body, "_")
			if len(parts) != 2 {
				log.Printf("[REQUEST LISTENER] Failed to parse request: %s", body)
				respond(conn, msg.UserId, "false")
			}

			targetUser := parts[1]
			ok, err := api.UserExists(targetUser)

			if err != nil {
				log.Printf("[REQUEST LISTENER] Failed to check if user exists: %v", err)
				ok = false
			}
			if err := respond(conn, msg.UserId, fmt.Sprintf("%t", ok)); err != nil {
				log.Printf("[REQUEST LISTENER] Failed to respond: %v", err)
			}
		}

		// Register user
		if strings.Contains(body, "register_") {
			parts := strings.Split(body, "_")
			if len(parts) != 3 {
				log.Printf("[REQUEST LISTENER] Failed to parse request: %s", body)
				respond(conn, msg.UserId, fmt.Sprintf("not 3 parts but %d", len(parts)))
			}

			targetUser := parts[1]
			targetPassword := parts[2]

			if err := api.CreateUser(ch, targetUser, targetPassword); err != nil {
				log.Printf("[REQUEST LISTENER] %v", err)
				respond(conn, msg.UserId, fmt.Sprintf("%v", err))
			}
		}

		// Create regular chat
		if strings.Contains(body, "createRegularChat_") {
			parts := strings.Split(body, "_")
			if len(parts) != 2 {
				log.Printf("[REQUEST LISTENER] Failed to parse request: %s", body)
				respond(conn, msg.UserId, fmt.Sprintf("not 2 parts but %d", len(parts)))
			}

			targetUser := parts[1]

			id, err := api.StartChat(ch, msg.UserId, targetUser)
			if err != nil {
				log.Printf("[REQUEST LISTENER] %v", err)
				respond(conn, msg.UserId, fmt.Sprintf("%v", err))
			}

			respond(conn, msg.UserId, fmt.Sprintf("ok_%s", id))
		}
	}
}

func createRequestsExchange(ch *amqp.Channel) error {
	err := ch.ExchangeDeclare(
		"request",
		"direct",
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return err
	}

	_, err = ch.QueueDeclare(
		"request",
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return err
	}

	err = ch.QueueBind(
		"request",
		"request",
		"request",
		false,
		nil,
	)
	if err != nil {
		return err
	}

	return nil
}

func respond(conn *amqp.Connection, user, message string) error {
	ch, err := conn.Channel()
	if err != nil {
		return fmt.Errorf("failed to open a channel: %v", err)
	}
	defer ch.Close()

	responseID := fmt.Sprintf("response:%s", user)

	err = ch.Publish(
		responseID,
		responseID,
		false,
		false,
		amqp.Publishing{
			ContentType: "text/plain",
			Body:        []byte(message),
		})

	if err != nil {
		return fmt.Errorf("failed to publish a message: %v", err)
	}

	return nil
}
