// main.go
package main

import (
	"context"
	"fmt"
	"log"

	"github.com/streadway/amqp"

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
}

func listenRequests(ctx context.Context, conn *amqp.Connection) {
	ch, err := conn.Channel()
	if consts.LogIfError(err, "Failed to open a channel") {
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
	if consts.LogIfError(err, "Failed to register a consumer") {
		return
	}

	select {
	case <-ctx.Done():
		return

	case msg := <-messages:
		log.Printf("Received a request: %s", msg.Body)
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
