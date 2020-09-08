package main

import (
	"fmt"
	. "gameswithgo/evolvingpictures/apt"
	"github.com/veandco/go-sdl2/sdl"
	"time"
)

const winWidth, winHeight, winDepth = 800, 600, 100

type audioState struct {
	explosionBytes []byte
	deviceID       sdl.AudioDeviceID
	audioSpec      *sdl.AudioSpec
}

type mouseState struct {
	leftButton  bool
	rightButton bool
	x, y        int
}

func getMouseState() mouseState {
	mouseX, mouseY, mouseButtonState := sdl.GetMouseState()
	leftButton := mouseButtonState & sdl.ButtonLMask() // using bitwise op
	rightButton := mouseButtonState & sdl.ButtonRMask()
	var result mouseState
	result.x = int(mouseX)
	result.y = int(mouseY)
	result.leftButton = !(leftButton == 0)
	result.rightButton = !(rightButton == 0)
	return result
}

type rgba struct {
	r, g, b byte
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

func APTToTexture(node1, node2 Node, w, h int, renderer *sdl.Renderer) *sdl.Texture {
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
			c := node1.Eval(x, y)
			c2 := node2.Eval(x, y)
			pixels[pixelIndex] = byte(c*scale - offset)
			pixelIndex++
			pixels[pixelIndex] = byte(c2*scale - offset)
			pixelIndex++
			pixels[pixelIndex] = 0//byte(c*scale - offset)
			pixelIndex++
			pixelIndex++ // skipping alpha
		}
	}
	return pixelsToTexture(renderer, pixels, w, h)
}

func main() {
	sdl.InitSubSystem(sdl.INIT_AUDIO)
	window, err := sdl.CreateWindow("Evolving Pictures", sdl.WINDOWPOS_UNDEFINED, sdl.WINDOWPOS_UNDEFINED,
		winWidth, winHeight, sdl.WINDOW_SHOWN)
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
	//currentMouseState := getMouseState()
	//prevMouseState := currentMouseState

	x := &OpX{}
	y := &OpY{}
	sine := &OpSin{}
	plus := &OpPlus{}

	sine.Child = x
	plus.LeftChild = sine
	plus.RightChild = y

	tex := APTToTexture(plus, sine, 800,600, renderer)

	// GAME LOOP
	for {
		frameStart := time.Now()
		for event := sdl.PollEvent(); event != nil; event = sdl.PollEvent() {
			switch event.(type) {
			case *sdl.QuitEvent:
				return
			}
		}
		//currentMouseState = getMouseState()

		renderer.Copy(tex, nil, nil)

		renderer.Present()

		elapsedTime = float32(time.Since(frameStart).Seconds()) * 1000
		//fmt.Println("ms per frame:", elapsedTime)
		if elapsedTime < 5 { // 5 milliseconds
			sdl.Delay(5 - uint32(elapsedTime))
			elapsedTime = float32(time.Since(frameStart).Seconds() * 1000)
		}
		//prevMouseState = currentMouseState
	}
}
