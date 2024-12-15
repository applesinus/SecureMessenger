package saved

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"math/big"
	diffiehellman "messengerClient/back/crypto/API/Diffie-Hellman"
	cryptocontext "messengerClient/back/crypto/API/symmetric"
	"messengerClient/back/crypto/constants/cryptoType"
	"messengerClient/back/crypto/constants/paddingType"
	"messengerClient/back/remoteServer"
	"messengerClient/consts"
	"messengerClient/types"
	"os"
	"strings"
	"sync"
	"time"
)

var SavedChats map[string]types.Chats

func RestoreChats() {
	file, err := os.OpenFile("back/saved/chats/chats.json", os.O_RDWR|os.O_CREATE, 0666)
	if err != nil {
		log.Printf("[BACKEND][CHATS RESTORE] Error opening file: %s", err)
	}
	defer file.Close()

	restoredChats := make(map[string]types.Chats)
	jsonDecoder := json.NewDecoder(file)
	err = jsonDecoder.Decode(&restoredChats)

	if err != nil {
		log.Printf("[BACKEND][CHATS RESTORE] Error decoding file: %s", err)
		restoredChats = make(map[string]types.Chats)
	}

	SavedChats = restoredChats

	SaveChats()
}

func SaveChats() {
	buff, err := json.MarshalIndent(SavedChats, "", "  ")

	if err != nil {
		log.Printf("[BACKEND][CHATS SAVE] Error encoding file: %s", err)
		return
	}

	file, err := os.OpenFile("back/saved/chats/chats.json", os.O_RDWR|os.O_CREATE, 0666)
	if err != nil {
		log.Fatalf("[BACKEND][CHATS SAVE] Error opening file: %s", err)
	}
	defer file.Close()

	file.Truncate(0)
	file.Seek(0, 0)
	file.Write(buff)
}

func CheckChats(user, password string) {
	if user == "" || user == "guest" {
		return
	}

	if _, ok := SavedChats[user]; !ok {
		SavedChats[user] = types.Chats{
			Chats: make(map[string]*types.ChatType),
			Mu:    new(sync.Mutex),
		}
	}

	chatsOnServer, err := remoteServer.GetUserChats(user, password)
	if err != nil {
		return
	}

	secretChatsOnServer, err := remoteServer.GetUserSecretChats(user, password)
	if err != nil {
		return
	}

	chats := make(map[string]struct{})

	SavedChats[user].Mu.Lock()
	defer SavedChats[user].Mu.Unlock()

	for _, chat := range chatsOnServer {
		chats[chat] = struct{}{}
	}
	for _, chat := range secretChatsOnServer {
		chats[chat] = struct{}{}
	}

	for chat := range SavedChats[user].Chats {
		if _, ok := chats[chat]; !ok {
			delete(SavedChats[user].Chats, chat)
		}
	}
}

// No Id needed
func AddMessage(user, chatId string, message types.Message) {
	SavedChats[user].Mu.Lock()
	defer SavedChats[user].Mu.Unlock()

	if _, ok := SavedChats[user]; !ok {
		log.Printf("Chats of user %s does not exist", user)
		return
	}

	if _, ok := SavedChats[user].Chats[chatId]; !ok {
		log.Printf("Chat %s of user %s does not exist", chatId, user)
		return
	}

	SavedChats[user].Chats[chatId].Messages = append(SavedChats[user].Chats[chatId].Messages, message)

	SaveChats()
}

func AddFile(user, chat, id string, fileContent *[]byte) {
	SavedChats[user].Mu.Lock()
	defer SavedChats[user].Mu.Unlock()

	if _, ok := SavedChats[user]; !ok {
		log.Printf("Chats of user %s does not exist", user)
		return
	}

	if _, ok := SavedChats[user].Chats[chat]; !ok {
		log.Printf("Chat %s of user %s does not exist", chat, user)
		return
	}

	timeOfMessage, err := time.Parse(time.StampNano, id)
	if err != nil {
		return
	}

	for idx, message := range SavedChats[user].Chats[chat].Messages {
		if message.Id == id {
			SavedChats[user].Chats[chat].Messages[idx].Message = *fileContent
			return
		}
		timeInChat, err := time.Parse(time.StampNano, message.Id)
		if err != nil {
			continue
		}

		if timeInChat.After(timeOfMessage) {
			log.Println("message not found")
			return
		}
	}

	log.Println("message not found")
}

