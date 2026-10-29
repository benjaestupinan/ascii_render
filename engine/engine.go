// Package engine corre un game loop que dibuja un Game en la terminal.
package engine

import (
	"errors"
	"fmt"
	"time"

	"github.com/benjaestupinan/ascii_render/render"
)

// Game es una escena que engine.Run anima. En cada tic, Run llama a Update
// y después a Draw.
type Game interface {
	// Update avanza la escena dt segundos. Devuelve ErrQuit para terminar
	// normalmente, o cualquier otro error para detener Run con ese error.
	Update(dt float32) error

	// Draw dibuja el estado actual de la escena en f, que llega en negro.
	Draw(f *render.Frame)
}

// ErrQuit lo devuelve Game.Update para terminar el game loop. Run lo trata
// como una salida normal y devuelve nil.
var ErrQuit = errors.New("engine: quit")

// Run toma la terminal y corre el game loop de g hasta que Update devuelva
// un error o el usuario presione Ctrl+C. Al salir, siempre deja la terminal
// como estaba. Devuelve nil si el juego terminó con ErrQuit o Ctrl+C.
func Run(g Game) error {
	frame := render.NewFrame(465, 109)
	ticker := time.NewTicker(time.Second / 30)
	defer ticker.Stop()

	fmt.Print("\033[H\033[2J")

	last := time.Now()
	for {
		<-ticker.C
		dt := time.Since(last)
		last = time.Now()

		err := g.Update(float32(dt.Seconds()))
		if err != nil {
			if errors.Is(err, ErrQuit) {
				return nil
			}
			return err
		}

		frame.Clear()

		g.Draw(frame)
		ascii := frame.String()

		fmt.Print("\033[H" + ascii)
	}
}
