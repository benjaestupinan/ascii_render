// A ball that falls by gravity, bounces off the floor and the walls, and
// loses energy until it comes to rest.
package main

import (
	"fmt"
	"math"

	"github.com/benjaestupinan/ascii_render/engine"
	"github.com/benjaestupinan/ascii_render/render"
)

// The world is measured in meters, not cells: the physics knows nothing about
// the screen. Draw takes care of converting meters to cells.
const (
	worldWidth  = 16.0 // meters
	worldHeight = 9.0  // meters

	gravity     = 9.8 // m/s², downwards
	restitution = 0.9 // fraction of the speed left after a bounce
	radius      = 0.8 // meters

	// terminal cells are 2.4 times taller than they are wide
	cellAspect = 2.4
)

type scene struct {
	// position of the center of the ball, with y = 0 at the floor and going up
	x, y   float32
	vx, vy float32 // velocity in m/s

	angle float32 // rotation of the ball, in radians

	stillTime float32 // seconds it has been without bouncing
}

func (s *scene) Update(dt float32) error {
	s.vy -= gravity * dt
	s.x += s.vx * dt
	s.y += s.vy * dt

	// spins as if rolling without slipping: vx / radius radians per second.
	// Moving right → spins clockwise → the angle decreases.
	s.angle -= s.vx / radius * dt

	// floor: if the ball goes through it, it is moved back to the surface and
	// its vertical velocity is reversed, losing energy
	if s.y-radius < 0 {
		s.y = radius
		s.vy = -s.vy * restitution
		s.vx *= 0.98 // a bit of friction
	}

	if s.y+radius > worldHeight {
		s.y = worldHeight - radius
		s.vy = -s.vy * restitution
	}

	// walls
	if s.x-radius < 0 {
		s.x = radius
		s.vx = -s.vx * restitution
	}
	if s.x+radius > worldWidth {
		s.x = worldWidth - radius
		s.vx = -s.vx * restitution
	}

	// the scene ends when the ball has barely moved for 1 second
	if s.y-radius < 0.01 && abs(s.vy) < 0.5 && abs(s.vx) < 0.5 {
		s.stillTime += dt
	} else {
		s.stillTime = 0
	}
	if s.stillTime > 1 {
		return engine.ErrQuit
	}
	return nil
}

func (s *scene) Draw(f *render.Frame) {
	// How many cells one meter takes. A cell is 2.4 times taller than it is
	// wide, so one meter takes 2.4 times more columns than rows.
	// The largest scale that fits the whole world in the frame is used.
	rowsPerMeter := min(
		float32(f.Height)/worldHeight,
		float32(f.Width)/(worldWidth*cellAspect),
	)
	colsPerMeter := rowsPerMeter * cellAspect

	// the world is drawn horizontally centered
	offsetCol := (float32(f.Width) - worldWidth*colsPerMeter) / 2

	// converts a cell (col, row) to the position of its center in meters
	toWorld := func(col, row int) (float32, float32) {
		wx := (float32(col) + 0.5 - offsetCol) / colsPerMeter
		wy := worldHeight - (float32(row)+0.5)/rowsPerMeter
		return wx, wy
	}

	// the brightest point spins around the center according to s.angle
	// (with angle = 0 it is on top): this makes the ball look round and "spin"
	hx := s.x - 0.5*radius*float32(math.Sin(float64(s.angle)))
	hy := s.y + 0.5*radius*float32(math.Cos(float64(s.angle)))

	for row := range f.Height {
		for col := range f.Width {
			wx, wy := toWorld(col, row)
			if wx < 0 || wx > worldWidth {
				continue // outside the world
			}

			// floor: only the first row below y = 0
			if wy < 0 {
				if wy > -1/rowsPerMeter {
					f.Set(col, row, 0.3)
				}
				continue
			}

			// ball: cells less than `radius` meters away from the center
			if dist(wx, wy, s.x, s.y) < radius {
				light := 1 - dist(wx, wy, hx, hy)/(radius*1.6)
				f.Set(col, row, max(light, 0.15))
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
	ball := &scene{
		x:  2,
		y:  worldHeight - radius,
		vx: 3,
	}
	if err := engine.Run(ball); err != nil {
		fmt.Println("error running the engine:", err)
	}
}
