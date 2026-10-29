package render_test

import (
	"strings"
	"testing"

	"github.com/benjaestupinan/ascii_render/render"
)

// misma rampa que usa el player hoy, de más oscuro a más claro
const ramp = ".'`^,:;Il!i><~+_-?][}{1)(|/tfjrxnuvczXYUJCLQ0OZmwqpdbkhao*#MW&8%B@$"

func TestNewCreaUnFrameNegro(t *testing.T) {
	f := render.NewFrame(4, 3)

	if f.Width != 4 || f.Height != 3 {
		t.Fatalf("tamaño = %dx%d, se esperaba 4x3", f.Width, f.Height)
	}
	if len(f.Light) != 4*3 {
		t.Fatalf("len(Light) = %d, se esperaba %d", len(f.Light), 4*3)
	}
	for i, v := range f.Light {
		if v != 0 {
			t.Errorf("Light[%d] = %v, se esperaba 0", i, v)
		}
	}
}

// Light se guarda fila por fila: la celda (x, y) está en Light[y*Width + x]
func TestSetEscribeFilaPorFila(t *testing.T) {
	f := render.NewFrame(4, 3)
	f.Set(1, 2, 0.5)

	if got := f.Light[2*4+1]; got != 0.5 {
		t.Errorf("Light[2*4+1] = %v, se esperaba 0.5", got)
	}
}

func TestSetLimitaElBrilloEntre0y1(t *testing.T) {
	f := render.NewFrame(2, 1)
	f.Set(0, 0, 1.7)
	f.Set(1, 0, -0.3)

	if f.Light[0] != 1 {
		t.Errorf("Set(1.7) guardó %v, se esperaba 1", f.Light[0])
	}
	if f.Light[1] != 0 {
		t.Errorf("Set(-0.3) guardó %v, se esperaba 0", f.Light[1])
	}
}

// El rasterizador va a dibujar triángulos que se salen de la pantalla,
// así que escribir fuera del frame tiene que ignorarse sin hacer panic.
func TestSetFueraDelFrameSeIgnora(t *testing.T) {
	f := render.NewFrame(4, 3)
	// (-1, 1) y (4, 1) caen dentro de Light si solo se revisa el índice
	fuera := [][2]int{{-1, 0}, {0, -1}, {4, 0}, {0, 3}, {100, 100}, {-1, 1}, {4, 1}}
	for _, p := range fuera {
		f.Set(p[0], p[1], 1)
	}

	for i, v := range f.Light {
		if v != 0 {
			t.Errorf("Light[%d] = %v, se esperaba 0 (Set fuera del frame escribió adentro)", i, v)
		}
	}
}

func TestClearDejaTodoNegro(t *testing.T) {
	f := render.NewFrame(3, 3)
	f.Set(0, 0, 1)
	f.Set(2, 2, 0.5)
	f.Clear()

	for i, v := range f.Light {
		if v != 0 {
			t.Errorf("Light[%d] = %v después de Clear, se esperaba 0", i, v)
		}
	}
}

func TestStringTieneUnaLineaPorFila(t *testing.T) {
	f := render.NewFrame(5, 3)
	lineas := strings.Split(f.String(), "\n")

	// cada fila termina en "\n", así que Split deja un "" al final
	if len(lineas) != 3+1 || lineas[3] != "" {
		t.Fatalf("String() = %q, se esperaban 3 líneas terminadas en \\n", f.String())
	}
	for i, l := range lineas[:3] {
		if len(l) != 5 {
			t.Errorf("línea %d tiene %d caracteres, se esperaban 5: %q", i, len(l), l)
		}
	}
}

func TestStringNegroEsPrimerCaracterYBlancoEsUltimo(t *testing.T) {
	f := render.NewFrame(2, 1)
	f.Set(0, 0, 0)
	f.Set(1, 0, 1)
	s := f.String()

	if s[0] != ramp[0] {
		t.Errorf("negro se dibujó como %q, se esperaba %q", s[0], ramp[0])
	}
	if s[1] != ramp[len(ramp)-1] {
		t.Errorf("blanco se dibujó como %q, se esperaba %q", s[1], ramp[len(ramp)-1])
	}
}

// Recorre 256 niveles de brillo: más brillo nunca puede dar un carácter más
// oscuro, y todos los caracteres de la rampa tienen que poder aparecer.
func TestStringUsaLaRampaCompletaEnOrden(t *testing.T) {
	f := render.NewFrame(256, 1)
	for x := 0; x < 256; x++ {
		f.Set(x, 0, float32(x)/255)
	}
	s := strings.TrimSuffix(f.String(), "\n")

	usados := map[byte]bool{}
	anterior := 0
	for x := 0; x < len(s); x++ {
		i := strings.IndexByte(ramp, s[x])
		if i < 0 {
			t.Fatalf("brillo %d/255 dio %q, que no está en la rampa", x, s[x])
		}
		if i < anterior {
			t.Errorf("brillo %d/255 dio %q, que es más oscuro que el anterior %q", x, s[x], ramp[anterior])
		}
		anterior = i
		usados[s[x]] = true
	}
	for i := 0; i < len(ramp); i++ {
		if !usados[ramp[i]] {
			t.Errorf("el carácter %q de la rampa nunca se usó", ramp[i])
		}
	}
}
