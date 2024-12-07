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
	"strconv"
	"strings"
	"time"
)

// Done
func chatsPage(w http.ResponseWriter, r *http.Request, data types.Data) {
	t, err := template.ParseFiles("front/pages/template.html", "front/pages/blocks_user.html", "front/pages/chats.html")
	if err != nil {
		log.Printf("[FRONT][TEMPLATE] Error parsing template: %s", err)
		return
	}

	password, err := r.Cookie("currentPassword")
	if err != nil {
		log.Printf("[FRONT][TEMPLATE] Error getting cookie: %s", err)
		http.Redirect(w, r, "/login", http.StatusSeeOther)
	}

	data.SecretChats, err = remoteserver.GetUserSecretChats(data.User, password.Value)
	if err != nil {
		log.Printf("[FRONT][TEMPLATE] Error getting secret chats: %s", err)
		http.Redirect(w, r, "/", http.StatusSeeOther)
	}

	data.RegularChats, err = remoteserver.GetUserChats(data.User, password.Value)
	if err != nil {
		log.Printf("[FRONT][TEMPLATE] Error getting regular chats: %s", err)
		http.Redirect(w, r, "/", http.StatusSeeOther)
	}

	data.Message = "Your chats"

	err = t.Execute(w, data)
	if err != nil {
		log.Printf("[FRONT][TEMPLATE] Error executing template: %s", err)
	}
}

// Done
func regularChatsPage(w http.ResponseWriter, r *http.Request, data types.Data) {
	t, err := template.ParseFiles("front/pages/template.html", "front/pages/blocks_user.html", "front/pages/chats.html")
	if err != nil {
		log.Printf("[FRONT][TEMPLATE] Error parsing template: %s", err)
	}

	password, err := r.Cookie("currentPassword")
	if err != nil {
		log.Printf("[FRONT][TEMPLATE] Error getting cookie: %s", err)
	}

	data.RegularChats, err = remoteserver.GetUserChats(data.User, password.Value)
	if err != nil {
		log.Printf("[FRONT][TEMPLATE] Error getting regular chats: %s", err)
		http.Redirect(w, r, "/", http.StatusSeeOther)
	}

	data.Message = "Regular chats"

	err = t.Execute(w, data)
	if err != nil {
		log.Printf("[FRONT][TEMPLATE] Error executing template: %s", err)
	}
}

// Done
func secretChatsPage(w http.ResponseWriter, r *http.Request, data types.Data) {
	t, err := template.ParseFiles("front/pages/template.html", "front/pages/blocks_user.html", "front/pages/chats.html")
	if err != nil {
		log.Printf("[FRONT][TEMPLATE] Error parsing template: %s", err)
	}

	password, err := r.Cookie("currentPassword")
	if err != nil {
		log.Printf("[FRONT][TEMPLATE] Error getting cookie: %s", err)
	}

	data.SecretChats, err = remoteserver.GetUserSecretChats(data.User, password.Value)
	if err != nil {
		log.Printf("[FRONT][TEMPLATE] Error getting secret chats: %s", err)
		http.Redirect(w, r, "/", http.StatusSeeOther)
	}

	data.Message = "Secret chats"

	err = t.Execute(w, data)
	if err != nil {
		log.Printf("[FRONT][TEMPLATE] Error executing template: %s", err)
	}
}

func newChatPage(w http.ResponseWriter, r *http.Request, data types.Data) {
	if r.Method == "POST" {
		password, err := r.Cookie("currentPassword")
		if err != nil {
			http.Redirect(w, r, "/chats/new?alert=You are not logged in", http.StatusSeeOther)
			return
		}

		reciever := r.FormValue("name")
		if reciever == "" {
			http.Redirect(w, r, "/chats/new?alert=Empty name", http.StatusSeeOther)
			return
		}

		chatType := r.FormValue("chatType")
		switch chatType {
		case "regular":
			chatType = consts.EncriptionNo
		case "magenta":
			chatType = consts.EncriptionMagenta
		case "rc6":
			chatType = consts.EncriptionRC6
		default:
			http.Redirect(w, r, "/chats/new?alert=Invalid chat type", http.StatusSeeOther)
			return
		}

		chatID, err := saved.NewChat(data.User, password.Value, reciever, chatType)
		if err != nil {
			http.Redirect(w, r, fmt.Sprintf("/chats/new?alert=Error creating chat: %s", err), http.StatusSeeOther)
			return
		}

		http.Redirect(w, r, "/chat?id="+chatID, http.StatusSeeOther)
		return
	}

	t, err := template.ParseFiles("front/pages/template.html", "front/pages/blocks_user.html", "front/pages/newChat.html")
	if err != nil {
		log.Printf("[FRONT][TEMPLATE] Error parsing template: %s", err)
	}

	data.Alert = r.URL.Query().Get("alert")

	err = t.Execute(w, data)
	if err != nil {
		log.Printf("[FRONT][TEMPLATE] Error executing template: %s", err)
	}
}

