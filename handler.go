package main

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"image"
	"image/color"
	"io/ioutil"
	"log"
	"net/http"
	"os/exec"
	"strings"
	"sync"
	"time"

	"github.com/nfnt/resize"
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
		http.Error(w, err.Error(), http.StatusInternalServerError)
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
		// get input
		b, err := ioutil.ReadAll(req.Body)
		defer req.Body.Close()
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}

		var uploaded imageData
		err = json.Unmarshal(b, &uploaded)
		if err != nil {
			err, status = errors.New(http.StatusText(http.StatusInternalServerError)), http.StatusInternalServerError
			return
		}

		imageURL := strings.Replace(uploaded.Image, "data:image/png;base64,", "", 1)
		unbased, err := base64.StdEncoding.DecodeString(imageURL)
		if err != nil {
			err, status = errors.New(http.StatusText(http.StatusInternalServerError)), http.StatusInternalServerError
			return
		}

		srcImage, format, err := image.Decode(bytes.NewReader(unbased))
		if err != nil {
			err, status = errors.New(http.StatusText(http.StatusInternalServerError)), http.StatusInternalServerError
			return
		}
		fmt.Println(srcImage.Bounds(), format)

		username := req.URL.Query().Get("username")
		uuid, err := exec.Command("uuidgen").Output()
		if err != nil {
			log.Fatal(err)
		}

		var wg sync.WaitGroup
		wg.Add(4)

		var thumbnail image.Image
		var t1 image.Image
		var t2 image.Image
		var t3 image.Image

		go func(thumbnail image.Image, wg *sync.WaitGroup) {
			defer wg.Done()
			thumbnail = resize.Thumbnail(100, 100, srcImage, resize.NearestNeighbor)
			path := fmt.Sprintf("%s/%s/thumbnail.%s", username, uuid, format)
			SetImage(thumbnail, path)

		}(thumbnail, &wg)

		go func(t1 image.Image, wg *sync.WaitGroup) {
			defer wg.Done()

			width := srcImage.Bounds().Size().X
			height := srcImage.Bounds().Size().Y
			fmt.Println(width, height)

			t1 = srcImage
			blue := color.RGBA{0, 0, 255, 255}
			for x := 0; x < width; x++ {
				for y := 0; y < height; y++ {
					t1.ColorModel().Convert(blue)
				}
			}

			path := fmt.Sprintf("%s/%s/t1.%s", username, uuid, format)
			SetImage(t1, path)
		}(t1, &wg)

		go func(t2 image.Image, wg *sync.WaitGroup) {
			defer wg.Done()
			width := srcImage.Bounds().Size().X
			height := srcImage.Bounds().Size().Y
			fmt.Println(width, height)

			t2 = srcImage
			green := color.RGBA{0, 255, 0, 255}
			for x := 0; x < width; x++ {
				for y := 0; y < height; y++ {
					t2.ColorModel().Convert(green)
				}
			}
			path := fmt.Sprintf("%s/%s/t2.%s", username, uuid, format)
			SetImage(t2, path)
		}(t2, &wg)

		go func(t3 image.Image, wg *sync.WaitGroup) {
			defer wg.Done()
			width := srcImage.Bounds().Size().X
			height := srcImage.Bounds().Size().Y
			fmt.Println(width, height)

			t3 = srcImage
			red := color.RGBA{255, 0, 0, 255}
			for x := 0; x < width; x++ {
				for y := 0; y < height; y++ {
					t3.ColorModel().Convert(red)
				}
			}
			path := fmt.Sprintf("%s/%s/t3.%s", username, uuid, format)
			SetImage(t3, path)
		}(t3, &wg)

		wg.Wait()

		return
	default:
		err, status = errors.New(http.StatusText(http.StatusInternalServerError)), http.StatusInternalServerError
		return
	}

}

func ValidateCookie(handler http.Handler) (out http.Handler) {

	out = http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {

		now := time.Now()
		err := error(nil)

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
