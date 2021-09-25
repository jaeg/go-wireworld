package main

import (
	"bufio"
	"image/color"
	"os"
	"strconv"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
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

type CellType int

const (
	Dead CellType = iota
	Wire
	Head
	Tail
)

type CursorMode int

const (
	CursorChange CursorMode = iota
	CursorSelect
	CursorPaste
)

type Game struct {
	title               string
	world               [][]CellType
	runningWorld        [][]CellType
	NumberOfTilesWidth  int
	NumberOfTilesHeight int
	Running             bool
	Tick                int
	SecondDelay         time.Duration
	StepMode            bool
	LastUpdated         time.Time
	ScrollX             int
	ScrollY             int
	SimulationWidth     int
	SimulationHeight    int
	CursorMode          CursorMode
	SelectStartX        int
	SelectStartY        int
	SelectEndX          int
	SelectEndY          int
	CopyBuffer          [][]CellType
}

func NewGame(width int, height int) (*Game, error) {
	g := &Game{title: Title + "Editing", NumberOfTilesWidth: width, NumberOfTilesHeight: height, SecondDelay: time.Second / 2, LastUpdated: time.Now(), CursorMode: CursorChange}

	g.world = CreateWorldArray(width, height)
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
	if screenCellX < g.NumberOfTilesWidth && screenCellY < g.NumberOfTilesHeight {
		if g.CursorMode == CursorChange {
			if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
				if g.Running {
					g.SetRunning(false)
				}

				switch g.world[screenCellX][screenCellY] {
				case Dead:
					g.world[screenCellX][screenCellY] = Wire
				case Wire:
					g.world[screenCellX][screenCellY] = Head
				case Head:
					g.world[screenCellX][screenCellY] = Tail
				case Tail:
					g.world[screenCellX][screenCellY] = Wire
				}
			} else if ebiten.IsMouseButtonPressed(ebiten.MouseButtonLeft) { //Allow for drag adding of wires.
				if g.Running {
					g.SetRunning(false)
				}

				if g.world[screenCellX][screenCellY] == Dead {
					g.world[screenCellX][screenCellY] = Wire
				}
			}

			//Kill cell
			if ebiten.IsMouseButtonPressed(ebiten.MouseButtonRight) {
				if g.Running {
					g.SetRunning(false)
				}

				g.world[screenCellX][screenCellY] = Dead
			}
		} else if g.CursorMode == CursorSelect {
			if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
				if g.SelectStartX == -1 || g.SelectEndX != -1 {
					g.SelectStartX = screenCellX
					g.SelectStartY = screenCellY
					g.SelectEndX = -1
					g.SelectEndY = -1
				} else {
					g.SelectEndX = screenCellX
					g.SelectEndY = screenCellY
					//g.CursorMode = CursorPaste
				}
			}

			if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonRight) {
				g.SelectStartX = -1
				g.SelectStartY = -1

				g.SelectEndX = -1
				g.SelectEndY = -1
			}

			//Copy selection
			if inpututil.IsKeyJustPressed(ebiten.KeyC) {
				g.CopyToBuffer(g.SelectStartX, g.SelectStartY, g.SelectEndX, g.SelectEndY)

				g.CursorMode = CursorPaste
			}

			//delete logic
			if inpututil.IsKeyJustPressed(ebiten.KeyX) {
				for x := g.SelectStartX; x <= g.SelectEndX; x++ {
					for y := g.SelectStartY; y <= g.SelectEndY; y++ {
						g.world[x][y] = Dead
					}
				}
				//Reset selection
				g.SelectStartX = -1
				g.SelectStartY = -1

				g.SelectEndX = -1
				g.SelectEndY = -1
			}

		} else if g.CursorMode == CursorPaste {
			//Copy from paste location
			if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
				g.PasteFromBuffer(screenCellX, screenCellY)
			}

			//Cancel the paste
			if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonRight) {
				g.SelectStartX = -1
				g.SelectStartY = -1

				g.SelectEndX = -1
				g.SelectEndY = -1
				g.CursorMode = CursorChange
			}
		}
	}

	//Pause/Run simulation
	if inpututil.IsKeyJustPressed(ebiten.KeySpace) {
		g.SetRunning(!g.Running)
	}

	//Speed settings
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

	// Move view controls
	if ebiten.IsKeyPressed(ebiten.KeyW) {
		if g.ScrollY > 0 {
			g.ScrollY--
		}
	}

	if ebiten.IsKeyPressed(ebiten.KeyS) {
		if g.ScrollY < g.NumberOfTilesHeight {
			g.ScrollY++
		}
	}

	if ebiten.IsKeyPressed(ebiten.KeyA) {
		if g.ScrollX > 0 {
			g.ScrollX--
		}
	}

	if ebiten.IsKeyPressed(ebiten.KeyD) {
		if g.ScrollX < g.NumberOfTilesWidth {
			g.ScrollX++
		}
	}

	// Editing tools
	if ebiten.IsKeyPressed(ebiten.KeyE) {
		g.SaveWorld("save.csv")
	}

	if ebiten.IsKeyPressed(ebiten.KeyL) {
		g.LoadWorld("save.csv")
	}

	if inpututil.IsKeyJustPressed(ebiten.KeyQ) {
		if g.CursorMode == CursorChange {
			g.CursorMode = CursorSelect
			g.SelectEndX = -1
			g.SelectEndY = -1
			g.SelectStartX = -1
			g.SelectEndY = -1
		} else {
			g.CursorMode = CursorChange
		}
	}

	g.UpdateSimulation()
	return nil
}

