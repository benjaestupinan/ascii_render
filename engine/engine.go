// Package engine runs a game loop that draws a Game in the terminal.
package engine

import (
	"errors"
	"fmt"
	"time"

	"github.com/benjaestupinan/ascii_render/render"
)

// Game is a scene animated by engine.Run. On every tick, Run calls Update
// and then Draw.
type Game interface {
	// Update advances the scene by dt seconds. It returns ErrQuit to finish
	// normally, or any other error to stop Run with that error.
	Update(dt float32) error

	// Draw draws the current state of the scene into f, which arrives black.
	Draw(f *render.Frame)
}

// ErrQuit is returned by Game.Update to stop the game loop. Run treats it as
// a normal exit and returns nil.
var ErrQuit = errors.New("engine: quit")

// Run takes over the terminal and runs g's game loop until Update returns an
// error or the user presses Ctrl+C. On exit it always leaves the terminal as
// it was. It returns nil if the game ended with ErrQuit or Ctrl+C.
func Run(g Game) error {
	frame := render.NewFrame(465, 109)
	ticker := time.NewTicker(time.Second / 30)
	defer ticker.Stop()

	fmt.Print("\033[?1049h\033[?25l")
	defer fmt.Print("\033[?25h\033[?1049l")

	last := time.Now()
	for {
		<-ticker.C
		dt := time.Since(last)
		last = time.Now()

		err := g.Update(float32(dt.Seconds()))
		if err != nil {
			if errors.Is(err, ErrQuit) {
				return nil
			}
			return err
		}

		frame.Clear()

		g.Draw(frame)
		ascii := frame.String()

		fmt.Print("\033[H" + ascii)
	}
}
