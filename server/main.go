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
	"time"

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

	err = api.CreateGuestUser(ch)
	if err != nil {
		log.Printf("Guest registration error: %v", err)
	}

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

	for {
		select {
		case <-ctx.Done():
			return

		case msg := <-messages:
			log.Printf("Received a request: %s", msg.Body)
			body := string(msg.Body)

			userId := msg.Headers["username"].(string)
			reqId := ""
			if userId == "guest" {
				reqId = msg.Headers["requestId"].(string)
			}

			// Check if user exists
			if strings.Contains(body, "userExists_") {
				parts := strings.Split(body, "_")
				if len(parts) != 2 {
					if err := respond(conn, userId, reqId, "false"); err != nil {
						log.Printf("[REQUEST LISTENER] Failed to respond: %v", err)
					}
					break
				}

				targetUser := parts[1]
				ok, err := api.UserExists(targetUser)

				if err != nil {
					log.Printf("[REQUEST LISTENER] Failed to check if user exists: %v", err)
					ok = false
				}
				if err := respond(conn, userId, reqId, fmt.Sprintf("%t", ok)); err != nil {
					log.Printf("[REQUEST LISTENER] Failed to respond: %v", err)
				}

				break
			}

			// Register user
			if strings.Contains(body, "register_") {
				parts := strings.Split(body, "_")
				if len(parts) != 3 {
					log.Printf("[REQUEST LISTENER] Failed to parse request: %s", body)
					respond(conn, userId, reqId, fmt.Sprintf("not 3 parts but %d", len(parts)))
				}

				targetUser := parts[1]
				targetPassword := parts[2]

				if err := api.CreateUser(ch, targetUser, targetPassword); err != nil {
					log.Printf("[REQUEST LISTENER] %v", err)
					respond(conn, userId, reqId, fmt.Sprintf("%v", err))
				}

				respond(conn, userId, reqId, "ok")
			}

			// Get regular chats
			if strings.Contains(body, "getUserChats") {
				chats, err := api.GetUserChats(ch, userId, false)
				if err != nil {
					log.Printf("[REQUEST LISTENER] %v", err)
					respond(conn, userId, reqId, fmt.Sprintf("%v", err))
				}

				respond(conn, userId, reqId, fmt.Sprintf("ok_%s", strings.Join(chats, "_")))
			}

			// Get secret chats
			if strings.Contains(body, "getUserSecretChats") {
				chats, err := api.GetUserChats(ch, userId, true)
				if err != nil {
					log.Printf("[REQUEST LISTENER] %v", err)
					respond(conn, userId, reqId, fmt.Sprintf("%v", err))
				}

				respond(conn, userId, reqId, fmt.Sprintf("ok_%s", strings.Join(chats, "_")))
			}

			// Create regular chat
			if strings.Contains(body, "createRegularChat_") {
				parts := strings.Split(body, "_")
				if len(parts) != 2 {
					log.Printf("[REQUEST LISTENER] Failed to parse request: %s", body)
					respond(conn, userId, reqId, fmt.Sprintf("not 2 parts but %d", len(parts)))
				}

				targetUser := parts[1]

				id, err := api.StartChat(ch, userId, targetUser, false)
				if err != nil {
					log.Printf("[REQUEST LISTENER] %v", err)
					respond(conn, userId, reqId, fmt.Sprintf("%v", err))
				}

				respond(conn, userId, reqId, fmt.Sprintf("ok_%s", id))
			}

			// Delete chat
			if strings.Contains(body, "deleteChat_") {
				parts := strings.Split(body, "_")
				if len(parts) != 2 {
					log.Printf("[REQUEST LISTENER] Failed to parse request: %s", body)
					respond(conn, userId, reqId, fmt.Sprintf("not 2 parts but %d", len(parts)))
				}

				chat := parts[1]
				parts = strings.Split(chat, "-")
				if len(parts) != 2 {
					log.Printf("[REQUEST LISTENER] Failed to parse request: %s", body)
					respond(conn, userId, reqId, fmt.Sprintf("not 2 parts in chatId but %d", len(parts)))
				}

				err := api.DeleteChat(ch, userId, parts[0], parts[1])
				if err != nil {
					log.Printf("[REQUEST LISTENER] %v", err)
					respond(conn, userId, reqId, fmt.Sprintf("%v", err))
				}

				respond(conn, userId, reqId, "ok")
			}

			// Kick user from the chat
			if strings.Contains(body, "kickUser_") {
				parts := strings.Split(body, "_")
				if len(parts) != 2 {
					log.Printf("[REQUEST LISTENER] Failed to parse request: %s", body)
					respond(conn, userId, reqId, fmt.Sprintf("not 2 parts but %d", len(parts)))
				}

				chat := parts[1]
				parts = strings.Split(chat, "-")
				if len(parts) != 2 {
					log.Printf("[REQUEST LISTENER] Failed to parse request: %s", body)
					respond(conn, userId, reqId, fmt.Sprintf("not 2 parts in chatId but %d", len(parts)))
				}

				err := api.KickUser(ch, userId, parts[0], parts[1])
				if err != nil {
					log.Printf("[REQUEST LISTENER] %v", err)
					respond(conn, userId, reqId, fmt.Sprintf("%v", err))
				}

				respond(conn, userId, reqId, "ok")
			}
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

func respond(conn *amqp.Connection, user, reqId, message string) error {
	ch, err := conn.Channel()
	if err != nil {
		return fmt.Errorf("failed to open a channel: %v", err)
	}
	defer ch.Close()

	responseChatName := api.CreateChatName(user, "response", reqId)
	if user != "guest" {
		responseChatName = fmt.Sprintf("%s-%s", user, "response")
	}

	if user == "guest" {
		err = api.CreateExchange(ch, responseChatName, "", "guest")
		if err != nil {
			return fmt.Errorf("failed to create exchange: %v", err)
		}

		defer func() {
			go func() {
				ch, err := conn.Channel()
				if err != nil {
					log.Printf("[SERVER][RABBIT] Failed to open a channel: %v", err)
					return
				}
				defer ch.Close()

				time.Sleep(time.Minute)

				_, err = ch.QueueDelete(responseChatName, false, false, false)
				if err != nil {
					log.Printf("[SERVER][RABBIT] Failed to delete queue: %v", err)
				}
				err = ch.ExchangeDelete(responseChatName, false, false)
				if err != nil {
					log.Printf("[SERVER][RABBIT] Failed to delete exchange: %v", err)
				}

				api.RevokePermissions("guest", responseChatName)
			}()
		}()

		err = api.CreateQueue(ch, responseChatName)
		if err != nil {
			return fmt.Errorf("failed to create queue: %v", err)
		}
	}

	err = ch.Publish(
		responseChatName,
		responseChatName,
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
