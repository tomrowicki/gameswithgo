package ui2d

import (
	"bufio"
	"fmt"
	. "gameswithgo/rpg/game"
	"github.com/veandco/go-sdl2/mix"
	"github.com/veandco/go-sdl2/sdl"
	"github.com/veandco/go-sdl2/ttf"
	"image/png"
	"math/rand"
	"os"
	"strconv"
	"strings"
)

const itemSizeRatio = .033

type mouseState struct {
	leftButton  bool
	rightButton bool
	pos         Pos
}

func getMouseState() *mouseState {
	mouseX, mouseY, mouseButtonState := sdl.GetMouseState()
	leftButton := mouseButtonState & sdl.ButtonLMask() // using bitwise op
	rightButton := mouseButtonState & sdl.ButtonRMask()
	var result mouseState
	result.pos = Pos{int(mouseX), int(mouseY)}
	result.leftButton = !(leftButton == 0)
	result.rightButton = !(rightButton == 0)
	return &result
}

type sounds struct {
	openingDoors []*mix.Chunk
	footsteps    []*mix.Chunk
}

func playRandomSound(chunks []*mix.Chunk, volume int) {
	chunkIndex := rand.Intn(len(chunks))
	chunks[chunkIndex].Volume(volume)
	chunks[chunkIndex].Play(-1, 0)
}

type uiState int

const (
	UIMain uiState = iota
	UIInventory
)

type ui struct {
	state uiState

	draggedItem *Item

	sounds sounds

	winWidth  int
	winHeight int

	renderer *sdl.Renderer
	window   *sdl.Window

	// tile of interest = x or y px from image / 32
	textureAtlas *sdl.Texture
	textureIndex map[rune][]sdl.Rect

	prevKeyboardState []uint8
	keyboardState     []uint8

	centerX int
	centerY int

	r         *rand.Rand
	levelChan chan *Level
	inputChan chan *Input

	fontSmall  *ttf.Font
	fontMedium *ttf.Font
	fontLarge  *ttf.Font

	eventBackground           *sdl.Texture
	groundInventoryBackground *sdl.Texture

	str2TexSmall map[string]*sdl.Texture
	str2TexMed   map[string]*sdl.Texture
	str2TexLarge map[string]*sdl.Texture

	currMouseState *mouseState
	prevMouseState *mouseState
}

