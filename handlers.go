package main

import (
	"html/template"
	"net/http"
	"time"

	"src/github.com/julienschmidt/httprouter"
)

func mainGetHandler(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	c, err := r.Cookie("username")
	if err == http.ErrNoCookie {
		tpl, err := template.ParseFiles("templates/noCookie.html", "templates/footer.html", "templates/noCookieHeader.html")
		if err != nil {
			panic(err)
		}

		err = tpl.ExecuteTemplate(w, "noCookie", nil)
		if err != nil {
			panic(err)
		}
	} else {
		username := c.Value
		http.Redirect(w, r, "/users/"+username, http.StatusFound)
		return
	}
}

func mainPostHandler(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	usernameCookie, err := r.Cookie("username")
	if err != http.ErrNoCookie {
		http.Redirect(w, r, "/users/"+usernameCookie.Value, http.StatusFound)
		return
	}
	if tryToLogIn(r.FormValue("username"), r.FormValue("password")) == Correct {
		usernameCookie := &http.Cookie{
			Name:    "username",
			Value:   r.FormValue("username"),
			Expires: time.Now().Add(10 * time.Minute),
			Path:    "/",
		}
		http.SetCookie(w, usernameCookie)
		http.Redirect(w, r, "/users/"+usernameCookie.Value, http.StatusFound)
		return
	}
	if tryToLogIn(r.FormValue("username"), r.FormValue("password")) == NoMatch {
		http.Redirect(w, r, "/incorrectPassword", http.StatusFound)
		return
	}
	if tryToLogIn(r.FormValue("username"), r.FormValue("password")) == WrongPassword {
		http.Redirect(w, r, "/incorrectPassword", http.StatusFound)
	}
}

func logoutHandler(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	_, err := r.Cookie("username")
	if err == http.ErrNoCookie {
		http.Redirect(w, r, "/", http.StatusFound)
		return
	}
	c := &http.Cookie{
		Name:   "username",
		MaxAge: -1,
	}
	http.SetCookie(w, c)
	http.Redirect(w, r, "/", http.StatusFound)
	return
}

func newPostGetHandler(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	_, err := r.Cookie("username")
	if err == http.ErrNoCookie {
		http.Redirect(w, r, "/", http.StatusFound)
		return
	}
	tpl, err := template.ParseFiles("templates/header.html", "templates/newPost.html")
	if err != nil {
		panic(err)
	}
	us := getUser(ps.ByName("username"))

	err = tpl.ExecuteTemplate(w, "newPost", us)
	if err != nil {
		panic(err)
	}
}

func newPostPostHandler(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	_, err := r.Cookie("username")
	if err == http.ErrNoCookie {
		http.Redirect(w, r, "/", http.StatusFound)
		return
	}
	username := ps.ByName("username")

	newPost := Post{
		Title: r.FormValue("title"),
		Body:  r.FormValue("body"),
		Date:  time.Now().Format(timeFormat),
	}

	err = getUser(username).addPost(newPost)
	if err != nil {
		http.Redirect(w, r, "/users/"+username+"/newPostInvalidSymbols", http.StatusFound)
		return
	}
	err = getUser(username).refreshUserInfo()
	if err != nil {
		panic(err)
	}

	http.Redirect(w, r, "/users/"+username, http.StatusFound)
}

func newPostInvalidSymbolsGetHandler(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	_, err := r.Cookie("username")
	if err == http.ErrNoCookie {
		http.Redirect(w, r, "/", http.StatusFound)
		return
	}
	tpl, err := template.ParseFiles("templates/header.html", "templates/newPostInvalidSymbols.html")
	if err != nil {
		panic(err)
	}
	us := getUser(ps.ByName("username"))

	err = tpl.ExecuteTemplate(w, "newPostInvalidSymbols", us)
	if err != nil {
		panic(err)
	}
}

func registerGetHandler(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	usernameCookie, err := r.Cookie("username")
	if err != http.ErrNoCookie {
		http.Redirect(w, r, "/users/"+usernameCookie.Value, http.StatusFound)
		return
	}
	tpl, err := template.ParseFiles("templates/noCookieHeader.html", "templates/register.html")
	if err != nil {
		panic(err)
	}

	err = tpl.ExecuteTemplate(w, "register", nil)
	if err != nil {
		panic(err)
	}
}

