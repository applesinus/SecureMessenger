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
		responseExchange := CreateExchangeName(name, "response")

		err = CreateExchange(ch, responseExchange, name, "")
		if err != nil {
			return fmt.Errorf("[USER REGISTER] failed to create response exchange: %w", err)
		}

		err = CreateQueue(ch, responseExchange, responseExchange, responseExchange)
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

func CreateQueue(ch *amqp.Channel, exchange, queue, routingKey string) error {
	log.Printf("[QUEUE] Creating queue: '%s'", queue)

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

func CreateExchange(ch *amqp.Channel, exchangeName, user1, user2 string) error {
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

	err = setPermissions(user1, exchangeName, true, true)
	if err != nil {
		return fmt.Errorf("[CHAT CREATOR] Error adding permission for user %s: %s", user1, err)
	}

	if user2 == "" {
		return nil
	}
	err = setPermissions(user2, exchangeName, true, true)
	if err != nil {
		return fmt.Errorf("[CHAT CREATOR] Error adding permission for user %s: %s", user2, err)
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

	exchanges, err := getAllExchanges()
	if err != nil {
		return "", err
	}

	exchangeName := ""
	channelId := -1

	e12 := CreateExchangeName(user1, user2)
	e21 := CreateExchangeName(user2, user1)

	for _, chat := range exchanges {
		if chat == e12 || chat == e21 {
			exchangeName = chat
			break
		}
	}

	if exchangeName == "" {
		exchangeName = e12
		err := CreateExchange(ch, exchangeName, user1, user2)
		if err != nil {
			return "", err
		}
	} else {
		channels, err := getExchangeQueues(exchangeName)
		if err != nil {
			return "", err
		}

		for _, channel := range channels {
			chInt, err := strconv.Atoi(channel)
			if err != nil {
				return "", err
			}

			if chInt > channelId {
				channelId = chInt
			}
		}
	}

	channelId++
	queueName := CreateQueueName(exchangeName, fmt.Sprintf("%d", channelId))

	err = CreateQueue(
		ch,
		exchangeName,
		queueName,
		queueName,
	)
	if err != nil {
		return "", fmt.Errorf("[CHAT CREATOR] Error creating channel '%s' between %s and %s: %s", queueName, user1, user2, err)
	}

	return fmt.Sprintf("%d", channelId), nil
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

func getExchangeQueues(exchange string) ([]string, error) {
	url := fmt.Sprintf("%s/queues/%s", consts.RabbitmqAPI, consts.Vhost)

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

	log.Printf("[QUEUE] Got %v", queues)

	var queueNames []string
	for _, queue := range queues {
		if strings.HasPrefix(queue.Name, exchange) {
			queueNames = append(queueNames, queue.Name)
		}
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
	exchanges, err := getAllExchanges()
	if err != nil {
		return nil, fmt.Errorf("failed to get all queues: %w", err)
	}

	userChats := []string{}

	for _, exchange := range exchanges {
		users := strings.Split(exchange, "-")
		if len(users) != 2 {
			continue
		}

		if users[0] == username || users[1] == username {
			var reciever string
			if users[0] == username {
				reciever = users[1]
			} else {
				reciever = users[0]
			}

			if reciever == "response" {
				continue
			}

			queues, err := getExchangeQueues(exchange)
			if err != nil {
				return nil, fmt.Errorf("failed to get exchange queues: %w", err)
			}

			for _, queue := range queues {
				if secret && strings.Contains(strings.TrimPrefix(queue, exchange), "S") ||
					!secret && !strings.Contains(strings.TrimPrefix(queue, exchange), "S") {
					userChats = append(userChats, fmt.Sprintf("%s%s", reciever, strings.TrimPrefix(queue, exchange)))
				}
			}
		}
	}

	return userChats, nil
}

func CreateExchangeName(user1, user2 string) string {
	return fmt.Sprintf("%s-%s", user1, user2)
}

func CreateQueueName(exchangeName, queueId string) string {
	return fmt.Sprintf("%s-%s", exchangeName, queueId)
}
