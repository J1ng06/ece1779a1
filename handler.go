package main

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"image"
	"image/color"
	"image/jpeg"
	"image/png"
	"io/ioutil"
	"log"
	"net/http"
	"os/exec"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/nfnt/resize"
)

func HandleUser(w http.ResponseWriter, req *http.Request) {

	//Timer
	now := time.Now()
	//Output
	data, status, err := []byte(nil), http.StatusOK, error(nil)

	//req.Header.Get("Cookie")

	//Log Request
	log.Printf("Request [Method: %s] [Path: %s]", req.Method, req.URL.Path)

	//Write response
	defer func() {

		//Response
		if status == http.StatusOK {
			w.Header().Set("Content-Type", "application/json")
			w.Write(data)
		} else {
			w.WriteHeader(status)
		}

	}()

	//Log Response
	defer func() {
		log.Printf("Response [Status: %d] [Path: %s] [Roundtrip: %d ms] [Error: %v]", status, req.URL.Path, (time.Now().Sub(now))/time.Millisecond, err)
	}()

	// check http method
	if req.Method != "POST" {
		err, status = errors.New(http.StatusText(http.StatusMethodNotAllowed)), http.StatusMethodNotAllowed
		return
	}

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

	username := userReq.Username
	password := userReq.Password

	if username == "" || password == "" {
		err, status = errors.New(http.StatusText(http.StatusBadRequest)), http.StatusBadRequest
		return
	}

	db, err := NewConnection()
	if err != nil {
		err, status = errors.New(http.StatusText(http.StatusInternalServerError)), http.StatusInternalServerError
		return
	}
	defer db.Close()

	user := new(User)
	switch action {
	case "login":

		db.Find(user, "username = ?", username)
		if user.Authentication(password) {

			w.Header().Set("Location", fmt.Sprintf("upload.html?username=%s", username))

			existing := sessions.Get(username)
			if existing == nil {

				value, err := NewCookieValue()
				if err != nil {
					err, status = errors.New(http.StatusText(http.StatusInternalServerError)), http.StatusInternalServerError
					return
				}

				newCookie := &Cookie{
					Name:  username,
					Value: value,
				}

				sessions.Set(newCookie)

				data, err = json.Marshal(newCookie)
				if err != nil {
					err, status = errors.New(http.StatusText(http.StatusInternalServerError)), http.StatusInternalServerError
					return
				}

			} else {

				data, err = json.Marshal(existing)
				if err != nil {
					err, status = errors.New(http.StatusText(http.StatusInternalServerError)), http.StatusInternalServerError
					return
				}

			}

		} else {
			err, status = errors.New(http.StatusText(http.StatusNotFound)), http.StatusNotFound
			w.Header().Set("Location", "/")
			return
		}

	case "register":
		db.Find(user, "username = ?", username)

		if user.Password != "" {
			err, status = errors.New(http.StatusText(http.StatusBadRequest)), http.StatusBadRequest
			return
		}

		err = user.RandomSalt()
		if err != nil {
			err, status = errors.New(http.StatusText(http.StatusInternalServerError)), http.StatusInternalServerError
			return
		}

		user.EncPass(password)
		db.Create(&user)
		w.Header().Set("Location", "/")

	case "logout":
		existing := sessions.Get(username)
		sessions.Del(existing)
		w.Header().Set("Location", "/")

	default:
		err, status = errors.New(http.StatusText(http.StatusBadRequest)), http.StatusBadRequest
		return
	}

}

