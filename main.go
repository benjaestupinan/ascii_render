package main

import (
	"fmt"
	"image"
	"image/draw"
	_ "image/png"
	"os"
	"strconv"

)

func ImgToGreyScale(img *image.NRGBA) *image.NRGBA {
	// Changes colors to be grayscale by setting red, green and blue values to the average
	// of this 3 values (255, 100, 200) -> (185, 185, 185)
	var newImg *image.NRGBA = image.NewNRGBA(img.Rect)
	var h int = img.Rect.Dy()
	var w int = img.Rect.Dx()

	for i := 0; i < h; i++ {

		jSrc := i * img.Stride
		jDst := i * newImg.Stride

		for k := 0; k < w; k++ {

			r := img.Pix[jSrc]
			g := img.Pix[jSrc+1]
			b := img.Pix[jSrc+2]

			var gray uint8 = uint8((int(r) + int(g) + int(b)) / 3)

			newImg.Pix[jDst] = gray
			newImg.Pix[jDst+1] = gray
			newImg.Pix[jDst+2] = gray
			newImg.Pix[jDst+3] = img.Pix[jSrc+3]

			jDst += 4
			jSrc += 4

		}
	}
	return newImg
}

func mapColorToAscii(col uint, ascii string) byte {

	// col esta entre [0, 255]

	length := len(ascii)

	index := (col * uint(length)) / 255

	return ascii[index]
}

func GrayScaleToAscii(img *image.NRGBA, ascii string, width int) string {
	img_h := img.Rect.Dy()
	img_w := img.Rect.Dx()

	// cuántos pixeles reales corresponden a un caracter
	block_w := float64(img_w) / float64(width)
	block_h := block_w * 2 // mantener bloques cuadrados
	height := int(float64(img_h) / block_h)

	var ret_ascii string

	for i := 0; i < height; i++ {
		for j := 0; j < width; j++ {

			// rangos de pixeles de ESTE carácter
			start_x := int(float64(j) * block_w)
			end_x := int(float64(j+1) * block_w)

			start_y := int(float64(i) * block_h)
			end_y := int(float64(i+1) * block_h)

			// corregir out-of-bounds
			if end_x > img_w {
				end_x = img_w
			}
			if end_y > img_h {
				end_y = img_h
			}

			// acumular promedio
			var sum float64
			var count int

			for y := start_y; y < end_y; y++ {
				row := y * img.Stride
				for x := start_x; x < end_x; x++ {
					off := row + x*4
					gray := img.Pix[off] // ya está en gris
					sum += float64(gray)
					count++
				}
			}

			avg := sum / float64(count)

			ret_ascii += string(mapColorToAscii(uint(avg), ascii))
		}
		ret_ascii += "\n"
	}

	return ret_ascii
}

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
	const brightness_ascii string = ".'`^,:;Il!i><~+_-?][}{1)(|/tfjrxnuvczXYUJCLQ0OZmwqpdbkhao*#MW&8%B@$"

	println(GrayScaleToAscii(gray_in_mem_img, brightness_ascii, resolution))

}
