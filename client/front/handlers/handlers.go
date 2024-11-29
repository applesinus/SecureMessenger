package handlers

import (
	"messengerClient/back/users"
	"net/http"
)

const (
	expireTime = 604800

	errNoUser        = "no user"
	errWrongPassword = "wrong password"
	errNotLogged     = "not logged in"
)

func Init() {
	http.HandleFunc("/login", loginPage)
	http.HandleFunc("/register", registerPage)

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path
		if path[len(path)-1] == '/' {
			path = path[:len(path)-1]
		}

		ok, username := isLoggedIn(w, r)
		if !ok {
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}

		data := data{
			User: username,
		}

		switch path {

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

		case "", "/main":
			mainPage(w, r, data)
		default:
			http.Redirect(w, r, "/main", http.StatusSeeOther)
		}
	})

	http.Handle("/getPage/", http.StripPrefix("/getPage/", http.FileServer(http.Dir("front/pages"))))
	http.Handle("/getpage/", http.StripPrefix("/getpage/", http.FileServer(http.Dir("front/pages"))))

	http.Handle("/getFile/", http.StripPrefix("/getFile/", http.FileServer(http.Dir("back/saved"))))
	http.Handle("/getfile/", http.StripPrefix("/getfile/", http.FileServer(http.Dir("back/saved"))))
}

func isLoggedIn(w http.ResponseWriter, r *http.Request) (bool, string) {
	user, erruser := r.Cookie("currentUser")
	password, errpassword := r.Cookie("currentPassword")

	if erruser != nil || errpassword != nil {
		updateCookie(w, "currentUser", "", expireTime)
		updateCookie(w, "currentPassword", "", expireTime)
		return false, ""
	}

	ok := users.CheckLogin(user.Value, password.Value)

	return ok, user.Value
}

func updateCookie(w http.ResponseWriter, cookieName, newVal string, expTime int) {
	cookie := &http.Cookie{
		Name:   cookieName,
		Value:  newVal,
		MaxAge: expTime,
	}
	http.SetCookie(w, cookie)
}
