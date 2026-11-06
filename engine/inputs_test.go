package engine

import (
	"errors"
	"slices"
	"testing"
	"time"

	"github.com/nsf/termbox-go"
)

// Characters are named after themselves, so KeyActionMap{"a": ...} runs on 'a'.
func TestKeyNameOfCharacters(t *testing.T) {
	for _, ch := range []rune{'a', 'A', 'w', '1', '?'} {
		ev := termbox.Event{Type: termbox.EventKey, Ch: ch}
		if got := keyName(ev); got != string(ch) {
			t.Errorf("keyName(%q) = %q, want %q", ch, got, string(ch))
		}
	}
}

// For special keys termbox leaves Ch at 0 and puts the key in Key instead,
// so string(ev.Ch) would give the same "\x00" for all of them.
func TestKeyNameOfSpecialKeys(t *testing.T) {
	cases := []struct {
		key  termbox.Key
		want string
	}{
		{termbox.KeyArrowUp, "up"},
		{termbox.KeyArrowDown, "down"},
		{termbox.KeyArrowLeft, "left"},
		{termbox.KeyArrowRight, "right"},
		{termbox.KeySpace, "space"},
		{termbox.KeyEnter, "enter"},
		{termbox.KeyEsc, "esc"},
		{termbox.KeyCtrlC, "ctrl+c"},
	}
	for _, c := range cases {
		ev := termbox.Event{Type: termbox.EventKey, Key: c.key}
		if got := keyName(ev); got != c.want {
			t.Errorf("keyName(Key %d) = %q, want %q", c.key, got, c.want)
		}
	}
}

// Function and navigation keys have no ASCII code at all: termbox gives them
// values near 0xFFFF in Key, and Ch stays at 0.
func TestKeyNameOfKeysWithoutASCIICode(t *testing.T) {
	cases := []struct {
		key  termbox.Key
		want string
	}{
		{termbox.KeyF1, "f1"},
		{termbox.KeyF2, "f2"},
		{termbox.KeyF3, "f3"},
		{termbox.KeyF4, "f4"},
		{termbox.KeyF5, "f5"},
		{termbox.KeyF6, "f6"},
		{termbox.KeyF7, "f7"},
		{termbox.KeyF8, "f8"},
		{termbox.KeyF9, "f9"},
		{termbox.KeyF10, "f10"},
		{termbox.KeyF11, "f11"},
		{termbox.KeyF12, "f12"},
		{termbox.KeyInsert, "insert"},
		{termbox.KeyDelete, "delete"},
		{termbox.KeyHome, "home"},
		{termbox.KeyEnd, "end"},
		{termbox.KeyPgup, "pgup"},
		{termbox.KeyPgdn, "pgdn"},
	}
	for _, c := range cases {
		ev := termbox.Event{Type: termbox.EventKey, Key: c.key}
		if got := keyName(ev); got != c.want {
			t.Errorf("keyName(Key %#x) = %q, want %q", c.key, got, c.want)
		}
	}
}

// Resizing the terminal or moving the mouse also produces events. They are
// not keys, so they get no name, and the input loop has to keep going.
func TestKeyNameOfNonKeyEventsIsEmpty(t *testing.T) {
	for _, typ := range []termbox.EventType{termbox.EventResize, termbox.EventMouse} {
		ev := termbox.Event{Type: typ}
		if got := keyName(ev); got != "" {
			t.Errorf("keyName(event type %d) = %q, want \"\"", typ, got)
		}
	}
}

func TestExecuteKeyActionRunsTheActionOfTheKey(t *testing.T) {
	jumps, shots := 0, 0
	actions := KeyActionMap{
		"space": func() error { jumps++; return nil },
		"x":     func() error { shots++; return nil },
	}

	if err := executeKeyAction(actions, "space"); err != nil {
		t.Fatalf("executeKeyAction returned %v, want nil", err)
	}
	if jumps != 1 {
		t.Errorf("the space action ran %d times, want 1", jumps)
	}
	if shots != 0 {
		t.Errorf("the x action ran %d times, want 0 (only the pressed key should run)", shots)
	}
}

// Most keys have nothing assigned: pressing them can't stop the game.
func TestExecuteKeyActionIgnoresKeysWithoutAction(t *testing.T) {
	ran := false
	actions := KeyActionMap{"a": func() error { ran = true; return nil }}

	for _, key := range []string{"b", "up", ""} {
		if err := executeKeyAction(actions, key); err != nil {
			t.Errorf("executeKeyAction(%q) returned %v, want nil", key, err)
		}
	}
	if ran {
		t.Error("an action ran for a key that isn't in the map")
	}
}

