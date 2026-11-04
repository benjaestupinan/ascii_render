// The simplest engine example: a comet crosses the screen from left to right
// and the scene ends when it leaves through the other side.
package main

import (
	"fmt"

	"github.com/benjaestupinan/ascii_render/engine"
	"github.com/benjaestupinan/ascii_render/render"
)

const (
	duration   = 4.0 // seconds it takes to cross the screen
	tailLength = 12  // length of the tail, in cells
)

type scene struct {
	// position as a fraction of the screen width: 0 is the left edge and 1
	// the right one. This way Update doesn't need to know how wide it is.
	x float32
}

func (s *scene) Update(dt float32) error {
	s.x += dt / duration
	if s.x > 1 {
		return engine.ErrQuit
	}
	return nil
}

func (s *scene) Draw(f *render.Frame) {
	col := int(s.x * float32(f.Width))
	row := f.Height / 2

	// the head is the brightest cell and the tail fades out behind it.
	// Set ignores the cells that fall outside the frame.
	for i := range tailLength {
		light := 1 - float32(i)/tailLength
		f.Set(col-i, row, light)
	}
}

func main() {
	if err := engine.Run(&scene{}); err != nil {
		fmt.Println("error running the engine:", err)
	}
}
