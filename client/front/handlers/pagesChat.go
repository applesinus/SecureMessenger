package handlers

import (
	"fmt"
	"html/template"
	"log"
	"net/http"
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

	//data.Message = "FBI OPEN UP"

	err = t.Execute(w, data)
	if err != nil {
		fmt.Println(err.Error())
	}
}

func chatPage(w http.ResponseWriter, r *http.Request, data data) {
	t, err := template.ParseFiles("front/pages/template.html", "front/pages/blocks_user.html", "front/pages/chat.html")
	if err != nil {
		log.Println(err.Error())
	}

	data.Name = "General Grievous"
	data.Messages = []message{
		{
			Author:  "kekus",
			Message: "Hello there",
		},
		{
			Author:  "General Grievous",
			Message: "General Kenobi! You are a bold one.",
		},
		{
			Author:  "General Grievous",
			Message: "Kill him!",
		},
		{
			Author:  "Battle Droids",
			Message: "[Droids fail to kill Obi-Wan]",
		},
		{
			Author:  "Battle Droids",
			Message: "[Other droids surround him]",
		},
		{
			Author:  "General Grievous",
			Message: "Back away! I will deal with this Jedi scum myself!",
		},
		{
			Author:  "kekus",
			Message: "Your move!",
		},
	}

	err = t.Execute(w, data)
	if err != nil {
		fmt.Println(err.Error())
	}
}