func NewUI(inputChan chan *Input, levelChan chan *Level) *ui {
	ui := &ui{}
	ui.state = UIMain
	ui.str2TexLarge = make(map[string]*sdl.Texture)
	ui.str2TexMed = make(map[string]*sdl.Texture)
	ui.str2TexSmall = make(map[string]*sdl.Texture)
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
	//sdl.SetHint(sdl.HINT_RENDER_SCALE_QUALITY, "1")

	ui.textureAtlas = ui.imgFileToTexture("rpg/ui2d/assets/tiles.png")
	ui.loadTextureIndex()

	ui.keyboardState = sdl.GetKeyboardState()
	ui.prevKeyboardState = make([]uint8, len(ui.keyboardState))
	for i, v := range ui.keyboardState {
		ui.prevKeyboardState[i] = v
	}

	ui.centerX = -1
	ui.centerY = -1

	ui.fontSmall, err = ttf.OpenFont("rpg/ui2d/assets/Kingthings_Foundation.ttf", int(float64(ui.winHeight)*.02))
	//ui.fontSmall, err = ttf.OpenFont("rpg/ui2d/assets/Kingthings_Foundation.ttf", 18)
	if err != nil {
		panic(err)
	}

	ui.fontMedium, err = ttf.OpenFont("rpg/ui2d/assets/Kingthings_Foundation.ttf", 32)
	if err != nil {
		panic(err)
	}

	ui.fontLarge, err = ttf.OpenFont("rpg/ui2d/assets/Kingthings_Foundation.ttf", 64)
	if err != nil {
		panic(err)
	}

	ui.eventBackground = ui.GetSinglePixelTex(sdl.Color{0, 0, 0, 128})
	ui.eventBackground.SetBlendMode(sdl.BLENDMODE_BLEND)

	ui.groundInventoryBackground = ui.GetSinglePixelTex(sdl.Color{149, 84, 19, 128})
	ui.groundInventoryBackground.SetBlendMode(sdl.BLENDMODE_BLEND)

	err = mix.OpenAudio(22050, mix.DEFAULT_FORMAT, 2, 4096)
	if err != nil {
		panic(err)
	}
	mus, err := mix.LoadMUS("rpg/ui2d/assets/ambient.ogg")
	if err != nil {
		panic(err)
	}
	err = mus.Play(-1)
	if err != nil {
		panic(err)
	}

	footstepBase := "rpg/ui2d/assets/footstep0"
	for i := 0; i < 10; i++ {
		footstepFile := footstepBase + strconv.Itoa(i) + ".ogg"
		footstep, err := mix.LoadWAV(footstepFile)
		if err != nil {
			panic(err)
		}
		ui.sounds.footsteps = append(ui.sounds.footsteps, footstep)
	}
	door1, err := mix.LoadWAV("rpg/ui2d/assets/doorOpen_1.ogg")
	if err != nil {
		panic(err)
	}
	ui.sounds.openingDoors = append(ui.sounds.openingDoors, door1)
	door2, err := mix.LoadWAV("rpg/ui2d/assets/doorOpen_2.ogg")
	if err != nil {
		panic(err)
	}
	ui.sounds.openingDoors = append(ui.sounds.openingDoors, door2)

	return ui
}

type FontSize int

const (
	FontSmall FontSize = iota
	FontMedium
	FontLarge
)

func (ui *ui) stringToTexture(s string, color sdl.Color, size FontSize) *sdl.Texture {
	var font *ttf.Font
	switch size {
	case FontSmall:
		font = ui.fontSmall
		tex, exists := ui.str2TexSmall[s]
		if exists {
			return tex
		}
	case FontMedium:
		font = ui.fontMedium
		tex, exists := ui.str2TexMed[s]
		if exists {
			return tex
		}
	case FontLarge:
		font = ui.fontLarge
		tex, exists := ui.str2TexLarge[s]
		if exists {
			return tex
		}
	}

	fontSurface, err := font.RenderUTF8Blended(s, color)
	if err != nil {
		panic(err)
	}
	tex, err := ui.renderer.CreateTextureFromSurface(fontSurface)
	if err != nil {
		panic(err)
	}

	switch size {
	case FontSmall:
		ui.str2TexSmall[s] = tex
	case FontMedium:
		ui.str2TexMed[s] = tex
	case FontLarge:
		ui.str2TexLarge[s] = tex
	}

	return tex
}

