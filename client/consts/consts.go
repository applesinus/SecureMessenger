package consts

import (
	"errors"
	"fmt"
	"log"
	"messengerClient/types"
	"strings"
)

var EventListeners types.EventsType
var Recievers types.RecievedType

func AddListener(user, chatId, eventID string, listener chan int) {
	EventListeners.Mu.Lock()
	if _, ok := EventListeners.Events[user]; !ok {
		EventListeners.Events[user] = make(map[string]map[string]chan int)
	}
	if _, ok := EventListeners.Events[user][chatId]; !ok {
		EventListeners.Events[user][chatId] = make(map[string]chan int)
	}

	EventListeners.Events[user][chatId][eventID] = listener
	EventListeners.Mu.Unlock()
}

func RemoveListener(user, chatId, eventID string) {
	EventListeners.Mu.Lock()
	log.Printf("Removing listener of user %s in chat %s: '%s'", user, chatId, eventID)
	if EventListeners.Events[user][chatId][eventID] != nil {
		close(EventListeners.Events[user][chatId][eventID])
	}
	delete(EventListeners.Events[user][chatId], eventID)
	EventListeners.Mu.Unlock()
}

func ErrOnServer(err error) error {
	return fmt.Errorf("error on server: %w", err)
}

func IsErrOnServer(err error) bool {
	return err != nil && strings.HasPrefix(err.Error(), "error on server: ")
}

const (
	LocalHost = "localhost:"
	LocalIP   = "127.0.0.1:"

	EncriptionNo      = "no"
	EncriptionMagenta = "magenta"
	EncriptionRC6     = "rc6"

	AlgorithmECB         = "ecb"
	AlgorithmCBC         = "cbc"
	AlgorithmPCBC        = "pcbc"
	AlgorithmOFB         = "ofb"
	AlgorithmCFB         = "cfb"
	AlgorithmCTR         = "ctr"
	AlgorithmRandomDelta = "randomdelta"

	PaddingZeros    = "zeros"
	PaddingPKCS7    = "pkcs7"
	PaddingANSIX923 = "ansix923"
	PaddingISO10126 = "iso10126"
)

var (
	ErrNoChat = errors.New("no chat")
)
