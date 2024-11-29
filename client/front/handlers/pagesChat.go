package handlers

import (
	"fmt"
	"html/template"
	"log"
	remoteserver "messengerClient/back/remoteServer"
	"messengerClient/back/saved"
	"messengerClient/consts"
	"messengerClient/types"
	"net/http"
)

func chatsPage(w http.ResponseWriter, r *http.Request, data types.Data) {
	t, err := template.ParseFiles("front/pages/template.html", "front/pages/blocks_user.html", "front/pages/chats.html")
	if err != nil {
		log.Println(err.Error())
	}

	data.RegularChats = []string{
		"Vupsen",
		"Pupsen",
	}

	data.SecretChats = []string{
		"General Grievous",
		"Chewbacca",
	}

	data.Message = "Your chats"

	err = t.Execute(w, data)
	if err != nil {
		fmt.Println(err.Error())
	}
}

func regularChatsPage(w http.ResponseWriter, r *http.Request, data types.Data) {
	t, err := template.ParseFiles("front/pages/template.html", "front/pages/blocks_user.html", "front/pages/chats.html")
	if err != nil {
		log.Println(err.Error())
	}

	if len(data.SecretChats) > 0 {
		data.SecretChats = []string{}
	}

	data.RegularChats = []string{
		"Vupsen",
		"Pupsen",
	}

	data.Message = "Regular chats"

	err = t.Execute(w, data)
	if err != nil {
		fmt.Println(err.Error())
	}
}

func secretChatsPage(w http.ResponseWriter, r *http.Request, data types.Data) {
	t, err := template.ParseFiles("front/pages/template.html", "front/pages/blocks_user.html", "front/pages/chats.html")
	if err != nil {
		log.Println(err.Error())
	}

	if len(data.RegularChats) > 0 {
		data.SecretChats = []string{}
	}

	data.SecretChats = []string{
		"General Grievous",
		"Chewbacca",
	}

	data.Message = "Secret chats"

	err = t.Execute(w, data)
	if err != nil {
		fmt.Println(err.Error())
	}
}

func newChatPage(w http.ResponseWriter, r *http.Request, data types.Data) {
	t, err := template.ParseFiles("front/pages/template.html", "front/pages/blocks_user.html", "front/pages/newChat.html")
	if err != nil {
		log.Println(err.Error())
	}

	err = t.Execute(w, data)
	if err != nil {
		fmt.Println(err.Error())
	}
}

func chatPage(w http.ResponseWriter, r *http.Request, data types.Data) {
	chatID := r.URL.Query().Get("id")
	if chatID == "" {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	if r.Method == "POST" {
		var progressChan chan int

		switch r.FormValue("formID") {
		case "sendMessage":
			progressChan = remoteserver.SendMessage(chatID, r.FormValue("message"))

		case "sendFile":
			// TODO
			//progressChan = remoteserver.SendFile(chatID, os.Open(r.FormValue("file")))

		default:
			data.Alert = "Unknown form"
		}

		if progressChan != nil {
			consts.AddListener(data.User, chatID, progressChan)
		}
	}

	t, err := template.ParseFiles("front/pages/template.html", "front/pages/blocks_user.html", "front/pages/chat.html")
	if err != nil {
		log.Println(err.Error())
	}

	data.Messages = saved.GetMessages(chatID)
	if data.Messages == nil {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	// TEMP
	data.Listeners = []string{
		chatID + ".0",
	}

	progressChan := remoteserver.SendMessage(chatID, r.FormValue("message"))
	if progressChan != nil {
		consts.AddListener(data.User, chatID+".0", progressChan)
		log.Printf("%v\n", chatID+".0")
	}
	// TEMP END

	for idx := range data.Messages {
		data.Messages[idx].Id = fmt.Sprint(idx)
	}

	data.Message = r.URL.Query().Get("id")
	if data.Message == "" {
		data.Message = "-1"
	}

	err = t.Execute(w, data)
	if err != nil {
		fmt.Println(err.Error())
	}
}

func updateChatsPage(w http.ResponseWriter, r *http.Request, data types.Data) {
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	flusher, ok := w.(http.Flusher)
	if !ok {
		fmt.Fprintf(w, "data: -1\n\n")
		return
	}
	defer flusher.Flush()

	user := data.User
	eventID := r.URL.Query().Get("name")
	if eventID == "" || user == "" {
		fmt.Fprintf(w, "data: -1\n\n")
		flusher.Flush()

		return
	}

	userUpdates := consts.EventListeners.Events[user]
	if userUpdates == nil {
		consts.EventListeners.Mu.Lock()
		consts.EventListeners.Events[user] = make(map[string]chan int)
		userUpdates = consts.EventListeners.Events[user]
		consts.EventListeners.Mu.Unlock()
	}

	if userUpdates[eventID] == nil {
		fmt.Fprintf(w, "data: 0\n\n")
		flusher.Flush()
		return
	}

	for {
		if userUpdates[eventID] == nil {
			fmt.Fprintf(w, "data: 0\n\n")
			flusher.Flush()

			break
		}

		progress := <-userUpdates[eventID]
		fmt.Fprintf(w, "data: %d\n\n", progress)
		flusher.Flush()

		if progress == 0 {
			break
		}
	}
}
