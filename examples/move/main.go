// El ejemplo más simple de engine: un cometa cruza la pantalla de izquierda a
// derecha y la escena termina cuando sale por el otro lado.
package main

import (
	"fmt"

	"github.com/benjaestupinan/ascii_render/engine"
	"github.com/benjaestupinan/ascii_render/render"
)

const (
	duracion = 4.0 // segundos que tarda en cruzar la pantalla
	cola     = 12  // largo de la cola, en celdas
)

type scene struct {
	// posición como fracción del ancho de la pantalla: 0 es el borde
	// izquierdo y 1 el derecho. Así Update no necesita saber cuánto mide.
	x float32
}

func (s *scene) Update(dt float32) error {
	s.x += dt / duracion
	if s.x > 1 {
		return engine.ErrQuit
	}
	return nil
}

func (s *scene) Draw(f *render.Frame) {
	col := int(s.x * float32(f.Width))
	row := f.Height / 2

	// la cabeza es la celda más brillante y la cola se va apagando hacia
	// atrás. Las celdas fuera del frame las ignora Set.
	for i := range cola {
		brillo := 1 - float32(i)/cola
		f.Set(col-i, row, brillo)
	}
}

func main() {
	if err := engine.Run(&scene{}); err != nil {
		fmt.Println("error corriendo el motor:", err)
	}
}