func registerPostHandler(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	usernameCookie, err := r.Cookie("username")
	if err != http.ErrNoCookie {
		http.Redirect(w, r, "/users/"+usernameCookie.Value, http.StatusFound)
		return
	}
	incAccount := r.FormValue("account")
	incPassword := r.FormValue("password")
	if UsernameExists(incAccount) {
		http.Redirect(w, r, "/registerUsernameAlreadyTaken", http.StatusFound)
		return
	}
	err = addUserToServer(incAccount, incPassword)
	if err != nil {
		http.Redirect(w, r, "/registerInvalidSymbols", http.StatusFound)
		return
	}
	err = addUserToUsers(incAccount, incPassword)
	if err != nil {
		http.Redirect(w, r, "/registerInvalidSymbols", http.StatusFound)
		return
	}

	setID()
	registerSuccessCookie := http.Cookie{
		Name:   "registerSuccess",
		Value:  "true",
		MaxAge: 60,
	}
	http.SetCookie(w, &registerSuccessCookie)
	http.Redirect(w, r, "/registerSuccess", http.StatusFound)
}

func registerUsernameAlreadyTakenGetHandler(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	usernameCookie, err := r.Cookie("username")
	if err != http.ErrNoCookie {
		http.Redirect(w, r, "/users/"+usernameCookie.Value, http.StatusFound)
		return
	}
	tpl, err := template.ParseFiles("templates/noCookieHeader.html", "templates/registerUsernameAlreadyTaken.html")
	if err != nil {
		panic(err)
	}

	err = tpl.ExecuteTemplate(w, "registerUsernameAlreadyTaken", nil)
	if err != nil {
		panic(err)
	}
}

func registerInvalidSymbolsGetHandler(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	usernameCookie, err := r.Cookie("username")
	if err != http.ErrNoCookie {
		http.Redirect(w, r, "/users/"+usernameCookie.Value, http.StatusFound)
		return
	}
	tpl, err := template.ParseFiles("templates/noCookieHeader.html", "templates/registerInvalidSymbols.html")
	if err != nil {
		panic(err)
	}

	err = tpl.ExecuteTemplate(w, "registerInvalidSymbols", nil)
	if err != nil {
		panic(err)
	}
}

func registerSuccessHandler(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	usernameCookie, err := r.Cookie("username")
	if err != http.ErrNoCookie {
		http.Redirect(w, r, "/users/"+usernameCookie.Value, http.StatusFound)
		return
	}
	registerSuccessCookie, err := r.Cookie("registerSuccess")
	if err == http.ErrNoCookie {
		http.Redirect(w, r, "/", http.StatusFound)
		return
	}
	registerSuccessCookie.MaxAge = -1
	tpl, err := template.ParseFiles("templates/noCookieHeader.html", "templates/registerSuccess.html")
	if err != nil {
		panic(err)
	}

	err = tpl.ExecuteTemplate(w, "registerSuccess", nil)
	if err != nil {
		panic(err)
	}
}

func incorrectPasswordGetHandler(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	usernameCookie, err := r.Cookie("username")
	if err != http.ErrNoCookie {
		http.Redirect(w, r, "/users/"+usernameCookie.Value, http.StatusFound)
		return
	}
	tpl, err := template.ParseFiles("templates/noCookieHeader.html", "templates/incorrectPassword.html")
	if err != nil {
		panic(err)
	}

	err = tpl.ExecuteTemplate(w, "incorrectPassword", nil)
	if err != nil {
		panic(err)
	}
}

func userListHandler(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	_, err := r.Cookie("username")
	if err == http.ErrNoCookie {
		http.Redirect(w, r, "/", http.StatusFound)
		return
	}
	tpl, err := template.ParseFiles("templates/header.html", "templates/userList.html")
	if err != nil {
		panic(err)
	}

	err = tpl.ExecuteTemplate(w, "userList", users)
	if err != nil {
		panic(err)
	}
}

func usersHandler(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	username := ps.ByName("username")
	usernameCookie, err := r.Cookie("username")
	if err == http.ErrNoCookie {
		http.Redirect(w, r, "/", http.StatusFound)
		return
	}
	if usernameCookie.Value != username {
		user := getUser(username)
		tpl, err := template.ParseFiles("templates/header.html", "templates/userPage.html")
		if err != nil {
			panic(err)
		}

		err = tpl.ExecuteTemplate(w, "userPage", user)
		if err != nil {
			panic(err)
		}
	} else {
		user := getUser(username)
		tpl, err := template.ParseFiles("templates/header.html", "templates/homePage.html")
		if err != nil {
			panic(err)
		}

		err = tpl.ExecuteTemplate(w, "homePage", user)
		if err != nil {
			panic(err)
		}
	}
}