func (ui *ui) loadTextureIndex() {
	ui.textureIndex = make(map[rune][]sdl.Rect)
	infile, err := os.Open("rpg/ui2d/assets/atlas-index.txt")
	if err != nil {
		panic(err)
	}
	defer infile.Close()
	scanner := bufio.NewScanner(infile)
	for scanner.Scan() {
		line := scanner.Text()
		line = strings.TrimSpace(line)
		tileRune := rune(line[0]) // turning a character from file to a proper tile!
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

	err = ttf.Init()
	if err != nil {
		panic(err)
	}

	err = mix.Init(mix.INIT_OGG)
	if err != nil {
		panic(err)
	}
}

func (ui *ui) DrawInventory(level *Level) {
	playerSrcRect := ui.textureIndex[level.Player.Rune][0]
	invRect := ui.getInventoryRect()
	ui.renderer.Copy(ui.groundInventoryBackground, nil, invRect)
	ui.renderer.Copy(ui.textureAtlas, &playerSrcRect, &sdl.Rect{invRect.X + invRect.X/4, invRect.Y, invRect.W / 2, invRect.H / 2})

	for i, item := range level.Player.Items {
		itemSrcRect := ui.textureIndex[item.Rune][0]
		if item == ui.draggedItem {
			itemSize := int32(float32(ui.winWidth) * itemSizeRatio)
			ui.renderer.Copy(ui.textureAtlas, &itemSrcRect, &sdl.Rect{int32(ui.currMouseState.pos.X), int32(ui.currMouseState.pos.Y), itemSize, itemSize})
		} else {
			ui.renderer.Copy(ui.textureAtlas, &itemSrcRect, ui.getInventoryItemRect(i))
		}
	}
}

func (ui *ui) getInventoryRect() *sdl.Rect {
	invWidth := int32(float32(ui.winWidth) * .40)
	invHeight := int32(float32(ui.winHeight) * .75)
	offsetX := (int32(ui.winWidth) - invWidth) / 2
	offsetY := (int32(ui.winHeight) - invHeight) / 2
	return &sdl.Rect{offsetX, offsetY, invWidth, invHeight}
}

func (ui *ui) getInventoryItemRect(i int) *sdl.Rect {
	invRect := ui.getInventoryRect()
	itemSize := int32(float32(ui.winWidth) * itemSizeRatio)
	return &sdl.Rect{invRect.X + int32(i)*itemSize, invRect.Y + invRect.H - itemSize, itemSize, itemSize}
}

func (ui *ui) CheckDroppedItem(level *Level) *Item {
	invRect := ui.getInventoryRect()
	mousePos := ui.currMouseState.pos
	if invRect.HasIntersection(&sdl.Rect{int32(mousePos.X), int32(mousePos.Y), 1, 1}) {
		return nil
	}
	return ui.draggedItem
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
		diff := level.Player.X - (ui.centerX + limit)
		ui.centerX += diff
	} else if level.Player.X < ui.centerX-limit {
		diff := (ui.centerX - limit) - level.Player.X
		ui.centerX -= diff
	} else if level.Player.Y > ui.centerY+limit {
		diff := level.Player.Y - (ui.centerY + limit)
		ui.centerY += diff
	} else if level.Player.Y < ui.centerY-limit {
		diff := (ui.centerY - limit) - level.Player.Y
		ui.centerY -= diff
	}

	// used for camera movement
	offsetX := int32((ui.winWidth / 2) - ui.centerX*32)
	offsetY := int32((ui.winHeight / 2) - ui.centerY*32)

	ui.renderer.Clear()
	ui.r.Seed(1)
	for y, row := range level.Map {
		for x, tile := range row {

			if tile.Rune != Blank {
				srcRects := ui.textureIndex[tile.Rune]
				srcRect := srcRects[ui.r.Intn(len(srcRects))]

				if tile.Visible || tile.Seen {
					dstRect := sdl.Rect{int32(x*32) + offsetX, int32(y*32) + offsetY, 32, 32}
					pos := Pos{x, y}
					if level.Debug[pos] { // does map containt the position to draw?
						ui.textureAtlas.SetColorMod(128, 0, 0) // enhances color on every copy(?)
					} else if tile.Seen && !tile.Visible {
						ui.textureAtlas.SetColorMod(128, 128, 128)
					} else {
						ui.textureAtlas.SetColorMod(255, 255, 255)
					}
					ui.renderer.Copy(ui.textureAtlas, &srcRect, &dstRect)
					// TODO different variants for overlay images?
					if tile.OverlayRune != Blank {
						srcRect := ui.textureIndex[tile.OverlayRune][0]
						ui.renderer.Copy(ui.textureAtlas, &srcRect, &dstRect)
					}
				}
			}
		}
	}
	ui.textureAtlas.SetColorMod(255, 255, 255) // prevents monster being drawn greyed-out due to fog of war mechanic
	for pos, monster := range level.Monsters {
		if level.Map[pos.Y][pos.X].Visible {
			monsterSrcRect := ui.textureIndex[monster.Rune][0]
			ui.renderer.Copy(ui.textureAtlas, &monsterSrcRect,
				&sdl.Rect{int32(pos.X)*32 + offsetX, int32(pos.Y)*32 + offsetY, 32, 32})
		}
	}

	// Render Items
	for pos, items := range level.Items {
		if level.Map[pos.Y][pos.X].Visible {
			for _, item := range items {
				itemSrcRect := ui.textureIndex[item.Rune][0]
				ui.renderer.Copy(ui.textureAtlas, &itemSrcRect,
					&sdl.Rect{int32(pos.X)*32 + offsetX, int32(pos.Y)*32 + offsetY, 32, 32})
			}
		}
	}

	// Render Player
	playerSrcRect := ui.textureIndex[level.Player.Rune][0]
	ui.renderer.Copy(ui.textureAtlas, &playerSrcRect,
		&sdl.Rect{int32(level.Player.X)*32 + offsetX, int32(level.Player.Y)*32 + offsetY, 32, 32})

	// Events UI
	textStartY := int32(float64(ui.winHeight) * .68) // allows to add spacing between lines
	textWidth := int32(float64(ui.winWidth) * .25)

	ui.renderer.Copy(ui.eventBackground, nil, &sdl.Rect{0, textStartY, textWidth, int32(ui.winHeight) - textStartY})

	i := level.EventPos
	count := 0
	_, fontSizeY, _ := ui.fontSmall.SizeUTF8("A") // mosta letters have the same height but there are exceptions
	for {
		event := level.Events[i]
		if event != "" {
			tex := ui.stringToTexture(event, sdl.Color{255, 0, 0, 0}, FontSmall)
			_, _, w, h, err := tex.Query()
			if err != nil {
				panic(err)
			}
			ui.renderer.Copy(tex, nil, &sdl.Rect{5, int32(count*fontSizeY) + textStartY, w, h})
		}
		i = (i + 1) % len(level.Events)
		count++
		if i == level.EventPos {
			break
		}
	}

	// Inventory UI
	groundInvStart := int32(float64(ui.winWidth) * .9)
	groundInvWidth := int32(ui.winWidth) - groundInvStart
	itemSize := int32(itemSizeRatio * float32(ui.winWidth))
	ui.renderer.Copy(ui.groundInventoryBackground, nil,
		&sdl.Rect{groundInvStart, int32(ui.winHeight) - itemSize, groundInvWidth, itemSize})
	items := level.Items[level.Player.Pos]
	for i, item := range items {
		itemSrcRect := ui.textureIndex[item.Rune][0]
		ui.renderer.Copy(ui.textureAtlas, &itemSrcRect, ui.getGroundItemRect(i))
	}
}

