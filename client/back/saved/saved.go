package saved

import (
	"encoding/json"
	"log"
	"messengerClient/consts"
	"messengerClient/types"
	"os"
	"strconv"
	"sync"
)

var SavedChats types.Chats

func RestoreChats() {
	file, err := os.OpenFile("back/saved/chats/chats.json", os.O_RDWR|os.O_CREATE, 0666)
	if err != nil {
		log.Printf("[CHATS RESTORE] Error opening file: %s", err)
	}
	defer file.Close()

	restoredChats := make([]types.ChatType, 0)
	jsonDecoder := json.NewDecoder(file)
	err = jsonDecoder.Decode(&restoredChats)

	if err != nil {
		log.Printf("[CHATS RESTORE] Error decoding file: %s", err)
		restoredChats = make([]types.ChatType, 0)
	}

	SavedChats = types.Chats{
		Mu:    new(sync.Mutex),
		Chats: make(map[string]*types.ChatType),
	}

	for _, chat := range restoredChats {
		SavedChats.Chats[chat.Id] = &chat
	}

	/*msgs := make([]Message, 0)
	msgs = append(msgs, []Message{
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
	}...)

	SavedChats.chats["1"] = &types.ChatType{
		Id:         "1",
		Reciever:   "General Grievous",
		Encryption: encriptionNo,
		Messages:   msgs,
	}*/
}

func SaveChats() {
	SavedChats.Mu.Lock()
	defer SavedChats.Mu.Unlock()

	backingUpChats := make([]types.ChatType, 0)

	for _, chat := range SavedChats.Chats {
		for idx := range chat.Messages {
			chat.Messages[idx].Id = strconv.Itoa(idx)
		}
		backingUpChats = append(backingUpChats, *chat)
	}

	buff, err := json.MarshalIndent(backingUpChats, "", "  ")

	if err != nil {
		log.Printf("[CHATS SAVE] Error encoding file: %s", err)
		return
	}

	file, err := os.OpenFile("back/saved/chats/chats.json", os.O_RDWR|os.O_CREATE, 0666)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	file.Truncate(0)
	file.Write(buff)
}

func AddMessage(chatId string, message types.Message) {
	SavedChats.Mu.Lock()
	defer SavedChats.Mu.Unlock()

	for _, chat := range SavedChats.Chats {
		if chat.Id == chatId {
			message.Id = strconv.Itoa(len(chat.Messages))
			chat.Messages = append(chat.Messages, message)
		}
	}
}

func GetChatsNames(chatype ...string) []string {
	chatNames := make([]string, 0)
	if len(chatype) == 0 && chatype[0] != "regular" && chatype[0] != "secret" {
		for _, chat := range SavedChats.Chats {
			chatNames = append(chatNames, chat.Reciever)
		}
	} else {
		if chatype[0] == "regular" {
			for _, chat := range SavedChats.Chats {
				if chat.Encryption == consts.EncriptionNo {
					chatNames = append(chatNames, chat.Reciever)
				}
			}
		} else if chatype[0] == "secret" {
			for _, chat := range SavedChats.Chats {
				if !(chat.Encryption == consts.EncriptionNo) {
					chatNames = append(chatNames, chat.Reciever)
				}
			}
		}
	}
	return chatNames
}

func GetMessages(chatId string) []types.Message {
	if SavedChats.Chats[chatId] == nil {
		return nil
	}

	return SavedChats.Chats[chatId].Messages
}
