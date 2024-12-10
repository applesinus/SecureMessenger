package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"messengerServer/consts"
	"messengerServer/types"
	"net/http"
	"strconv"
	"strings"

	"github.com/streadway/amqp"
)

func CreateUser(ch *amqp.Channel, name, password string) error {
	// Check user existence
	ok, err := UserExists(name)
	if err != nil {
		return fmt.Errorf("[USER REGISTER] failed to check user existence: %w", err)
	}
	if ok {
		return fmt.Errorf("[USER REGISTER] user %s already exists", name)
	}

	// Create user
	err = setUser(name, password)
	if err != nil {
		return fmt.Errorf("[USER REGISTER] failed to set user: %w", err)
	}

	// Set vhost permissions
	err = setVhostPermission(name, types.Permission{Configure: "", Write: ".*", Read: ".*"})
	if err != nil {
		return fmt.Errorf("[USER REGISTER] failed to set vhost permission: %w", err)
	}

	// Set request chat permissions
	err = setPermissions(name, "request", true, true)
	if err != nil {
		return fmt.Errorf("[USER REGISTER] failed to set request permission: %w", err)
	}

	// Create response channel & set permissions
	if name != "guest" {
		responseExchange := fmt.Sprintf("%s-response", name)

		err = CreateExchange(ch, responseExchange, "", name)
		if err != nil {
			return fmt.Errorf("[USER REGISTER] failed to create response exchange: %w", err)
		}

		err = CreateQueue(ch, responseExchange)
		if err != nil {
			return fmt.Errorf("[USER REGISTER] failed to create response channel: %w", err)
		}

		err = setPermissions(name, responseExchange, true, true)
		if err != nil {
			return fmt.Errorf("[USER REGISTER] failed to set chat permission: %w", err)
		}
	}

	// Logging new user registration
	log.Printf("[SERVER][USERS] User registered: %s", name)

	return nil
}

func setUser(name, password string) error {
	url := fmt.Sprintf("%s/users/%s", consts.RabbitmqAPI, name)
	log.Println(url)

	user := types.User{
		Name:     name,
		Password: password,
		Tags:     "",
	}

	body, err := json.Marshal(user)
	if err != nil {
		return fmt.Errorf("failed to marshal user: %w", err)
	}

	req, err := http.NewRequest(http.MethodPut, url, bytes.NewBuffer(body))
	if err != nil {
		return fmt.Errorf("failed to create HTTP request: %w", err)
	}

	req.SetBasicAuth(consts.RabbitmqUser, consts.RabbitmqPassword)
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send HTTP request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNoContent && resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("unexpected response: %s, status: %d", body, resp.StatusCode)
	}

	return nil
}

func setVhostPermission(name string, permission types.Permission) error {
	url := fmt.Sprintf("%s/permissions/%s/%s", consts.RabbitmqAPI, consts.Vhost, name)
	log.Println(url)

	body, err := json.Marshal(permission)
	if err != nil {
		return fmt.Errorf("failed to marshal permissions: %w", err)
	}

	req, err := http.NewRequest(http.MethodPut, url, bytes.NewBuffer(body))
	if err != nil {
		return fmt.Errorf("failed to create HTTP request: %w", err)
	}

	req.SetBasicAuth(consts.RabbitmqUser, consts.RabbitmqPassword)
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send HTTP request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNoContent && resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("failed to set permissions: %s, status: %d", body, resp.StatusCode)
	}

	return nil
}

func setPermissions(user, exchange string, write, read bool) error {
	url := fmt.Sprintf("%s/topic-permissions/%s/%s", consts.RabbitmqAPI, consts.Vhost, user)

	permission := types.TopicPermissions{
		Exchange: exchange,
		Write:    "",
		Read:     "",
	}

	if write {
		permission.Write = ".*"
	}

	if read {
		permission.Read = ".*"
	}

	body, err := json.Marshal(permission)
	if err != nil {
		return fmt.Errorf("failed to marshal permissions: %w", err)
	}

	req, err := http.NewRequest(http.MethodPut, url, bytes.NewBuffer(body))
	if err != nil {
		return fmt.Errorf("failed to create HTTP request: %w", err)
	}

	req.SetBasicAuth(consts.RabbitmqUser, consts.RabbitmqPassword)
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send HTTP request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNoContent && resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("failed to set permissions: %s, status: %d", body, resp.StatusCode)
	}

	return nil
}

