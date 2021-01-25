// Copyright ©2020 The go-p5 Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"flag"
	"fmt"
	"image/color"
	"math/rand"
	"time"

	"github.com/go-p5/p5"
)

func main() {
	rand.Seed(time.Now().Unix())

	flag.IntVar(&width, "width", width, "window width")
	flag.IntVar(&height, "height", height, "window height")
	flag.Float64Var(&size, "cell", size, "cell size")
	flag.Float64Var(&speed, "speed", speed, "initial snake speed")
	flag.IntVar(&start, "snake", start, "initial snake length")
	flag.IntVar(&foods, "food", foods, "number of food pieces")
	flag.Parse()

	p5.Run(setup, draw)
}

type Point struct {
	X, Y float64
}

func (p Point) String() string {
	return fmt.Sprintf("(%3.2f,%3.2f)", p.X, p.Y)
}

func (p Point) Move(x, y float64) Point {
	p.X += x
	p.Y += y
	return p
}

func (p Point) toRect(w, h float64) Rect {
	w /= 2
	h /= 2

	return Rect{
		Min: Point{X: p.X - w, Y: p.Y - h},
		Max: Point{X: p.X + w, Y: p.Y + h},
	}
}

func (p Point) toSquare(sz float64) Rect {
	cz := sz / 2

	return Rect{
		Min: Point{X: p.X - cz, Y: p.Y - cz},
		Max: Point{X: p.X + cz, Y: p.Y + cz},
	}
}

type Rect struct {
	Min, Max Point
}

func (r Rect) String() string {
	return fmt.Sprintf("%v-%v", r.Min, r.Max)
}

func (r Rect) Move(x, y float64) Rect {
	r.Min.X += x
	r.Max.X += x
	r.Min.Y += y
	r.Max.Y += y
	return r
}

func (r Rect) Contains(p Point) bool {
	return p.X >= r.Min.X && p.X < r.Max.X &&
		p.Y >= r.Min.Y && p.Y < r.Max.Y
}

func (r Rect) Overlaps(r2 Rect) bool {
	if r.Max.X <= r2.Min.X || r2.Max.X <= r.Min.X {
		return false
	}

	if r.Max.Y <= r2.Min.Y || r2.Max.Y <= r.Min.Y {
		return false
	}

	return true
}

type List []Point

func (l List) Overlaps(p Point, sz float64) int {
	r := p.toSquare(sz)

	for i, c := range l {
		if c.X == 0 && c.Y == 0 {
			continue
		}

		q := c.toSquare(sz)
		if q.Overlaps(r) {
			//fmt.Println(r, "overlaps", i, q)
			return i
		}
	}

	return -1
}

func (l List) OverlapsRect(r Rect, sz float64) int {
	for i, c := range l {
		if c.X == 0 && c.Y == 0 {
			continue
		}

		q := c.toSquare(sz)
		if q.Overlaps(r) {
			//fmt.Println(r, "overlaps", i, q)
			return i
		}
	}

	return -1
}

type Dir int

const (
	None Dir = iota
	Up
	Down
	Left
	Right

	Die
)

var (
	width  = 1800
	height = 1200
	size   = 50.0
	speed  = 5.0
	foods  = 10
	start  = 10

	snake List
	food  List

	dir Dir
	cur Point

	score int
	max   int

	foodColor  = color.RGBA{R: 255, A: 255}
	deadColor  = color.RGBA{R: 255, G: 165}
	snakeColor = color.RGBA{G: 255}
)

func makefood() {
	for i, c := range food {
		if c.X == 0 && c.Y == 0 {
			for i := 0; i < 1000; i++ {
				c.X = size + rand.Float64()*(p5.Event.Width-size-size)
				c.Y = size + rand.Float64()*(p5.Event.Height-size-size)

				if food.Overlaps(c, size) < 0 && snake.Overlaps(c, size) < 0 {
					break
				}
			}

			food[i] = c
		}

		// p5.Square uses top/left
		// we use center point
		c = c.Move(-size/2, -size/2)

		p5.Fill(foodColor)
		p5.Square(c.X, c.Y, size)
	}
}

func setup() {
	p5.Canvas(width, height)
	p5.Stroke(nil)
	p5.Background(color.Gray{Y: 220})

	dir = Up
	cur = Point{X: float64(width / 2), Y: 0}

	snake = make(List, start)
	food = make(List, foods)
}

func draw() {
	if p5.Event.Key.Pressed {
		switch p5.Event.Key.Name {
		case "↑":
			dir = Up

		case "↓":
			dir = Down

		case "←":
			dir = Left

		case "→":
			dir = Right

		case "-", "_":
			if speed > 1 {
				speed--
			}

		case "+", "=":
			speed++

		case "F":
			food = append(food, Point{})
		}
	}

	var head Rect

	switch dir {
	case Up:
		cur.Y -= speed
		if cur.Y < 0 {
			cur.Y = p5.Event.Height
		}

		head = cur.toRect(size, 1).Move(0, -size/2)

	case Down:
		cur.Y += speed
		if cur.Y > p5.Event.Height {
			cur.Y = 0
		}

		head = cur.toRect(size, 1).Move(0, size/2)

	case Left:
		cur.X -= speed
		if cur.X < 0 {
			cur.X = p5.Event.Width
		}

		head = cur.toRect(1, size).Move(-size/2, 0)

	case Right:
		cur.X += speed
		if cur.X > p5.Event.Width {
			cur.X = 0
		}

		head = cur.toRect(1, size).Move(size/2, 0)
	}

	//fmt.Println("dir", dir, "head", head)

	eat := false

	if i := food.OverlapsRect(head, size); i >= 0 {
		food[i] = Point{}
		eat = true

		score++
		if score > max {
			max = score
		}

		fmt.Println("score", score, max)
	}

	if i := snake.OverlapsRect(head, size); i >= 0 { // snake touches itself, snake dies
		dir = Die
		score = 0
	}

	makefood()

	if eat {
		snake = append(snake, cur)
	} else {
		for i := 1; i < len(snake); i++ {
			snake[i-1] = snake[i]
		}

		if dir != Die {
			snake[len(snake)-1].X = cur.X
			snake[len(snake)-1].Y = cur.Y
		}
	}

	col := snakeColor
	if dir == Die {
		col = deadColor
	}

	for i, c := range snake {
		col.A = uint8(i * 180 / len(snake))

		p5.Fill(col)
		p5.Ellipse(c.X, c.Y, size, size)
	}
}
