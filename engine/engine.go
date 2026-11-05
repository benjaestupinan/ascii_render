// Package engine runs a game loop that draws a Game in the terminal.
package engine

import (
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/benjaestupinan/ascii_render/render"
	"github.com/nsf/termbox-go"
	"golang.org/x/term"
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

func processKeys(keys <-chan string, keyActions KeyActionMap) error {
	for {
		select {
		case k := <-keys:
			err := executeKeyAction(keyActions, k)
			if err != nil {
				if errors.Is(err, ErrQuit) {
					return ErrQuit
				}
				return fmt.Errorf("engine: key action execution failed: %w", err)
			}
		default:
			return nil
		}
	}
}

// Run takes over the terminal and runs g's game loop until Update returns an
// error or the user presses Ctrl+C. On exit it always leaves the terminal as
// it was. It returns nil if the game ended with ErrQuit or Ctrl+C.
func Run(g Game) error {
	width, height, err := term.GetSize(int(os.Stdout.Fd()))
	if err != nil {
		return fmt.Errorf("engine: unable to get terminal size: %w", err)
	}
	frame := render.NewFrame(width, height-1)
	ticker := time.NewTicker(time.Second / 30)
	defer ticker.Stop()

	fmt.Print("\033[?1049h\033[?25l")
	defer fmt.Print("\033[?25h\033[?1049l")

	inputs_channel := make(chan string, 16)
	var keyActions KeyActionMap
	if keyHandler, ok := g.(KeyHandler); ok {
		err := termbox.Init()
		if err != nil {
			return fmt.Errorf("engine: unable to initialize termbox: %w", err)
		}
		defer termbox.Close()
		defer termbox.Interrupt()
		keyActions = keyHandler.KeyActions()
		go ReadInput(inputs_channel)
	}

	last := time.Now()
	for {
		newWidth, newHeight, err := term.GetSize(int(os.Stdout.Fd()))
		if err != nil {
			return fmt.Errorf("engine: unable to get terminal size: %w", err)
		}
		if newWidth != frame.Width || newHeight != frame.Height+1 {
			fmt.Print("\033[2J")
			frame = render.NewFrame(newWidth, newHeight-1)
		}
		<-ticker.C
		dt := time.Since(last)
		last = time.Now()

		err = processKeys(inputs_channel, keyActions)
		if err != nil {
			return err
		}

		err = g.Update(float32(dt.Seconds()))
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
