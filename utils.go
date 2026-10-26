package main

import (
	"fmt"
	"image"
	"image/draw"
	_ "image/jpeg"
	_ "image/png"
	"os"
	"strings"
	"time"

	vidio "github.com/AlexEidt/Vidio"
)

const brightness_ascii string = ".'`^,:;Il!i><~+_-?][}{1)(|/tfjrxnuvczXYUJCLQ0OZmwqpdbkhao*#MW&8%B@$"

func ImgToGreyScale(img *image.NRGBA) *image.NRGBA {
	// Changes colors to be grayscale by setting red, green and blue values to the average
	// of this 3 values (255, 100, 200) -> (185, 185, 185)
	newImg := image.NewNRGBA(img.Rect)
	h := img.Rect.Dy()
	w := img.Rect.Dx()

	for i := 0; i < h; i++ {

		jSrc := i * img.Stride
		jDst := i * newImg.Stride

		for k := 0; k < w; k++ {

			r := img.Pix[jSrc]
			g := img.Pix[jSrc+1]
			b := img.Pix[jSrc+2]

			gray := uint8((int(r) + int(g) + int(b)) / 3)

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

	index := ( (int(col) * length - 1) / 255 )

	return ascii[index]
}

func GrayScaleToAscii(img *image.NRGBA, ascii string, width int) string {
	img_h := img.Rect.Dy()
	img_w := img.Rect.Dx()

	// cuántos pixeles reales corresponden a un caracter
	block_w := float64(img_w) / float64(width)
	block_h := block_w * 2.4 // ratio of width to height 1:2.4
	height := int(float64(img_h) / block_h)

	var ret_ascii strings.Builder

	for i := 0; i < height; i++ {
		for j := 0; j < width; j++ {

			// rangos de pixeles de ESTE carácter
			start_x := int(float64(j) * block_w)
			end_x := int(float64(j+1) * block_w)

			start_y := int(float64(i) * block_h)
			end_y := int(float64(i+1) * block_h)

			if end_x <= start_x {
				end_x = start_x + 1
			}
			if end_y <= start_y {
				end_y = start_y + 1
			}

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

			ret_ascii.WriteByte(mapColorToAscii(uint(avg), ascii))
		}
		ret_ascii.WriteByte('\n')
	}

	return ret_ascii.String()
}

func ImageToAscii(image_path string, resolution int) {

	reader, err := os.Open(image_path)
	if err != nil {
		println("error abriendo el archivo", err)
		return
	}
	defer reader.Close()
	img, _, err := image.Decode(reader)
	if err != nil {
		fmt.Println("error decodificando imagen:", err)
		return
	}

	in_mem_img := image.NewNRGBA(img.Bounds())
	draw.Draw(in_mem_img, img.Bounds(), img, image.Point{}, draw.Src)
	gray_in_mem_img := ImgToGreyScale(in_mem_img)

	// ordered ascii brightness string
	println(GrayScaleToAscii(gray_in_mem_img, brightness_ascii, resolution))
}

func setFrameRoutine(video *vidio.Video, frame_buffers_list []*image.NRGBA, ready_channel chan int, frames_channel chan int) {
	// cerrar el channel le avisa a VideoToAscii que no vienen más frames
	defer close(frames_channel)

	// espera un buffer libre, lee el siguiente frame y lo manda
	for frame_buffer_index := range ready_channel {
		err := video.SetFrameBuffer(frame_buffers_list[frame_buffer_index].Pix)
		if err != nil {
			fmt.Println("error setteando el framebuffer", err)
			return
		}

		if !video.Read() {
			return
		}
		frames_channel <- frame_buffer_index
	}
}

func VideoToAscii(video_path string, resolution int) {
	video, err := vidio.NewVideo(video_path)
	if err != nil {
		fmt.Println("error creando video", err)
		return
	}
	defer video.Close()

	w, h := video.Width(), video.Height()
	switch video.MetaData()["side_datum/display_matrix:rotation"] {
	case "90", "-90", "270", "-270":
		w, h = h, w
	}

	const N int = 5


	frames_list := make([]*image.NRGBA, N)
	frames_channel := make(chan int, N)
	ready_channel := make(chan int, N)
	for i := 0; i < N; i++ {
		frames_list[i] = image.NewNRGBA(image.Rect(0, 0, w, h))
	}

	// al principio todos los buffers están libres
	for i := 0; i < N; i++ {
		ready_channel <- i
	}

	// una sola goroutine es dueña de `video`
	go setFrameRoutine(video, frames_list, ready_channel, frames_channel)

	ticker := time.NewTicker(time.Duration(float64(time.Second) / video.FPS()))
	defer ticker.Stop()

	fmt.Print("\033[H\033[2J")
	for next_frame_index := range frames_channel {
		next_frame := frames_list[next_frame_index]
		gray_in_mem_img := ImgToGreyScale(next_frame)
		ready_channel <- next_frame_index
		ascii := GrayScaleToAscii(gray_in_mem_img, brightness_ascii, resolution)
		<-ticker.C
		fmt.Print("\033[H" + ascii)
	}
}
