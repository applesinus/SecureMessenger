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
	/*
		// TEMP

		text := "Alice to Bob! Attention! You're a stinky poo (:"

		// RC6 FULLY DONE

		// MAGENTA NOT DONE:
		// - ECB & 0/4
		// - CBC & 0/4
		// - PCBC & 0/4
		// - CFB & 0/4
		// - OFB & 0/4
		// - CTR & 0/4
		// - RandomDelta & 0/4

		// DiffieHellman NOT DONE

		cipherer, err := cryptocontext.NewSymmetricContext([]byte("secret"), cryptoType.ECB, paddingType.ANSIX923, RC6.NewRC6(), nil)
		if err != nil {
			log.Fatal(err)
		}

		log.Printf("[MAIN] Encrypting: '%s'", text)
		iv, encrypted, err := cipherer.Encrypt([]byte(text))
		if err != nil {
			log.Fatal(err)
		}
		log.Println("[MAIN] Encrypted")

		log.Printf("[MAIN] Decrypting:")
		log.Printf("%v", encrypted)
		log.Printf("iv=%s", string(iv))
		decrypted, err := cipherer.Decrypt(encrypted)
		if err != nil {
			log.Fatal(err)
		}

		log.Printf("[MAIN] Decrypted: '%s'", string(decrypted))
		log.Printf("[MAIN] Original: '%s' (%v)", string(text), string(decrypted) == text)

		return
		// END TEMP
	*/

	ok := remoteServer.RabbitIsConnected()
	if !ok {
		log.Fatal("[MAIN] RabbitMQ and remote server is not connected")
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	wg := new(sync.WaitGroup)

	go back.Start(ctx, wg)
	wg.Add(1)

	go users.RefreshUsers(ctx, wg)
	wg.Add(1)

	endCh := make(chan struct{})
	go lineReader(endCh)

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)

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
