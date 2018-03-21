package main

import (
	"fmt"
	"image"
	"image/jpeg"
	"image/png"
	"os"
	"path"
	"strings"
)

func GetImage(url string) (out image.Image, err error) {
	imagePath, _ := os.Open(url)
	defer imagePath.Close()

	srcImage, format, err := image.Decode(imagePath)

	fmt.Println("[DEBUG] ", url, " ", format)

	out = srcImage
	return
}

func SetImage(srcImage image.Image, url string) (err error) {

	out, err := os.Create(path.Join(url))
	if err != nil {
		fmt.Println(err)
		return
	}

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
