package handlers

import (
	"log"
	"messengerClient/back/remoteServer"
	"messengerClient/back/users"
	"messengerClient/consts"
	"messengerClient/types"
	"net/http"
	"strings"
	"sync"
)

const (
	expireTime = 604800
)

func Init() {
	avaliableUsers := users.GetUsers()

	consts.EventListeners = types.EventsType{
		Events: make(map[string]map[string]map[string]chan int),
		Mu:     new(sync.Mutex),
	}
	for users := range avaliableUsers {
		consts.EventListeners.Events[users] = make(map[string]map[string]chan int)
	}

	consts.Recievers = types.RecievedType{
		Events: make(map[string]map[string]chan []byte),
		Mu:     new(sync.Mutex),
	}
	for users := range avaliableUsers {
		consts.Recievers.Events[users] = make(map[string]chan []byte)
	}

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path
		if path[len(path)-1] == '/' {
			path = path[:len(path)-1]
		}

		log.Println("Request: " + path)

		if strings.Contains(path, "/getfile/") {
			http.StripPrefix("/getfile/", http.FileServer(http.Dir("back/saved"))).ServeHTTP(w, r)
			return
		}
		if strings.Contains(path, "/getFile/") {
			http.StripPrefix("/getFile/", http.FileServer(http.Dir("back/saved"))).ServeHTTP(w, r)
			return
		}

		if strings.Contains(path, "/getpage/") {
			http.StripPrefix("/getpage/", http.FileServer(http.Dir("front/pages"))).ServeHTTP(w, r)
			return
		}
		if strings.Contains(path, "/getPage/") {
			http.StripPrefix("/getPage/", http.FileServer(http.Dir("front/pages"))).ServeHTTP(w, r)
			return
		}

		ok, username := isLoggedIn(w, r)
		if !ok && path != "/login" && path != "/register" {
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}

		data := types.Data{
			User: username,
		}

		switch path {
		case "/login":
			loginPage(w, r)

		case "/register":
			registerPage(w, r)

		case "/chats":
			chatsPage(w, r, data)

		case "/chats/regular":
			regularChatsPage(w, r, data)

		case "/chats/secret":
			secretChatsPage(w, r, data)

		case "/chats/new":
			newChatPage(w, r, data)

		case "/chat":
			chatPage(w, r, data)

		case "/logout":
			logoutPage(w, r)

		case "/redirect":
			redirectPage(w, r)

		case "/chat/update":
			updateChatsPage(w, r, data)

		case "/chat/file":
			filePage(w, r, data)

		case "/chat/recieve":
			recieveChatPage(w, r, data)

		case "", "/main":
			mainPage(w, r, data)
		default:
			http.Redirect(w, r, "/main", http.StatusSeeOther)
		}
	})
}

func isLoggedIn(w http.ResponseWriter, r *http.Request) (bool, string) {
	user, erruser := r.Cookie("currentUser")
	password, errpassword := r.Cookie("currentPassword")

	if erruser != nil || errpassword != nil {
		updateCookie(w, "currentUser", "", expireTime)
		updateCookie(w, "currentPassword", "", expireTime)
		return false, ""
	}

	err := remoteServer.UserLogin(user.Value, password.Value)
	if err != nil {
		updateCookie(w, "currentUser", "", expireTime)
		updateCookie(w, "currentPassword", "", expireTime)
		return false, ""
	}

	return true, user.Value
}

func updateCookie(w http.ResponseWriter, cookieName, newVal string, expTime int) {
	cookie := &http.Cookie{
		Name:   cookieName,
		Value:  newVal,
		MaxAge: expTime,
	}
	http.SetCookie(w, cookie)
}
