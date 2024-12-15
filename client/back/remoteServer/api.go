package remoteServer

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"math/rand"
	"messengerClient/back/crypto"
	diffiehellman "messengerClient/back/crypto/API/Diffie-Hellman"
	cryptocontext "messengerClient/back/crypto/API/symmetric"
	"messengerClient/back/crypto/constants/cryptoType"
	"messengerClient/back/crypto/constants/paddingType"
	"messengerClient/types"
	"net/http"
	"strings"
	"time"

	"github.com/streadway/amqp"
)

func sendProgress(ch chan int, progress int) {
	select {
	case ch <- progress:
	default:
	}
}

func SendMessage(username, password, reciever, chatID string, message types.Message, chats ...types.ChatType) chan int {
	//log.Printf("[SERVER][RABBIT] Sending a message: from %s:%s '%s' to %s-%s", username, password, message, reciever, chatID)
	chProgress := make(chan int)

	go sendingWorker(chProgress, username, password, reciever, chatID, message, chats...)

	return chProgress
}

func encryptingWorker(chProgress chan int, message types.Message, chat types.ChatType) (types.Message, error) {
	keys := chat.Keys
	sessionKey := diffiehellman.ComputeSharedSecret(keys.MyPrivateKey, keys.RecieverPublicKey, keys.Prime).Bytes()
	//log.Printf("Session key: %s", sessionKey)

	sendProgress(chProgress, 1010)

	cypherer, err := cryptocontext.NewSymmetricContext(
		sessionKey,
		cryptoType.GetEncryptionMode(chat.Algorithm),
		paddingType.GetPaddingMode(chat.Padding),
		cryptocontext.GetSymmetricMode(chat.Encryption),
		nil,
	)

	if err != nil {
		return types.Message{}, err
	}

	sendProgress(chProgress, 1050)

	iv, encryptedMsg, err := cypherer.Encrypt(message.Message)

	if err != nil {
		return types.Message{}, err
	}

	sendProgress(chProgress, 1100)

	return types.Message{
		Id:      message.Id,
		Message: encryptedMsg,
		Iv:      iv,
		Type:    message.Type,
		Author:  message.Author,
	}, nil
}

func sendingWorker(chProgress chan int, username, password, reciever, chatID string, message types.Message, chats ...types.ChatType) {
	if strings.Contains(chatID, "S") && len(chats) > 0 {
		//log.Println("[SERVER][RABBIT] Encrypting a message")
		encryptedMsg, err := encryptingWorker(chProgress, message, chats[0])
		if err != nil {
			chProgress <- -1
			close(chProgress)
			return
		}
		message = encryptedMsg

		sendProgress(chProgress, 2000)
	}

	conn, err := connectToRabbit(username, password)
	if err != nil {
		chProgress <- -1
		close(chProgress)
		return
	}
	defer conn.Close()

	sendProgress(chProgress, 2010)

	ch, err := conn.Channel()
	if err != nil {
		chProgress <- -1
		close(chProgress)
		return
	}
	defer ch.Close()

	sendProgress(chProgress, 2020)

	marshalled, err := json.Marshal(message)
	if err != nil {
		chProgress <- -1
		close(chProgress)
		return
	}

	sendProgress(chProgress, 2030)

	chatName := fmt.Sprintf("%s-%s-%s", username, reciever, chatID)

	_, err = ch.QueueInspect(chatName)
	if err != nil {
		chProgress <- -1
		close(chProgress)
		return
	}

	sendProgress(chProgress, 2040)

	if err := ch.Confirm(false); err != nil {
		chProgress <- -1
		close(chProgress)
		return
	}

	err = ch.Publish(
		chatName,
		chatName,
		true,
		false,
		amqp.Publishing{
			Headers: amqp.Table{
				"requestId": "",
				"username":  username,
			},
			ContentType:  "text/plain",
			Body:         marshalled,
			DeliveryMode: amqp.Persistent,
		})

	if err != nil {
		chProgress <- -1
		close(chProgress)
		return
	}

	sendProgress(chProgress, 2090)

	confirmChan := ch.NotifyPublish(make(chan amqp.Confirmation, 1))

	resp := <-confirmChan
	if !resp.Ack {
		chProgress <- -1
		close(chProgress)
		return
	}

	sendProgress(chProgress, 2100)
	chProgress <- 0
}

