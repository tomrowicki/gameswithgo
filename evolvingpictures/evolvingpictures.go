package main

import (
	"fmt"
	. "gameswithgo/evolvingpictures/apt"
	. "gameswithgo/evolvingpictures/gui"
	"github.com/veandco/go-sdl2/sdl"
	"math/rand"
	"time"
)

var winWidth, winHeight = 800, 600
var rows, cols, numPics = 3, 3, rows * cols

type pixelResult struct {
	pixels []byte
	index  int
}

type audioState struct {
	explosionBytes []byte
	deviceID       sdl.AudioDeviceID
	audioSpec      *sdl.AudioSpec
}

type rgba struct {
	r, g, b byte
}

type picture struct {
	r Node
	g Node
	b Node
}

func (p *picture) String() string {
	return "R" + p.r.String() + "\n" + "G" + p.g.String() + "\n" + "B" + p.b.String()
}

func NewPicture() *picture {
	p := &picture{}

	p.r = GetRandomNode()
	p.g = GetRandomNode()
	p.b = GetRandomNode()

	num := rand.Intn(20) + 5
	for i := 0; i < num; i++ {
		p.r.AddRandom(GetRandomNode())
	}

	num = rand.Intn(20) + 5
	for i := 0; i < num; i++ {
		p.g.AddRandom(GetRandomNode())
	}

	num = rand.Intn(20) + 5
	for i := 0; i < num; i++ {
		p.b.AddRandom(GetRandomNode())
	}

	for p.r.AddLeaf(GetRandomLeaf()) {
	}
	for p.g.AddLeaf(GetRandomLeaf()) {
	}
	for p.b.AddLeaf(GetRandomLeaf()) {
	}

	return p
}

func (p *picture) Mutate() {
	r := rand.Intn(3)
	var nodeToMutate Node
	switch r {
	case 0:
		nodeToMutate = p.r
	case 1:
		nodeToMutate = p.g
	case 2:
		nodeToMutate = p.b
	}

	count := nodeToMutate.NodeCount()
	r = rand.Intn(count)
	nodeToMutate, count = GetNthNode(nodeToMutate, r, 0)
	mutation := Mutate(nodeToMutate)
	if nodeToMutate == p.r {
		p.r = mutation
	} else if nodeToMutate == p.g {
		p.g = mutation
	} else if nodeToMutate == p.b {
		p.b = mutation
	}
}

func setPixel(x, y int, c rgba, pixels []byte) {
	winWidthInt := winWidth
	index := (y*winWidthInt + x) * 4

	if index < len(pixels)-4 && index >= 0 {
		pixels[index] = c.r
		pixels[index+1] = c.g
		pixels[index+2] = c.b
	}
}

func pixelsToTexture(renderer *sdl.Renderer, pixels []byte, w, h int) *sdl.Texture {
	tex, err := renderer.CreateTexture(sdl.PIXELFORMAT_ABGR8888, sdl.TEXTUREACCESS_STREAMING, int32(w), int32(h))
	if err != nil {
		panic(err)
	}
	// pitch = bytes per line
	tex.Update(nil, pixels, w*4)
	return tex
}

func aptToPixels(pic *picture, w, h int, renderer *sdl.Renderer) []byte {
	// -1.0 and 1.0
	scale := float32(255 / 2)
	offset := -1.0 * scale
	pixels := make([]byte, w*h*4)
	pixelIndex := 0
	for yi := 0; yi < h; yi++ {
		y := float32(yi)/float32(h)*2 - 1
		for xi := 0; xi < w; xi++ {
			x := float32(xi)/float32(w)*2 - 1
			// color
			r := pic.r.Eval(x, y)
			g := pic.g.Eval(x, y)
			b := pic.b.Eval(x, y)
			pixels[pixelIndex] = byte(r*scale - offset)
			pixelIndex++
			pixels[pixelIndex] = byte(g*scale - offset)
			pixelIndex++
			pixels[pixelIndex] = byte(b*scale - offset)
			pixelIndex++
			pixelIndex++ // skipping alpha
		}
	}
	return pixels
}

