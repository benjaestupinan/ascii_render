package main

import (
	"fmt"
	"os"
	"strconv"
	"strings"
)

func main() {
	if len(os.Args) < 3 {
		fmt.Println("Falta argumento")
		return
	}


	res_s := os.Args[2]
	resolution, err := strconv.Atoi(res_s)
	if err != nil {
		fmt.Println("Argumento no es número")
		return
	}

	file_path := os.Args[1]
	switch true {
	case strings.HasSuffix(file_path, ".png"):
		ImageToAscii(file_path, resolution)
	case strings.HasSuffix(file_path, ".jpeg"):
		ImageToAscii(file_path, resolution)
	case strings.HasSuffix(file_path, ".mp4"):
		VideoToAscii(file_path, resolution)
	}

}
