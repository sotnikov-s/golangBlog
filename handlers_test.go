package main

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"src/github.com/julienschmidt/httprouter"
	"time"
	"strconv"
)

func TestUserDataValidation (t *testing.T) {
	if alright := addUserToServer("account", "password"); alright != nil {
		t.Errorf("TestUserDataValidation --> FAILED")
	}
	if shortPair := addUserToServer("sh", "sh"); shortPair == nil {
		t.Errorf("TestUserDataValidation --> FAILED")
	}
	if longPair := addUserToServer("moreThan16Symbols", "moreThan16Symbols"); longPair == nil {
		t.Errorf("TestUserDataValidation --> FAILED")
	}
	if notAlphabeticPair := addUserToServer("{|}#|@%@", "{|}#|@%@"); notAlphabeticPair == nil {
		t.Errorf("TestUserDataValidation --> FAILED")
	}
	if rusPair := addUserToServer("аккаунт", "пароль"); rusPair == nil {
		t.Errorf("TestUserDataValidation --> FAILED")
	}
	if nilPair := addUserToServer("", ""); nilPair == nil {
		t.Errorf("TestUserDataValidation --> FAILED")
	}
}

func TestPostValidation (t *testing.T) {
	testUser := User{}
	if alright := testUser.addPost(Post{"title", "body", "now"}); alright != nil {
		t.Errorf("TestPostValidation --> FAILED")
	}
	if nilPost := testUser.addPost(Post{"", "", ""}); nilPost == nil {
		t.Errorf("TestPostValidation --> FAILED")
	}
	if longPost := testUser.addPost(Post{"tooMuchLettersTooMuchLettersToo", "tooMuchLettersTooMuchLetters" +
		"TooMuchLettersTooMuchLettersTooMuchLettersTooMuchLettersTooMuchLettersTooMuchLettersTooMuchLetters" +
		"TooMuchLettersTooMuchLettersTooMuchLettersTooMuchLettersTooMuchLettersTooMuchLettersTooMuchLetters" +
		"TooMuchLettersTooMuchLettersTooMuchLettersTooMuchLettersTooMuchLettersTooMuchLettersTooMuch" +
		"Letters", "now"}); longPost == nil {
		t.Errorf("TestPostValidation --> FAILED")
	}
	if nonASCII := testUser.addPost(Post{"¡¡¡¡", "¡¡¡¡", "now"}); nonASCII == nil {
		t.Errorf("TestPostValidation --> FAILED")
	}
}

func TestMainGetHandlerWithCookie(t *testing.T) {
	url := "http://127.0.0.1/"
	req := httptest.NewRequest("GET", url, nil)
	req.AddCookie(&http.Cookie{Name: "username", Value: "testUser"})
	w := httptest.NewRecorder()

	mainGetHandler(w, req, nil)

	result := w.Result()

	l, _ := result.Location()
	if result.StatusCode != 302 || l.Path != "/users/testUser" {
		t.Errorf("TestMainGetHandlerWithCookie --> FAILED")
	}
}

func TestMainGetHandlerWithoutCookie(t *testing.T) {
	url := "http://127.0.0.1/"
	req := httptest.NewRequest("GET", url, nil)
	w := httptest.NewRecorder()

	mainGetHandler(w, req, nil)

	result := w.Result()

	if result.StatusCode != 200 {
		t.Errorf("TestMainGetHandlerWithoutCookie --> FAILED")
	}
}

func TestMainPostHandlerWithCookie(t *testing.T) {
	url := "http://127.0.0.1/"
	req := httptest.NewRequest("POST", url, nil)
	req.AddCookie(&http.Cookie{Name: "username", Value: "testUser"})
	w := httptest.NewRecorder()

	mainPostHandler(w, req, nil)

	result := w.Result()

	l, _ := result.Location()
	if result.StatusCode != 302 || l.Path != "/users/testUser" {
		t.Errorf("TestMainPostHandlerWithCookie --> FAILED")
	}
}

