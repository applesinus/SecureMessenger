package handlers

import (
	"fmt"
	"html/template"
	"log"
	"messengerClient/back/crypto"
	"messengerClient/back/users"
	"net/http"
)

func mainPage(w http.ResponseWriter, r *http.Request, data data) {
	t, err := template.ParseFiles("front/pages/template.html", "front/pages/blocks_user.html", "front/pages/main.html")
	if err != nil {
		log.Println(err.Error())
		return
	}

	err = t.Execute(w, data)
	if err != nil {
		log.Println(err.Error())
	}
}

func redirectPage(w http.ResponseWriter, r *http.Request) {
	redirectPath := r.URL.Query().Get("path")
	if redirectPath == "" {
		redirectPath = "/"
	}

	http.Redirect(w, r, redirectPath, http.StatusSeeOther)
}

func loginPage(w http.ResponseWriter, r *http.Request) {
	ok, _ := isLoggedIn(w, r)
	if ok {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	if r.Method == "POST" {
		username := r.FormValue("login")
		password := r.FormValue("password")
		password = crypto.Hash(password)
		ok := users.Login(username, password)
		if !ok {
			http.Redirect(w, r, "/login?error=Wrong login or password", http.StatusSeeOther)
			return
		}
		updateCookie(w, "currentUser", username, expireTime)
		updateCookie(w, "currentPassword", password, expireTime)
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	t, err := template.ParseFiles("front/pages/template.html", "front/pages/blocks_notuser.html", "front/pages/login.html")
	if err != nil {
		log.Println(err.Error())
	}

	err = t.Execute(w, nil)
	if err != nil {
		fmt.Println(err.Error())
	}
}

func registerPage(w http.ResponseWriter, r *http.Request) {
	ok, _ := isLoggedIn(w, r)
	if ok {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	if r.Method == "POST" {
		username := r.FormValue("login")
		if len(username) < 3 {
			http.Redirect(w, r, "/register?error=Too short login", http.StatusSeeOther)
			return
		}
		if len(username) > 15 {
			http.Redirect(w, r, "/register?error=Too long login", http.StatusSeeOther)
			return
		}
		for _, ch := range username {
			if !((ch >= 'a' && ch <= 'z') || (ch >= 'A' && ch <= 'Z') || (ch >= '0' && ch <= '9')) {
				http.Redirect(w, r, "/register?error=Invalid login", http.StatusSeeOther)
				return
			}
		}

		password := r.FormValue("password")
		if len(password) < 8 {
			http.Redirect(w, r, "/register?error=Too short password", http.StatusSeeOther)
			return
		}

		res := users.Register(w, username, password)
		if res != "ok" {
			http.Redirect(w, r, "/register?error="+res, http.StatusSeeOther)
			return
		}

		updateCookie(w, "currentUser", username, expireTime)
		updateCookie(w, "currentPassword", crypto.Hash(password), expireTime)
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	message := data{
		Message: "",
	}
	msg := r.URL.Query().Get("error")
	if msg != "" {
		message.Message = msg
	}

	t, err := template.ParseFiles("front/pages/template.html", "front/pages/blocks_notuser.html", "front/pages/register.html")
	if err != nil {
		log.Println(err.Error())
	}

	err = t.Execute(w, message)
	if err != nil {
		fmt.Println(err.Error())
	}
}

func logoutPage(w http.ResponseWriter, r *http.Request) {
	updateCookie(w, "currentUser", "", 1)
	updateCookie(w, "currentPassword", "", 1)

	http.Redirect(w, r, "/login", http.StatusSeeOther)
}
