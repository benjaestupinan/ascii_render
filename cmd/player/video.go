package main

import (
	"fmt"
	"image"
	"time"

	vidio "github.com/AlexEidt/Vidio"
	"github.com/benjaestupinan/ascii_render/render"
)

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

func videoToAscii(video_path string, resolution int) {
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
		f := render.FromImage(next_frame, resolution)
		ready_channel <- next_frame_index
		ascii := f.String()
		<-ticker.C
		fmt.Print("\033[H" + ascii)
	}
}
