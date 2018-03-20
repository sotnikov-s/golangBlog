package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
	"time"

	"src/github.com/julienschmidt/httprouter"
	"src/github.com/pkg/errors"
	"src/github.com/asaskevich/govalidator"
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
	Username  string	`valid:"alphanum, required, runelength(3|16)"`
	Password  string	`valid:"alphanum, required, runelength(3|16)"`
	ID        int		`json:"id, int"valid:"required"`
	PostCount int		`valid:"-"`
	Posts     []Post	`valid:"-"`
}

func getUser(username string) *User {
	for _, us := range users {
		if username == us.Username {
			return us
		}
	}
	return nil
}

func (u *User) addPost(post Post) error {
	_, err := govalidator.ValidateStruct(post)
	if err != nil {
		return errors.Wrap(err, fmt.Sprintf("user:%s ID:%v failed to add a new post: " +
			"it doesn't pass validation\n", u.Username, u.ID))
	}
	u.Posts = appendPost(u.Posts, post)
	u.PostCount++
	return nil
}

func (p Post) isEmpty() bool {
	if p.Title != "" && p.Body != "" && p.Date !="" {
		return false
	}
	return true
}

func appendPost(dst []Post, item Post) []Post {
	if len(dst) == 0 {
		dst = append(dst, item)
		return dst
	}
	newPosts := make([]Post, 0)
	newPosts = append(newPosts, item)
	for i := 0; i < len(dst); i++ {
		if dst[i].isEmpty() {
			dst = append(dst[:i], dst[i+1:]...)
			i--
		}
	}
	for i := 0; i < len(dst); i++ {
		newPosts = append(newPosts, dst[i])
	}
	fmt.Println("len:", len(newPosts), "cap:", cap(newPosts))
	return newPosts
}

func setID() {
	ID++
}

type Post struct {
	Title string	`valid:"required, ascii, runelength(1|30)"`
	Body  string	`valid:"required, ascii, runelength(1|300)"`
	Date  string	`valid:"-"`
}

func (u User) NoPosts() bool {
	return u.PostCount == 0
}

//func (p Post) isEmpty() bool {
//	if p.Title != "" && p.Body != "" && p.Date !="" {
//		return false
//	}
//	return true
//}
//
//func reversePosts(p []Post) []Post {
//	p2 := make([]Post, 0)
//	for i, v := range p {
//		if v.isEmpty() {
//			continue
//		}
//		p2 = append(p2, p[len(p)-1-i])
//	}
//	return p2
//}

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

		_, err = govalidator.ValidateStruct(us)
		if err != nil {
			log.Printf("can't upload data to program: user:%s ID:%v doesn't pass validation\n",
				us.Username, us.ID)
			return
		}

		users = append(users, &us)
		ID++
	}

	fmt.Print("Server started at", addr, "\n\n")
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
			Expires: time.Now().Add(10 * time.Minute),
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
		log.Println(err)
		return
	}
	err = getUser(username).refreshUserInfo()
	if err != nil {
		panic(err)
	}

	http.Redirect(w, r, "/", http.StatusFound)
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

func addUserToServer(incLogin string, incPassword string) error {
	newUser := &User{
		Username:  incLogin,
		Password:  incPassword,
		ID:        ID,
		PostCount: 0,
		Posts:     make([]Post, 0),
	}
	_, err := govalidator.ValidateStruct(newUser)
	if err != nil {
		return errors.Wrap(err, fmt.Sprintf("new user struct is invalid: %s", newUser))
	}

	parsedNewUser, err := json.Marshal(newUser)
	if err != nil {
		return errors.Wrap(err, "error while adding user data to server")
	}

	ioutil.WriteFile("data/accounts/"+incLogin+".txt", parsedNewUser, 0600)
	return nil
}

func (u User) refreshUserInfo() error {
	_, err := govalidator.ValidateStruct(u)
	if err != nil {
		return errors.Wrap(err, fmt.Sprintf("new user struct is invalid: %s", u))
	}

	parsedNewUser, err := json.Marshal(u)
	if err != nil {
		return errors.Wrap(err, "error while refreshing user's data on server")
	}

	ioutil.WriteFile("data/accounts/"+u.Username+".txt", parsedNewUser, 0600)
	return nil
}

func addUserToUsers(incLogin string, incPassword string) error {
	newUser := &User{
		Username:  incLogin,
		Password:  incPassword,
		ID:        ID,
		PostCount: 0,
		Posts:     make([]Post, 0),
	}
	_, err := govalidator.ValidateStruct(newUser)
	if err != nil {
		return errors.Wrap(err, fmt.Sprintf("new user struct is invalid: %s", newUser))
	}

	users = append(users, newUser)
	return nil
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
		Name:    "registerSuccess",
		Value:   "true",
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
			Expires: time.Now().Add(10 * time.Minute),
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

type Middleware struct {
	next http.Handler
}

func newMiddleware(next http.Handler) *Middleware {
	return &Middleware{next: next}
}

func (m *Middleware) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	fmt.Println("panicMiddleware", r.URL.Path)
	defer func() {
		if err := recover(); err != nil {
			log.Println("recovered", err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
		}
	}()
	fmt.Println("accessLogMiddleware", r.URL.Path)
	start := time.Now()
	m.next.ServeHTTP(w, r)
	fmt.Printf("[%s] %s, %s %s\n-\n", r.Method, r.RemoteAddr, r.URL.Path, time.Since(start))
}

func main() {
	httpMux := httprouter.New()

	httpMux.GET("/", mainGetHandler)
	httpMux.POST("/", mainPostHandler)
	httpMux.GET("/logout", logoutHandler)
	httpMux.GET("/users/:username/newPost", newPostGetHandler)
	httpMux.POST("/users/:username/newPost", newPostPostHandler)
	httpMux.GET("/users/:username/newPostInvalidSymbols", newPostInvalidSymbolsGetHandler)
	httpMux.POST("/users/:username/newPostInvalidSymbols", newPostPostHandler)
	httpMux.GET("/register", registerGetHandler)
	httpMux.POST("/register", registerPostHandler)
	httpMux.GET("/registerUsernameAlreadyTaken", registerUsernameAlreadyTakenGetHandler)
	httpMux.POST("/registerUsernameAlreadyTaken", registerPostHandler)
	httpMux.GET("/registerInvalidSymbols", registerInvalidSymbolsGetHandler)
	httpMux.POST("/registerInvalidSymbols", registerPostHandler)
	httpMux.GET("/incorrectPassword", incorrectPasswordGetHandler)
	httpMux.POST("/incorrectPassword", incorrectPasswordPostHandler)
	httpMux.GET("/registerSuccess", registerSuccessHandler)
	httpMux.GET("/userList", userListHandler)
	httpMux.GET("/users/:username/", usersHandler)
	httpMux.ServeFiles("/images/*filepath", http.Dir("./images"))
	httpMux.ServeFiles("/users/:username/images/*filepath", http.Dir("./images"))

	m := newMiddleware(httpMux)

	startServer(":8080", m)
}