func SendFile(username, password, reciever, chatID string, message types.Message, r *http.Request, chats ...types.ChatType) (chan int, chan *[]byte) {
	//log.Printf("[SERVER][RABBIT] Sending a message: from %s:%s '%s' to %s-%s", username, password, message, reciever, chatID)
	chProgress := make(chan int)
	chContent := make(chan *[]byte)

	go fileUploadWorker(chProgress, chContent, username, password, reciever, chatID, message, r, chats...)

	return chProgress, chContent
}

func fileUploadWorker(chProgress chan int, chContent chan *[]byte, username, password, reciever, chatID string, message types.Message, r *http.Request, chats ...types.ChatType) {
	// Max upload size is 50MB
	err := r.ParseMultipartForm(50 << 20)
	if err != nil {
		chProgress <- -1
		return
	}

	file, handler, err := r.FormFile("file")
	if err != nil {
		chProgress <- -1
		return
	}
	defer file.Close()

	fileBytes, err := readFileWithProgressChan(file, handler.Size, chProgress)
	if err != nil {
		chProgress <- -1
		return
	}

	message.Message = fileBytes
	go func() { chContent <- &fileBytes }()

	chProgress <- 3100

	sendingWorker(chProgress, username, password, reciever, chatID, message, chats...)
}

func readFileWithProgressChan(r io.Reader, totalSize int64, progressChan chan int) ([]byte, error) {
	progressReader := &progressReader{
		reader:       r,
		totalSize:    totalSize,
		progressChan: progressChan,
	}

	return io.ReadAll(progressReader)
}

type progressReader struct {
	reader       io.Reader
	totalSize    int64
	currentPos   int64
	progressChan chan int
}

func (pr *progressReader) Read(p []byte) (n int, err error) {
	n, err = pr.reader.Read(p)
	pr.currentPos += int64(n)

	percentage := 3000 + int(100*float64(pr.currentPos)/float64(pr.totalSize))

	sendProgress(pr.progressChan, percentage)

	return n, err
}

func GetChatMessages(username, password, reciever, chatID string) ([]types.Message, error) {
	conn, err := connectToRabbit(username, password)
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	ch, err := conn.Channel()
	if err != nil {
		return nil, err
	}
	defer ch.Close()

	chatName := fmt.Sprintf("%s-%s-%s", reciever, username, chatID)

	resp, err := ch.Consume(
		chatName,
		"",
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return nil, err
	}

	//log.Printf("[SERVER][RABBIT] Getting messages from exchange: %s", chatName)

	messages := make([]types.Message, 0)

	timer := time.NewTimer(3 * time.Second)
	defer timer.Stop()

	for {
		select {
		case <-timer.C:
			return messages, nil

		case msg := <-resp:
			message := types.Message{}
			err = json.Unmarshal(msg.Body, &message)
			if err != nil {
				return nil, err
			}
			timer.Reset(1 * time.Second)

			messages = append(messages, message)
		}
	}
}

func CreateChat(user, password string, reciever string) (string, error) {
	resp, err := makeRequest(user, password, fmt.Sprintf("createRegularChat_%s", reciever), "")
	if err != nil {
		return "", err
	}

	if !strings.HasPrefix(resp, "ok_") {
		return "", fmt.Errorf("failed to create chat: %s", resp)
	}

	chatId := strings.Replace(strings.TrimPrefix(resp, "ok_"), fmt.Sprintf("%s-", user), "", 1)

	return chatId, nil
}