func HandleImage(w http.ResponseWriter, req *http.Request) {

	now := time.Now()

	data, status, err := []byte(nil), http.StatusOK, error(nil)

	//Write response
	defer func() {

		//Response
		if status == http.StatusOK {
			w.Header().Set("Content-Type", "application/json")
			w.Write(data)
		} else {
			w.WriteHeader(status)
		}

	}()

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
	username := req.URL.Query().Get("username")

	db, _ := NewConnection()
	if err != nil {
		err, status = errors.New(http.StatusText(http.StatusInternalServerError)), http.StatusInternalServerError
		return
	}
	defer db.Close()

	user := &User{}
	db.Find(&user, "username = ?", username)

	pageSize := 8

	switch action {
	case "upload":
		// get input
		body, err := ioutil.ReadAll(req.Body)
		defer req.Body.Close()
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}

		var uploaded imageData
		err = json.Unmarshal(body, &uploaded)
		if err != nil {
			err, status = errors.New(http.StatusText(http.StatusInternalServerError)), http.StatusInternalServerError
			return
		}

		imageURL := strings.Replace(uploaded.Image, "data:image/png;base64,", "", 1)
		imageURL = strings.Replace(imageURL, "data:image/jpeg;base64,", "", 1)
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

		uuid, err := exec.Command("uuidgen").Output()
		if err != nil {
			log.Fatal(err)
		}
		uuid = uuid[:len(uuid)-1]

		var wg sync.WaitGroup
		wg.Add(4)

		var thumbnail image.Image
		b := srcImage.Bounds()
		t1, t2, t3 := image.NewRGBA(b), image.NewRGBA(b), image.NewRGBA(b)

		go func() {
			SetImage(srcImage, fmt.Sprintf("%s/%s/original.%s", username, uuid, format))
		}()

		go func(thumbnail image.Image, wg *sync.WaitGroup) {
			defer wg.Done()
			thumbnail = resize.Thumbnail(100, 100, srcImage, resize.NearestNeighbor)
			path := fmt.Sprintf("%s/%s/thumbnail.%s", username, uuid, format)
			SetImage(thumbnail, path)

		}(thumbnail, &wg)

		go func(t1 *image.RGBA, wg *sync.WaitGroup) {
			defer wg.Done()

			for y := b.Min.Y; y < b.Max.Y; y++ {
				for x := b.Min.X; x < b.Max.X; x++ {
					oldPixel := srcImage.At(x, y)
					pixel := color.NRGBAModel.Convert(oldPixel)
					t1.Set(x, y, pixel)
				}
			}

			path := fmt.Sprintf("%s/%s/t1.%s", username, uuid, format)
			SetImage(t1, path)
		}(t1, &wg)

		go func(t2 *image.RGBA, wg *sync.WaitGroup) {
			defer wg.Done()

			for y := b.Min.Y; y < b.Max.Y; y++ {
				for x := b.Min.X; x < b.Max.X; x++ {
					oldPixel := srcImage.At(x, y)
					pixel := color.GrayModel.Convert(oldPixel)
					t2.Set(x, y, pixel)
				}
			}
			path := fmt.Sprintf("%s/%s/t2.%s", username, uuid, format)
			SetImage(t2, path)
		}(t2, &wg)

		go func(t3 *image.RGBA, wg *sync.WaitGroup) {
			defer wg.Done()

			for y := b.Min.Y; y < b.Max.Y; y++ {
				for x := b.Min.X; x < b.Max.X; x++ {
					oldPixel := srcImage.At(x, y)
					pixel := color.CMYKModel.Convert(oldPixel)
					t3.Set(x, y, pixel)
				}
			}
			path := fmt.Sprintf("%s/%s/t3.%s", username, uuid, format)
			SetImage(t3, path)
		}(t3, &wg)

		wg.Wait()
		db.Create(&Image{
			User_ID:   user.ID,
			Original:  fmt.Sprintf("%s/%s/original.%s", username, uuid, format),
			Thumbnail: fmt.Sprintf("%s/%s/thumbnail.%s", username, uuid, format),
			T1:        fmt.Sprintf("%s/%s/t1.%s", username, uuid, format),
			T2:        fmt.Sprintf("%s/%s/t2.%s", username, uuid, format),
			T3:        fmt.Sprintf("%s/%s/t3.%s", username, uuid, format),
		})

		w.Header().Set("Location", fmt.Sprintf("upload.html?username=%s", username))

		return

	case "userimagecount":

		db.Where("id=?", user.ID).Preload("Images").Find(&user)

		// TODO: fix the json marshal with password and salt
		user.Password = ""
		user.Salt = ""

		data, err = json.Marshal(len(user.Images))
		if err != nil {
			err, status = errors.New(http.StatusText(http.StatusInternalServerError)), http.StatusInternalServerError
		}

		return

	case "userimages":

		page, err := strconv.Atoi(req.URL.Query().Get("page"))
		if err != nil {
			err, status = errors.New(http.StatusText(http.StatusBadRequest)), http.StatusBadRequest
			return
		}

		db.Where("id=?", user.ID).Preload("Images").Find(&user)

		// TODO: fix the json marshal with password and salt
		user.Password = ""
		user.Salt = ""
		start := page * pageSize
		end := (page + 1) * pageSize
		if end > len(user.Images) {
			end = len(user.Images)
		}
		data, err = json.Marshal(user.Images[start:end])
		if err != nil {
			err, status = errors.New(http.StatusText(http.StatusInternalServerError)), http.StatusInternalServerError
		}

		return
	case "getimage":

		location := req.URL.Query().Get("location")
		if location == "" {
			err, status = errors.New(http.StatusText(http.StatusBadRequest)), http.StatusBadRequest
			return
		}

		result, err := GetImage(location)
		if err != nil {
			err, status = errors.New(http.StatusText(http.StatusInternalServerError)), http.StatusInternalServerError
			return
		}

		buf := new(bytes.Buffer)
		encodedString := ""
		if strings.HasSuffix(location, ".png") {
			err = png.Encode(buf, result)
			encodedString = fmt.Sprintf("data:image/png;base64,%s", base64.StdEncoding.EncodeToString(buf.Bytes()))
		} else if strings.HasSuffix(location, ".jpeg") {
			err = jpeg.Encode(buf, result, &jpeg.Options{jpeg.DefaultQuality})
			encodedString = fmt.Sprintf("data:image/jpeg;base64,%s", base64.StdEncoding.EncodeToString(buf.Bytes()))
		}

		data = []byte(encodedString)

		return

	default:
		err, status = errors.New(http.StatusText(http.StatusInternalServerError)), http.StatusInternalServerError
		return
	}

}

func ValidateCookie(handler http.Handler) (out http.Handler) {

	out = http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		if (req.URL.Path != "/" && req.URL.Path != "/register.html") && strings.HasSuffix(req.URL.Path, ".html") {

			username := req.URL.Query().Get("username")

			if username == "" {
				return
			}

			cookie, err := req.Cookie(username)
			if err != nil {
				return
			}

			if c := sessions.Get(cookie.Name); c == nil {
				return
			}

		}
		handler.ServeHTTP(w, req)
	})

	return
}
