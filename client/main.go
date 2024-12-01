package main

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"messengerClient/back"
	"messengerClient/back/remoteServer"
	"messengerClient/back/users"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"
)

func main() {
	ok := remoteServer.RabbitIsConnected()
	if !ok {
		log.Fatal("[MAIN] RabbitMQ and remote server is not connected")
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	wg := new(sync.WaitGroup)

	go back.Start(ctx, wg)
	wg.Add(1)

	go users.LoadUsers(ctx, wg)
	wg.Add(1)

	endCh := make(chan struct{})
	go lineReader(endCh)

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	fmt.Print("\n=== MESSENGER APP IS RUNNING ===\n\n")
	defer fmt.Print("\n=== MESSENGER APP IS STOPPED ===\n")

	isStopping := false

	for {
		select {

		case <-endCh:
			cancel()
			isStopping = true

		case <-sigChan:
			endCh <- struct{}{}
		}

		if isStopping {
			break
		}
	}

	wg.Wait()

}

func help() {
	fmt.Println("You can exit the messenger by typing 'exit'")
	fmt.Println("You can show this message by typing 'help'")
}

func lineReader(ch chan struct{}) {
	line := ""

	in := bufio.NewReader(os.Stdin)
	for line != "exit" {
		fmt.Println("You can show all available commands by typing 'help'")

		line, err := in.ReadString('\n')
		line = strings.TrimSpace(line)
		line = strings.ToLower(line)

		if err != nil {
			log.Printf("[MAIN][LINE_READER] Error on reading line: %s", err)
		} else {
			switch line {
			case "exit":
				fmt.Println()
				ch <- struct{}{}
				return

			case "help":
				help()

			default:
				log.Printf("[MAIN][LINE_READER] Received unknown line: '%s'\n\n", line)
			}
		}
	}
}