func (ui *ui) getGroundItemRect(index int) *sdl.Rect {
	itemSize := int32(float32(ui.winWidth) * itemSizeRatio)
	return &sdl.Rect{int32(int32(ui.winWidth) - itemSize - int32(index)*itemSize), int32(ui.winHeight) - itemSize, itemSize, itemSize}
}

func (ui *ui) keyDownOnce(key uint8) bool {
	return ui.keyboardState[key] != 0 && ui.prevKeyboardState[key] == 0
}

// key pressed then released
func (ui *ui) keyPressed(key uint8) bool {
	return ui.keyboardState[key] == 0 && ui.prevKeyboardState[key] != 0
}

func (ui *ui) GetSinglePixelTex(color sdl.Color) *sdl.Texture {
	tex, err := ui.renderer.CreateTexture(sdl.PIXELFORMAT_ABGR8888, sdl.TEXTUREACCESS_STATIC, 1, 1)
	if err != nil {
		panic(err)
	}
	pixels := make([]byte, 4)
	pixels[0] = color.R
	pixels[1] = color.G
	pixels[2] = color.B
	pixels[3] = color.A
	// pitch = bytes per line
	tex.Update(nil, pixels, 4)
	return tex
}

func (ui *ui) CheckInventoryItems(level *Level) *Item {
	if ui.currMouseState.leftButton {
		mousePos := ui.currMouseState.pos
		for i, item := range level.Player.Items {
			itemRect := ui.getInventoryItemRect(i)
			// checking if click occurs within the item's rect
			if itemRect.HasIntersection(&sdl.Rect{int32(mousePos.X), int32(mousePos.Y), 1, 1}) {
				fmt.Println("clicked inventory item!")
				return item
			}
		}
	}
	return nil
}

