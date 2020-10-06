package ui2d

import (
	"bufio"
	. "gameswithgo/rpg/game"
	"github.com/veandco/go-sdl2/sdl"
	"image/png"
	"math/rand"
	"os"
	"strconv"
	"strings"
)

type ui struct {
	winWidth  int
	winHeight int

	renderer *sdl.Renderer
	window   *sdl.Window

	// tile of interest = x or y px from image / 32
	textureAtlas *sdl.Texture
	textureIndex map[Tile][]sdl.Rect

	prevKeyboardState []uint8
	keyboardState     []uint8

	centerX int
	centerY int

	r         *rand.Rand
	levelChan chan *Level
	inputChan chan *Input
}

func NewUI(inputChan chan *Input, levelChan chan *Level) *ui {
	ui := &ui{}
	ui.inputChan = inputChan
	ui.levelChan = levelChan
	ui.winHeight = 720
	ui.winWidth = 1280
	ui.r = rand.New(rand.NewSource(1))
	var err error
	ui.window, err = sdl.CreateWindow("RPG!!!", sdl.WINDOWPOS_UNDEFINED, sdl.WINDOWPOS_UNDEFINED,
		int32(ui.winWidth), int32(ui.winHeight), sdl.WINDOW_SHOWN)
	if err != nil {
		panic(err)
	}
	ui.renderer, err = sdl.CreateRenderer(ui.window, -1, sdl.RENDERER_ACCELERATED)
	if err != nil {
		panic(err)
	}

	// bilinear filtering through SDL
	sdl.SetHint(sdl.HINT_RENDER_SCALE_QUALITY, "1")

	ui.textureAtlas = ui.imgFileToTexture("rpg/ui2d/assets/tiles.png")
	ui.loadTextureIndex()

	ui.keyboardState = sdl.GetKeyboardState()
	ui.prevKeyboardState = make([]uint8, len(ui.keyboardState))
	for i, v := range ui.keyboardState {
		ui.prevKeyboardState[i] = v
	}

	ui.centerX = -1
	ui.centerY = -1

	return ui
}

func (ui *ui) loadTextureIndex() {
	ui.textureIndex = make(map[Tile][]sdl.Rect)
	infile, err := os.Open("rpg/ui2d/assets/atlas-index.txt")
	if err != nil {
		panic(err)
	}
	defer infile.Close()
	scanner := bufio.NewScanner(infile)
	for scanner.Scan() {
		line := scanner.Text()
		line = strings.TrimSpace(line)
		tileRune := Tile(line[0]) // turning a character from file to a proper tile!
		xy := line[1:]
		splitXYC := strings.Split(xy, ",") // x, y, count (of tiles of the same type, e.g. different walls)
		x, err := strconv.ParseInt(strings.TrimSpace(splitXYC[0]), 10, 64)
		if err != nil {
			panic(err)
		}
		y, err := strconv.ParseInt(strings.TrimSpace(splitXYC[1]), 10, 64)
		if err != nil {
			panic(err)
		}
		variationCount, err := strconv.ParseInt(strings.TrimSpace(splitXYC[2]), 10, 64)
		if err != nil {
			panic(err)
		}

		var rects []sdl.Rect
		for i := 0; i < int(variationCount); i++ {
			rects = append(rects, sdl.Rect{int32(x * 32), int32(y * 32), 32, 32})
			x++
			if x > 62 {
				x = 0
				y++
			}
		}

		ui.textureIndex[tileRune] = rects
	}
}

func (ui *ui) imgFileToTexture(filename string) *sdl.Texture {
	infile, err := os.Open(filename)
	if err != nil {
		panic(err)
	}
	defer infile.Close()

	img, err := png.Decode(infile)
	if err != nil {
		panic(err)
	}

	w := img.Bounds().Max.X
	h := img.Bounds().Max.Y

	pixels := make([]byte, w*h*4)
	bIndex := 0
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			r, g, b, a := img.At(x, y).RGBA()
			pixels[bIndex] = byte(r / 256)
			bIndex++
			pixels[bIndex] = byte(g / 256)
			bIndex++
			pixels[bIndex] = byte(b / 256)
			bIndex++
			pixels[bIndex] = byte(a / 256)
			bIndex++
		}
	}
	tex, err := ui.renderer.CreateTexture(sdl.PIXELFORMAT_ABGR8888, sdl.TEXTUREACCESS_STATIC, int32(w), int32(h))
	if err != nil {
		panic(err)
	}
	err = tex.SetBlendMode(sdl.BLENDMODE_BLEND)
	if err != nil {
		panic(err)
	}
	err = tex.Update(nil, pixels, w*4)
	// pitch = bytes per line
	if err != nil {
		panic(err)
	}

	return tex
}

func init() {
	sdl.LogSetAllPriority(sdl.LOG_PRIORITY_VERBOSE)

	err := sdl.Init(sdl.INIT_EVERYTHING)
	if err != nil {
		panic(err)
	}
}