func RevokePermissions(user, exchange string) error {
	return setPermissions(user, exchange, false, false)
}

func CreateQueue(ch *amqp.Channel, name string) error {
	log.Printf("[QUEUE] Creating queue: '%s'", name)

	_, err := ch.QueueDeclare(
		name,
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return fmt.Errorf("failed to declare queue: %w", err)
	}

	if err := ch.QueueBind(
		name,
		name,
		name,
		false,
		nil,
	); err != nil {
		return fmt.Errorf("failed to bind queue: %w", err)
	}

	if strings.Contains(name, "guest") {
		setPermissions("guest", name, false, true)
	}

	return nil
}

func CreateExchange(ch *amqp.Channel, exchangeName, writer, reader string) error {
	err := ch.ExchangeDeclare(
		exchangeName,
		"direct",
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return fmt.Errorf("failed to declare exchange: %w", err)
	}

	if reader == "" {
		return fmt.Errorf("[CHAT CREATOR] Reader cannot be empty")
	}

	if writer != "" {
		err = setPermissions(writer, exchangeName, true, false)
		if err != nil {
			return fmt.Errorf("[CHAT CREATOR] Error adding write permission for user %s: %s", writer, err)
		}
	}

	err = setPermissions(reader, exchangeName, false, true)
	if err != nil {
		return fmt.Errorf("[CHAT CREATOR] Error adding read permission for user %s: %s", reader, err)
	}

	return nil
}

func StartChat(ch *amqp.Channel, user1, user2 string, secret bool) (string, error) {
	if user1 == user2 {
		return "", fmt.Errorf("[CHAT CREATOR] You cannot start a chat with yourself")
	}

	if ok, err := UserExists(user1); !ok {
		if err != nil {
			return "", fmt.Errorf("[CHAT CREATOR] %s", err)
		} else {
			return "", fmt.Errorf("[CHAT CREATOR] User %s does not exist", user1)
		}
	}
	if ok, err := UserExists(user2); !ok {
		if err != nil {
			return "", fmt.Errorf("[CHAT CREATOR] %s", err)
		} else {
			return "", fmt.Errorf("[CHAT CREATOR] User %s does not exist", user1)
		}
	}

	chats, err := getAllExchanges()
	if err != nil {
		return "", err
	}

	chatName := ""
	chatId := -1

	chatPreffix := CreateChatPreffix(user1, user2)

	for _, chat := range chats {
		if strings.HasPrefix(chat, chatPreffix) {
			chatId, err = strconv.Atoi(strings.Split(chat, "-")[2])
			if err != nil {
				return "", err
			}
		}
	}
	chatId++

	ch12 := CreateChatName(user1, user2, fmt.Sprintf("%d", chatId))
	ch21 := CreateChatName(user2, user1, fmt.Sprintf("%d", chatId))

	err = CreateExchange(ch, ch12, user1, user2)
	if err != nil {
		return "", fmt.Errorf("[CHAT CREATOR] Error creating exchange '%s' between %s and %s: %s", chatName, user1, user2, err)
	}
	err = CreateExchange(ch, ch21, user2, user1)
	if err != nil {
		return "", fmt.Errorf("[CHAT CREATOR] Error creating exchange '%s' between %s and %s: %s", chatName, user2, user1, err)
	}

	err = CreateQueue(ch, ch12)
	if err != nil {
		return "", fmt.Errorf("[CHAT CREATOR] Error creating queue '%s' between %s and %s: %s", chatName, user1, user2, err)
	}

	err = CreateQueue(ch, ch21)
	if err != nil {
		return "", fmt.Errorf("[CHAT CREATOR] Error creating queue '%s' between %s and %s: %s", chatName, user2, user1, err)
	}

	return fmt.Sprintf("%d", chatId), nil
}

