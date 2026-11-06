// An example to try out the keyboard: a snake of constant length that moves
// on its own and wraps around the edges of the screen.
//
//	arrows / wasd  turn
//	q / esc        quit
package main

import (
	"fmt"

	"github.com/benjaestupinan/ascii_render/engine"
	"github.com/benjaestupinan/ascii_render/render"
)

const (
	length = 10   // in segments
	speed  = 10.0 // segments per second
)

type point struct{ x, y int }

type snake struct {
	body []point // body[0] is the head
	dir  point   // how much the head moves on every step
	wait float32 // seconds until the next step
}

// KeyActions makes snake an engine.KeyHandler.
func (s *snake) KeyActions() engine.KeyActionMap {
	// turn returns an action that changes the direction, unless it is the
	// opposite one: the head can't go back over the body
	turn := func(dir point) engine.Action {
		return func() error {
			if dir != (point{-s.dir.x, -s.dir.y}) {
				s.dir = dir
			}
			return nil
		}
	}
	left, right := turn(point{-1, 0}), turn(point{1, 0})
	up, down := turn(point{0, -1}), turn(point{0, 1})
	quit := func() error { return engine.ErrQuit }

	return engine.KeyActionMap{
		"left": left, "a": left,
		"right": right, "d": right,
		"up": up, "w": up,
		"down": down, "s": down,
		"q": quit, "esc": quit,
	}
}

func (s *snake) Update(dt float32) error {
	s.wait -= dt
	if s.wait > 0 {
		return nil
	}
	s.wait += 1 / speed

	// the new head goes in front, and the last segment is dropped
	head := point{s.body[0].x + s.dir.x, s.body[0].y + s.dir.y}
	s.body = append([]point{head}, s.body[:len(s.body)-1]...)
	return nil
}

func (s *snake) Draw(f *render.Frame) {
	// each segment is 2 columns wide, so that it looks more or less square
	cols, rows := f.Width/2, f.Height
	for _, p := range s.body {
		// wrap around the edges (% can be negative in Go)
		x := (p.x%cols + cols) % cols
		y := (p.y%rows + rows) % rows
		f.Set(2*x, y, 1)
		f.Set(2*x+1, y, 1)
	}
}

func main() {
	// all the segments start at (0, 0), and the snake unrolls as it moves
	s := &snake{body: make([]point, length), dir: point{1, 0}}
	if err := engine.Run(s); err != nil {
		fmt.Println(err)
	}
}