func TestMainPostHandlerNoMatch(t *testing.T) {
	defer func() {
		users = nil
	}()
	url := "http://127.0.0.1/"
	req := httptest.NewRequest("POST", url, nil)
	users = append(users, &User{Username: "testUser", Password: "testPassword"})
	w := httptest.NewRecorder()

	req.ParseForm()
	req.PostForm.Set("username", "testUser")
	req.PostForm.Set("password", "wrongPassword")

	mainPostHandler(w, req, nil)

	result := w.Result()

	l, _ := result.Location()
	if result.StatusCode != 302 || l.Path != "/incorrectPassword" || len(result.Cookies()) != 0 {
		t.Errorf("TestMainPostHandlerNoMatch --> FAILED")
	}
}

func TestMainPostHandlerIncorrectPassword (t *testing.T) {
	defer func() {
		users = nil
	}()
	url := "http://127.0.0.1/"
	req := httptest.NewRequest("POST", url, nil)
	users = append(users, &User{Username:"testUser", Password: "testPassword"})
	w := httptest.NewRecorder()

	req.ParseForm()
	req.Form.Set("username", "testUser")
	req.Form.Set("password", "wrongPassword")

	mainPostHandler(w, req, nil)

	result := w.Result()

	l, _ := result.Location()
	if result.StatusCode != 302 || l.Path != "/incorrectPassword" || len(result.Cookies()) != 0 {
		t.Errorf("TestMainPostHandlerIncorrectPassword --> FAILED")
	}

}

func TestMainPostHandlerCorrectData(t *testing.T) {
	defer func() {
		users = nil
	}()
	url := "http://127.0.0.1/"
	req := httptest.NewRequest("POST", url, nil)
	users = append(users, &User{Username: "testUser", Password: "testPassword"})
	w := httptest.NewRecorder()

	req.ParseForm()
	req.Form.Set("username", "testUser")
	req.Form.Set("password", "testPassword")

	mainPostHandler(w, req, nil)

	result := w.Result()

	l, _ := result.Location()
	if result.StatusCode != 302 || l.Path != "/users/testUser" || result.Cookies()[0].Value != "testUser" {
		t.Errorf("TestMainPostHandlerCorrectData --> FAILED")
	}
}

func TestLogoutHandlerWithCookie(t *testing.T) {
	url := "http://127.0.0.1/logout"
	req := httptest.NewRequest("GET", url, nil)
	req.AddCookie(&http.Cookie{Name: "username", Value: "testUser"})
	w := httptest.NewRecorder()

	logoutHandler(w, req, nil)

	result := w.Result()

	if result.StatusCode != 302 || result.Cookies()[0].Value != "" {
		t.Errorf("TestLogoutHandlerWithCookie --> FAILED")
	}
}

func TestLogoutHandlerWithoutCookie(t *testing.T) {
	url := "http://127.0.0.1/logout"
	req := httptest.NewRequest("GET", url, nil)
	w := httptest.NewRecorder()

	logoutHandler(w, req, nil)

	result := w.Result()

	if result.StatusCode != 302 {
		t.Errorf("TestLogoutHandlerWithoutCookie --> FAILED")
	}
}

func TestNewPostGetHandlerWithoutCookie(t *testing.T) {
	url := "http://127.0.0.1/testUset/newPost"
	req := httptest.NewRequest("GET", url, nil)
	w := httptest.NewRecorder()

	newPostGetHandler(w, req, nil)

	result := w.Result()

	l, _ := result.Location()
	if result.StatusCode != 302 || l.Path != "/" {
		t.Errorf("TestNewPostGetHandlerWithoutCookie --> FAILED")
	}
}

func TestNewPostGetHandlerWithCookie(t *testing.T) {
	defer func() {
		users = nil
	}()
	url := "http://127.0.0.1/testUset/newPost"
	req := httptest.NewRequest("GET", url, nil)
	req.AddCookie(&http.Cookie{Name:"username", Value:"testUser"})
	users = append(users, &User{Username:"testUser", Password:"testPassword"})

	w := httptest.NewRecorder()

	newPostGetHandler(w, req, httprouter.Params{{"username", "testUser"}})

	result := w.Result()

	if result.StatusCode != 200{
		t.Errorf("TestNewPostGetHandlerWithCookie --> FAILED")
	}
}