func UserExists(username string) (bool, error) {
	url := fmt.Sprintf("%s/users", consts.RabbitmqAPI)

	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return false, fmt.Errorf("failed to create HTTP request: %w", err)
	}

	req.SetBasicAuth(consts.RabbitmqUser, consts.RabbitmqPassword)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return false, fmt.Errorf("failed to send HTTP request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return false, fmt.Errorf("failed to get users: %s, status: %d", body, resp.StatusCode)
	}

	var users []struct {
		Name string `json:"name"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&users); err != nil {
		return false, fmt.Errorf("failed to decode response: %w", err)
	}

	for _, user := range users {
		if user.Name == username {
			return true, nil
		}
	}

	return false, nil
}

func getAllExchanges() ([]string, error) {
	url := fmt.Sprintf("%s/exchanges/%s", consts.RabbitmqAPI, consts.Vhost)

	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create HTTP request: %w", err)
	}

	req.SetBasicAuth(consts.RabbitmqUser, consts.RabbitmqPassword)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send HTTP request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("failed to get queues: %s, status: %d", body, resp.StatusCode)
	}

	var exchanges []struct {
		Name string `json:"name"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&exchanges); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	log.Printf("[EXCHANGE] Got %v", exchanges)

	var exchangeNames []string
	for _, queue := range exchanges {
		exchangeNames = append(exchangeNames, queue.Name)
	}

	return exchangeNames, nil
}

func CreateGuestUser(ch *amqp.Channel) error {
	err := CreateUser(ch, "guest", "guest")
	if err != nil {
		return fmt.Errorf("failed to create guest user: %w", err)
	}

	return nil
}

func GetUserChats(ch *amqp.Channel, username string, secret bool) ([]string, error) {
	exchanges, err := getAllExchanges()
	if err != nil {
		return nil, fmt.Errorf("failed to get all queues: %w", err)
	}

	userChats := []string{}

	for _, chat := range exchanges {
		parts := strings.Split(chat, "-")
		if len(parts) != 3 {
			continue
		}
		user := parts[0]

		if user == username {
			if parts[1] == "response" {
				continue
			}

			if secret && strings.Contains(parts[2], "S") ||
				!secret && !strings.Contains(parts[2], "S") {
				userChats = append(userChats, strings.TrimPrefix(chat, fmt.Sprintf("%s-", user)))
			}
		}
	}

	return userChats, nil
}

func CreateChatName(user1, user2, chatId string) string {
	return fmt.Sprintf("%s-%s", CreateChatPreffix(user1, user2), chatId)
}

func CreateChatPreffix(user1, user2 string) string {
	return fmt.Sprintf("%s-%s", user1, user2)
}

func DeleteChat(ch *amqp.Channel, username, reciever, chatId string) error {
	errorString := ""

	err := ch.ExchangeDelete(CreateChatName(reciever, username, chatId), false, false)
	if err != nil {
		errorString += fmt.Sprintf("failed to delete exchange: %s", err.Error())
	}

	err = ch.ExchangeDelete(CreateChatName(username, reciever, chatId), false, false)
	if err != nil {
		if errorString != "" {
			errorString += ", "
		}
		errorString += fmt.Sprintf("failed to delete exchange: %s", err.Error())
	}

	_, err = ch.QueueDelete(CreateChatName(reciever, username, chatId), false, false, false)
	if err != nil {
		if errorString != "" {
			errorString += ", "
		}
		errorString += fmt.Sprintf("failed to delete queue: %s", err.Error())
	}

	_, err = ch.QueueDelete(CreateChatName(username, reciever, chatId), false, false, false)
	if err != nil {
		if errorString != "" {
			errorString += ", "
		}
		errorString += fmt.Sprintf("failed to delete queue: %s", err.Error())
	}

	if errorString != "" {
		return fmt.Errorf("%s", errorString)
	}

	return nil
}

func KickUser(ch *amqp.Channel, username, reciever, chatId string) error {
	return DeleteChat(ch, username, reciever, chatId)
}