func (g *Game) Draw(screen *ebiten.Image) {
	if !g.Running {
		cX, cY := getTileUnderMouse()

		ebiten.SetWindowTitle(g.title + " FPS: " + strconv.Itoa(int(ebiten.CurrentFPS())) + " X: " + strconv.Itoa(g.ScrollX+cX) + " Y: " + strconv.Itoa(g.ScrollY+cY))
		g.DrawWorldArray(g.world, screen)
	} else {
		ebiten.SetWindowTitle(g.title + " FPS: " + strconv.Itoa(int(ebiten.CurrentFPS())) + " : " + strconv.Itoa(g.Tick))
		g.DrawWorldArray(g.runningWorld, screen)
	}

}

func (g *Game) Layout(outsideWidth, outsideHeight int) (int, int) {
	g.SimulationWidth = outsideWidth - GUIWidth
	g.SimulationHeight = outsideHeight
	return outsideWidth, outsideHeight
}

func (g *Game) SetRunning(status bool) {
	g.Running = status
	if status {
		temp := g.world
		g.runningWorld = temp
		g.Tick = 0
		g.title = Title + "Running"
	} else {
		g.title = Title + "Editing"
	}
}

func (g *Game) DrawWorldArray(world [][]CellType, screen *ebiten.Image) {
	width := float64(g.SimulationWidth) / TileSize
	height := float64(g.SimulationHeight) / TileSize

	for x := 0.0; x < width; x++ {
		for y := 0.0; y < height; y++ {
			screenX := x * TileSize
			screenY := y * TileSize

			cellX := g.ScrollX + int(x)
			cellY := g.ScrollY + int(y)

			c := color.RGBA{R: 0, G: 0, B: 0, A: 255}

			if cellX >= 0 && cellX < g.NumberOfTilesWidth && cellY >= 0 && cellY < g.NumberOfTilesHeight {
				switch world[cellX][cellY] {
				case Wire:
					c = WireColor
				case Head:
					c = HeadColor
				case Tail:
					c = TailColor
				case Dead:
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
			if g.CursorMode == CursorSelect || g.CursorMode == CursorPaste {
				//Highlight cell if its selected
				if cellX >= g.SelectStartX && cellX <= g.SelectEndX && cellY >= g.SelectStartY && cellY <= g.SelectEndY {
					ebitenutil.DrawRect(screen, screenX, screenY, TileSize-1, TileSize-1, CursorSelectedColor)
				}

				if g.SelectStartX != -1 && g.SelectEndY == -1 {
					if cellX >= g.SelectStartX+g.ScrollX && cellX <= cX+g.ScrollX && cellY >= g.SelectStartY && cellY <= cY+g.ScrollY {
						ebitenutil.DrawRect(screen, screenX, screenY, TileSize-1, TileSize-1, CursorSelectedColor)
					}
				}
			}

			//Paste
			if g.CursorMode == CursorPaste {
				//Highlight cell if its selected
				if cellX >= cX+g.ScrollX && cellX <= g.SelectEndX-g.SelectStartX+cX+g.ScrollX && cellY >= cY+g.ScrollY && cellY <= g.SelectEndY-g.SelectStartY+cY+g.ScrollY {
					ebitenutil.DrawRect(screen, screenX, screenY, TileSize-1, TileSize-1, CursorPasteColor)
				}
			}
		}
	}
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

func (g *Game) SaveWorld(fileName string) error {
	f, err := os.OpenFile(fileName, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0755)

	if err != nil {
		return err
	}
	defer f.Close()
	for y := 0; y < g.NumberOfTilesHeight; y++ {
		for x := 0; x < g.NumberOfTilesWidth; x++ {
			f.WriteString(strconv.Itoa(int(g.world[x][y])))
		}
		f.WriteString("\n")
	}

	return nil
}

func (g *Game) LoadWorld(fileName string) error {
	f, err := os.OpenFile(fileName, os.O_RDONLY, 0755)

	if err != nil {
		return err
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	// optionally, resize scanner's capacity for lines over 64K, see next example
	y := 0
	for scanner.Scan() {
		l := scanner.Text()
		for x, c := range l {
			if string(c) == "0" {
				g.world[x][y] = 0
			}

			if string(c) == "1" {
				g.world[x][y] = 1
			}
		}
		y++
	}

	if err := scanner.Err(); err != nil {
		return err
	}

	return nil
}

func (g *Game) CopyToBuffer(sX, sY, eX, eY int) {
	g.CopyBuffer = CreateWorldArray(eX-sX+1, eY-sY+1)

	for x := sX; x <= eX; x++ {
		for y := sY; y <= eY; y++ {
			g.CopyBuffer[x-sX][y-sY] = g.world[x][y]
		}
	}
}

func (g *Game) PasteFromBuffer(startX, startY int) {
	for x := 0; x < len(g.CopyBuffer); x++ {
		for y := 0; y < len(g.CopyBuffer[0]); y++ {
			dstX := x + startX
			dstY := y + startY
			if dstX > 0 && dstX < g.NumberOfTilesWidth && dstY > 0 && dstY < g.NumberOfTilesHeight {
				g.world[dstX][dstY] = g.CopyBuffer[x][y]
			}
		}
	}
}

func getTileUnderMouse() (int, int) {
	cX, cY := ebiten.CursorPosition()
	temp := int(TileSize)
	return cX / temp, cY / temp
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
