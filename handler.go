package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
	"time"
)

func HandleUser(w http.ResponseWriter, req *http.Request) {

	//Timer
	now := time.Now()
	//Output
	status, err := http.StatusOK, error(nil)

	//req.Header.Get("Cookie")

	//Log Request
	log.Printf("Request [Method: %s] [Path: %s]", req.Method, req.URL.Path)

	// check http method
	if req.Method != "POST" {
		err, status = errors.New(http.StatusText(http.StatusMethodNotAllowed)), http.StatusMethodNotAllowed
		return
	}

	//Log Response
	defer func() {
		log.Printf("Response [Status: %d] [Path: %s] [Roundtrip: %d ms] [Error: %v]", status, req.URL.Path, (time.Now().Sub(now))/time.Millisecond, err)
	}()

	var action string
	var temp string
	fmt.Sscanf(strings.Replace(req.URL.Path, "/", " ", -1), "%s %s", &temp, &action)

	// get input
	b, err := ioutil.ReadAll(req.Body)
	defer req.Body.Close()
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	var userReq User
	err = json.Unmarshal(b, &userReq)
	if err != nil {
		err, status = errors.New(http.StatusText(http.StatusInternalServerError)), http.StatusInternalServerError
		return
	}
	//fmt.Printf("[DEBUG] %+v\n", userReq)
	username := userReq.Username
	password := userReq.Password

	if username == "" || password == "" {
		err, status = errors.New(http.StatusText(http.StatusBadRequest)), http.StatusBadRequest
		return
	}

	db, _ := NewConnection()
	if err != nil {
		err, status = errors.New(http.StatusText(http.StatusInternalServerError)), http.StatusInternalServerError
		return
	}
	defer db.Close()

	user := &User{Username: username}
	switch action {
	case "login":

		db.Find(user)
		if user.Authentication(password) {
			existing := sessions.Get(username)
			if existing == nil {

				value, err := NewCookieValue()
				if err != nil {
					err, status = errors.New(http.StatusText(http.StatusInternalServerError)), http.StatusInternalServerError
					return
				}

				newCookie := &CookieSlim{
					Name:  username,
					Value: value,
				}

				sessions.Set(newCookie)

				//fmt.Printf("%v", newCookie)
				data, err := json.Marshal(newCookie)
				if err != nil {
					err, status = errors.New(http.StatusText(http.StatusInternalServerError)), http.StatusInternalServerError
					return
				}

				w.Header().Set("Content-Type", "application/json")
				w.Write(data)

			} else {

				data, err := json.Marshal(existing)
				if err != nil {
					err, status = errors.New(http.StatusText(http.StatusInternalServerError)), http.StatusInternalServerError
					return
				}

				w.Header().Set("Content-Type", "application/json")
				w.Write(data)

			}
		}

	case "register":

		err = user.RandomSalt()
		if err != nil {
			err, status = errors.New(http.StatusText(http.StatusInternalServerError)), http.StatusInternalServerError
			return
		}

		user.EncPass(password)
		db.Create(&user)

	default:
		err, status = errors.New(http.StatusText(http.StatusBadRequest)), http.StatusBadRequest
		return
	}

}

func HandleImage(w http.ResponseWriter, req *http.Request) {

	now := time.Now()

	status, err := http.StatusOK, error(nil)

	defer func() {
		log.Printf("Response [Status: %d] [Path: %s] [Roundtrip: %d ms] [Error: %v]", status, req.URL.Path, (time.Now().Sub(now))/time.Millisecond, err)
	}()

	if req.Method != "POST" && req.Method != "GET" {
		err, status = errors.New(http.StatusText(http.StatusMethodNotAllowed)), http.StatusMethodNotAllowed
		return
	}

	var action string
	var temp string
	fmt.Sscanf(strings.Replace(req.URL.Path, "/", " ", -1), "%s %s", &temp, &action)

	switch action {
	case "upload":

	default:
		err, status = errors.New(http.StatusText(http.StatusBadRequest)), http.StatusBadRequest
		return
	}

}

func ValidateCookie(handler http.Handler) (out http.Handler) {

	now := time.Now()
	err := error(nil)
	out = http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		//Log Response
		defer func() {

			log.Printf("Response [Path: %s] [Roundtrip: %d ms] [Error: %v]", req.URL.Path, (time.Now().Sub(now))/time.Millisecond, err)

		}()

		if req.URL.Path != "/" && strings.HasSuffix(req.URL.Path, ".html") {

			username := req.URL.Query().Get("username")

			if username == "" {
				return
			}

			cookie, err := req.Cookie(username)
			if err != nil {
				return
			}
			fmt.Println("[DEBUG] ", cookie.Name, "Value", cookie.Value)
			if c := sessions.Get(cookie.Name); c == nil {
				return
			}

		}
		handler.ServeHTTP(w, req)
	})

	return
}
