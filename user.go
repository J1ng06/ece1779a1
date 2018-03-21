package main

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
)

type User struct {
	ID       int64  `gorm:"primary_key"`
	Username string `json:"username"`
	Password string `json:"password"`
	Salt     string
}

func (u User) Authentication(pwd string) bool {

	sum := sha256.Sum256(append([]byte(pwd), []byte(u.Salt)...))
	return u.Password == base64.URLEncoding.EncodeToString(sum[:])

}

func (u *User) RandomSalt() (err error) {

	randomBytes := make([]byte, 32)
	_, err = rand.Read(randomBytes)
	if err != nil {
		return
	}

	u.Salt = base64.URLEncoding.EncodeToString(randomBytes)

	return
}

func (u *User) EncPass(pwd string) {
	sum := sha256.Sum256(append([]byte(pwd), []byte(u.Salt)...))
	u.Password = base64.URLEncoding.EncodeToString(sum[:])
}

type UserImages struct {
	ID        int64 `gorm:"primary_key"`
	userId    int64
	Original  string `json:"original"`
	Thumbnail string `json:"thumbnail"`
	T1        string `json:"t1"`
	T2        string `json:"t2"`
	T3        string `json:"t3"`
}
