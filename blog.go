package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
	"src/github.com/julienschmidt/httprouter"
	"src/github.com/pkg/errors"
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

type Post struct {
	Title string
	Body  string
	Date  string
}

func (u User) NoPosts() bool {
	return u.PostCount == 0
}

func startServer(addr string, handler http.Handler) {
	files, err := ioutil.ReadDir("data/accounts")
	if err != nil {
		log.Println(err, "reading directory error")
		return
	}
	for _, file := range files {
		data, err := ioutil.ReadFile("data/accounts/" + file.Name())
		if err != nil {
			log.Println(err, "reading file error")
			return
		}
		us := User{}
		err = json.Unmarshal(data, &us)
		if err != nil {
			log.Println(err, "unmarshal error")
			return
		}
		users = append(users, &us)
		ID++
	}

	fmt.Println("Server started at", addr, "\n")
	http.ListenAndServe(addr, handler)
}

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
			Expires: time.Now().Add(10 * time.Hour),
			Path:    "/",
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

func newPostGetHandler(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	_, err := r.Cookie("username")
	if err == http.ErrNoCookie {
		http.Redirect(w, r, "/", http.StatusFound)
		return
	}
	tpl, err := template.ParseFiles("templates/header.html", "templates/newPost.html")
	if err != nil {
		panic(err)
	}

	err = tpl.ExecuteTemplate(w, "newPost", nil)
	if err != nil {
		panic(err)
	}
}

func newPostPostHandler(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	usernameCookie, err := r.Cookie("username")
	if err == http.ErrNoCookie {
		http.Redirect(w, r, "/", http.StatusFound)
		return
	}
	username := usernameCookie.Value
	newPost := Post{
		Title: r.FormValue("title"),
		Body:  r.FormValue("body"),
		Date:  time.Now().Format(timeFormat),
	}
	err = getUser(username).addUserToServer()
	if err != nil {
		panic(err)
	}
	getUser(username).addPost(newPost)

	http.Redirect(w, r, "/", http.StatusFound)
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

func addUserToServer(incLogin string, incPassword string) error {
	newUser := &User{
		Username:  incLogin,
		Password:  incPassword,
		ID:        ID,
		PostCount: 0,
		Posts:     make([]Post, 0),
	}
	parsedNewUser, err := json.Marshal(newUser)
	if err != nil {
		return errors.Wrap(err, "error while adding user data to server")
	}

	ioutil.WriteFile("data/accounts/"+incLogin+".txt", parsedNewUser, 0600)
	return nil
}

func (u User) addUserToServer() error {
	parsedNewUser, err := json.Marshal(u)
	if err != nil {
		return errors.Wrap(err, "error while adding user data to server")
	}

	ioutil.WriteFile("data/accounts/"+u.Username+".txt", parsedNewUser, 0600)
	return nil
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
		http.Redirect(w, r, "/registerAlreadyTaken", http.StatusFound)
		return
	}
	err = addUserToServer(incAccount, incPassword)
	if err != nil {
		panic(err)
	}

	addUserToUsers(incAccount, incPassword)
	setID()
	registerSuccessCookie := http.Cookie{
		Name:   "registerSuccess",
		Value:  "true",
		MaxAge: 1,
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

func registerUsernameAlreadyTakenPostHandler(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	usernameCookie, err := r.Cookie("username")
	if err != http.ErrNoCookie {
		http.Redirect(w, r, "/users/"+usernameCookie.Value, http.StatusFound)
		return
	}
	incAccount := r.FormValue("account")
	incPassword := r.FormValue("password")
	if UsernameExists(incAccount) {
		http.Redirect(w, r, "/registerAlreadyTaken", http.StatusFound)
		return
	}
	err = addUserToServer(incAccount, incPassword)
	if err != nil {
		panic(err)
	}
	addUserToUsers(incAccount, incPassword)
	setID()
	registerSuccessCookie := http.Cookie{
		Name:    "registerSuccess",
		Value:   "true",
		Expires: time.Now().Add(10 * time.Second),
	}
	http.SetCookie(w, &registerSuccessCookie)
	http.Redirect(w, r, "/registerSuccess", http.StatusFound)
}

func UsernameExists(username string) bool {
	for _, us := range users {
		if username == us.Username {
			return true
		}
	}
	return false
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

func incorrectPasswordPostHandler(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	usernameCookie, err := r.Cookie("username")
	if err != http.ErrNoCookie {
		http.Redirect(w, r, "/users/"+usernameCookie.Value, http.StatusFound)
		return
	}
	if tryToLogIn(r.FormValue("username"), r.FormValue("password")) == Correct {
		usernameCookie := &http.Cookie{
			Name:    "username",
			Value:   r.FormValue("username"),
			Expires: time.Now().Add(10 * time.Hour),
			Path:    "/",
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

func usersHandler(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
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

		err = tpl.ExecuteTemplate(w, "userPage", user)
		if err != nil {
			panic(err)
		}
	} else {
		user := getUser(username)
		tpl, err := template.ParseFiles("templates/homePage.html", "templates/header.html")
		if err != nil {
			panic(err)
		}

		err = tpl.ExecuteTemplate(w, "homePage", user)
		if err != nil {
			panic(err)
		}
	}
}

func panicMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Println("panicMiddleware", r.URL.Path)
		defer func() {
			if err := recover(); err != nil {
				log.Println("recovered", err)
				http.Error(w, "Internal server error", http.StatusInternalServerError)
			}
		}()
		next.ServeHTTP(w, r)
	})
}

func accessLogMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Println("accessLogMiddleware", r.URL.Path)
		start := time.Now()
		next.ServeHTTP(w, r)
		fmt.Printf("[%s] %s, %s %s\n-\n", r.Method, r.RemoteAddr, r.URL.Path, time.Since(start))
	})
}

func main() {
	siteMux := http.NewServeMux()
	httpMux := httprouter.New()

	httpMux.GET("/", mainGetHandler)
	httpMux.POST("/", mainPostHandler)
	httpMux.GET("/logout", logoutHandler)
	httpMux.GET("/newPost", newPostGetHandler)
	httpMux.POST("/newPost", newPostPostHandler)
	httpMux.GET("/register", registerGetHandler)
	httpMux.POST("/register", registerPostHandler)
	httpMux.GET("/registerAlreadyTaken", registerUsernameAlreadyTakenGetHandler)
	httpMux.POST("/registerAlreadyTaken", registerUsernameAlreadyTakenPostHandler)
	httpMux.GET("/incorrectPassword", incorrectPasswordGetHandler)
	httpMux.POST("/incorrectPassword", incorrectPasswordPostHandler)
	httpMux.GET("/registerSuccess", registerSuccessHandler)
	httpMux.GET("/userList", userListHandler)
	httpMux.GET("/users/", usersHandler)
	httpMux.ServeFiles("/images/*filepath", http.Dir("./images"))
	httpMux.ServeFiles("/users/images/*filepath", http.Dir("./images"))

	siteMux.Handle("/", httpMux)
	siteMux.Handle("/logout", httpMux)
	siteMux.Handle("/newPost", httpMux)
	siteMux.Handle("/register", httpMux)
	siteMux.Handle("/registerAlreadyTaken", httpMux)
	siteMux.Handle("/incorrectPassword", httpMux)
	siteMux.Handle("/registerSuccess", httpMux)
	siteMux.Handle("/userList", httpMux)
	siteMux.Handle("/users/", httpMux)

	siteHandler := accessLogMiddleware(siteMux)
	siteHandler = panicMiddleware(siteHandler)

	startServer(":8080", httpMux)
}
