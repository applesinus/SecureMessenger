package handlers

import (
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"strings"
)

func chatsPage(w http.ResponseWriter, r *http.Request, data data) {
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

func regularChatsPage(w http.ResponseWriter, r *http.Request, data data) {
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

func secretChatsPage(w http.ResponseWriter, r *http.Request, data data) {
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

func newChatPage(w http.ResponseWriter, r *http.Request, data data) {
	t, err := template.ParseFiles("front/pages/template.html", "front/pages/blocks_user.html", "front/pages/newChat.html")
	if err != nil {
		log.Println(err.Error())
	}

	err = t.Execute(w, data)
	if err != nil {
		fmt.Println(err.Error())
	}
}

func chatPage(w http.ResponseWriter, r *http.Request, data data) {
	if r.Method == "POST" {
		switch r.FormValue("formID") {
		case "sendMessage":
			msg := r.FormValue("message")
			data.Messages = append(data.Messages, message{
				Author:  data.User,
				Message: msg,
			})

		case "sendFile":
			data.Alert = "File uploading is not supported yet"

		default:
			data.Alert = "Unknown form"
		}
	}

	t, err := template.ParseFiles("front/pages/template.html", "front/pages/blocks_user.html", "front/pages/chat.html")
	if err != nil {
		log.Println(err.Error())
	}

	// TEMP
	data.Name = "General Grievous"

	meme, err := os.Open("back/saved/meme.jpeg")
	if err != nil {
		log.Println(err.Error())
	} else {
		data.Messages = append(data.Messages, message{
			Author:  data.User,
			Message: strings.TrimPrefix(meme.Name(), "back/saved/"),
			Type:    "image",
		})

		data.Messages = append(data.Messages, message{
			Author:  data.User,
			Message: strings.TrimPrefix(meme.Name(), "back/saved/"),
			Type:    "file",
		})
	}

	data.Messages = append(data.Messages, []message{
		{
			Author:  data.User,
			Message: "Hello there",
			Type:    "text",
		},
		{
			Author:  "General Grievous",
			Message: "General Kenobi! You are a bold one.",
			Type:    "text",
		},
		{
			Author:  "General Grievous",
			Message: "Kill him!",
			Type:    "text",
		},
		{
			Author:  "Battle Droids",
			Message: "[Droids fail to kill Obi-Wan]",
			Type:    "text",
		},
		{
			Author:  "Battle Droids",
			Message: "[Other droids surround him]",
			Type:    "text",
		},
		{
			Author:  "General Grievous",
			Message: "Back away! I will deal with this Jedi scum myself!",
			Type:    "text",
		},
		{
			Author:  data.User,
			Message: "Your move!",
			Type:    "text",
		},
	}...)
	// TEMP END

	data.Message = r.URL.Query().Get("id")
	if data.Message == "" {
		data.Message = "-1"
	}

	err = t.Execute(w, data)
	if err != nil {
		fmt.Println(err.Error())
	}
}