func (ui *ui) CheckGroundItems(level *Level) *Item {
	if !ui.currMouseState.leftButton && ui.prevMouseState.leftButton {
		items := level.Items[level.Player.Pos]
		mousePos := ui.currMouseState.pos
		for i, item := range items {
			itemRect := ui.getGroundItemRect(i)
			// checking if click occurs within the item's rect
			if itemRect.HasIntersection(&sdl.Rect{int32(mousePos.X), int32(mousePos.Y), 1, 1}) {
				return item
			}
		}
	}
	return nil
}

func (ui *ui) Run() {
	var newLevel *Level
	ui.prevMouseState = getMouseState()
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

		ui.currMouseState = getMouseState()
		var input Input
		var ok bool
		select {
		case newLevel, ok = <-ui.levelChan:
			if ok {
				switch newLevel.LastEvent {
				case Move:
					playRandomSound(ui.sounds.footsteps, 16)
				case DoorOpen:
					playRandomSound(ui.sounds.openingDoors, 32)
				default:
					// add more sounds
				}
			}
		default:
		}
		ui.Draw(newLevel)

		if ui.state == UIInventory {
			// have we stopped dragging?
			if ui.draggedItem != nil && !ui.currMouseState.leftButton && ui.prevMouseState.leftButton {
				item := ui.CheckDroppedItem(newLevel)
				if item != nil {
					input.Typ = DropItem
					input.Item = ui.draggedItem
					ui.draggedItem = nil
				}
			}
			if ui.currMouseState.leftButton && ui.draggedItem != nil {

			} else {
				ui.draggedItem = ui.CheckInventoryItems(newLevel)
			}
			ui.DrawInventory(newLevel)
		}
		ui.renderer.Present()

		item := ui.CheckGroundItems(newLevel)
		if item != nil {
			input.Typ = TakeItem
			input.Item = item
		}

		if sdl.GetKeyboardFocus() == ui.window || sdl.GetMouseFocus() == ui.window {

			if ui.keyDownOnce(sdl.SCANCODE_UP) {
				input.Typ = Up
			}
			if ui.keyDownOnce(sdl.SCANCODE_DOWN) {
				input.Typ = Down
			}
			if ui.keyDownOnce(sdl.SCANCODE_LEFT) {
				input.Typ = Left
			}
			if ui.keyDownOnce(sdl.SCANCODE_RIGHT) {
				input.Typ = Right
			}
			if ui.keyDownOnce(sdl.SCANCODE_T) {
				input.Typ = TakeAll
			}
			if ui.keyDownOnce(sdl.SCANCODE_I) {
				if ui.state == UIMain {
					ui.state = UIInventory
				} else {
					ui.state = UIMain
				}
			}
			//if ui.keyboardState[sdl.SCANCODE_S] == 0 && ui.prevKeyboardState[sdl.SCANCODE_S] != 0 {
			//	input.Typ = Search
			//}

			for i, v := range ui.keyboardState {
				ui.prevKeyboardState[i] = v
			}

			if input.Typ != None {
				//fmt.Println("Sending input...")
				ui.inputChan <- &input
			}
		}
		ui.prevMouseState = ui.currMouseState
		sdl.Delay(20) // CPU is not getting eaten when waiting for inputs

	}
}
