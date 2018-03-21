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

	// srcImage, err := GetImage("/Users/jingnanchen/Desktop/test.png")
	// if err != nil {
	// 	fmt.Println(err)
	// }
	// SetImage(srcImage, "/Users/jingnanchen/Desktop/testout.png")

	http.HandleFunc("/user/", HandleUser)
	http.Handle("/image/", ValidateCookie(http.HandlerFunc(HandleImage)))
	http.Handle("/", ValidateCookie(http.FileServer(http.Dir("static"))))
	http.ListenAndServe(":8080", nil)

}
