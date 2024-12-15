package main

import (
	"context"
	"log"
	"messengerClient/back"
	"messengerClient/back/remoteServer"
	"messengerClient/back/users"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

func main() {
	/*
		// TEMP

		p, _ := diffiehellman.GeneratePrime(256)
		g, _ := diffiehellman.GeneratePrimitiveRoot(p)

		myPrivateKey, _ := diffiehellman.GeneratePrivateKey(p)
		otherPrivateKey, _ := diffiehellman.GeneratePrivateKey(p)
		otherPublicKey := diffiehellman.GeneratePublicKey(otherPrivateKey, g, p)

		key := diffiehellman.ComputeSharedSecret(myPrivateKey, otherPublicKey, p)

		text := "Alice to Bob! Attention! You're a stinky poo"

		cipherer, err := cryptocontext.NewSymmetricContext(key.Bytes(), cryptoType.ECB, paddingType.ANSIX923, magenta.NewMagenta(), nil)
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
	idx := 0
	for !ok {
		idx++
		if idx > 30 {
			log.Println("[REQUEST LISTENER] Failed to connect to RabbitMQ in 30 attempts")
			return
		}

		time.Sleep(time.Second)
		ok = remoteServer.RabbitIsConnected()
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	wg := new(sync.WaitGroup)

	go back.Start(ctx, wg)
	wg.Add(1)

	go users.RefreshUsers(ctx, wg)
	wg.Add(1)

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)

	isStopping := false

	for {
		select {
		case <-ctx.Done():
			isStopping = true

		case <-sigChan:
			cancel()
			isStopping = true
		}

		if isStopping {
			break
		}
	}

	wg.Wait()

}