// chatype can be "regular" or "secret" (or empty for all)
func GetChatsNames(user string, chatype ...string) [][]string {
	chatNames := make([][]string, 0)

	if len(chatype) == 0 && chatype[0] != "regular" && chatype[0] != "secret" {
		for _, chat := range SavedChats[user].Chats {
			chatNames = append(chatNames, []string{chat.Reciever, chat.Id})
		}
	} else {
		if chatype[0] == "regular" {
			for _, chat := range SavedChats[user].Chats {
				if chat.Encryption == consts.EncriptionNo {
					chatNames = append(chatNames, []string{chat.Reciever, chat.Id})
				}
			}
		} else if chatype[0] == "secret" {
			for _, chat := range SavedChats[user].Chats {
				if !(chat.Encryption == consts.EncriptionNo) {
					chatNames = append(chatNames, []string{chat.Reciever, chat.Id})
				}
			}
		}
	}
	return chatNames
}

func GetMessages(user, password, reciever, chatId string) ([]types.Message, error) {
	var err error

	messagesOnServer, err := remoteServer.GetChatMessages(user, password, reciever, chatId)
	if err != nil {
		err = consts.ErrOnServer(err)
		log.Printf("[BACKEND][GET_MESSAGES] Error getting messages from server: %s", err)
		messagesOnServer = make([]types.Message, 0)
	}

	_, ok := SavedChats[user]
	if !ok {
		SavedChats[user] = types.Chats{
			Mu:    new(sync.Mutex),
			Chats: make(map[string]*types.ChatType),
		}

		if err != nil {
			return nil, consts.ErrNoChat
		}
	}

	if chatId[0] == 'S' && len(messagesOnServer) != 0 {
		firstMessage := messagesOnServer[0]
		//log.Printf("[BACKEND][GET_MESSAGES] First message: %v", firstMessage)
		lowerType := strings.ToLower(firstMessage.Type)

		if strings.HasPrefix(lowerType, "magenta-") || strings.HasPrefix(lowerType, "rc6-") {
			keys := types.Keys{}
			err = json.Unmarshal(firstMessage.Message, &keys)
			if err != nil {
				return nil, err
			}

			keys.MyPrivateKey, err = diffiehellman.GeneratePrivateKey(keys.Prime)
			myPublicKey := diffiehellman.GeneratePublicKey(keys.MyPrivateKey, keys.PrimitiveRoot, keys.Prime)

			parts := strings.Split(lowerType, "-")
			if len(parts) != 3 {
				return nil, fmt.Errorf("invalid encryption type: %s", lowerType)
			}

			chatName := fmt.Sprintf("%s-%s", reciever, chatId)

			encryption := parts[0]
			algorithm := parts[1]
			padding := parts[2]

			SavedChats[user].Mu.Lock()
			SavedChats[user].Chats[chatName] = &types.ChatType{
				Id:         chatId,
				Reciever:   reciever,
				Encryption: encryption,
				Algorithm:  algorithm,
				Padding:    padding,
				Keys:       keys,
				Messages:   make([]types.Message, 0),
			}
			SavedChats[user].Mu.Unlock()
			SaveChats()

			response := types.Message{
				Type:    "response-key",
				Author:  user,
				Message: myPublicKey.Bytes(),
			}

			ch := remoteServer.SendMessage(user, password, reciever, chatId, response)

			for {
				val := <-ch
				if val == -1 {
					return nil, errors.New("connection closed")
				}
				if val == 0 || val == -1000 {
					break
				}
			}

			if len(messagesOnServer) == 1 {
				return make([]types.Message, 0), nil
			}
			messagesOnServer = messagesOnServer[1:]
		} else if lowerType == "response-key" {
			recieverPublicKey := big.NewInt(0)
			recieverPublicKey.SetBytes(firstMessage.Message)

			SavedChats[user].Mu.Lock()
			SavedChats[user].Chats[fmt.Sprintf("%s-%s", reciever, chatId)].Keys.RecieverPublicKey = recieverPublicKey
			SavedChats[user].Mu.Unlock()
			SaveChats()

			if len(messagesOnServer) == 1 {
				return make([]types.Message, 0), nil
			}
			messagesOnServer = messagesOnServer[1:]
		} else {
			keys := SavedChats[user].Chats[fmt.Sprintf("%s-%s", reciever, chatId)].Keys
			encryption := SavedChats[user].Chats[fmt.Sprintf("%s-%s", reciever, chatId)].Encryption
			algorithm := SavedChats[user].Chats[fmt.Sprintf("%s-%s", reciever, chatId)].Algorithm
			padding := SavedChats[user].Chats[fmt.Sprintf("%s-%s", reciever, chatId)].Padding

			sessionKey := diffiehellman.ComputeSharedSecret(keys.MyPrivateKey, keys.RecieverPublicKey, keys.Prime)
			//log.Printf("Session key: %s", sessionKey.String())

			for i, message := range messagesOnServer {
				iv := message.Iv

				decypher, err := cryptocontext.NewSymmetricContext(
					sessionKey.Bytes(),
					cryptoType.GetEncryptionMode(algorithm),
					paddingType.GetPaddingMode(padding),
					cryptocontext.GetSymmetricMode(encryption),
					iv)

				if err != nil {
					return nil, err
				}

				messagesOnServer[i].Message, err = decypher.Decrypt(message.Message)
				if err != nil {
					return nil, err
				}
			}
		}
	}

	chatOnDisk, ok := SavedChats[user].Chats[fmt.Sprintf("%s-%s", reciever, chatId)]
	if !ok {
		if err != nil {
			return nil, consts.ErrNoChat
		} else {
			SavedChats[user].Chats[fmt.Sprintf("%s-%s", reciever, chatId)] = &types.ChatType{
				Id:         chatId,
				Reciever:   reciever,
				Encryption: consts.EncriptionNo,
				Messages:   messagesOnServer,
			}

			return messagesOnServer, nil
		}
	}

	messagesInChat := messagesOnServer
	for _, message := range chatOnDisk.Messages {
		after := -1
		for i, msg := range messagesInChat {
			timeInChat, err := time.Parse(time.StampNano, msg.Id)
			if err != nil {
				continue
			}

			timeOnServer, err := time.Parse(time.StampNano, message.Id)
			if err != nil {
				continue
			}

			if timeOnServer.After(timeInChat) {
				after = i
			}
		}

		if after < len(messagesInChat)-1 {
			messagesInChat = append(messagesInChat[:after+1], append([]types.Message{message}, messagesInChat[after+1:]...)...)
		} else {
			messagesInChat = append(messagesInChat, message)
		}
	}

	SavedChats[user].Mu.Lock()
	SavedChats[user].Chats[fmt.Sprintf("%s-%s", reciever, chatId)].Messages = messagesInChat
	SaveChats()
	defer SavedChats[user].Mu.Unlock()

	return messagesInChat, nil
}