func main() {
	sdl.InitSubSystem(sdl.INIT_AUDIO)
	window, err := sdl.CreateWindow("Evolving Pictures", sdl.WINDOWPOS_UNDEFINED, sdl.WINDOWPOS_UNDEFINED,
		int32(winWidth), int32(winHeight), sdl.WINDOW_SHOWN)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer window.Destroy()

	renderer, err := sdl.CreateRenderer(window, -1, sdl.RENDERER_ACCELERATED)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer renderer.Destroy()

	//explosionBytes, audioSpec := sdl.LoadWAV("balloons2/explode.wav")
	//audioId, err := sdl.OpenAudioDevice("", false, audioSpec, nil, 0)
	//if err != nil {
	//	panic(err)
	//}
	//defer sdl.FreeWAV(explosionBytes)
	//
	//audioState := audioState{explosionBytes, audioId, audioSpec}

	// bilinear filtering through SDL
	sdl.SetHint(sdl.HINT_RENDER_SCALE_QUALITY, "1")

	var elapsedTime float32

	rand.Seed(time.Now().Unix())

	picTrees := make([]*picture, numPics)
	for i := range picTrees {
		picTrees[i] = NewPicture()
	}

	picWidth := int(float32(winWidth/cols) * float32(.9)) // leaving some gap between pics (remaining 0.1)
	picHeight := int(float32(winHeight/rows) * float32(.9))

	pixelsChannel := make(chan pixelResult, numPics)
	buttons := make([]*ImageButton, numPics)
	for i := range picTrees {
		go func(i int) {
			pixels := aptToPixels(picTrees[i], picWidth, picHeight, renderer)
			pixelsChannel <- pixelResult{pixels, i}
		}(i)
	}

	keyboardState := sdl.GetKeyboardState()
	mouseState := GetMouseState()

	// GAME LOOP
	for {
		frameStart := time.Now()
		mouseState.Update()
		for event := sdl.PollEvent(); event != nil; event = sdl.PollEvent() {
			switch event.(type) {
			case *sdl.QuitEvent:
				return
			}
		}

		if keyboardState[sdl.SCANCODE_ESCAPE] != 0 {
			return
		}

		select {
		case pixelsAndIndex, ok := <-pixelsChannel:
			if ok {
				tex := pixelsToTexture(renderer, pixelsAndIndex.pixels, picWidth, picHeight)
				xi := pixelsAndIndex.index % cols
				yi := (pixelsAndIndex.index - xi) / cols
				x := int32(xi * picWidth)
				y := int32(yi * picHeight)
				xPad := int32(float32(winWidth) * .1 / float32(cols+1))
				yPad := int32(float32(winHeight) * .1 / float32(rows+1))
				x += xPad * (int32(xi) + 1)
				y += yPad * (int32(yi) + 1)
				rect := sdl.Rect{x, y, int32(picWidth), int32(picHeight)}
				button := NewImageButton(renderer, tex, rect, sdl.Color{255, 255, 255, 0})
				buttons[pixelsAndIndex.index] = button
			}
		default:

		}
		renderer.Clear()
		for _, button := range buttons {
			if button != nil {
				button.Update(mouseState)
				if button.WasLeftClicked {
					button.IsSelected = !button.IsSelected
				}
				button.Draw(renderer)
			}
		}
		renderer.Present()

		elapsedTime = float32(time.Since(frameStart).Seconds()) * 1000
		//fmt.Println("ms per frame:", elapsedTime)
		if elapsedTime < 5 { // 5 milliseconds
			sdl.Delay(5 - uint32(elapsedTime))
			elapsedTime = float32(time.Since(frameStart).Seconds() * 1000)
		}
	}
}
