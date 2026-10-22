package main

import (
	"fmt"
	"image/draw"
	"image"
	"os"
	"strconv"
)

func main() {
	if len(os.Args) < 3 {
		fmt.Println("Falta argumento")
		return
	}

	img_path := os.Args[1]
	reader, err := os.Open(img_path)
	if err != nil {
		println("error abriendo el archivo")
		return
	}
	defer reader.Close()

	res_s := os.Args[2]
	resolution, err := strconv.Atoi(res_s)
	if err != nil {
		fmt.Println("Argumento no es número")
		return
	}
	img, _, err := image.Decode(reader)
	if err != nil {
		fmt.Println("error decodificando imagen:", err)
		return
	}

	in_mem_img := image.NewNRGBA(img.Bounds())
	draw.Draw(in_mem_img, img.Bounds(), img, image.Point{}, draw.Src)
	gray_in_mem_img := ImgToGreyScale(in_mem_img)

	//ordered ascii brightness string
	println(GrayScaleToAscii(gray_in_mem_img, brightness_ascii, resolution))

}
