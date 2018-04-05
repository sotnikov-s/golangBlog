package main

import (
	"fmt"
	"encoding/json"
	"io/ioutil"

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
	Username  string `valid:"alphanum, required, runelength(3|16)"`
	Password  string `valid:"alphanum, required, runelength(3|16)"`
	ID        int    `json:"id, int"valid:"required"`
	PostCount int    `valid:"-"`
	Posts     []Post `valid:"-"`
}

func getUser(username string) *User {
	for _, us := range users {
		if username == us.Username {
			return us
		}
	}
	return nil
}

type Post struct {
	Title string `valid:"required, ascii, runelength(1|30)"`
	Body  string `valid:"required, ascii, runelength(1|300)"`
	Date  string `valid:"-"`
}

func (u User) NoPosts() bool {
	return u.PostCount == 0
}

func (u *User) addPost(post Post) error {
	_, err := govalidator.ValidateStruct(post)
	if err != nil {
		return errors.Wrap(err, fmt.Sprintf("user:%s ID:%v failed to add a new post: "+
			"it doesn't pass validation\n", u.Username, u.ID))
	}
	u.Posts = appendPost(u.Posts, post)
	u.PostCount++
	return nil
}

func (p Post) isEmpty() bool {
	if p.Title != "" && p.Body != "" && p.Date != "" {
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
		newPosts = append(newPosts, dst[i])
	}
	return newPosts
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
		return errors.Wrap(err, fmt.Sprintf("new user struct is invalid: %v", newUser))
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
		return errors.Wrap(err, fmt.Sprintf("new user struct is invalid: %v", u))
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
		return errors.Wrap(err, fmt.Sprintf("new user struct is invalid: %v", newUser))
	}

	users = append(users, newUser)
	return nil
}

func UsernameExists(username string) bool {
	for _, us := range users {
		if username == us.Username {
			return true
		}
	}
	return false
}

func setID() {
	ID++
}
