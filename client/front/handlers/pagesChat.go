package handlers

import (
	"fmt"
	"html/template"
	"io"
	"log"
	remoteserver "messengerClient/back/remoteServer"
	"messengerClient/back/saved"
	"messengerClient/consts"
	"messengerClient/types"
	"net/http"
	"os"
	"strings"
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
			http.Redirect(w, r, "/newChat?alert=You are not logged in", http.StatusSeeOther)
			return
		}

		recipient := r.FormValue("name")
		if recipient == "" {
			http.Redirect(w, r, "/newChat?alert=Empty name", http.StatusSeeOther)
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
			http.Redirect(w, r, "/newChat?alert=Invalid chat type", http.StatusSeeOther)
			return
		}

		chatID, err := saved.NewChat(data.User, password.Value, recipient, chatType)
		if err != nil {
			http.Redirect(w, r, fmt.Sprintf("/newChat?alert=Error creating chat: %s", err), http.StatusSeeOther)
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

	password, err := r.Cookie("currentPassword")
	if err != nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	if r.Method == "POST" {
		var progressChan chan int

		if r.FormValue("message") != "" {
			progressChan = remoteserver.SendMessage(chatID, r.FormValue("message"))
			saved.AddMessage(data.User, chatID, types.Message{Author: data.User, Type: "text", Message: r.FormValue("message")})
		}

		if r.FormValue("file") != "" {
			r.ParseMultipartForm(10 << 20)
			file, handler, err := r.FormFile("file")

			if err != nil {
				log.Printf("[BACKEND][FILE UPLOAD] Error opening file: %s", err)
				return
			}
			defer file.Close()

			fileBytes, err := io.ReadAll(file)
			if err != nil {
				log.Printf("[BACKEND][FILE UPLOAD] Error reading file: %s", err)
				return
			}

			parts := strings.Split(handler.Filename, ".")
			filename := strings.Join(parts[:len(parts)-1], ".")
			extension := parts[len(parts)-1]
			if _, err := os.Stat(fmt.Sprintf("back/saved/%s.%s", filename, extension)); err == nil {
				for i := 1; ; i++ {
					if _, err := os.Stat(fmt.Sprintf("back/saved/%s(%d).%s", filename, i, extension)); err == nil {
						i++
					} else {
						filename = fmt.Sprintf("%s(%d).%s", filename, i, extension)
						break
					}
				}
			} else {
				filename = filename + "." + extension
			}

			savedFile, err := os.Create("back/saved/" + filename)
			if err != nil {
				log.Printf("[BACKEND][FILE UPLOAD] Error creating file: %s", err)
			}
			defer savedFile.Close()
			savedFile.Write(fileBytes)

			progressChan = remoteserver.SendFile(chatID, savedFile)
			saved.AddMessage(data.User, chatID, types.Message{Author: data.User, Type: "file", Message: filename})
		}

		if progressChan != nil {
			consts.AddListener(data.User, chatID, progressChan)
		}
	}

	t, err := template.ParseFiles("front/pages/template.html", "front/pages/blocks_user.html", "front/pages/chat.html")
	if err != nil {
		log.Printf("[FRONT][TEMPLATE] Error parsing template: %s", err)
	}

	data.Messages, err = saved.GetMessages(data.User, password.Value, chatID)
	if err != nil {
		log.Printf("[FRONT][TEMPLATE] Error getting messages: %s", err)
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	data.Name = strings.TrimRight(data.Name, ":")
	if data.Messages == nil {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	// TEMP
	/*data.Listeners = []string{
		chatID + ".0",
	}

	progressChan := remoteserver.SendMessage(chatID, r.FormValue("message"))
	if progressChan != nil {
		consts.AddListener(data.User, chatID+".0", progressChan)
	}*/
	// TEMP END

	data.Message = r.URL.Query().Get("id")
	if data.Message == "" {
		data.Message = "-1"
	}

	err = t.Execute(w, data)
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
