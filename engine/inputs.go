package engine

import (
	"fmt"

	"github.com/nsf/termbox-go"
)

type Action func() error

type KeyActionMap map[string]Action

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

func ReadInput(inputs_channel chan string) error {
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
