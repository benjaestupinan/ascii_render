package render

import (
	"strings"
)

const brightness_ascii string = ".'`^,:;Il!i><~+_-?][}{1)(|/tfjrxnuvczXYUJCLQ0OZmwqpdbkhao*#MW&8%B@$"

func mapColorToAscii(light float32, ascii string) byte {

	length := len(ascii)

	index := int(light * float32(length-1))

	return ascii[index]
}

type Frame struct {
	Width, Height int
	Light         []float32 // row by row, cell (x, y) would be .Light[x + .Width*y]
}

// New creates a new frame of corresponding .Width and .Height with all black cells
func New(width, height int) *Frame {
	ret_frame := Frame{Width: width, Height: height, Light: make([]float32, width*height)}
	return &ret_frame
}

// Set sets the light a the cell (x, y) to the given light value in [0-1]
func (f *Frame) Set(x, y int, light float32) {

	if x < 0 || x >= f.Width || y < 0 || y >= f.Height {
		return
	}

	index := x + y*f.Width
	f.Light[index] = max(0, min(1, light))
}

// Clear clears the frame and leaves it completely black
func (f *Frame) Clear() {
	for i := range f.Light {
		f.Light[i] = 0
	}
}

// String converts a frame into its ascii form
func (f *Frame) String() string {
	var ret_ascii strings.Builder
	for i := range f.Height {
		for j := range f.Width {
			ret_ascii.WriteByte(mapColorToAscii(f.Light[j+i*f.Width], brightness_ascii))
		}
		ret_ascii.WriteByte('\n')
	}

	return ret_ascii.String()
}
