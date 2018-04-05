package main

import (
	"net/http"
	"fmt"
	"log"
	"time"
	"encoding/json"
	"io/ioutil"

	"src/github.com/julienschmidt/httprouter"
	"src/github.com/asaskevich/govalidator"
)

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

	fmt.Print("Server started at ", addr, "\n\n")
	http.ListenAndServe(addr, handler)
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
	httpMux.POST("/incorrectPassword", mainPostHandler)
	httpMux.GET("/registerSuccess", registerSuccessHandler)
	httpMux.GET("/userList", userListHandler)
	httpMux.GET("/users/:username/", usersHandler)
	httpMux.ServeFiles("/images/*filepath", http.Dir("./images"))
	httpMux.ServeFiles("/users/:username/images/*filepath", http.Dir("./images"))

	m := newMiddleware(httpMux)

	startServer(":8080", m)
}
