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
// and then Draw. A Game that also implements KeyHandler can use the keyboard.
type Game interface {
	// Update advances the scene by dt seconds. It returns ErrQuit to finish
	// normally, or any other error to stop Run with that error.
	Update(dt float32) error

	// Draw draws the current state of the scene into f, which arrives black.
	Draw(f *render.Frame)
}

// ErrQuit is returned by Game.Update or by an Action to stop the game loop.
// Run treats it as a normal exit and returns nil.
var ErrQuit = errors.New("engine: quit")

// processKeys runs the action of every key waiting in keys, in order, and
// returns as soon as the channel is empty, without waiting for new keys. It
// stops at the first action that fails: ErrQuit is returned as is, so Run can
// tell a normal exit apart, and other errors are wrapped. "ctrl+c" returns
// ErrQuit even if keyActions has no action for it.
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
// error or the user presses Ctrl+C. If g is a KeyHandler, Run also reads the
// keyboard and runs the actions of the keys pressed before each Update. On
// exit it always leaves the terminal as it was. It returns nil if the game
// ended with ErrQuit or Ctrl+C.
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
		go readInput(inputs_channel)
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
