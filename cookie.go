package main

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"sync"
	"time"
)

type Cookie struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

type SessionManager struct {
	Cookies map[Cookie]struct{}
	lock    sync.Mutex
}

func NewSessionManager() *SessionManager {
	return &SessionManager{Cookies: make(map[Cookie]struct{})}
}

func (session *SessionManager) Get(name string) *Cookie {

	session.lock.Lock()
	defer session.lock.Unlock()

	for k := range session.Cookies {
		if k.Name == name {
			return &k
		}
	}
	return nil
}

func (session *SessionManager) Set(cookie *Cookie) {

	session.lock.Lock()
	defer session.lock.Unlock()

	session.Cookies[*cookie] = struct{}{}

	time.AfterFunc(time.Duration(cookieLifeTime*24*3600)*time.Second, func() {
		session.Del(cookie)
	})
}

func (session *SessionManager) Del(cookie *Cookie) {

	session.lock.Lock()
	defer session.lock.Unlock()

	delete(session.Cookies, *cookie)

}

func NewCookieValue() (value string, err error) {

	randomBytes := make([]byte, 32)
	_, err = rand.Read(randomBytes)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("%s/%s", "ece1779", base64.URLEncoding.EncodeToString(randomBytes)), nil

}
