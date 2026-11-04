package render_test

import (
	"strings"
	"testing"

	"github.com/benjaestupinan/ascii_render/render"
)

// same ramp the player uses today, from darkest to brightest
const ramp = ".'`^,:;Il!i><~+_-?][}{1)(|/tfjrxnuvczXYUJCLQ0OZmwqpdbkhao*#MW&8%B@$"

func TestNewCreatesBlackFrame(t *testing.T) {
	f := render.NewFrame(4, 3)

	if f.Width != 4 || f.Height != 3 {
		t.Fatalf("size = %dx%d, want 4x3", f.Width, f.Height)
	}
	if len(f.Light) != 4*3 {
		t.Fatalf("len(Light) = %d, want %d", len(f.Light), 4*3)
	}
	for i, v := range f.Light {
		if v != 0 {
			t.Errorf("Light[%d] = %v, want 0", i, v)
		}
	}
}

// Light is stored row by row: cell (x, y) is at Light[y*Width + x]
func TestSetWritesRowByRow(t *testing.T) {
	f := render.NewFrame(4, 3)
	f.Set(1, 2, 0.5)

	if got := f.Light[2*4+1]; got != 0.5 {
		t.Errorf("Light[2*4+1] = %v, want 0.5", got)
	}
}

func TestSetClampsLightBetween0And1(t *testing.T) {
	f := render.NewFrame(2, 1)
	f.Set(0, 0, 1.7)
	f.Set(1, 0, -0.3)

	if f.Light[0] != 1 {
		t.Errorf("Set(1.7) stored %v, want 1", f.Light[0])
	}
	if f.Light[1] != 0 {
		t.Errorf("Set(-0.3) stored %v, want 0", f.Light[1])
	}
}

// The rasterizer will draw triangles that go off-screen, so writing outside
// the frame has to be ignored without panicking.
func TestSetOutsideFrameIsIgnored(t *testing.T) {
	f := render.NewFrame(4, 3)
	// (-1, 1) and (4, 1) land inside Light if only the index is checked
	outside := [][2]int{{-1, 0}, {0, -1}, {4, 0}, {0, 3}, {100, 100}, {-1, 1}, {4, 1}}
	for _, p := range outside {
		f.Set(p[0], p[1], 1)
	}

	for i, v := range f.Light {
		if v != 0 {
			t.Errorf("Light[%d] = %v, want 0 (Set outside the frame wrote inside it)", i, v)
		}
	}
}

func TestClearMakesEverythingBlack(t *testing.T) {
	f := render.NewFrame(3, 3)
	f.Set(0, 0, 1)
	f.Set(2, 2, 0.5)
	f.Clear()

	for i, v := range f.Light {
		if v != 0 {
			t.Errorf("Light[%d] = %v after Clear, want 0", i, v)
		}
	}
}

func TestStringHasOneLinePerRow(t *testing.T) {
	f := render.NewFrame(5, 3)
	lines := strings.Split(f.String(), "\n")

	// every row ends in "\n", so Split leaves a "" at the end
	if len(lines) != 3+1 || lines[3] != "" {
		t.Fatalf("String() = %q, want 3 lines ending in \\n", f.String())
	}
	for i, l := range lines[:3] {
		if len(l) != 5 {
			t.Errorf("line %d has %d characters, want 5: %q", i, len(l), l)
		}
	}
}

func TestStringBlackIsFirstCharAndWhiteIsLast(t *testing.T) {
	f := render.NewFrame(2, 1)
	f.Set(0, 0, 0)
	f.Set(1, 0, 1)
	s := f.String()

	if s[0] != ramp[0] {
		t.Errorf("black was drawn as %q, want %q", s[0], ramp[0])
	}
	if s[1] != ramp[len(ramp)-1] {
		t.Errorf("white was drawn as %q, want %q", s[1], ramp[len(ramp)-1])
	}
}

// Goes through 256 light levels: more light can never give a darker
// character, and every character of the ramp must be able to appear.
func TestStringUsesWholeRampInOrder(t *testing.T) {
	f := render.NewFrame(256, 1)
	for x := 0; x < 256; x++ {
		f.Set(x, 0, float32(x)/255)
	}
	s := strings.TrimSuffix(f.String(), "\n")

	used := map[byte]bool{}
	previous := 0
	for x := 0; x < len(s); x++ {
		i := strings.IndexByte(ramp, s[x])
		if i < 0 {
			t.Fatalf("light %d/255 gave %q, which is not in the ramp", x, s[x])
		}
		if i < previous {
			t.Errorf("light %d/255 gave %q, which is darker than the previous %q", x, s[x], ramp[previous])
		}
		previous = i
		used[s[x]] = true
	}
	for i := 0; i < len(ramp); i++ {
		if !used[ramp[i]] {
			t.Errorf("ramp character %q was never used", ramp[i])
		}
	}
}
