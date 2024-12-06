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

	SavedChats = make(map[string]types.Chats)

	for user := range restoredChats {
		SavedChats[user] = restoredChats[user]
	}

	SaveChats()

	// TEMP
	/*chats := types.Chats{
		Mu:    new(sync.Mutex),
		Chats: make(map[string]*types.ChatType),
	}

	msgs := make([]types.Message, 0)
	msgs = append(msgs, []types.Message{
		{
			Id:      "0",
			Author:  "kekus",
			Message: "Hello there",
			Type:    "text",
		},
		{
			Id:      "1",
			Author:  "General Grievous",
			Message: "General Kenobi! You are a bold one.",
			Type:    "text",
		},
		{
			Id:      "2",
			Author:  "General Grievous",
			Message: "Kill him!",
			Type:    "text",
		},
		{
			Id:      "3",
			Author:  "Battle Droids",
			Message: "[Droids fail to kill Obi-Wan]",
			Type:    "text",
		},
		{
			Id:      "4",
			Author:  "Battle Droids",
			Message: "[Other droids surround him]",
			Type:    "text",
		},
		{
			Id:      "5",
			Author:  "General Grievous",
			Message: "Back away! I will deal with this Jedi scum myself!",
			Type:    "text",
		},
		{
			Id:      "6",
			Author:  "kekus",
			Message: "Your move!",
			Type:    "text",
		},
		{
			Id:      "7",
			Author:  "kekus",
			Message: "meme.jpeg",
			Type:    "image",
		},
	}...)

	chats.Chats["1"] = &types.ChatType{
		Id:         "1",
		Reciever:   "General Grievous",
		Encryption: consts.EncriptionNo,
		Messages:   msgs,
	}

	SavedChats["kekus"] = chats*/
	// TEMP END
}

func SaveChats() {
	backingUpChats := make(map[string]types.Chats, 0)

	for user := range SavedChats {
		backingUpChats[user] = SavedChats[user]
	}

	buff, err := json.MarshalIndent(backingUpChats, "", "  ")

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

	for _, chat := range SavedChats[user].Chats {
		if chat.Id == chatId {
			chat.Messages = append(chat.Messages, message)
		}
	}

	SaveChats()
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

func GetMessages(user, password, chatId string) ([]types.Message, error) {
	var err error

	messagesOnServer, err := remoteServer.GetChatMessages(user, password, chatId)
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

	chatOnDisk, ok := SavedChats[user].Chats[chatId]
	if !ok {
		if err != nil {
			return nil, consts.ErrNoChat
		} else {
			SavedChats[user].Chats[chatId] = &types.ChatType{
				Id:         chatId,
				Reciever:   messagesOnServer[0].Author,
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
			if message.Id == msg.Id {
				break
			}

			timeInChat, err := time.Parse(time.RFC3339, msg.Id)
			if err != nil {
				continue
			}

			timeOnServer, err := time.Parse(time.RFC3339, message.Id)
			if err != nil {
				continue
			}

			if timeOnServer.After(timeInChat) {
				after = i
			}
		}

		if after < len(messagesInChat)-1 {
			messagesInChat = append(messagesInChat[:after], append(messagesInChat[after+1:], message)...)
		} else {
			messagesInChat = append(messagesInChat, message)
		}
	}

	return messagesInChat, nil
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

	SavedChats[user].Chats[fmt.Sprintf("%s:%s", reciever, id)] = &newChat

	SaveChats()

	return id, nil
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