// The error of an action has to reach whoever called executeKeyAction, so that an
// action can end the game by returning ErrQuit.
func TestExecuteKeyActionReturnsTheErrorOfTheAction(t *testing.T) {
	actions := KeyActionMap{"q": func() error { return ErrQuit }}

	if err := executeKeyAction(actions, "q"); !errors.Is(err, ErrQuit) {
		t.Errorf("executeKeyAction returned %v, want ErrQuit", err)
	}
}

// pressed returns a channel that already holds keys, as if they had been
// pressed during the last tick.
func pressed(keys ...string) chan string {
	ch := make(chan string, len(keys)+1)
	for _, k := range keys {
		ch <- k
	}
	return ch
}

// recorder returns actions that write down the name of their key in *log
// every time they run.
func recorder(log *[]string, keys ...string) KeyActionMap {
	actions := KeyActionMap{}
	for _, k := range keys {
		actions[k] = func() error { *log = append(*log, k); return nil }
	}
	return actions
}

// process calls processKeys, but fails the test if it hasn't returned after
// one second. Without this, a processKeys that waits for keys or loops forever
// would hang go test without saying which test is stuck.
func process(t *testing.T, keys <-chan string, actions KeyActionMap) error {
	t.Helper()
	done := make(chan error, 1)
	go func() { done <- processKeys(keys, actions) }()

	select {
	case err := <-done:
		return err
	case <-time.After(time.Second):
		t.Fatal("processKeys is still running after 1 second: it's waiting for more keys, or looping forever (is there a break inside the select?)")
		return nil
	}
}

// Every key pressed since the last tick has to run, not just the first one,
// and in the order they were pressed.
func TestProcessKeysRunsEveryPendingKeyInOrder(t *testing.T) {
	var log []string
	actions := recorder(&log, "a", "b", "up")

	if err := process(t, pressed("up", "a", "b", "a"), actions); err != nil {
		t.Fatalf("processKeys returned %v, want nil", err)
	}
	if want := []string{"up", "a", "b", "a"}; !slices.Equal(log, want) {
		t.Errorf("actions ran for %q, want %q", log, want)
	}
}

func TestProcessKeysEmptiesTheChannel(t *testing.T) {
	keys := pressed("a", "b", "c")
	process(t, keys, KeyActionMap{})

	if len(keys) != 0 {
		t.Errorf("%d keys left in the channel, want 0", len(keys))
	}
}

// If no key was pressed, processKeys has to return right away: the game
// can't freeze waiting for the player.
func TestProcessKeysDoesNotWaitForKeys(t *testing.T) {
	if err := process(t, pressed(), KeyActionMap{}); err != nil {
		t.Errorf("processKeys returned %v, want nil", err)
	}
}

// A key without an action doesn't stop the keys that come after it.
func TestProcessKeysSkipsKeysWithoutAction(t *testing.T) {
	var log []string
	actions := recorder(&log, "a")

	if err := process(t, pressed("z", "a", "left"), actions); err != nil {
		t.Fatalf("processKeys returned %v, want nil", err)
	}
	if want := []string{"a"}; !slices.Equal(log, want) {
		t.Errorf("actions ran for %q, want %q", log, want)
	}
}

// A game without keyboard has a nil KeyActionMap. Reading a nil map is
// allowed in Go, so this must not panic.
func TestProcessKeysWithNilActions(t *testing.T) {
	if err := process(t, pressed("a", "space"), nil); err != nil {
		t.Errorf("processKeys returned %v, want nil", err)
	}
}

// When an action fails (for example q returning ErrQuit), the game is over:
// the keys pressed after it must not run.
func TestProcessKeysStopsAtTheFirstError(t *testing.T) {
	var log []string
	actions := recorder(&log, "a", "b")
	actions["q"] = func() error { return ErrQuit }

	err := process(t, pressed("a", "q", "b"), actions)
	if !errors.Is(err, ErrQuit) {
		t.Errorf("processKeys returned %v, want ErrQuit", err)
	}
	if want := []string{"a"}; !slices.Equal(log, want) {
		t.Errorf("actions ran for %q, want %q (b was pressed after q)", log, want)
	}
}

// termbox turns Ctrl+C into a normal key, so the engine has to handle it:
// it always quits, even if the game didn't assign it to anything.
func TestProcessKeysQuitsOnCtrlC(t *testing.T) {
	var log []string
	actions := recorder(&log, "a")

	err := process(t, pressed("ctrl+c", "a"), actions)
	if !errors.Is(err, ErrQuit) {
		t.Errorf("processKeys returned %v, want ErrQuit", err)
	}
	if len(log) != 0 {
		t.Errorf("actions ran for %q after ctrl+c, want none", log)
	}
}
