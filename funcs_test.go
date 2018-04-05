package main

import (
	"testing"
	"time"
)

func TestGetUser (t *testing.T) {
	defer func() {
		users = nil
	}()
	users = append(users, &User{Username:"testUser", Password:"testPassword", ID:1})
	users = append(users, &User{Username:"testUser1", Password:"testPassword1", ID:2})
	users = append(users, &User{Username:"testUser2", Password:"testPassword2", ID:3})
	users = append(users, &User{Username:"testUser3", Password:"testPassword3", ID:4})
	desUser := &User{Username:"testUser3", Password:"testPassword3", ID:4}
	checkUser := getUser("testUser3")
	if desUser.Username != checkUser.Username || desUser.Password != checkUser.Password || desUser.ID != checkUser.ID ||
		desUser.PostCount != checkUser.PostCount {
		t.Errorf("TestGetUser --> FAILED")
	}
	for i:=0; i < desUser.PostCount; i++ {
		if desUser.Posts[i] != checkUser.Posts[i] {
			t.Errorf("TestGetUser --> FAILED")
		}
	}
}

func TestGetNonexistentUser (t *testing.T) {
	defer func() {
		users = nil
	}()
	users = append(users, &User{Username:"testUser", Password:"testPassword", ID:1})
	users = append(users, &User{Username:"testUser1", Password:"testPassword1", ID:2})
	users = append(users, &User{Username:"testUser2", Password:"testPassword2", ID:3})
	users = append(users, &User{Username:"testUser3", Password:"testPassword3", ID:4})

	checkUser := getUser("testUser4")

	if checkUser != nil {
		t.Errorf("TestGetNonexistentUser --> FAILED")
	}
}

func TestIsEmpty (t *testing.T) {
	nonEmpty := Post{
		Title: "title",
		Body: "body",
		Date: time.Now().Format(timeFormat),
	}
	empty := Post{
		Title: "",
		Body: "",
		Date: "",
	}

	if !empty.isEmpty() {
		t.Errorf("TestIsEmpty --> FAILED")
	}
	if nonEmpty.isEmpty() {
		t.Errorf("TestIsEmpty --> FAILED")
	}
}

func TestTryToLogin (t *testing.T) {
	defer func() {
		users = nil
	}()
	users = append(users, &User{Username:"testUser0", Password:"testPass0"})
	users = append(users, &User{Username:"testUser1", Password:"testPass1"})
	users = append(users, &User{Username:"testUser3", Password:"testPass3"})
	un0 := "testUser0"
	ps0 := "testPass0"		// ok
	un1 := "testUser1"
	ps1 := "testPass1_err"	// Wrong password
	un2 := "testUser2"
	ps2 := "testPass2_err"	// No such user
	un3 := "testUser3"
	ps3 := "testPass3"		// ok

	if tryToLogIn(un0, ps0) != Correct {
		t.Errorf("TestTryToLogin --> FAILED")
	}
	if tryToLogIn(un1, ps1) != WrongPassword {
		t.Errorf("TestTryToLogin --> FAILED")
	}
	if tryToLogIn(un2, ps2) != NoMatch {
		t.Errorf("TestTryToLogin --> FAILED")
	}
	if tryToLogIn(un3, ps3) != Correct {
		t.Errorf("TestTryToLogin --> FAILED")
	}
}

func TestAddUserToUsers (t *testing.T) {
	defer func() {
		users = nil
	}()
	desUser := &User{Username:"testUser", Password:"testPassword", ID: 1}
	addUserToUsers("testUser", "testPassword")
	checkUser := users[0]

	if desUser.Username != checkUser.Username || desUser.Password != checkUser.Password || desUser.ID != checkUser.ID ||
		desUser.PostCount != checkUser.PostCount {
		t.Errorf("TestGetUser --> FAILED")
	}
	for i:=0; i < desUser.PostCount; i++ {
		if desUser.Posts[i] != checkUser.Posts[i] {
			t.Errorf("TestGetUser --> FAILED")
		}
	}
}