func CreateSecretChat(user, password, reciever, cipherType string) (string, *types.Keys, error) {
	resp, err := makeRequest(user, password, fmt.Sprintf("createSecretChat_%s", reciever), "")
	if err != nil {
		return "", nil, err
	}

	if !strings.HasPrefix(resp, "ok_") {
		return "", nil, fmt.Errorf("failed to create chat: %s", resp)
	}

	chatId := strings.Replace(strings.TrimPrefix(resp, "ok_"), fmt.Sprintf("%s-", user), "", 1)

	//log.Println("Created chat:", chatId)

	p, err := diffiehellman.GeneratePrime(256)
	if err != nil {
		return "", nil, err
	}

	g, err := diffiehellman.GeneratePrimitiveRoot(p)
	if err != nil {
		return "", nil, err
	}

	myPrivateKey, err := diffiehellman.GeneratePrivateKey(p)
	if err != nil {
		return "", nil, err
	}

	myPublicKey := diffiehellman.GeneratePublicKey(myPrivateKey, g, p)

	keys := &types.Keys{
		Prime:             p,
		PrimitiveRoot:     g,
		MyPrivateKey:      nil,
		RecieverPublicKey: myPublicKey,
	}

	keysJson, err := json.Marshal(keys)
	if err != nil {
		return "", nil, err
	}

	ch := SendMessage(
		user, password, reciever, chatId,
		types.Message{
			Message: keysJson,
			Type:    cipherType,
		},
	)

	for {
		val := <-ch
		if val == -1 {
			return "", nil, fmt.Errorf("failed to create secret chat")
		}
		if val == 0 || val == -1000 {
			break
		}
	}

	keys.MyPrivateKey = myPrivateKey
	keys.RecieverPublicKey = nil

	return chatId, keys, nil
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

	conn, err := amqp.Dial(fmt.Sprintf("amqp://%s:%s@rabbitmq:5672/", username, password))
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
	conn, err := amqp.Dial(fmt.Sprintf("amqp://%s:%s@rabbitmq:5672/", username, password))
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

	//log.Printf("[SERVER][RABBIT] Sending a request: %s from %s:%s", request, username, password)

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

	responseChannel := fmt.Sprintf("%s-response-%s", username, requestId)
	if requestId == "" {
		responseChannel = fmt.Sprintf("%s-response", username)
	}

	//log.Printf("[SERVER][RABBIT] Waiting for a response: %s", responseChannel)

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
	//log.Printf("[SERVER][RABBIT] Waiting for a response...")
	go func() {
		select {
		case <-ctx.Done():
			respCh <- nil
		case msg := <-resp:
			respCh <- &msg
			//log.Printf("Resp got: %s", msg.Body)
		}
	}()

	response := <-respCh
	if response == nil {
		return "", fmt.Errorf("request timed out after 30 seconds")
	}

	return string(response.Body), nil
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

	channelName := fmt.Sprintf("guest-response:%s", username)

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
	conn, err := amqp.Dial("amqp://guest:guest@rabbitmq:5672/")
	if err != nil {
		return false
	}
	conn.Close()
	return true
}

func createRequestId() string {
	return crypto.Hash(fmt.Sprintf("%d", time.Now().UnixNano()+rand.Int63()))
}

func DeleteChat(user, password, reciever, chatId string) error {
	resp, err := makeRequest(user, password, fmt.Sprintf("deleteChat_%s-%s", reciever, chatId), "")
	if err != nil {
		return err
	}
	if resp != "ok" {
		return fmt.Errorf("failed to delete chat: %s", resp)
	}
	return nil
}

func KickUserFromChat(user, password, reciever, chatId string) error {
	resp, err := makeRequest(user, password, fmt.Sprintf("kickUser_%s-%s", reciever, chatId), "")
	if err != nil {
		return err
	}
	if resp != "ok" {
		return fmt.Errorf("failed to kick user from chat: %s", resp)
	}
	return nil
}
