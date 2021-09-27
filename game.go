package main

import (
	"image/color"
	"strconv"
	"time"

	"github.com/gen2brain/dlgs"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/jaeg/go-wireworld/ww"
)

const Title = "Wireworld - "

const ScreenWidth = 640
const ScreenHeight = 480

const GUIWidth = 0

var TileSize = 16.0

//Cell Colors
var WireColor = color.RGBA{R: 0, G: 0, B: 100, A: 255}
var HeadColor = color.RGBA{R: 100, G: 0, B: 0, A: 255}
var TailColor = color.RGBA{R: 0, G: 100, B: 0, A: 255}
var DeadColor = color.RGBA{R: 100, G: 100, B: 100, A: 255}

var CursorColor = color.RGBA{R: 100, G: 0, B: 0, A: 100}
var CursorSelectColor = color.RGBA{R: 0, G: 100, B: 0, A: 100}
var CursorSelectedColor = color.RGBA{R: 0, G: 255, B: 0, A: 10}
var CursorPasteColor = color.RGBA{R: 100, G: 0, B: 0, A: 10}

type CursorMode int

const (
	CursorChange CursorMode = iota
	CursorSelect
	CursorPaste
)

type Game struct {
	title            string
	ww               *ww.WireWorld
	SecondDelay      time.Duration
	StepMode         bool
	LastUpdated      time.Time
	ScrollX          int
	ScrollY          int
	SimulationWidth  int
	SimulationHeight int
	CursorMode       CursorMode
	SelectStartX     int
	SelectStartY     int
	SelectEndX       int
	SelectEndY       int
}

func NewGame(width int, height int) (*Game, error) {
	g := &Game{title: Title + "Editing", ww: ww.CreateNewWireWorld(width, height), SecondDelay: time.Second / 2, LastUpdated: time.Now(), CursorMode: CursorChange}

	ebiten.SetWindowResizable(true)
	ebiten.SetWindowSize(ScreenWidth, ScreenHeight)
	return g, nil
}

func (g *Game) Run() error {
	err := ebiten.RunGame(g)
	return err
}

