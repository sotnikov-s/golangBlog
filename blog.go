package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"io/ioutil"
	"net/http"
	"time"
)

// base format: Mon Jan 2 15:04:05 -0700 MST 2006
var timeFormat = "01.02.2006 15:04:05"
var ID = 1
var users []*User

const (
	NoMatch       = "no match"
	WrongPassword = "wrong password"
	Correct       = "correct"
)

type User struct {
	Username  string
	Password  string
	ID        int
	PostCount int
	Posts     []Post
}

func getUser(username string) *User {
	for _, us := range users {
		if username == us.Username {
			return us
		}
	}
	return nil
}

func (u *User) addPost(post Post) {
	u.Posts = append(u.Posts, post)
	u.PostCount++
}

func (u User) reversePosts() User {
	rp := make([]Post, len(u.Posts), cap(u.Posts))
	for i := range u.Posts {
		rp = append(rp, u.Posts[len(u.Posts)-1-i])
	}
	u.Posts = rp
	return u
}

type Post struct {
	Title string
	Body  string
	Date  string
}

func (u User) NoPosts() bool {
	return u.PostCount == 0
}

func startServer(addr string) {
	files, err := ioutil.ReadDir("data/accounts")
	if err != nil {
		panic(err)
	}
	for _, file := range files {
		data, err := ioutil.ReadFile("data/accounts/" + file.Name())
		us := User{}
		err = json.Unmarshal(data, &us)
		if err != nil {
			panic(err)
		}
		users = append(users, &us)
		ID++
	}

	fmt.Println("Starting server at", addr)
	http.ListenAndServe(addr, nil)
}

func mainHandler(w http.ResponseWriter, r *http.Request) {
	c, err := r.Cookie("username")
	if err == http.ErrNoCookie || c.Value == "" {
		mainNoCookieHandler(w, r)
	} else {
		username := c.Value
		http.Redirect(w, r, "/users/"+username, http.StatusFound)
		return
	}
}

func mainNoCookieHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		mainNoCookiePostHandler(w, r)
	} else {
		tpl, err := template.ParseFiles("templates/noCookie.html", "templates/footer.html", "templates/noCookieHeader.html")
		if err != nil {
			panic(err)
		}
		tpl.ExecuteTemplate(w, "noCookie", nil)
		return
	}
}

