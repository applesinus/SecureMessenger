package saved

import (
	"encoding/json"
	"fmt"
	"log"
	"messengerClient/back/remoteServer"
	"messengerClient/consts"
	"messengerClient/types"
	"os"
	"strconv"
)

var SavedChats map[string]types.Chats

func RestoreChats() {
	// TEMP
	// clear file to debug
	fileT, err := os.OpenFile("back/saved/chats/chats.json", os.O_RDWR|os.O_CREATE, 0666)
	if err != nil {
		log.Printf("[BACKEND][CHATS RESTORE] Error opening file: %s", err)
	}
	fileT.Truncate(0)
	fileT.Close()
	// TEMP END

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
	file.Write(buff)
}

// No Id needed
func AddMessage(user, chatId string, message types.Message) {
	SavedChats[user].Mu.Lock()
	defer SavedChats[user].Mu.Unlock()

	for _, chat := range SavedChats[user].Chats {
		if chat.Id == chatId {
			message.Id = strconv.Itoa(len(chat.Messages))
			chat.Messages = append(chat.Messages, message)
		}
	}
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

func GetMessages(user, chatId string) ([]types.Message, string) {
	if SavedChats[user].Chats[chatId] == nil {
		return nil, ""
	}

	return SavedChats[user].Chats[chatId].Messages, SavedChats[user].Chats[chatId].Reciever
}

func NewChat(user, password, reciever, encryption string) (string, error) {
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

	SavedChats[user].Chats[id] = &newChat

	return id, nil
}
