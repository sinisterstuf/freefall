// Copyright 2021 Siôn le Roux.  All rights reserved.
// Use of this source code is subject to an MIT-style
// licence which can be found in the LICENSE file.

package main

import (
	"errors"
	"image"
	"log"
	"math/rand"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/sinisterstuf/freefall/nokia"
)

func main() {
	windowScale := 10
	ebiten.SetWindowSize(nokia.GameSize.X*windowScale, nokia.GameSize.Y*windowScale)
	ebiten.SetWindowTitle("Freefall")

	game := &Game{
		Size: nokia.GameSize,
		Player: &Player{
			Coords: image.Pt(nokia.GameSize.X/2, nokia.GameSize.Y/2),
		},
		Dusts: Dusts{},
	}

	if err := ebiten.RunGame(game); err != nil {
		log.Fatal(err)
	}
}

// Game represents the main game state
type Game struct {
	Size   image.Point
	Player *Player
	Dusts  Dusts
	Tick   int64
}

// Layout is hardcoded for now, may be made dynamic in future
func (g *Game) Layout(outsideWidth int, outsideHeight int) (screenWidth int, screenHeight int) {
	return g.Size.X, g.Size.Y
}

// Update calculates game logic
func (g *Game) Update() error {
	g.Tick++

	// Pressing Q any time quits immediately
	if ebiten.IsKeyPressed(ebiten.KeyQ) {
		return errors.New("game quit by player")
	}

	// Pressing F toggles full-screen
	if inpututil.IsKeyJustPressed(ebiten.KeyF) {
		if ebiten.IsFullscreen() {
			ebiten.SetFullscreen(false)
		} else {
			ebiten.SetFullscreen(true)
		}
	}

	if g.Player.Chute {
		if g.Tick%2 == 0 {
			g.Dusts.Update()
		}
	} else {
		g.Dusts.Update()
	}

	// Movement controls
	if inpututil.IsKeyJustPressed(ebiten.KeySpace) {
		g.Player.Pull()
	}

	return nil
}

// Draw draws the game screen by one frame
func (g *Game) Draw(screen *ebiten.Image) {
	screen.Fill(nokia.PaletteOriginal.Dark())

	for _, d := range g.Dusts {
		ebitenutil.DrawRect(
			screen,
			float64(d.Coords.X), float64(d.Coords.Y),
			1, 1,
			nokia.PaletteOriginal.Light(),
		)
	}

	ebitenutil.DrawRect(
		screen,
		float64(g.Player.Coords.X),
		float64(g.Player.Coords.Y),
		5,
		5,
		nokia.PaletteOriginal.Light(),
	)
}

// Player is the player character in the game
type Player struct {
	Coords image.Point
	Chute  bool
}

// Move moves the player upwards
func (p *Player) Pull() {
	p.Chute = !p.Chute
}

// Dust is decorative dirt on the screen to give the illusion of motion
type Dust struct {
	Coords image.Point
}

func (d *Dust) Update() {
	// Move dusts up
	log.Println(d.Coords.X, d.Coords.Y)
	d.Coords.Y--
	log.Println(d.Coords.X, d.Coords.Y)
}

type Dusts []*Dust

func (ds *Dusts) Update() {
	const maxDusts = 5

	if len(*ds) < maxDusts {
		dsX := rand.Intn(nokia.GameSize.X)
		*ds = append(*ds, &Dust{
			image.Pt(dsX, nokia.GameSize.Y+1),
		})
	}

	for i, d := range *ds {
		d.Update()
		if d.Coords.Y < 0 {
			*ds = append((*ds)[:i], (*ds)[i+1:]...)
		}
	}
}

type Entity interface {
	Update()
}
