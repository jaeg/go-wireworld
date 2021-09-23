package main

import (
	"image/color"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
)

const TileSize = 16

type CellType int

const (
	Dead CellType = iota
	Wire
	Head
	Tail
)

type Game struct {
	world               [][]CellType
	runningWorld        [][]CellType
	ScreenWidth         int
	ScreenHeight        int
	NumberOfTilesWidth  int
	NumberOfTilesHeight int
	Running             bool
	Tick                int
	SecondDelay         time.Duration
	LastUpdated         time.Time
	ScrollX             int
	ScrollY             int
}

func NewGame(width int, height int, screenWidth int, screenHeight int) (*Game, error) {
	g := &Game{ScreenWidth: screenWidth, ScreenHeight: screenHeight, NumberOfTilesWidth: width, NumberOfTilesHeight: height, SecondDelay: time.Second / 2, LastUpdated: time.Now()}

	g.world = CreateWorldArray(width, height)
	return g, nil
}

func (g *Game) Run() error {
	err := ebiten.RunGame(g)
	return err
}

func (g *Game) Update() error {
	//Adjust Tile Logic
	cX, cY := getTileUnderMouse()
	if cX < g.NumberOfTilesWidth && cY < g.NumberOfTilesHeight {
		if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
			switch g.world[cX][cY] {
			case Dead:
				g.world[cX][cY] = Wire
			case Wire:
				g.world[cX][cY] = Head
			case Head:
				g.world[cX][cY] = Tail
			case Tail:
				g.world[cX][cY] = Wire
			}

		}

		if ebiten.IsMouseButtonPressed(ebiten.MouseButtonRight) {
			g.Running = false

			g.world[cX][cY] = Dead
		}
	}

	if inpututil.IsKeyJustPressed(ebiten.KeySpace) {
		g.Running = !g.Running
		if g.Running {
			temp := g.world
			g.runningWorld = temp
			g.Tick = 0
			ebiten.SetWindowTitle("Running")
		} else {
			ebiten.SetWindowTitle("Editing")
		}

	}

	if inpututil.IsKeyJustPressed(ebiten.Key1) {
		g.SecondDelay = time.Millisecond
	}

	if inpututil.IsKeyJustPressed(ebiten.Key2) {
		g.SecondDelay = time.Second / 2
	}

	if inpututil.IsKeyJustPressed(ebiten.Key3) {
		g.SecondDelay = time.Second
	}

	if inpututil.IsKeyJustPressed(ebiten.Key4) {
		g.SecondDelay = time.Second * 2
	}

	if inpututil.IsKeyJustPressed(ebiten.KeyW) {
		if g.ScrollY > 0 {
			g.ScrollY--
		}

		if g.ScrollY < g.NumberOfTilesHeight {
			g.ScrollY++
		}

		if g.ScrollX > 0 {
			g.ScrollX--
		}

		if g.ScrollX < g.NumberOfTilesWidth {
			g.ScrollX++
		}
	}

	g.UpdateSimulation()
	return nil
}

func (g *Game) Draw(screen *ebiten.Image) {
	if !g.Running {
		for cellX := 0; cellX < g.NumberOfTilesWidth; cellX++ {
			for cellY := 0; cellY < g.NumberOfTilesHeight; cellY++ {
				x := float64(cellX) * TileSize
				y := float64(cellY) * TileSize

				c := color.RGBA{R: 100, G: 100, B: 100, A: 255}

				switch g.world[cellX][cellY] {
				case Wire:
					c = color.RGBA{R: 0, G: 0, B: 100, A: 255}
				case Head:
					c = color.RGBA{R: 100, G: 0, B: 0, A: 255}
				case Tail:
					c = color.RGBA{R: 0, G: 100, B: 0, A: 255}
				}

				cX, cY := getTileUnderMouse()
				if cX == cellX && cY == cellY {
					c.R += 100
				}
				ebitenutil.DrawRect(screen, x, y, TileSize-1, TileSize-1, c)
			}
		}
	} else {
		for cellX := 0; cellX < g.NumberOfTilesWidth; cellX++ {
			for cellY := 0; cellY < g.NumberOfTilesHeight; cellY++ {
				x := float64(cellX) * TileSize
				y := float64(cellY) * TileSize

				c := color.RGBA{R: 100, G: 100, B: 100, A: 255}

				switch g.runningWorld[cellX][cellY] {
				case Wire:
					c = color.RGBA{R: 0, G: 0, B: 100, A: 255}
				case Head:
					c = color.RGBA{R: 100, G: 0, B: 0, A: 255}
				case Tail:
					c = color.RGBA{R: 0, G: 100, B: 0, A: 255}
				}

				cX, cY := getTileUnderMouse()
				if cX == cellX && cY == cellY {
					c.R += 100
				}
				ebitenutil.DrawRect(screen, x, y, TileSize-1, TileSize-1, c)
			}
		}
	}

}

func (g *Game) Layout(outsideWidth, outsideHeight int) (int, int) {
	return g.ScreenWidth, g.ScreenHeight
}

func getTileUnderMouse() (int, int) {
	cX, cY := ebiten.CursorPosition()
	return cX / TileSize, cY / TileSize
}

func CreateWorldArray(width int, height int) [][]CellType {
	data := make([][]CellType, width)
	for x := 0; x < width; x++ {
		col := []CellType{}
		for y := 0; y < height; y++ {
			col = append(col, Dead)
		}
		data[x] = append(data[x], col...)
	}

	return data
}

func (g *Game) UpdateSimulation() {
	if g.Running && time.Now().After(g.LastUpdated.Add(g.SecondDelay)) {
		//Process the rules
		nextWorld := CreateWorldArray(g.NumberOfTilesWidth, g.NumberOfTilesHeight)
		for cellX := 0; cellX < g.NumberOfTilesWidth; cellX++ {
			for cellY := 0; cellY < g.NumberOfTilesHeight; cellY++ {
				nextWorld[cellX][cellY] = g.runningWorld[cellX][cellY]
				//The head turns all blocks around it into heads
				if g.runningWorld[cellX][cellY] == Wire {
					count := 0
					for nX := cellX - 1; nX <= cellX+1; nX++ {
						if nX < 0 || nX >= g.NumberOfTilesWidth {
							continue
						}
						for nY := cellY - 1; nY <= cellY+1; nY++ {
							if nY < 0 || nY >= g.NumberOfTilesHeight {
								continue
							}

							//Convert to head
							if g.runningWorld[nX][nY] == Head {
								count++
							}
						}
					}
					if count <= 2 && count > 0 {
						nextWorld[cellX][cellY] = Head
					}
				}

				if g.runningWorld[cellX][cellY] == Dead {
					continue
				}
				if g.runningWorld[cellX][cellY] == Head {
					nextWorld[cellX][cellY] = Tail
					continue
				}
				if g.runningWorld[cellX][cellY] == Tail {
					nextWorld[cellX][cellY] = Wire
					continue
				}

			}
		}

		g.runningWorld = nextWorld
		g.Tick++
		g.LastUpdated = time.Now()
	}
}