func (g *Game) Update() error {
	//Adjust Tile Logic
	cX, cY := getTileUnderMouse()
	screenCellX := cX + g.ScrollX
	screenCellY := cY + g.ScrollY
	if screenCellX < g.ww.NumberOfTilesWidth && screenCellY < g.ww.NumberOfTilesHeight {
		if g.CursorMode == CursorChange {
			if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
				if g.ww.Running {
					g.SetRunning(false)
				}

				switch g.ww.GetEditorTileAt(screenCellX, screenCellY) {
				case ww.Dead:
					g.ww.SetEditorTile(screenCellX, screenCellY, ww.Dead)
				case ww.Wire:
					g.ww.SetEditorTile(screenCellX, screenCellY, ww.Head)
				case ww.Head:
					g.ww.SetEditorTile(screenCellX, screenCellY, ww.Tail)
				case ww.Tail:
					g.ww.SetEditorTile(screenCellX, screenCellY, ww.Wire)
				}
			} else if ebiten.IsMouseButtonPressed(ebiten.MouseButtonLeft) { //Allow for drag adding of wires.
				if g.ww.Running {
					g.SetRunning(false)
				}

				if g.ww.GetEditorTileAt(screenCellX, screenCellY) == ww.Dead {
					g.ww.SetEditorTile(screenCellX, screenCellY, ww.Wire)
				}
			}

			//Kill cell
			if ebiten.IsMouseButtonPressed(ebiten.MouseButtonRight) {
				if g.ww.Running {
					g.SetRunning(false)
				}
				g.ww.SetEditorTile(screenCellX, screenCellY, ww.Dead)
			}
		} else if g.CursorMode == CursorSelect {
			if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
				if g.SelectStartX == -1 || g.SelectEndX != -1 {
					g.SelectStartX = screenCellX
					g.SelectStartY = screenCellY
					g.SelectEndX = -1
					g.SelectEndY = -1
				}
			}

			if inpututil.IsMouseButtonJustReleased(ebiten.MouseButtonLeft) {
				if g.SelectStartX != -1 {
					g.SelectEndX = screenCellX
					g.SelectEndY = screenCellY
				}
			}

			if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonRight) {
				g.SelectStartX = -1
				g.SelectStartY = -1

				g.SelectEndX = -1
				g.SelectEndY = -1
			}

		}
	}

	//Pause/Run simulation
	if inpututil.IsKeyJustPressed(ebiten.KeySpace) {
		g.SetRunning(!g.ww.Running)
	}

	//Speed settings
	if inpututil.IsKeyJustPressed(ebiten.Key1) {
		g.SecondDelay = time.Millisecond
	}

	if inpututil.IsKeyJustPressed(ebiten.Key2) {
		g.SecondDelay = time.Second / 4
	}

	if inpututil.IsKeyJustPressed(ebiten.Key3) {
		g.SecondDelay = time.Second / 2
	}

	if inpututil.IsKeyJustPressed(ebiten.Key4) {
		g.SecondDelay = time.Second
	}

	if inpututil.IsKeyJustPressed(ebiten.Key5) {
		g.SecondDelay = time.Second * 2
	}

	//Tile size settings
	if inpututil.IsKeyJustPressed(ebiten.KeyMinus) {
		if TileSize > 8 {
			TileSize = TileSize / 2
		}
	}

	if inpututil.IsKeyJustPressed(ebiten.KeyEqual) {
		if TileSize < 62 {
			TileSize = TileSize * 2
		}
	}

	if ebiten.IsKeyPressed(ebiten.KeyControl) {
		// Loading tools
		if inpututil.IsKeyJustPressed(ebiten.KeyS) {
			filename, success, err := dlgs.Entry("Save world", "Enter name of file", "save.ww")
			if err != nil {
				panic(err)
			}
			if success {
				g.ww.SaveWorld(filename)
			}
		}

		if inpututil.IsKeyJustPressed(ebiten.KeyO) {
			filename, success, err := dlgs.File("Load world", "*.ww", false)
			if err != nil {
				panic(err)
			}
			if success {
				g.ww.LoadWorld(filename)
			}
		}

		//Copy selection
		if inpututil.IsKeyJustPressed(ebiten.KeyC) {
			if g.SelectStartX != -1 && g.SelectStartY != -1 && g.SelectEndX != -1 && g.SelectEndY != -1 {
				g.ww.CopyToBuffer(g.SelectStartX, g.SelectStartY, g.SelectEndX, g.SelectEndY)
			}
		}

		if inpututil.IsKeyJustPressed(ebiten.KeyV) {
			bufferWidth, _ := g.ww.GetCopyBufferDimensions()
			if bufferWidth > 0 {
				g.ww.PasteFromBuffer(screenCellX, screenCellY)
			}
		}
	} else {
		// Move view controls
		if ebiten.IsKeyPressed(ebiten.KeyW) {
			if g.ScrollY > 0 {
				g.ScrollY--
			}
		}

		if ebiten.IsKeyPressed(ebiten.KeyS) {
			if g.ScrollY < g.ww.NumberOfTilesHeight {
				g.ScrollY++
			}
		}

		if ebiten.IsKeyPressed(ebiten.KeyA) {
			if g.ScrollX > 0 {
				g.ScrollX--
			}
		}

		if ebiten.IsKeyPressed(ebiten.KeyD) {
			if g.ScrollX < g.ww.NumberOfTilesWidth {
				g.ScrollX++
			}
		}
	}

	//delete logic
	if inpututil.IsKeyJustPressed(ebiten.KeyDelete) {
		if g.SelectStartX != -1 && g.SelectStartY != -1 && g.SelectEndX != -1 && g.SelectEndY != -1 {
			for x := g.SelectStartX; x <= g.SelectEndX; x++ {
				for y := g.SelectStartY; y <= g.SelectEndY; y++ {
					g.ww.SetEditorTile(screenCellX, screenCellY, ww.Dead)
				}
			}
			//Reset selection
			g.SelectStartX = -1
			g.SelectStartY = -1

			g.SelectEndX = -1
			g.SelectEndY = -1
		}
	}

	if ebiten.IsKeyPressed(ebiten.KeyShift) {
		if g.CursorMode != CursorSelect {
			//Reset selection if there was one partionally started
			if g.SelectStartX != -1 && g.SelectEndX == -1 {
				g.SelectStartX = -1
				g.SelectStartY = -1
				g.SelectEndX = -1
				g.SelectEndY = -1
			}
			g.CursorMode = CursorSelect
		}
	} else {
		g.CursorMode = CursorChange
	}

	if time.Now().After(g.LastUpdated.Add(g.SecondDelay)) && g.ww.Running {
		g.ww.UpdateSimulation()
		g.LastUpdated = time.Now()
	}
	return nil
}

