package main

import (
	"fmt"
	"image"
	_ "image/jpeg"
	_ "image/png"
	"os"

	"github.com/benjaestupinan/ascii_render/render"
)

func imageToAscii(image_path string, resolution int) {

	reader, err := os.Open(image_path)
	if err != nil {
		fmt.Println("error opening the file:", err)
		return
	}
	defer reader.Close()
	img, _, err := image.Decode(reader)
	if err != nil {
		fmt.Println("error decoding image:", err)
		return
	}

	frame := render.FromImage(img, resolution)
	fmt.Print(frame.String())
}
