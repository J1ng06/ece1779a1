package main

import (
	"net/http"

	"gopkg.in/gographics/imagick.v2/imagick"
)

var (
	sessions       = NewSessionManager()
	cookieLifeTime = 7
)

func init() {
	imagick.Initialize()
	defer imagick.Terminate()
}

func main() {

	http.HandleFunc("/user/", HandleUser)
	http.Handle("/image/", ValidateCookie(http.HandlerFunc(HandleImage)))
	http.Handle("/", ValidateCookie(http.FileServer(http.Dir("static"))))
	http.ListenAndServe(":8080", nil)

}
