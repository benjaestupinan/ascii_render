package render_test

import (
	"image"
	"image/color"
	"math"
	"testing"

	"github.com/benjaestupinan/ascii_render/render"
)

// imagen de w×h pintada entera de un color
func imagenDeColor(w, h int, c color.Color) *image.NRGBA {
	img := image.NewNRGBA(image.Rect(0, 0, w, h))
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			img.Set(x, y, c)
		}
	}
	return img
}

// imagen de w×h con la mitad izquierda blanca y la derecha negra
func imagenMitades(w, h int) *image.NRGBA {
	img := imagenDeColor(w, h, color.Black)
	for y := 0; y < h; y++ {
		for x := 0; x < w/2; x++ {
			img.Set(x, y, color.White)
		}
	}
	return img
}

func casiIgual(a, b float32) bool {
	return math.Abs(float64(a-b)) < 0.01
}

// Cada carácter cubre un bloque de (ancho/width) píxeles de ancho y 2.4 veces
// eso de alto, porque un carácter de la terminal es unas 2.4 veces más alto
// que ancho. Las filas incompletas al final se descartan.
func TestFromImageCalculaElTamano(t *testing.T) {
	casos := []struct {
		imgW, imgH, width int
		wantW, wantH      int
	}{
		{100, 50, 10, 10, 2},     // bloques de 10×24 → 50/24 ≈ 2.08 → 2 filas
		{1920, 1080, 80, 80, 18}, // bloques de 24×57.6 → 1080/57.6 = 18.75 → 18 filas
		{576, 1024, 60, 60, 44},  // tu video vertical: bloques de 9.6×23.04 → 44.4 → 44 filas
	}
	for _, c := range casos {
		f := render.FromImage(imagenDeColor(c.imgW, c.imgH, color.Black), c.width)
		if f.Width != c.wantW || f.Height != c.wantH {
			t.Errorf("imagen %dx%d con width %d → frame %dx%d, se esperaba %dx%d",
				c.imgW, c.imgH, c.width, f.Width, f.Height, c.wantW, c.wantH)
		}
	}
}

func TestFromImageBlancoYNegro(t *testing.T) {
	blanco := render.FromImage(imagenDeColor(40, 40, color.White), 4)
	for i, v := range blanco.Light {
		if !casiIgual(v, 1) {
			t.Errorf("imagen blanca: Light[%d] = %v, se esperaba 1", i, v)
		}
	}

	negro := render.FromImage(imagenDeColor(40, 40, color.Black), 4)
	for i, v := range negro.Light {
		if !casiIgual(v, 0) {
			t.Errorf("imagen negra: Light[%d] = %v, se esperaba 0", i, v)
		}
	}
}

// El brillo es el promedio de rojo, verde y azul: (255 + 0 + 0) / 3 = 85 → 85/255
func TestFromImagePromediaLosColores(t *testing.T) {
	rojo := color.NRGBA{R: 255, G: 0, B: 0, A: 255}
	f := render.FromImage(imagenDeColor(40, 40, rojo), 4)

	for i, v := range f.Light {
		if !casiIgual(v, 85.0/255) {
			t.Errorf("imagen roja: Light[%d] = %v, se esperaba %v", i, v, 85.0/255)
		}
	}
}

// Comprueba que cada celda promedia SU bloque de la imagen y no otro.
func TestFromImageCadaCeldaMiraSuBloque(t *testing.T) {
	f := render.FromImage(imagenMitades(40, 48), 4) // bloques de 10×24 → 4×2 celdas

	for y := 0; y < f.Height; y++ {
		for x := 0; x < f.Width; x++ {
			got := f.Light[y*f.Width+x]
			want := float32(0)
			if x < 2 {
				want = 1
			}
			if !casiIgual(got, want) {
				t.Errorf("celda (%d, %d) = %v, se esperaba %v", x, y, got, want)
			}
		}
	}
}

// Era el bug del NaN: con más columnas que píxeles, había bloques vacíos.
func TestFromImageConMasColumnasQuePixeles(t *testing.T) {
	gris := color.NRGBA{R: 128, G: 128, B: 128, A: 255}
	f := render.FromImage(imagenDeColor(4, 4, gris), 10)

	if f.Width != 10 {
		t.Fatalf("Width = %d, se esperaba 10", f.Width)
	}
	for i, v := range f.Light {
		if !casiIgual(v, 128.0/255) {
			t.Errorf("Light[%d] = %v, se esperaba %v", i, v, 128.0/255)
		}
	}
}

// Una imagen no siempre empieza en (0, 0): SubImage devuelve un pedazo de
// otra imagen que conserva sus coordenadas originales. Aquí la imagen va de
// x=20 a x=40, así que leer desde x=0 sería leer fuera de ella.
func TestFromImageRespetaElOrigenDeLaImagen(t *testing.T) {
	img := imagenDeColor(40, 40, color.Black)
	for y := 0; y < 40; y++ {
		for x := 20; x < 40; x++ {
			img.Set(x, y, color.White)
		}
	}
	mitadDerecha := img.SubImage(image.Rect(20, 0, 40, 40)) // toda blanca

	f := render.FromImage(mitadDerecha, 2)
	for i, v := range f.Light {
		if !casiIgual(v, 1) {
			t.Errorf("Light[%d] = %v, se esperaba 1 (¿se leyó desde (0, 0) en vez de Bounds().Min?)", i, v)
		}
	}
}

// Una librería no puede elegir qué tipo de imagen le pasan: PNG en escala de
// grises da *image.Gray, JPEG da *image.YCbCr, etc.
func TestFromImageAceptaCualquierTipoDeImagen(t *testing.T) {
	gris := image.NewGray(image.Rect(0, 0, 40, 40))
	for i := range gris.Pix {
		gris.Pix[i] = 128
	}
	rgba := image.NewRGBA(image.Rect(0, 0, 40, 40))
	for i := range rgba.Pix {
		rgba.Pix[i] = 128
		if i%4 == 3 {
			rgba.Pix[i] = 255 // alfa opaco
		}
	}

	casos := map[string]image.Image{"*image.Gray": gris, "*image.RGBA": rgba}
	for nombre, img := range casos {
		f := render.FromImage(img, 4)
		for i, v := range f.Light {
			if !casiIgual(v, 128.0/255) {
				t.Errorf("%s: Light[%d] = %v, se esperaba %v", nombre, i, v, 128.0/255)
				break
			}
		}
	}
}

// No es un test que falle: mide cuánto tarda. Córrelo con
//
//	go test -bench=. ./render
//
// Para video a 30 fps, cada frame tiene 33 ms en total.
func BenchmarkFromImage1080p(b *testing.B) {
	img := imagenMitades(1920, 1080)
	for b.Loop() {
		render.FromImage(img, 200)
	}
}