func GetMessage(user, chat, id string) (types.Message, error) {
	_, ok := SavedChats[user]
	if !ok {
		return types.Message{}, consts.ErrNoChat
	}

	_, ok = SavedChats[user].Chats[chat]
	if !ok {
		return types.Message{}, consts.ErrNoChat
	}

	timeOfMessage, err := time.Parse(time.StampNano, id)
	if err != nil {
		return types.Message{}, err
	}

	for _, message := range SavedChats[user].Chats[chat].Messages {
		if message.Id == id {
			return message, nil
		}
		timeInChat, err := time.Parse(time.StampNano, message.Id)
		if err != nil {
			continue
		}

		if timeInChat.After(timeOfMessage) {
			return types.Message{}, fmt.Errorf("message not found")
		}
	}

	return types.Message{}, fmt.Errorf("message not found")
}

func NewChat(user, password, reciever, encryption string) (string, error) {
	if _, ok := SavedChats[user]; !ok {
		SavedChats[user] = types.Chats{
			Chats: make(map[string]*types.ChatType),
			Mu:    new(sync.Mutex),
		}
	}

	SavedChats[user].Mu.Lock()
	defer SavedChats[user].Mu.Unlock()

	id := ""
	var err error

	algorithm := ""
	padding := ""
	var keys *types.Keys = nil

	if encryption == consts.EncriptionNo {
		id, err = remoteServer.CreateChat(user, password, reciever)
		if err != nil {
			return "", err
		}
	} else if strings.Contains(encryption, consts.EncriptionMagenta) || strings.Contains(encryption, consts.EncriptionRC6) {
		parts := strings.Split(encryption, "-")
		if len(parts) != 3 {
			return "", fmt.Errorf("invalid encryption type: %s", encryption)
		}

		id, keys, err = remoteServer.CreateSecretChat(user, password, reciever, encryption)
		if err != nil {
			return "", err
		}

		encryption = parts[0]
		algorithm = parts[1]
		padding = parts[2]
	} else {
		return "", fmt.Errorf("unknown encryption type: %s", encryption)
	}

	newChat := types.ChatType{
		Id:         id,
		Reciever:   reciever,
		Encryption: encryption,
		Algorithm:  algorithm,
		Padding:    padding,
		Messages:   make([]types.Message, 0),
	}

	if keys != nil {
		newChat.Keys = *keys
	}

	if SavedChats[user].Chats[id] != nil {
		return "", fmt.Errorf("chat with id %s already exists on local device but not on server, please contact admin", id)
	}

	SavedChats[user].Chats[fmt.Sprintf("%s-%s", reciever, id)] = &newChat

	SaveChats()

	return fmt.Sprintf("%s-%s", reciever, id), nil
}

