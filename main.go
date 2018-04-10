package main

import "net/http"

var (
	sessions       = NewSessionManager()
	cookieLifeTime = 7
)

func main() {

	http.HandleFunc("/user/", HandleUser)
	http.Handle("/image/", ValidateCookie(http.HandlerFunc(HandleImage)))
	http.Handle("/", ValidateCookie(http.FileServer(http.Dir("static"))))
	http.ListenAndServe(":8080", nil)

}