func mainNoCookiePostHandler(w http.ResponseWriter, r *http.Request) {
	if tryToLogIn(r.FormValue("username"), r.FormValue("password")) == Correct {
		usernameCookie := &http.Cookie{
			Name:    "username",
			Value:   r.FormValue("username"),
			Expires: time.Now().Add(10 * time.Hour),
			Path: "/",
		}
		http.SetCookie(w, usernameCookie)
		http.Redirect(w, r, "/", http.StatusFound)
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

func logoutHandler(w http.ResponseWriter, r *http.Request) {
	_, err := r.Cookie("username")
	if err == http.ErrNoCookie {
		http.Redirect(w, r, "/", http.StatusFound)
		return
	}
	c := &http.Cookie{
		Name:    "username",
		MaxAge: -1,
	}
	http.SetCookie(w, c)
	http.Redirect(w, r, "/", http.StatusFound)
	return
}

func newPostHandler(w http.ResponseWriter, r *http.Request) {
	usernameCookie, err := r.Cookie("username")
	if err == http.ErrNoCookie {
		http.Redirect(w, r, "/", http.StatusFound)
		return
	}
	if r.Method != http.MethodPost {
		tpl, err := template.ParseFiles("templates/header.html", "templates/newPost.html")
		if err != nil {
			panic(err)
		}
		tpl.ExecuteTemplate(w, "newPost", nil)
	} else {
		username := usernameCookie.Value
		newPost := Post{
			Title: r.FormValue("title"),
			Body:  r.FormValue("body"),
			Date:  time.Now().Format(timeFormat),
		}
		getUser(username).addPost(newPost)
		getUser(username).addUserToServer()
		http.Redirect(w, r, "/", http.StatusFound)
	}
}

func tryToLogIn(incLogin string, incPassword string) string {
	if !UsernameExists(incLogin) {
		return NoMatch
	} else {
		for _, us := range users {
			if incLogin != us.Username {
				continue
			} else {
				if us.Password != incPassword {
					return WrongPassword
				}
			}
		}
	}
	return Correct
}

func setID() {
	ID++
}

func addUserToServer(incLogin string, incPassword string) {
	newUser := &User{
		Username:  incLogin,
		Password:  incPassword,
		ID:        ID,
		PostCount: 0,
		Posts:     make([]Post, 0),
	}
	parsedNewUser, err := json.Marshal(newUser)
	if err != nil {
		panic(err)
	}

	ioutil.WriteFile("data/accounts/"+incLogin+".txt", parsedNewUser, 0600)
}

func (u User) addUserToServer() {
	parsedNewUser, err := json.Marshal(u)
	if err != nil {
		panic(err)
	}

	ioutil.WriteFile("data/accounts/"+u.Username+".txt", parsedNewUser, 0600)
}

func addUserToUsers(incLogin string, incPassword string) {
	newUser := &User{
		Username:  incLogin,
		Password:  incPassword,
		ID:        ID,
		PostCount: 0,
		Posts:     make([]Post, 0),
	}

	users = append(users, newUser)
}

func registerHandler(w http.ResponseWriter, r *http.Request) {
	usernameCookie, err := r.Cookie("username")
	if err != http.ErrNoCookie {
		http.Redirect(w, r, "/users/"+usernameCookie.Value, http.StatusFound)
		return
	}
	if r.Method != http.MethodPost {
		tpl, err := template.ParseFiles("templates/noCookieHeader.html", "templates/register.html")
		if err != nil {
			panic(err)
		}
		tpl.ExecuteTemplate(w, "register", nil)
	} else {
		incAccount := r.FormValue("account")
		incPassword := r.FormValue("password")
		if UsernameExists(incAccount) {
			http.Redirect(w, r, "/registerAlreadyTaken", http.StatusFound)
			return
		}
		addUserToServer(incAccount, incPassword)
		addUserToUsers(incAccount, incPassword)
		setID()
		registerSuccessCookie := http.Cookie{
			Name:    "registerSuccess",
			Value:   "true",
			MaxAge: 1,
		}
		http.SetCookie(w, &registerSuccessCookie)
		http.Redirect(w, r, "/registerSuccess", http.StatusFound)
	}
}

func registerUsernameAlreadyTakenHandler(w http.ResponseWriter, r *http.Request) {
	usernameCookie, err := r.Cookie("username")
	if err != http.ErrNoCookie {
		http.Redirect(w, r, "/users/"+usernameCookie.Value, http.StatusFound)
		return
	}
	if r.Method != http.MethodPost {
		tpl, err := template.ParseFiles("templates/noCookieHeader.html", "templates/registerUsernameAlreadyTaken.html")
		if err != nil {
			panic(err)
		}
		tpl.ExecuteTemplate(w, "registerUsernameAlreadyTaken", nil)
	} else {
		incAccount := r.FormValue("account")
		incPassword := r.FormValue("password")
		if UsernameExists(incAccount) {
			http.Redirect(w, r, "/registerAlreadyTaken", http.StatusFound)
			return
		}
		addUserToServer(incAccount, incPassword)
		addUserToUsers(incAccount, incPassword)
		setID()
		registerSuccessCookie := http.Cookie{
			Name:    "registerSuccess",
			Value:   "true",
			Expires: time.Now().Add(1 * time.Second),
		}
		http.SetCookie(w, &registerSuccessCookie)
		http.Redirect(w, r, "/registerSuccess", http.StatusFound)
	}
}

func UsernameExists(username string) bool {
	for _, us := range users {
		if username == us.Username {
			return true
		}
	}
	return false
}

func registerSuccessHandler(w http.ResponseWriter, r *http.Request) {
	usernameCookie, err := r.Cookie("username")
	if err != http.ErrNoCookie {
		http.Redirect(w, r, "/users/"+usernameCookie.Value, http.StatusFound)
		return
	}
	_, err = r.Cookie("registerSuccess")
	if err == http.ErrNoCookie {
		http.Redirect(w, r, "/", http.StatusFound)
		return
	}
	tpl, err := template.ParseFiles("templates/noCookieHeader.html", "templates/registerSuccess.html")
	if err != nil {
		panic(err)
	}
	tpl.ExecuteTemplate(w, "registerSuccess", nil)

}

func incorrectPasswordHandler(w http.ResponseWriter, r *http.Request) {
	usernameCookie, err := r.Cookie("username")
	if err != http.ErrNoCookie {
		http.Redirect(w, r, "/users/"+usernameCookie.Value, http.StatusFound)
		return
	}
	tpl, err := template.ParseFiles("templates/noCookieHeader.html", "templates/incorrectPassword.html")
	if err != nil {
		panic(err)
	}
	tpl.ExecuteTemplate(w, "incorrectPassword", nil)
}

func userListHandler(w http.ResponseWriter, r *http.Request) {
	_, err := r.Cookie("username")
	if err == http.ErrNoCookie {
		http.Redirect(w, r, "/", http.StatusFound)
		return
	}
	tpl, err := template.ParseFiles("templates/header.html", "templates/userList.html")
	if err != nil {
		panic(err)
	}
	tpl.ExecuteTemplate(w, "userList", users)
}

func usersHandler(w http.ResponseWriter, r *http.Request) {
	username := r.URL.String()[7:]
	usernameCookie, err := r.Cookie("username")
	if err == http.ErrNoCookie {
		http.Redirect(w, r, "/", http.StatusFound)
		return
	}
	if usernameCookie.Value != username {
		user := getUser(username)
		tpl, err := template.ParseFiles("templates/userPage.html", "templates/header.html")
		if err != nil {
			panic(err)
		}
		tpl.ExecuteTemplate(w, "userPage", user)
	} else {
		user := getUser(username)
		tpl, err := template.ParseFiles("templates/homePage.html", "templates/header.html")
		if err != nil {
			panic(err)
		}
		tpl.ExecuteTemplate(w, "homePage", user)
	}
}

func main() {
	http.HandleFunc("/", mainHandler)
	http.HandleFunc("/newPost", newPostHandler)
	http.HandleFunc("/logout", logoutHandler)
	http.HandleFunc("/register", registerHandler)
	http.HandleFunc("/registerAlreadyTaken", registerUsernameAlreadyTakenHandler)
	http.HandleFunc("/incorrectPassword", incorrectPasswordHandler)
	http.HandleFunc("/registerSuccess", registerSuccessHandler)
	http.HandleFunc("/userList", userListHandler)
	http.HandleFunc("/users/", usersHandler)
	http.Handle("/images/", http.StripPrefix("/images", http.FileServer(http.Dir("./images"))))
	http.Handle("/users/images/", http.StripPrefix("/users/images", http.FileServer(http.Dir("./images"))))

	startServer(":8080")
}