func chatPage(w http.ResponseWriter, r *http.Request, data types.Data) {
	chatID := r.URL.Query().Get("id")
	if chatID == "" {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	parts := strings.Split(chatID, "-")
	if len(parts) != 2 {
		log.Println("Invalid chatID:", chatID)
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	password, err := r.Cookie("currentPassword")
	if err != nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	if r.Method == "POST" {
		var progressChan chan int

		if r.FormValue("message") != "" {
			message := types.Message{
				Author:  data.User,
				Message: []byte(r.FormValue("message")),
				Type:    "text",
				Id:      fmt.Sprintf("%v", time.Now().UTC().Format(time.StampNano)),
			}

			log.Printf("[BACKEND][MESSAGE] Sending message: %s", message.Message)

			progressChan = remoteserver.SendMessage(data.User, password.Value, parts[0], parts[1], message)
			saved.AddMessage(data.User, chatID, message)
			consts.AddListener(data.User, chatID, message.Id, progressChan)
		}

		if file, _, _ := r.FormFile("file"); file != nil {
			_, handler, err := r.FormFile("file")
			if err != nil {
				return
			}

			msgType := "file/"
			if handler.Header.Get("Content-Type") == "image/jpeg" {
				msgType = "image/"
			}

			message := types.Message{
				Author: data.User,
				Type:   fmt.Sprintf("%s%s", msgType, handler.Filename),
				Id:     fmt.Sprintf("%v", time.Now().UTC().Format(time.StampNano)),
			}

			log.Println("[BACKEND][MESSAGE] Sending file")

			progressChan, chBytes := remoteserver.SendFile(data.User, password.Value, parts[0], parts[1], message, r)

			saved.AddMessage(data.User, chatID, message)
			consts.AddListener(data.User, chatID, message.Id, progressChan)

			go func() {
				fileContents := <-chBytes
				saved.AddFile(data.User, chatID, message.Id, fileContents)
				close(chBytes)
			}()
		}
	}

	t, err := template.ParseFiles("front/pages/template.html", "front/pages/blocks_user.html", "front/pages/chat.html")
	if err != nil {
		log.Printf("[FRONT][TEMPLATE] Error parsing template: %s", err)
	}

	log.Printf("[FRONT][TEMPLATE] Chat id: %s", chatID)

	data.Name = parts[0]
	data.Message = parts[1]

	data.Messages, err = saved.GetMessages(data.User, password.Value, parts[0], parts[1])
	if err != nil {
		log.Printf("[FRONT][TEMPLATE] Error getting messages: %s", err)
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	data.Listeners = make([]string, 0)
	for key := range consts.EventListeners.Events[data.User][chatID] {
		data.Listeners = append(data.Listeners, key)
		log.Printf("[FRONT][TEMPLATE] Listener: %s", key)
	}

	funcMap := template.FuncMap{
		"contains": strings.Contains,
	}

	err = t.Funcs(funcMap).Execute(w, data)
	if err != nil {
		log.Printf("[FRONT][TEMPLATE] Error executing template: %s", err)
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

	chatId := r.URL.Query().Get("chat")
	if chatId == "" {
		fmt.Fprintf(w, "data: -1\n\n")
		flusher.Flush()

		return
	}

	if consts.EventListeners.Events[user] == nil {
		consts.EventListeners.Mu.Lock()
		consts.EventListeners.Events[user] = make(map[string]map[string]chan int)
		consts.EventListeners.Mu.Unlock()
	}

	if consts.EventListeners.Events[user][chatId] == nil {
		consts.EventListeners.Mu.Lock()
		consts.EventListeners.Events[user][chatId] = make(map[string]chan int)
		consts.EventListeners.Mu.Unlock()
	}

	if _, ok := consts.EventListeners.Events[user][chatId][eventID]; !ok {
		fmt.Fprintf(w, "data:\n\n")
		flusher.Flush()
		return
	}

	if consts.EventListeners.Events[user][chatId][eventID] == nil {
		fmt.Fprintf(w, "data: 0\n\n")
		flusher.Flush()
		consts.RemoveListener(user, chatId, eventID)
		return
	}

	progress := <-consts.EventListeners.Events[user][chatId][eventID]
	fmt.Fprintf(w, "data: %d\n\n", progress)
	if progress == 0 {
		consts.RemoveListener(user, chatId, eventID)
	}
	flusher.Flush()
}

func recieveChatPage(w http.ResponseWriter, r *http.Request, data types.Data) {
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	flusher, ok := w.(http.Flusher)
	if !ok {
		log.Printf("[FRONT][RECIEVER] Error on creating flusher: %s", ok)
		return
	}
	defer flusher.Flush()

	user := data.User
	chatID := r.URL.Query().Get("id")

	chatUpdates := consts.Recievers.Events[user][chatID]
	for chatUpdates != nil {
		message := <-chatUpdates
		fmt.Fprintf(w, "data: %s\n\n", message)
		flusher.Flush()
	}
}

func filePage(w http.ResponseWriter, r *http.Request, data types.Data) {
	chat := r.URL.Query().Get("chat")
	if chat == "" {
		return
	}

	msgId := r.URL.Query().Get("id")
	if msgId == "" {
		return
	}

	msg, err := saved.GetMessage(data.User, chat, msgId)
	if err != nil {
		return
	}

	filename := msg.Type
	if strings.HasPrefix(filename, "image/") {
		filename = strings.TrimPrefix(filename, "image/")
	} else if strings.HasPrefix(filename, "file/") {
		filename = strings.TrimPrefix(filename, "file/")
	} else {
		return
	}

	body := []byte(msg.Message)

	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", filename))

	contentType := http.DetectContentType(body)
	w.Header().Set("Content-Type", contentType)

	w.Header().Set("Content-Length", strconv.Itoa(len(body)))
	w.WriteHeader(http.StatusOK)

	_, err = w.Write(body)
	if err != nil {
		log.Printf("[FRONT][FILE] Error writing file: %s", err)
	}
}
