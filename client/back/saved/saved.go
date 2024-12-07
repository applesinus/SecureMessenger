package saved

import (
	"encoding/json"
	"fmt"
	"log"
	"messengerClient/back/remoteServer"
	"messengerClient/consts"
	"messengerClient/types"
	"os"
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

	switch encryption {
	case consts.EncriptionNo:
		id, err = remoteServer.CreateChat(user, password, reciever)
		if err != nil {
			return "", err
		}

	case consts.EncriptionMagenta:
		// TODO

	case consts.EncriptionRC6:
		// TODO

	default:
		return "", fmt.Errorf("unknown encryption type: %s", encryption)
	}

	newChat := types.ChatType{
		Id:         id,
		Reciever:   reciever,
		Encryption: encryption,
		Messages:   make([]types.Message, 0),
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
