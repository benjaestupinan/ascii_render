package engine

import (
	"fmt"

	"github.com/nsf/termbox-go"
)

// Action is what a key does when it is pressed, usually a closure that
// changes the game. Run calls it once per key press, in the order the keys
// were pressed, at the start of the next tick and before Update. It runs in
// the same goroutine as Update and Draw, so it can change the game freely.
//
// Returning ErrQuit ends the game normally, and any other error stops Run
// with that error. Keys pressed after an Action that fails are not run.
type Action func() error

// KeyActionMap assigns an Action to each key name (see KeyHandler for the
// names). Keys that are not in the map are ignored, and the same Action can
// be assigned to more than one key.
type KeyActionMap map[string]Action

// KeyHandler is implemented by games that use the keyboard. It is optional:
// if the Game passed to Run is also a KeyHandler, Run reads the keyboard and
// calls KeyActions once, before the first tick.
//
// Keys are named after their character ("a", "A", "1", "?"), and special
// keys by lowercase names: "up", "down", "left", "right", "space", "enter",
// "esc", "f1" to "f12", "insert", "delete", "home", "end", "pgup" and "pgdn".
// Ctrl+C always quits, whatever the map says.
type KeyHandler interface {
	KeyActions() KeyActionMap
}

func keyName(ev termbox.Event) string {

	return string(ev.Ch)
}

func executeKeyAction(keyActions KeyActionMap, key string) error {
	action, ok := keyActions[key]
	if ok {
		err := action()
		if err != nil {
			return err
		}
	}

	return nil
}

// readInput waits for termbox events and sends the name of every key pressed
// to inputs_channel. It runs in its own goroutine, because PollEvent blocks,
// and returns when termbox.Interrupt is called or termbox fails.
func readInput(inputs_channel chan string) error {
	for {
		ev := termbox.PollEvent()
		switch ev.Type {

		case termbox.EventError:
			return fmt.Errorf("engine: event returned error: %w", ev.Err)
		default:
			inputs_channel <- keyName(ev)
		}
	}
}
