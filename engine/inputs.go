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
	// fmt.Printf("%+v\n", ev) // usefull to tell "what Key is each key"
	if ev.Type != 0{
		return ""
	}
	switch ev.Key {
	default:
		return string(ev.Ch)
	case 3:
		return "ctrl+c"
	case 32:
		return "space"
	case 13:
		return "enter"
	case 27:
		return "esc"
	case 65517:
		return "up"
	case 65516:
		return "down"
	case 65515:
		return "left"
	case 65514:
		return "right"
	case 127:
		return "delete"
	case 65522:
		return "delete"
	case 65523:
		return "insert"
	case 65521:
		return "home"
	case 65520:
		return "end"
	case 65519:
		return "pgup"
	case 65518:
		return "pgdn"
	case 65535:
		return "f1"
	case 65534:
		return "f2"
	case 65533:
		return "f3"
	case 65532:
		return "f4"
	case 65531:
		return "f5"
	case 65530:
		return "f6"
	case 65529:
		return "f7"
	case 65528:
		return "f8"
	case 65527:
		return "f9"
	case 65526:
		return "f10"
	case 65525:
		return "f11"
	case 65524:
		return "f12"
	}
}

func executeKeyAction(keyActions KeyActionMap, key string) error {

	if key == "ctrl+c" {
		return ErrQuit
	}

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
