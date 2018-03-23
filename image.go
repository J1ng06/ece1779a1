package main

import (
	"fmt"
	i "image"
	"image/jpeg"
	"image/png"
	"os"
	"path"
	"strings"
)

type imageData struct {
	Name  string `json:"name"`
	Image string `json:"image"`
}

func GetImage(url string) (out i.Image, err error) {
	imageFile, _ := os.Open(url)
	defer imageFile.Close()

	srcImage, format, err := i.Decode(imageFile)

	fmt.Println("[DEBUG] ", url, " ", format)

	out = srcImage
	return
}

func SetImage(srcImage i.Image, url string) (err error) {

	err = os.MkdirAll(path.Dir(url), os.ModePerm)
	if err != nil {
		return
	}

	out, err := os.Create(url)
	if err != nil {
		return
	}
	defer out.Close()

	if strings.HasSuffix(url, ".jpg") {
		err = jpeg.Encode(out, srcImage, &jpeg.Options{jpeg.DefaultQuality})
	} else if strings.HasSuffix(url, ".png") {
		err = png.Encode(out, srcImage)
	}

	if err != nil {
		fmt.Println(err)
		return
	}

	return
}
