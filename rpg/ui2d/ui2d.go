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

const winWidth, winHeight = 800, 600

var renderer *sdl.Renderer

// tile of interest = x or y px from image / 32
var textureAtlas *sdl.Texture
var textureIndex map[Tile][]sdl.Rect


var prevKeyboardState []uint8
var keyboardState []uint8

var centerX int
var centerY int

func loadTextureIndex() {
	textureIndex = make(map[Tile][]sdl.Rect)
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

		textureIndex[tileRune] = rects
	}
}

func imgFileToTexture(filename string) *sdl.Texture {
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
	tex, err := renderer.CreateTexture(sdl.PIXELFORMAT_ABGR8888, sdl.TEXTUREACCESS_STATIC, int32(w), int32(h))
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
	window, err := sdl.CreateWindow("RPG!!!", sdl.WINDOWPOS_UNDEFINED, sdl.WINDOWPOS_UNDEFINED,
		winWidth, winHeight, sdl.WINDOW_SHOWN)
	if err != nil {
		panic(err)
	}
	renderer, err = sdl.CreateRenderer(window, -1, sdl.RENDERER_ACCELERATED)
	if err != nil {
		panic(err)
	}

	// bilinear filtering through SDL
	sdl.SetHint(sdl.HINT_RENDER_SCALE_QUALITY, "1")

	textureAtlas = imgFileToTexture("rpg/ui2d/assets/tiles.png")
	loadTextureIndex()

	keyboardState = sdl.GetKeyboardState()
	prevKeyboardState = make([]uint8, len(keyboardState))
	for i, v := range keyboardState {
		prevKeyboardState[i] = v
	}

	centerX = -1
	centerY = -1
}

type UI2d struct {
}

func (ui *UI2d) Draw(level *Level) {
	if centerX == -1 && centerY == -1 {
		centerX = level.Player.X
		centerY = level.Player.Y
	}

	//dx := level.Player.X - centerX
	//dy := level.Player.Y - centerY
	//distFromCenter := math.Sqrt(float64(dx*dx+dy*dy))
	limit := 5
	if level.Player.X > centerX + limit {
		centerX++
	} else if level.Player.X < centerX - limit {
		centerX--
	} else if level.Player.Y > centerY + limit {
		centerY++
	} else if level.Player.Y < centerY - limit {
		centerY--
	}

	// used for camera movement
	offsetX := int32((winWidth/2) - centerX*32)
	offsetY := int32((winHeight/2) - centerY*32)

	renderer.Clear()
	rand.Seed(1)
	for y, row := range level.Map {
		for x, tile := range row {
			if tile != Blank {
				srcRects := textureIndex[tile]
				srcRect := srcRects[rand.Intn(len(srcRects))]
				dstRect := sdl.Rect{int32(x * 32) + offsetX, int32(y * 32) + offsetY, 32, 32}
				renderer.Copy(textureAtlas, &srcRect, &dstRect)
			}
		}
	}
	renderer.Copy(textureAtlas, &sdl.Rect{21 * 32, 59 * 32, 32, 32},
	&sdl.Rect{int32(level.Player.X) *32 + offsetX, int32(level.Player.Y) * 32 + offsetY, 32, 32})
	renderer.Present()
}

func (ui *UI2d) GetInput() *Input {
	// Comment from YT
	//I'm not sure if you discover this later, but for the keyboard events: the event has "Type" and "Repeat" members.  So, if "Type" is "sdl.KEYDOWN" and "Repeat" is 0, then this is the initial press of that key.
	//
	//if e.Type == "sdl.KEYDOWN" && e.Repeat == 0 {
	//	if e.Keysym.Sym == <Whatever> {
	//		<do stuff>
	//	}
	//}

	for event := sdl.PollEvent(); event != nil; event = sdl.PollEvent() {
		switch event.(type) {
		case *sdl.QuitEvent:
			return &Input{Quit}
		}
	}

	var input Input
	if keyboardState[sdl.SCANCODE_UP] == 0 && prevKeyboardState[sdl.SCANCODE_UP] != 0 {
		input.Typ = Up
	}
	if keyboardState[sdl.SCANCODE_DOWN] == 0 && prevKeyboardState[sdl.SCANCODE_DOWN] != 0 {
		input.Typ = Down
	}
	if keyboardState[sdl.SCANCODE_LEFT] == 0 && prevKeyboardState[sdl.SCANCODE_LEFT] != 0 {
		input.Typ = Left
	}
	if keyboardState[sdl.SCANCODE_RIGHT] == 0 && prevKeyboardState[sdl.SCANCODE_RIGHT] != 0 {
		input.Typ = Right
	}

	for i, v := range keyboardState {
		prevKeyboardState[i] = v
	}

	if input.Typ != None {
		return &input
	}
	return nil
}
