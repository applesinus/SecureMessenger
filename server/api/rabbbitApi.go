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
		responseName := fmt.Sprintf("response:%s", name)
		err = CreateChannel(ch, responseName, responseName, responseName)
		if err != nil {
			return fmt.Errorf("[USER REGISTER] failed to create response channel: %w", err)
		}

		err = setPermissions(name, responseName, true, true)
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

func CreateChannel(ch *amqp.Channel, exchange, queue, routingKey string) error {
	if err := ch.ExchangeDeclare(
		exchange,
		"direct",
		true,
		false,
		false,
		false,
		nil,
	); err != nil {
		return fmt.Errorf("failed to declare exchange: %w", err)
	}

	_, err := ch.QueueDeclare(
		queue,
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
		queue,
		routingKey,
		exchange,
		false,
		nil,
	); err != nil {
		return fmt.Errorf("failed to bind queue: %w", err)
	}

	if strings.Contains(exchange, "guest") {
		setPermissions("guest", exchange, false, true)
	}

	return nil
}

func StartChat(ch *amqp.Channel, user1, user2 string, secret bool) (string, error) {
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

	chats, err := getAllQueues()
	if err != nil {
		return "", err
	}

	newChatId := -1

	for _, chat := range chats {
		if strings.Contains(chat, fmt.Sprintf("%s-%s", user1, user2)) || strings.Contains(chat, fmt.Sprintf("%s-%s", user2, user1)) {
			parts := strings.Split(chat, ":")
			if len(parts) != 2 {
				return "", fmt.Errorf("[CHAT CREATOR] failed to get chat id: %s", chat)
			}

			id, err := strconv.Atoi(parts[1])
			if err != nil {
				return "", fmt.Errorf("[CHAT CREATOR] %s", err)
			}

			if id > newChatId {
				newChatId = id
			}
		}
	}

	newChatId += 1
	strId := ""
	if secret {
		strId = fmt.Sprintf("S%d", newChatId)
	} else {
		strId = fmt.Sprintf("%d", newChatId)
	}

	channel12 := fmt.Sprintf("%s-%s:%s", user1, user2, strId)
	channel21 := fmt.Sprintf("%s-%s:%s", user2, user1, strId)

	err = CreateChannel(
		ch,
		channel12,
		channel12,
		channel12,
	)
	if err != nil {
		return "", fmt.Errorf("[CHAT CREATOR] Error creating channel between %s and %s: %s", user1, user2, err)
	}

	//err = addPermission(user1, chat12exchange, true)
	err = setPermissions(user1, channel12, true, true)
	if err != nil {
		return "", fmt.Errorf("[CHAT CREATOR] Error adding permission for user %s: %s", user1, err)
	}
	//err = addPermission(user2, chat12exchange, false)
	err = setPermissions(user2, channel12, false, true)
	if err != nil {
		return "", fmt.Errorf("[CHAT CREATOR] Error adding permission for user %s: %s", user2, err)
	}

	err = CreateChannel(
		ch,
		channel21,
		channel21,
		channel21,
	)
	if err != nil {
		return "", fmt.Errorf("[CHAT CREATOR] Error creating channel between %s and %s: %s", user2, user1, err)
	}

	//err = addPermission(user1, chat21exchange, false)
	err = setPermissions(user1, channel21, false, true)
	if err != nil {
		return "", fmt.Errorf("[CHAT CREATOR] Error adding permission for user %s: %s", user1, err)
	}
	//err = addPermission(user2, chat21exchange, true)
	err = setPermissions(user2, channel21, true, true)
	if err != nil {
		return "", fmt.Errorf("[CHAT CREATOR] Error adding permission for user %s: %s", user2, err)
	}

	return strId, nil
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

func getAllQueues() ([]string, error) {
	url := fmt.Sprintf("%s/queues", consts.RabbitmqAPI)

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

	var queues []struct {
		Name string `json:"name"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&queues); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	var queueNames []string
	for _, queue := range queues {
		queueNames = append(queueNames, queue.Name)
	}

	return queueNames, nil
}

func CreateGuestUser(ch *amqp.Channel) error {
	err := CreateUser(ch, "guest", "guest")
	if err != nil {
		return fmt.Errorf("failed to create guest user: %w", err)
	}

	return nil
}

func GetUserChats(ch *amqp.Channel, username string, secret bool) ([]string, error) {
	queues, err := getAllQueues()
	if err != nil {
		return nil, fmt.Errorf("failed to get all queues: %w", err)
	}

	userChats := []string{}

	for _, queue := range queues {
		parts := strings.Split(queue, ":")
		if len(parts) != 2 {
			continue
		}

		users := strings.Split(parts[0], "-")
		if len(users) != 2 {
			continue
		}

		if users[0] == username {
			if secret && strings.Contains(parts[1], "S") {
				userChats = append(userChats, fmt.Sprintf("%s:%s", users[1], parts[1]))
			}
			if !secret && !strings.Contains(parts[1], "S") {
				userChats = append(userChats, fmt.Sprintf("%s:%s", users[1], parts[1]))
			}
		}
	}

	return userChats, nil
}
