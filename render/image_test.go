package render_test

import (
	"image"
	"image/color"
	"math"
	"testing"

	"github.com/benjaestupinan/ascii_render/render"
)

// w×h image filled entirely with one color
func solidColorImage(w, h int, c color.Color) *image.NRGBA {
	img := image.NewNRGBA(image.Rect(0, 0, w, h))
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			img.Set(x, y, c)
		}
	}
	return img
}

// w×h image with the left half white and the right half black
func halvesImage(w, h int) *image.NRGBA {
	img := solidColorImage(w, h, color.Black)
	for y := 0; y < h; y++ {
		for x := 0; x < w/2; x++ {
			img.Set(x, y, color.White)
		}
	}
	return img
}

func almostEqual(a, b float32) bool {
	return math.Abs(float64(a-b)) < 0.01
}

// Each character covers a block (image width / width) pixels wide and 2.4
// times that tall, because a terminal character is about 2.4 times taller
// than it is wide. Incomplete rows at the bottom are dropped.
func TestFromImageComputesSize(t *testing.T) {
	cases := []struct {
		imgW, imgH, width int
		wantW, wantH      int
	}{
		{100, 50, 10, 10, 2},     // 10×24 blocks → 50/24 ≈ 2.08 → 2 rows
		{1920, 1080, 80, 80, 18}, // 24×57.6 blocks → 1080/57.6 = 18.75 → 18 rows
		{576, 1024, 60, 60, 44},  // vertical phone video: 9.6×23.04 blocks → 44.4 → 44 rows
	}
	for _, c := range cases {
		f := render.FromImage(solidColorImage(c.imgW, c.imgH, color.Black), c.width)
		if f.Width != c.wantW || f.Height != c.wantH {
			t.Errorf("%dx%d image with width %d → %dx%d frame, want %dx%d",
				c.imgW, c.imgH, c.width, f.Width, f.Height, c.wantW, c.wantH)
		}
	}
}

func TestFromImageWhiteAndBlack(t *testing.T) {
	white := render.FromImage(solidColorImage(40, 40, color.White), 4)
	for i, v := range white.Light {
		if !almostEqual(v, 1) {
			t.Errorf("white image: Light[%d] = %v, want 1", i, v)
		}
	}

	black := render.FromImage(solidColorImage(40, 40, color.Black), 4)
	for i, v := range black.Light {
		if !almostEqual(v, 0) {
			t.Errorf("black image: Light[%d] = %v, want 0", i, v)
		}
	}
}

// Light is the average of red, green and blue: (255 + 0 + 0) / 3 = 85 → 85/255
func TestFromImageAveragesColors(t *testing.T) {
	red := color.NRGBA{R: 255, G: 0, B: 0, A: 255}
	f := render.FromImage(solidColorImage(40, 40, red), 4)

	for i, v := range f.Light {
		if !almostEqual(v, 85.0/255) {
			t.Errorf("red image: Light[%d] = %v, want %v", i, v, 85.0/255)
		}
	}
}

// Checks that each cell averages ITS block of the image and not another one.
func TestFromImageEachCellReadsItsBlock(t *testing.T) {
	f := render.FromImage(halvesImage(40, 48), 4) // 10×24 blocks → 4×2 cells

	for y := 0; y < f.Height; y++ {
		for x := 0; x < f.Width; x++ {
			got := f.Light[y*f.Width+x]
			want := float32(0)
			if x < 2 {
				want = 1
			}
			if !almostEqual(got, want) {
				t.Errorf("cell (%d, %d) = %v, want %v", x, y, got, want)
			}
		}
	}
}

// This was the NaN bug: with more columns than pixels, some blocks were empty.
func TestFromImageWithMoreColumnsThanPixels(t *testing.T) {
	gray := color.NRGBA{R: 128, G: 128, B: 128, A: 255}
	f := render.FromImage(solidColorImage(4, 4, gray), 10)

	if f.Width != 10 {
		t.Fatalf("Width = %d, want 10", f.Width)
	}
	for i, v := range f.Light {
		if !almostEqual(v, 128.0/255) {
			t.Errorf("Light[%d] = %v, want %v", i, v, 128.0/255)
		}
	}
}

// An image doesn't always start at (0, 0): SubImage returns a piece of
// another image that keeps its original coordinates. Here the image goes
// from x=20 to x=40, so reading from x=0 would read outside of it.
func TestFromImageRespectsImageOrigin(t *testing.T) {
	img := solidColorImage(40, 40, color.Black)
	for y := 0; y < 40; y++ {
		for x := 20; x < 40; x++ {
			img.Set(x, y, color.White)
		}
	}
	rightHalf := img.SubImage(image.Rect(20, 0, 40, 40)) // all white

	f := render.FromImage(rightHalf, 2)
	for i, v := range f.Light {
		if !almostEqual(v, 1) {
			t.Errorf("Light[%d] = %v, want 1 (was it read from (0, 0) instead of Bounds().Min?)", i, v)
		}
	}
}

// A library can't choose what kind of image it gets: grayscale PNGs give
// *image.Gray, JPEGs give *image.YCbCr, etc.
func TestFromImageAcceptsAnyImageType(t *testing.T) {
	gray := image.NewGray(image.Rect(0, 0, 40, 40))
	for i := range gray.Pix {
		gray.Pix[i] = 128
	}
	rgba := image.NewRGBA(image.Rect(0, 0, 40, 40))
	for i := range rgba.Pix {
		rgba.Pix[i] = 128
		if i%4 == 3 {
			rgba.Pix[i] = 255 // opaque alpha
		}
	}

	cases := map[string]image.Image{"*image.Gray": gray, "*image.RGBA": rgba}
	for name, img := range cases {
		f := render.FromImage(img, 4)
		for i, v := range f.Light {
			if !almostEqual(v, 128.0/255) {
				t.Errorf("%s: Light[%d] = %v, want %v", name, i, v, 128.0/255)
				break
			}
		}
	}
}

// Not a test that fails: it measures how long FromImage takes. Run it with
//
//	go test -bench=. ./render
//
// For 30 fps video, each frame has 33 ms in total.
func BenchmarkFromImage1080p(b *testing.B) {
	img := halvesImage(1920, 1080)
	for b.Loop() {
		render.FromImage(img, 200)
	}
}
