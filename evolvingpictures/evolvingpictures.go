package main

import (
	"fmt"
	. "gameswithgo/evolvingpictures/apt"
	. "gameswithgo/evolvingpictures/gui"
	"github.com/veandco/go-sdl2/sdl"
	"io/ioutil"
	"math/rand"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

var winWidth, winHeight = 800, 600
var rows, cols, numPics = 3, 3, rows * cols

type pixelResult struct {
	pixels []byte
	index  int
}

type guiState struct {
	zoom      bool
	zoomImage *sdl.Texture
	zoomTree  *picture
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
	return "( Picture\n" + p.r.String() + "\n" + p.g.String() + "\n" + p.b.String() + " )"
}

func NewPicture() *picture {
	p := &picture{}

	p.r = GetRandomNode()
	p.g = GetRandomNode()
	p.b = GetRandomNode()

	num := rand.Intn(1) + 5
	for i := 0; i < num; i++ {
		p.r.AddRandom(GetRandomNode())
	}

	num = rand.Intn(1) + 5
	for i := 0; i < num; i++ {
		p.g.AddRandom(GetRandomNode())
	}

	num = rand.Intn(1) + 5
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

func (p *picture) pickRandomColor() Node {
	r := rand.Intn(3)
	switch r {
	case 0:
		return p.r
	case 1:
		return p.g
	case 2:
		return p.b
	default:
		panic("Wrong random value!")
	}
}

func cross(a, b *picture) *picture {
	aCopy := &picture{CopyTree(a.r, nil), CopyTree(a.g, nil), CopyTree(a.b, nil)}
	aColor := aCopy.pickRandomColor()
	bColor := b.pickRandomColor()

	aIndex := rand.Intn(aColor.NodeCount())
	aNode, _ := GetNthNode(aColor, aIndex, 0)

	bIndex := rand.Intn(bColor.NodeCount())
	bNode, _ := GetNthNode(bColor, bIndex, 0)
	bNodeCopy := CopyTree(bNode, bNode.GetParent())

	ReplaceNode(aNode, bNodeCopy)
	return aCopy
}

func evolve(survivors []*picture) []*picture {
	newPics := make([]*picture, numPics)
	i := 0
	for i < len(survivors) {
		a := survivors[i]
		b := survivors[rand.Intn(len(survivors))]
		newPics[i] = cross(a, b)
		i++
	}

	for i < len(newPics) {
		a := survivors[rand.Intn(len(survivors))]
		b := survivors[rand.Intn(len(survivors))]
		newPics[i] = cross(a, b)
		i++
	}

	for _, pic := range newPics {
		r := rand.Intn(4)
		for i := 0; i < r; i++ {
			pic.mutate()
		}
	}

	return newPics
}

func (p *picture) mutate() {
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

	if nodeToMutate == nil {
		panic("Node to mutate is nil!")
	}

	count := nodeToMutate.NodeCount()
	r = rand.Intn(count)
	//fmt.Println("mutation count and rand:", count, r)
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

func aptToPixels(pic *picture, w, h int) []byte {
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

func saveTree(p *picture) {
	files, err := ioutil.ReadDir("./evolvingpictures")
	if err != nil {
		fmt.Println(err)
	}
	biggestNumber := 0
	for _, f := range files {
		name := f.Name()
		if strings.HasSuffix(name, ".apt") {
			numberStr := strings.TrimSuffix(name, ".apt")
			num, err := strconv.Atoi(numberStr)
			if err == nil {
				if num > biggestNumber {
					biggestNumber = num
				}
			}

		}
	}
	saveName := strconv.Itoa(biggestNumber+1) + ".apt"
	file, err := os.Create(filepath.Join("evolvingpictures", saveName))
	if err != nil {
		fmt.Println(err)
	}
	defer file.Close()
	fmt.Fprintf(file, p.String())
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
	// generating new pictures
	for i := range picTrees {
		picTrees[i] = NewPicture()
	}

	picWidth := int(float32(winWidth/cols) * float32(.9)) // leaving some gap between pics (remaining 0.1)
	picHeight := int(float32(winHeight/rows) * float32(.8))

	pixelsChannel := make(chan pixelResult, numPics)
	buttons := make([]*ImageButton, numPics)

	evolveButtonTex := GetSinglePixelTex(renderer, sdl.Color{255, 255, 255, 0})
	evolveRect := sdl.Rect{
		int32(float32(winWidth)/2 - float32(picWidth)/2),
		int32(float32(winHeight) - (float32(winHeight) * .10)),
		int32(picWidth), int32(float32(winHeight) * .08)}
	evolveButton := NewImageButton(renderer, evolveButtonTex, evolveRect, sdl.Color{255, 255, 255, 0})

	// applying random functions and translating to pixels
	for i := range picTrees {
		go func(i int) {
			pixels := aptToPixels(picTrees[i], picWidth*2, picHeight*2)
			pixelsChannel <- pixelResult{pixels, i}
		}(i)
	}

	keyboardState := sdl.GetKeyboardState()
	prevKeyboardState := make([]uint8, len(keyboardState))

	mouseState := GetMouseState()
	state := guiState{false, nil, nil}

	args := os.Args
	if len(args) > 1 {
		fileBytes, err := ioutil.ReadFile(args[1])
		if err != nil {
			panic(err)
		}
		fileStr := string(fileBytes)
		pictureNode := BeginLexing(fileStr)
		p := &picture{pictureNode.GetChildren()[0], pictureNode.GetChildren()[1], pictureNode.GetChildren()[2]}
		pixels := aptToPixels(p, winWidth, winHeight)
		tex := pixelsToTexture(renderer, pixels, winWidth, winHeight)
		state.zoom = true
		state.zoomImage = tex
		state.zoomTree = p
	}

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

		if !state.zoom {

			select {
			case pixelsAndIndex, ok := <-pixelsChannel:
				if ok {
					tex := pixelsToTexture(renderer, pixelsAndIndex.pixels, picWidth*2, picHeight*2)
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
			for i, button := range buttons {
				if button != nil {
					button.Update(mouseState)
					if button.WasLeftClicked {
						button.IsSelected = !button.IsSelected
					} else if button.WasRightClicked {
						//fmt.Println(picTrees[i])
						zoomPixels := aptToPixels(picTrees[i], winWidth*2, winHeight*2)
						zoomTex := pixelsToTexture(renderer, zoomPixels, winWidth*2, winHeight*2)
						state.zoomImage = zoomTex
						state.zoomTree = picTrees[i]
						state.zoom = true
					}
					button.Draw(renderer)
				}
			}
			evolveButton.Update(mouseState)
			if evolveButton.WasLeftClicked {
				selectedPictures := make([]*picture, 0)
				for i, button := range buttons {
					if button.IsSelected {
						selectedPictures = append(selectedPictures, picTrees[i])
					}
				}
				if len(selectedPictures) != 0 {
					for i := range buttons {
						buttons[i] = nil
					}
					picTrees = evolve(selectedPictures)
					for i := range picTrees {
						go func(i int) {
							pixels := aptToPixels(picTrees[i], picWidth*2, picHeight*2)
							pixelsChannel <- pixelResult{pixels, i}
						}(i)
					}
				}
			}
			evolveButton.Draw(renderer)
		} else {
			if !mouseState.RightButton && mouseState.PrevRightButton {
				state.zoom = false
			}
			if keyboardState[sdl.SCANCODE_S] == 0 && prevKeyboardState[sdl.SCANCODE_S] != 0 {
				saveTree(state.zoomTree)
			}
			renderer.Copy(state.zoomImage, nil, nil)
		}

		renderer.Present()
		for i, v := range keyboardState {
			prevKeyboardState[i] = v
		}

		elapsedTime = float32(time.Since(frameStart).Seconds()) * 1000
		//fmt.Println("ms per frame:", elapsedTime)
		if elapsedTime < 5 { // 5 milliseconds
			sdl.Delay(5 - uint32(elapsedTime))
			elapsedTime = float32(time.Since(frameStart).Seconds() * 1000)
		}
	}
}