func TestNewPostPostHandlerWithoutCookie(t *testing.T) {
	url := "http://127.0.0.1/users/testUser/newPost"
	req := httptest.NewRequest("POST", url, nil)
	w := httptest.NewRecorder()

	newPostPostHandler(w, req, nil)

	result := w.Result()

	l, _ := result.Location()
	if result.StatusCode != 302 || l.Path != "/" {
		t.Errorf("TestNewPostPostHandlerWithoutCookie --> FAILED")
	}
}

func TestNewPostPostHandlerInvalidSymbols (t *testing.T) {
	url := "http://127.0.0.1/users/testUser/newPost"
	req := httptest.NewRequest("POST", url, nil)
	req.AddCookie(&http.Cookie{Name:"username", Value:"testUser"})
	req.ParseForm()
	req.Form.Set("title", "¡¡¡¡")
	req.Form.Set("body", "¡¡¡¡")
	users = append(users, &User{Username:"testUser", Password:"testPassword"})
	w := httptest.NewRecorder()

	newPostPostHandler(w, req, httprouter.Params{{"username", "testUser"}})

	result := w.Result()
	l, _ := result.Location()
	if result.StatusCode != 302 || l.Path != "/users/testUser/newPostInvalidSymbols" {
		t.Errorf("TestNewPostPostHandlerInvalidSymbols --> FAILED")
	}
}

func TestNewPostPostHandlerSuccess(t *testing.T) {
	defer func() {
		users = nil
	}()
	url := "http://127.0.0.1/testUset/newPost"
	req := httptest.NewRequest("POST", url, nil)
	req.AddCookie(&http.Cookie{Name:"username", Value:"testUser"})
	req.ParseForm()
	users = append(users, &User{Username:"testUser", Password:"testPassword"})
	us := getUser("testUser")
	us.ID = 1

	w := httptest.NewRecorder()

	for i:= 0; i < 10; i++ {
		pc := us.PostCount
		newPost := Post{"title"+strconv.Itoa(i), "body"+strconv.Itoa(i), time.Now().Format(timeFormat)}
		req.Form.Set("title", "title"+strconv.Itoa(i))
		req.Form.Set("body", "body"+strconv.Itoa(i))
		newPostPostHandler(w, req, httprouter.Params{{"username", "testUser"}})
		result := w.Result()
		l, _ := result.Location()
		if result.StatusCode != 302 || l.Path != "/users/testUser" || us.Posts[0] != newPost || us.PostCount != pc+1 {
			t.Errorf("TestNewPostPostHandlerSuccess --> FAILED")
			return
		}
	}
}

func TestNewPostInvalidSymbolsGetHandlerWithoutCookie (t *testing.T) {
	url := "http://127.0.0.1/testUset/newPostInvalidSymbols"
	req := httptest.NewRequest("GET", url, nil)
	w := httptest.NewRecorder()

	newPostInvalidSymbolsGetHandler(w, req, nil)

	result := w.Result()

	l, _ := result.Location()
	if result.StatusCode != 302 || l.Path != "/" {
		t.Errorf("TestNewPostInvalidSymbolsGetHandlerWithoutCookie --> FAILED")
	}
}

func TestNewPostInvalidSymbolsGetHandlerWithCookie(t *testing.T) {
	defer func() {
		users = nil
	}()
	url := "http://127.0.0.1/testUset/newPostInvalidSymbols"
	req := httptest.NewRequest("GET", url, nil)
	req.AddCookie(&http.Cookie{Name:"username", Value:"testUser"})
	req.ParseForm()
	users = append(users, &User{Username:"testUser", Password:"testPassword"})
	us := getUser("testUser")
	us.ID = 1

	w := httptest.NewRecorder()

	newPostInvalidSymbolsGetHandler(w, req, httprouter.Params{{"username", "testUser"}})

	result := w.Result()

	if result.StatusCode != 200 {
		t.Errorf("TestNewPostInvalidSymbolsGetHandlerWithCookie --> FAILED")
	}
}