func (ui *ui) Draw(level *Level) {

	if ui.centerX == -1 && ui.centerY == -1 {
		ui.centerX = level.Player.X
		ui.centerY = level.Player.Y
	}

	//dx := level.Player.X - centerX
	//dy := level.Player.Y - centerY
	//distFromCenter := math.Sqrt(float64(dx*dx+dy*dy))
	limit := 5
	if level.Player.X > ui.centerX+limit {
		ui.centerX++
	} else if level.Player.X < ui.centerX-limit {
		ui.centerX--
	} else if level.Player.Y > ui.centerY+limit {
		ui.centerY++
	} else if level.Player.Y < ui.centerY-limit {
		ui.centerY--
	}

	// used for camera movement
	offsetX := int32((ui.winWidth / 2) - ui.centerX*32)
	offsetY := int32((ui.winHeight / 2) - ui.centerY*32)

	ui.renderer.Clear()
	ui.r.Seed(1)
	for y, row := range level.Map {
		for x, tile := range row {
			if tile != Blank {

				srcRects := ui.textureIndex[tile]
				srcRect := srcRects[ui.r.Intn(len(srcRects))]
				dstRect := sdl.Rect{int32(x*32) + offsetX, int32(y*32) + offsetY, 32, 32}

				pos := Pos{x, y}
				if level.Debug[pos] { // does map containt the position to draw?
					ui.textureAtlas.SetColorMod(128, 0, 0) // enhances color on every copy(?)
				} else {
					ui.textureAtlas.SetColorMod(255, 255, 255)
				}

				ui.renderer.Copy(ui.textureAtlas, &srcRect, &dstRect)
			}
		}
	}
	for pos, monster := range level.Monsters {
		monsterSrcRect := ui.textureIndex[Tile(monster.Rune)][0]
		ui.renderer.Copy(ui.textureAtlas, &monsterSrcRect,
			&sdl.Rect{int32(pos.X)*32 + offsetX, int32(pos.Y)*32 + offsetY, 32, 32})
	}
	playerSrcRect := ui.textureIndex['@'][0]
	ui.renderer.Copy(ui.textureAtlas, &playerSrcRect,
		&sdl.Rect{int32(level.Player.X)*32 + offsetX, int32(level.Player.Y)*32 + offsetY, 32, 32})
	ui.renderer.Present()
}

func (ui *ui) Run() {
	// Comment from YT
	//I'm not sure if you discover this later, but for the keyboard events: the event has "Type" and "Repeat" members.  So, if "Type" is "sdl.KEYDOWN" and "Repeat" is 0, then this is the initial press of that key.
	//
	//if e.Type == "sdl.KEYDOWN" && e.Repeat == 0 {
	//	if e.Keysym.Sym == <Whatever> {
	//		<do stuff>
	//	}
	//}
	for {
		for event := sdl.PollEvent(); event != nil; event = sdl.PollEvent() {
			switch e := event.(type) {
			case *sdl.QuitEvent:
				ui.inputChan <- &Input{Typ: QuitGame, LevelChannel: ui.levelChan}
			case *sdl.WindowEvent:
				if e.Event == sdl.WINDOWEVENT_CLOSE {
					ui.inputChan <- &Input{Typ: CloseWindow}
				}
			}
		}

		select {
		case newLevel, ok := <-ui.levelChan:
			if ok {
				ui.Draw(newLevel)
			}
		default:
		}
		// TODO make a function to ask if a key has been pressed
		if sdl.GetKeyboardFocus() == ui.window || sdl.GetMouseFocus() == ui.window {
			var input Input
			if ui.keyboardState[sdl.SCANCODE_UP] != 0 && ui.prevKeyboardState[sdl.SCANCODE_UP] == 0 {
				input.Typ = Up
			}
			if ui.keyboardState[sdl.SCANCODE_DOWN] != 0 && ui.prevKeyboardState[sdl.SCANCODE_DOWN] == 0 {
				input.Typ = Down
			}
			if ui.keyboardState[sdl.SCANCODE_LEFT] != 0 && ui.prevKeyboardState[sdl.SCANCODE_LEFT] == 0 {
				input.Typ = Left
			}
			if ui.keyboardState[sdl.SCANCODE_RIGHT] != 0 && ui.prevKeyboardState[sdl.SCANCODE_RIGHT] == 0 {
				input.Typ = Right
			}
			if ui.keyboardState[sdl.SCANCODE_S] == 0 && ui.prevKeyboardState[sdl.SCANCODE_S] != 0 {
				input.Typ = Search
			}

			for i, v := range ui.keyboardState {
				ui.prevKeyboardState[i] = v
			}

			if input.Typ != None {
				//fmt.Println("Sending input...")
				ui.inputChan <- &input
			}
		}
		sdl.Delay(20) // CPU is not getting eaten when waiting for inputs
	}
}