func ClearChats(user string) {
	if _, ok := SavedChats[user]; !ok {
		return
	}

	SavedChats[user].Mu.Lock()
	defer SavedChats[user].Mu.Unlock()
	delete(SavedChats, user)

	SaveChats()
}

func DeleteChat(user, password, chat string) error {
	parts := strings.Split(chat, "-")
	if len(parts) != 2 {
		return fmt.Errorf("invalid chat id")
	}

	reciever, chatId := parts[0], parts[1]

	SavedChats[user].Mu.Lock()
	if SavedChats[user].Chats[chat] != nil {
		delete(SavedChats[user].Chats, chat)
	}
	SavedChats[user].Mu.Unlock()

	consts.EventListeners.Mu.Lock()
	if consts.EventListeners.Events[user] != nil && consts.EventListeners.Events[user][chat] != nil {
		delete(consts.EventListeners.Events[user], chat)
	}
	consts.EventListeners.Mu.Unlock()

	return remoteServer.DeleteChat(user, password, reciever, chatId)
}

func KickUserFromChat(user, password, chat string) error {
	log.Println("Kicking user from chat")

	var err error

	parts := strings.Split(chat, "-")
	if len(parts) != 2 {
		return fmt.Errorf("invalid chat id")
	}

	reciever, chatId := parts[0], parts[1]

	err = remoteServer.KickUserFromChat(user, password, reciever, chatId)
	if err != nil {
		return err
	}

	consts.EventListeners.Mu.Lock()
	if consts.EventListeners.Events[user] != nil && consts.EventListeners.Events[user][chat] != nil {
		delete(consts.EventListeners.Events[user], chat)
	}
	consts.EventListeners.Mu.Unlock()

	var savingChat types.ChatType
	SavedChats[user].Mu.Lock()
	if SavedChats[user].Chats[chat] != nil {
		savingChat = *SavedChats[user].Chats[chat]
		delete(SavedChats[user].Chats, chat)
	} else {
		savingChat = types.ChatType{
			Id:         chatId,
			Reciever:   reciever,
			Encryption: consts.EncriptionNo,
			Messages:   make([]types.Message, 0),
		}
	}
	SavedChats[user].Mu.Unlock()

	var bytesChat []byte
	if savingChat.Encryption == consts.EncriptionNo {
		bytesChat, err = json.Marshal(savingChat)
		if err != nil {
			return err
		}
	}
	// TODO add other encryptions

	return os.WriteFile(fmt.Sprintf("back/saved/%s-%s.json", user, chat), bytesChat, 0644)
}