func TestRegisterGetHandlerWithoutCookie(t *testing.T) {
	url := "http://127.0.0.1/register"
	req := httptest.NewRequest("GET", url, nil)
	w := httptest.NewRecorder()

	registerGetHandler(w, req, nil)

	result := w.Result()

	if result.StatusCode != 200 {
		t.Errorf("TestRegisterGetHandlerWithoutCookie --> FAILED")
	}
}

func TestRegisterGetHandlerWithCookie (t *testing.T) {
	url := "http://127.0.0.1/register"
	req := httptest.NewRequest("GET", url, nil)
	req.AddCookie(&http.Cookie{Name:"username", Value:"testUser"})
	w := httptest.NewRecorder()

	registerGetHandler(w, req, nil)

	result := w.Result()

	l, _ := result.Location()
	if result.StatusCode != 302 || l.Path != "/users/testUser"{
		t.Errorf("TestRegisterGetHandlerWithCookie --> FAILED")
	}
}

func TestRegisterPostHandlerWithCookie (t *testing.T) {
	url := "http://127.0.0.1/register"
	req := httptest.NewRequest("POST", url, nil)
	req.AddCookie(&http.Cookie{Name:"username", Value:"testUser"})
	w := httptest.NewRecorder()

	registerPostHandler(w, req, nil)

	result := w.Result()

	l, _ := result.Location()
	if result.StatusCode != 302 || l.Path != "/users/testUser" {
		t.Errorf("TestRegisterPostHandlerWithCookie --> FAILED")
	}
}

func TestRegisterPostHandlerSuccess (t *testing.T) {
	defer func() {
		users = nil
	}()
	url := "http://127.0.0.1/register"
	req := httptest.NewRequest("POST", url, nil)
	req.ParseForm()
	req.Form.Set("account", "testUser")
	req.Form.Set("password", "testPassword")

	w := httptest.NewRecorder()

	registerPostHandler(w, req, nil)

	result := w.Result()

	desUser := &User{Username:"testUser", Password:"testPassword", ID:1}
	if desUser.Username != users[0].Username || desUser.Password != users[0].Password || desUser.ID != users[0].ID ||
		desUser.PostCount != users[0].PostCount {
		t.Errorf("TestRegisterPostHandlerSuccess --> FAILED")
	}
	for i:=0; i < desUser.PostCount; i++ {
		if desUser.Posts[i] != users[0].Posts[i] {
			t.Errorf("TestRegisterPostHandlerSuccess --> FAILED")
		}
	}

	l, _ := result.Location()
	if result.StatusCode != 302 || l.Path != "/registerSuccess" || result.Cookies()[0].Value != "true" {
		t.Errorf("TestRegisterPostHandlerSuccess --> FAILED")
	}
}

func TestRegisterSuccessHandlerWithoutSuccessCookie (t *testing.T) {
	url := "http://127.0.0.1/registerSuccess"
	req := httptest.NewRequest("GET", url, nil)
	w := httptest.NewRecorder()

	registerSuccessHandler(w, req, nil)

	result := w.Result()

	l, _ := result.Location()
	if result.StatusCode != 302 || l.Path != "/" {
		t.Errorf("TestRegisterSuccessHandlerWithoutSuccessCookie --> FAILED")
	}
}

func TestRegisterSuccessHandlerWithUsernameCookie (t *testing.T) {
	url := "http://127.0.0.1/registerSuccess"
	req := httptest.NewRequest("GET", url, nil)
	req.AddCookie(&http.Cookie{Name: "username", Value: "testUser"})
	w := httptest.NewRecorder()

	registerSuccessHandler(w, req, nil)

	result := w.Result()

	l, _ := result.Location()
	if result.StatusCode != 302 || l.Path != "/users/testUser" {
		t.Errorf("TestRegisterSuccessHandlerWithUsernameCookie --> FAILED")
	}
}

