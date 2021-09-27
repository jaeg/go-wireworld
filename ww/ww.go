package ww

import (
	"bufio"
	"os"
	"strconv"
)

type CellType int

const (
	Dead CellType = iota
	Wire
	Head
	Tail
)

type WireWorld struct {
	world               [][]CellType
	runningWorld        [][]CellType
	NumberOfTilesWidth  int
	NumberOfTilesHeight int
	Running             bool
	Tick                int
	CopyBuffer          [][]CellType
}

func CreateNewWireWorld(width int, height int) *WireWorld {
	ww := &WireWorld{NumberOfTilesWidth: width, NumberOfTilesHeight: height}
	ww.world = CreateWorldArray(width, height)
	return ww
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

func (g *WireWorld) SaveWorld(fileName string) error {
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

func (g *WireWorld) LoadWorld(fileName string) error {
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

func (g *WireWorld) CopyToBuffer(sX, sY, eX, eY int) {
	g.CopyBuffer = CreateWorldArray(eX-sX+1, eY-sY+1)

	for x := sX; x <= eX; x++ {
		for y := sY; y <= eY; y++ {
			g.CopyBuffer[x-sX][y-sY] = g.world[x][y]
		}
	}
}

func (g *WireWorld) PasteFromBuffer(startX, startY int) {
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

func (g *WireWorld) UpdateSimulation() {
	if g.Running {
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
	}
}

func (w *WireWorld) GetEditorTileAt(x int, y int) CellType {
	return w.world[x][y]
}

func (w *WireWorld) SetEditorTile(x int, y int, t CellType) {
	w.world[x][y] = t
}

func (w *WireWorld) GetRunningTileAt(x int, y int) CellType {
	return w.runningWorld[x][y]
}

func (w *WireWorld) GetCopyBufferAt(x int, y int) CellType {
	return w.CopyBuffer[x][y]
}

func (w *WireWorld) GetCopyBufferDimensions() (int, int) {
	if len(w.CopyBuffer) > 0 {
		return len(w.CopyBuffer), len(w.CopyBuffer[0])
	}
	return 0, 0

}

func (g *WireWorld) SetRunning(status bool) {
	g.Running = status
	if status {
		temp := g.world
		g.runningWorld = temp
		g.Tick = 0
	}
}

func (w *WireWorld) ReturnWorld() [][]CellType {
	return w.world
}

func (w *WireWorld) ReturnRunningWorld() [][]CellType {
	return w.runningWorld
}
