package main

import (
	"fmt"
	"gameswithgo/noise"
	"gameswithgo/vec3"
	"github.com/veandco/go-sdl2/sdl"
	"image/png"
	"math"
	"math/rand"
	"os"
	"sort"
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

type balloon struct {
	tex  *sdl.Texture
	pos  vec3.Vector3
	dir  vec3.Vector3
	w, h int

	exploding         bool
	exploded          bool
	explosionStart    time.Time
	explosionInterval float32
	explosionTexture  *sdl.Texture
}

func newBalloon(tex *sdl.Texture, pos, dir vec3.Vector3, explosionTexture *sdl.Texture) *balloon {
	_, _, w, h, err := tex.Query()
	if err != nil {
		panic(err)
	}
	return &balloon{tex, pos, dir, int(w), int(h),
		false, false, time.Now(),
		50, explosionTexture}
}

type balloonArray []*balloon

func (balloons balloonArray) Len() int {
	return len(balloons)
}

func (balloons balloonArray) Swap(i, j int) {
	balloons[i], balloons[j] = balloons[j], balloons[i]
}

func (balloons balloonArray) Less(i, j int) bool {
	diff := balloons[i].pos.Z - balloons[j].pos.Z
	return diff < -0.5
}

func (balloon *balloon) getScale() float32 {
	return (balloon.pos.Z/200 + 1) / 2
}

func (balloon *balloon) getCircle() (x, y, r float32) {
	x = balloon.pos.X
	y = balloon.pos.Y - 30*balloon.getScale()
	r = float32(balloon.w/2) * balloon.getScale()
	return x, y, r
}

func (balloon *balloon) update(elapsedTime float32, currMouseState, prevMouseState mouseState,
	audioState *audioState) {

	numAnimations := 16
	animationElapsed := float32(time.Since(balloon.explosionStart).Seconds() * 1000)
	// counting from bottom right
	animationIndex := numAnimations - 1 - int(animationElapsed/balloon.explosionInterval)
	if animationIndex < 0 {
		balloon.exploding = false
		balloon.exploded = true
	}

	// registering lmb click
	if !prevMouseState.leftButton && currMouseState.leftButton {
		x, y, r := balloon.getCircle()
		mouseX := currMouseState.x
		mouseY := currMouseState.y
		xDiff := float32(mouseX) - x
		yDiff := float32(mouseY) - y
		dist := float32(math.Sqrt(float64(xDiff*xDiff + yDiff*yDiff)))
		if dist < r {
			sdl.ClearQueuedAudio(audioState.deviceID)
			sdl.QueueAudio(audioState.deviceID, audioState.explosionBytes)
			// the most quirky way to "push play"
			sdl.PauseAudioDevice(audioState.deviceID, false)
			balloon.exploding = true
			balloon.explosionStart = time.Now()
		}
	}

	// calculating new position
	p := vec3.Add(balloon.pos, vec3.Mult(balloon.dir, elapsedTime))

	if p.X < 0 || p.X > float32(winWidth) {
		balloon.dir.X = -balloon.dir.X
	}

	if p.Y < 0 || p.Y > float32(winHeight) {
		balloon.dir.Y = -balloon.dir.Y
	}

	if p.Z < 0 || p.Z > float32(winDepth) {
		balloon.dir.Z = -balloon.dir.Z
	}

	balloon.pos = vec3.Add(balloon.pos, vec3.Mult(balloon.dir, elapsedTime))
}

func (balloon *balloon) draw(renderer *sdl.Renderer) {
	scale := balloon.getScale()
	newW := int32(float32(balloon.w) * scale)
	newH := int32(float32(balloon.h) * scale)
	x := int32(balloon.pos.X - float32(newW)/2)
	y := int32(balloon.pos.Y - float32(newH)/2)
	rect := &sdl.Rect{x, y, newW, newH}
	renderer.Copy(balloon.tex, nil, rect)

	if balloon.exploding {
		numAnimations := 16
		animationElapsed := float32(time.Since(balloon.explosionStart).Seconds() * 1000)
		// counting from bottom right
		animationIndex := numAnimations - 1 - int(animationElapsed/balloon.explosionInterval)
		animationX := animationIndex % 4
		animationY := 64 * ((animationIndex - animationX) / 4)
		animationX *= 64
		animationRect := &sdl.Rect{int32(animationX), int32(animationY), 64, 64}
		// enlarging explosions
		rect.X -= rect.W / 2
		rect.Y -= rect.H / 2
		rect.W *= 2
		rect.H *= 2
		err := renderer.Copy(balloon.explosionTexture, animationRect, rect)
		if err != nil {
			panic(err)
		}
	}
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

func imgFileToTexture(renderer *sdl.Renderer, filename string) *sdl.Texture {
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
	tex := pixelsToTexture(renderer, pixels, w, h)
	err = tex.SetBlendMode(sdl.BLENDMODE_BLEND)
	if err != nil {
		panic(err)
	}

	return tex
}

func loadBalloons(renderer *sdl.Renderer, numAllBalloons int) []*balloon {
	explosionTexture := imgFileToTexture(renderer, "balloons2/explosion.png")

	balloonStrs := []string{"balloons/balloon_red.png", "balloons/balloon_green.png", "balloons/balloon_blue.png"}
	balloonTextures := make([]*sdl.Texture, len(balloonStrs))

	for i, bstr := range balloonStrs {
		balloonTextures[i] = imgFileToTexture(renderer, bstr)
	}
	balloons := make([]*balloon, numAllBalloons)
	for i := range balloons {
		tex := balloonTextures[i%3]
		pos := vec3.Vector3{rand.Float32() * winWidth, rand.Float32() * winHeight, rand.Float32() * winDepth}
		dir := vec3.Vector3{rand.Float32()*.5 - .25, rand.Float32()*.5 - .25, rand.Float32()*.25 - .25/2}

		balloons[i] = newBalloon(tex, pos, dir, explosionTexture)
	}
	return balloons
}

func clear(pixels []byte) {
	for i := range pixels {
		pixels[i] = 0
	}
}

func lerp(b1, b2 byte, pct float32) byte {
	return byte(float32(b1) + pct*(float32(b2)-float32(b1)))
}

func colorLerp(c1, c2 rgba, pct float32) rgba {
	return rgba{lerp(c1.r, c2.r, pct), lerp(c1.g, c2.g, pct), lerp(c1.b, c2.b, pct)}
}

func getGradient(c1, c2 rgba) []rgba {
	result := make([]rgba, 256)
	for i := range result {
		pct := float32(i) / float32(255)
		result[i] = colorLerp(c1, c2, pct)
	}
	return result
}

func getDualGradient(c1, c2, c3, c4 rgba) []rgba {
	result := make([]rgba, 256)
	for i := range result {
		pct := float32(i) / float32(255)
		if pct < 0.5 {
			result[i] = colorLerp(c1, c2, pct*float32(2))
		} else {
			result[i] = colorLerp(c3, c4, pct*float32(1.5)-float32(0.5))
		}
	}
	return result
}

func clamp(min, max, v int) int {
	if v < min {
		v = min
	} else if v > max {
		v = max
	}
	return v
}

func rescaleAndDraw(noise []float32, min, max float32, gradient []rgba, w, h int) []byte {
	result := make([]byte, w*h*4)
	scale := 255.0 / (max - min)
	offset := min * scale

	for i := range noise {
		noise[i] = noise[i]*scale - offset

		c := gradient[clamp(0, 255, int(noise[i]))]
		p := i * 4
		result[p] = c.r
		result[p+1] = c.g
		result[p+2] = c.b
	}
	return result
}

func main() {
	sdl.InitSubSystem(sdl.INIT_AUDIO)
	window, err := sdl.CreateWindow("Exploding Balloons", sdl.WINDOWPOS_UNDEFINED, sdl.WINDOWPOS_UNDEFINED,
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

	//var audioSpec sdl.AudioSpec
	explosionBytes, audioSpec := sdl.LoadWAV("balloons2/explode.wav")
	audioId, err := sdl.OpenAudioDevice("", false, audioSpec, nil, 0)
	if err != nil {
		panic(err)
	}
	defer sdl.FreeWAV(explosionBytes)

	audioState := audioState{explosionBytes, audioId, audioSpec}

	// bilinear filtering through SDL
	sdl.SetHint(sdl.HINT_RENDER_SCALE_QUALITY, "1")

	cloudNoise, min, max := noise.MakeNoise(noise.FBM, .009, .5, 3, 3, winWidth, winHeight)
	cloudGradient := getGradient(rgba{0, 0, 255}, rgba{255, 255, 255})
	cloudPixels := rescaleAndDraw(cloudNoise, min, max, cloudGradient, winWidth, winHeight)
	cloudTexture := pixelsToTexture(renderer, cloudPixels, winWidth, winHeight)

	balloons := loadBalloons(renderer, 10)
	var elapsedTime float32
	currentMouseState := getMouseState()
	prevMouseState := currentMouseState

	// GAME LOOP
	for {
		frameStart := time.Now()
		for event := sdl.PollEvent(); event != nil; event = sdl.PollEvent() {
			switch event.(type) {
			case *sdl.QuitEvent:
				return
			}
		}
		currentMouseState = getMouseState()

		renderer.Copy(cloudTexture, nil, nil)

		for _, balloon := range balloons {
			balloon.update(elapsedTime, currentMouseState, prevMouseState, &audioState)
		}
		sort.Stable(balloonArray(balloons))
		for _, balloon := range balloons {
			balloon.draw(renderer)
		}

		renderer.Present()

		elapsedTime = float32(time.Since(frameStart).Seconds()) * 1000
		//fmt.Println("ms per frame:", elapsedTime)
		if elapsedTime < 5 { // 5 milliseconds
			sdl.Delay(5 - uint32(elapsedTime))
			elapsedTime = float32(time.Since(frameStart).Seconds() * 1000)
		}
		prevMouseState = currentMouseState
	}
}