func TestRegisterSuccessHandler (t *testing.T) {
	url := "http://127.0.0.1/registerSuccess"
	req := httptest.NewRequest("GET", url, nil)
	req.AddCookie(&http.Cookie{Name:"registerSuccess", Value:"true"})
	w := httptest.NewRecorder()

	registerSuccessHandler(w, req, nil)

	result := w.Result()

	if result.StatusCode != 200 || len(result.Cookies()) != 0 {
		t.Errorf("TestRegisterSuccessHandler --> FAILED")
	}
}

func TestRegisterGetHandlerUsernameAlreadyTakenWithCookie (t *testing.T) {
	url := "http://127.0.0.1/registerUsernameAlreadyTaken"
	req := httptest.NewRequest("GET", url, nil)
	req.AddCookie(&http.Cookie{Name:"username", Value:"testUser"})
	w := httptest.NewRecorder()

	registerUsernameAlreadyTakenGetHandler(w, req, nil)

	result := w.Result()

	l, _ := result.Location()
	if result.StatusCode != 302 || l.Path != "/users/testUser" {
		t.Errorf("TestRegisterGetHandlerUsernameAlreadyTakenWithCookie --> FAILED")
	}
}

func TestRegisterUsernameAlreadyTakenGetHandler (t *testing.T) {
	url := "http://127.0.0.1/registerUsernameAlreadyTaken"
	req := httptest.NewRequest("GET", url, nil)
	w := httptest.NewRecorder()

	registerUsernameAlreadyTakenGetHandler(w, req, nil)

	result := w.Result()

	if result.StatusCode != 200 {
		t.Errorf("TestRegisterUsernameAlreadyTakenGetHandler --> FAILED")
	}
}

func TestRegisterInvalidSymbolsGetHandler (t *testing.T) {
	url := "http://127.0.0.1/registerInvalidSymbols"
	req := httptest.NewRequest("GET", url, nil)
	w := httptest.NewRecorder()

	registerInvalidSymbolsGetHandler(w, req, nil)

	result := w.Result()

	if result.StatusCode != 200 {
		t.Errorf("TestRegisterInvalidSymbolsGetHandler --> FAILED")
	}
}

func TestRegisterPostHandlerUsernameAlreadyTaken (t *testing.T) {
	defer func() {
		users = nil
	}()
	url := "http://127.0.0.1/register"
	req := httptest.NewRequest("POST", url, nil)
	w := httptest.NewRecorder()
	users = append(users, &User{Username:"testUser", Password:"testUser", ID: 1})
	req.ParseForm()
	req.Form.Set("account", "testUser")
	req.Form.Set("password", "testUser")

	registerPostHandler(w, req, nil)

	result := w.Result()

	l, _ := result.Location()
	if result.StatusCode != 302 || l.Path != "/registerUsernameAlreadyTaken" {
		t.Errorf("TestRegisterPostHandlerUsernameAlreadyTaken --> FAILED")
	}
}

func TestRegisterGetHandlerInvalidSymbolsWithCookie (t *testing.T) {
	url := "http://127.0.0.1/registerInvalidSymbols"
	req := httptest.NewRequest("GET", url, nil)
	req.AddCookie(&http.Cookie{Name:"username", Value:"testUser"})
	w := httptest.NewRecorder()

	registerInvalidSymbolsGetHandler(w, req, nil)

	result := w.Result()

	l, _ := result.Location()
	if result.StatusCode != 302 || l.Path != "/users/testUser" {
		t.Errorf("TestRegisterGetHandlerInvalidSymbolsWithCookie --> FAILED")
	}
}

func TestRegisterPostHandlerInvalidSymbols (t *testing.T) {
	url := "http://127.0.0.1/register"
	req := httptest.NewRequest("POST", url, nil)
	w := httptest.NewRecorder()
	testCases := [][]string {{"a", "a"}, {"moreThan16Symbols", "moreThan16Symbols"}, {"@#%@@#%@", "@#%@@#%@"},
	{"аккаунт", "пароль"}}

	req.ParseForm()
	for _, v := range testCases {
		req.Form.Set("account", v[0])
		req.Form.Set("password", v[1])
		registerPostHandler(w, req, nil)
		result := w.Result()
		l, _ := result.Location()
		if result.StatusCode != 302 || l.Path != "/registerInvalidSymbols" {
			t.Errorf("TestRegisterPostHandlerInvalidSymbols --> FAILED")
		}
	}
}

