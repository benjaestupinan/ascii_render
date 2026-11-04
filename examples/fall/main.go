// Una pelota que cae por gravedad, rebota en el suelo y en las paredes, y
// va perdiendo energía hasta quedarse quieta.
package main

import (
	"fmt"
	"math"

	"github.com/benjaestupinan/ascii_render/engine"
	"github.com/benjaestupinan/ascii_render/render"
)

// El mundo se mide en metros, no en celdas: la física no sabe nada de la
// pantalla. Draw se encarga de convertir metros a celdas.
const (
	worldWidth  = 16.0 // metros
	worldHeight = 9.0  // metros

	gravity     = 9.8 // m/s², hacia abajo
	restitution = 0.9 // fracción de la velocidad que queda después de un rebote
	radius      = 0.8 // metros

	// las celdas de la terminal son 2.4 veces más altas que anchas
	cellAspect = 2.4
)

type scene struct {
	// posición del centro de la pelota, con y = 0 en el suelo y hacia arriba
	x, y   float32
	vx, vy float32 // velocidad en m/s

	angulo float32 // giro de la pelota, en radianes

	quieta float32 // segundos que lleva sin rebotar
}

func (s *scene) Update(dt float32) error {
	s.vy -= gravity * dt
	s.x += s.vx * dt
	s.y += s.vy * dt

	// gira como si rodara sin resbalar: vx / radius radianes por segundo.
	// Va hacia la derecha → gira en sentido horario → el ángulo disminuye.
	s.angulo -= s.vx / radius * dt

	// suelo: si la pelota lo atraviesa, se devuelve a la superficie y la
	// velocidad vertical se invierte perdiendo energía
	if s.y-radius < 0 {
		s.y = radius
		s.vy = -s.vy * restitution
		s.vx *= 0.98 // un poco de roce
	}

	if s.y+radius > worldHeight {
		s.y = worldHeight - radius
		s.vy = -s.vy * restitution
	}

	// paredes
	if s.x-radius < 0 {
		s.x = radius
		s.vx = -s.vx * restitution
	}
	if s.x+radius > worldWidth {
		s.x = worldWidth - radius
		s.vx = -s.vx * restitution
	}

	// la escena termina cuando la pelota lleva 1 segundo casi sin moverse
	if s.y-radius < 0.01 && abs(s.vy) < 0.5 && abs(s.vx) < 0.5 {
		s.quieta += dt
	} else {
		s.quieta = 0
	}
	if s.quieta > 1 {
		return engine.ErrQuit
	}
	return nil
}

func (s *scene) Draw(f *render.Frame) {
	// Cuántas celdas mide un metro. Una celda es 2.4 veces más alta que
	// ancha, así que un metro ocupa 2.4 veces más columnas que filas.
	// Se elige la escala más grande que hace caber el mundo en el frame.
	rowsPerMeter := min(
		float32(f.Height)/worldHeight,
		float32(f.Width)/(worldWidth*cellAspect),
	)
	colsPerMeter := rowsPerMeter * cellAspect

	// el mundo se dibuja centrado horizontalmente
	offsetCol := (float32(f.Width) - worldWidth*colsPerMeter) / 2

	// convierte una celda (col, row) a la posición en metros de su centro
	toWorld := func(col, row int) (float32, float32) {
		wx := (float32(col) + 0.5 - offsetCol) / colsPerMeter
		wy := worldHeight - (float32(row)+0.5)/rowsPerMeter
		return wx, wy
	}

	// el punto más brillante gira alrededor del centro según s.angulo
	// (con angulo = 0 está arriba): así se ve redonda y además "gira"
	hx := s.x - 0.5 * radius*float32(math.Sin(float64(s.angulo)))
	hy := s.y + 0.5 * radius*float32(math.Cos(float64(s.angulo)))

	for row := range f.Height {
		for col := range f.Width {
			wx, wy := toWorld(col, row)
			if wx < 0 || wx > worldWidth {
				continue // fuera del mundo
			}

			// suelo: solo la primera fila bajo y = 0
			if wy < 0 {
				if wy > -1/rowsPerMeter {
					f.Set(col, row, 0.3)
				}
				continue
			}

			// pelota: celdas a menos de `radius` metros del centro
			if dist(wx, wy, s.x, s.y) < radius {
				brillo := 1 - dist(wx, wy, hx, hy)/(radius*1.6)
				f.Set(col, row, max(brillo, 0.15))
			}
		}
	}
}

func dist(x1, y1, x2, y2 float32) float32 {
	dx, dy := x1-x2, y1-y2
	return float32(math.Sqrt(float64(dx*dx + dy*dy)))
}

func abs(v float32) float32 {
	return float32(math.Abs(float64(v)))
}

func main() {
	pelota := &scene{
		x:  2,
		y:  worldHeight - radius,
		vx: 3,
	}
	if err := engine.Run(pelota); err != nil {
		fmt.Println("error corriendo el motor:", err)
	}
}