func (g *Game) Draw(screen *ebiten.Image) {
	if !g.ww.Running {
		cX, cY := getTileUnderMouse()

		ebiten.SetWindowTitle(g.title + " FPS: " + strconv.Itoa(int(ebiten.CurrentFPS())) + " X: " + strconv.Itoa(g.ScrollX+cX) + " Y: " + strconv.Itoa(g.ScrollY+cY))
	} else {
		ebiten.SetWindowTitle(g.title + " FPS: " + strconv.Itoa(int(ebiten.CurrentFPS())) + " : " + strconv.Itoa(g.ww.Tick))
	}
	g.DrawWorldArray(screen)

}

func (g *Game) Layout(outsideWidth, outsideHeight int) (int, int) {
	g.SimulationWidth = outsideWidth - GUIWidth
	g.SimulationHeight = outsideHeight
	return outsideWidth, outsideHeight
}

func (g *Game) SetRunning(status bool) {
	g.ww.SetRunning(status)
	if status {
		g.title = Title + "Running"
	} else {
		g.title = Title + "Editing"
	}
}

func (g *Game) DrawWorldArray(screen *ebiten.Image) {
	width := float64(g.SimulationWidth) / TileSize
	height := float64(g.SimulationHeight) / TileSize

	for x := 0.0; x < width; x++ {
		for y := 0.0; y < height; y++ {
			screenX := x * TileSize
			screenY := y * TileSize

			cellX := g.ScrollX + int(x)
			cellY := g.ScrollY + int(y)

			c := color.RGBA{R: 0, G: 0, B: 0, A: 255}

			if cellX >= 0 && cellX < g.ww.NumberOfTilesWidth && cellY >= 0 && cellY < g.ww.NumberOfTilesHeight {
				var cell ww.CellType
				if g.ww.Running {
					cell = g.ww.GetRunningTileAt(cellX, cellY)
				} else {
					cell = g.ww.GetEditorTileAt(cellX, cellY)
				}
				switch cell {
				case ww.Wire:
					c = WireColor
				case ww.Head:
					c = HeadColor
				case ww.Tail:
					c = TailColor
				case ww.Dead:
					c = DeadColor
				}

			}

			ebitenutil.DrawRect(screen, screenX, screenY, TileSize-1, TileSize-1, c)

			//Highlight cell - note that it's using cellX and cellY not the screen coords
			cX, cY := getTileUnderMouse()
			if cX == int(x) && cY == int(y) {
				c := CursorColor
				if g.CursorMode == CursorSelect {
					c = CursorSelectColor
				} else if g.CursorMode == CursorPaste {
					c = CursorPasteColor
				}
				ebitenutil.DrawRect(screen, screenX, screenY, TileSize-1, TileSize-1, c)
			}

			//Render selection
			if g.CursorMode == CursorSelect || (g.SelectStartX != -1 && g.SelectStartY != -1 && g.SelectEndX != -1 && g.SelectEndY != -1) {
				//Highlight cell if its selected
				if cellX >= g.SelectStartX && cellX <= g.SelectEndX && cellY >= g.SelectStartY && cellY <= g.SelectEndY {
					ebitenutil.DrawRect(screen, screenX, screenY, TileSize-1, TileSize-1, CursorSelectedColor)
				}

				if g.SelectStartX != -1 && g.SelectEndX == -1 {
					if cellX >= g.SelectStartX+g.ScrollX && cellX <= cX+g.ScrollX && cellY >= g.SelectStartY && cellY <= cY+g.ScrollY {
						ebitenutil.DrawRect(screen, screenX, screenY, TileSize-1, TileSize-1, CursorSelectedColor)
					}
				}
			}

			//Render Paste
			bufferWidth, bufferHeight := g.ww.GetCopyBufferDimensions()
			if ebiten.IsKeyPressed(ebiten.KeyControl) && bufferWidth > 0 {
				if cellX >= cX+g.ScrollX && cellX < bufferWidth+cX+g.ScrollX && cellY >= cY+g.ScrollY && cellY < bufferHeight+cY+g.ScrollY {
					c = CursorPasteColor
					switch g.ww.GetCopyBufferAt(cellX-cX, cellY-cY) {
					case ww.Wire:
						c = WireColor
					case ww.Head:
						c = HeadColor
					case ww.Tail:
						c = TailColor
					}

					c.A = 100

					ebitenutil.DrawRect(screen, screenX, screenY, TileSize-1, TileSize-1, c)

				}
			}

		}
	}
}

func getTileUnderMouse() (int, int) {
	cX, cY := ebiten.CursorPosition()
	temp := int(TileSize)
	return cX / temp, cY / temp
}