func TestIncorrectPasswordGetHandlerWithCookie(t *testing.T) {
	url := "http://127.0.0.1/"
	req := httptest.NewRequest("GET", url, nil)
	req.AddCookie(&http.Cookie{Name:"username", Value:"testUser"})
	w := httptest.NewRecorder()

	incorrectPasswordGetHandler(w, req, nil)

	result := w.Result()

	l, _ := result.Location()

	if result.StatusCode != 302 || l.Path != "/users/testUser" {
		t.Errorf("TestIncorrectPasswordGetHandlerWithCookie --> FAILED")
	}
}

func TestIncorrectPasswordGetHandler(t *testing.T) {
	url := "http://127.0.0.1/"
	req := httptest.NewRequest("GET", url, nil)
	w := httptest.NewRecorder()

	incorrectPasswordGetHandler(w, req, nil)

	result := w.Result()

	if result.StatusCode != 200 {
		t.Errorf("TestIncorrectPasswordGetHandler --> FAILED")
	}
}

func TestUserListHandlerWithoutCookie (t *testing.T) {
	url := "http://127.0.0.1/userList"
	req := httptest.NewRequest("GET", url, nil)
	w := httptest.NewRecorder()

	userListHandler(w, req, nil)

	result := w.Result()

	l, _ := result.Location()

	if result.StatusCode != 302 || l.Path != "/" {
		t.Errorf("TestUserListHandlerWithoutCookie --> FAILED")
	}
}

func TestUserListHandler (t *testing.T) {
	url := "http://127.0.0.1/userList"
	req := httptest.NewRequest("GET", url, nil)
	req.AddCookie(&http.Cookie{Name:"username", Value:"testUser"})
	w := httptest.NewRecorder()

	userListHandler(w, req, nil)

	result := w.Result()

	if result.StatusCode != 200 {
		t.Errorf("TestUserListHandlerWithoutCookie --> FAILED")
	}
}

func TestUsersHandlerWithoutCookie (t *testing.T) {
	url := "http://127.0.0.1/users/testuser"
	req := httptest.NewRequest("GET", url, nil)
	w := httptest.NewRecorder()

	usersHandler(w, req, httprouter.Params{{"username", "testUser"}})

	result := w.Result()

	l, _ := result.Location()

	if result.StatusCode != 302 || l.Path != "/" {
		t.Errorf("TestUsersHandlerWithoutCookie --> FAILED")
	}
}

func TestUsersHandlerHomepage (t *testing.T) {
	defer func() {
		users = nil
	}()
	url := "http://127.0.0.1/users/testUser"
	req := httptest.NewRequest("GET", url, nil)
	req.AddCookie(&http.Cookie{Name:"username", Value:"testUser"})
	users = append(users, &User{Username:"testUser", Password:"testPassword", ID:1})
	w := httptest.NewRecorder()

	usersHandler(w, req, httprouter.Params{{"username", "testUser"}})

	result := w.Result()

	if result.StatusCode != 200 {
		t.Errorf("TestUsersHandlerHomepage --> FAILED")
	}
}

func TestUsersHandlerUserPage (t *testing.T) {
	defer func() {
		users = nil
	}()
	url := "http://127.0.0.1/users/AnotherTestUser"
	req := httptest.NewRequest("GET", url, nil)
	req.AddCookie(&http.Cookie{Name:"username", Value:"testUser"})
	users = append(users, &User{Username:"AnotherTestUser", Password:"testPassword", ID:1})
	w := httptest.NewRecorder()

	usersHandler(w, req, httprouter.Params{{"username", "AnotherTestUser"}})

	result := w.Result()

	if result.StatusCode != 200 {
		t.Errorf("TestUsersHandlerHomepage --> FAILED")
	}
}

