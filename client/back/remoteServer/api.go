package remoteserver

import (
	"log"
	"os"
	"time"
)

func SendMessage(chatID string, message string) chan int {
	progress := make(chan int)

	percentage := 1000
	go func() {
		defer close(progress)

		for {
			log.Printf("Sending message... %v", percentage)
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

func CreateChat(sender string, recipient string) error {
	return nil
}

func CreateSecretChat(sender string, recipient string, cipherType string) error {
	return nil
}
