package main

import (
	"fmt"
	"gameswithgo/gogl"
	"github.com/go-gl/gl/v3.3-core/gl"
	"github.com/veandco/go-sdl2/sdl"
)

// https://learnopengl.com/ - nice OpenGL learning source (requires C knowledge)

const winWidth = 1280
const winHeight = 720

func main() {
	err := sdl.Init(sdl.INIT_EVERYTHING)
	if err != nil {
		panic(err)
	}
	defer sdl.Quit()

	window, err := sdl.CreateWindow("Hello triangle!", 200, 200, winWidth, winHeight, sdl.WINDOW_OPENGL)
	if err != nil {
		panic(err)
	}

	_, err = window.GLCreateContext()
	if err != nil {
		panic(err)
	}
	defer window.Destroy()

	gl.Init()

	fmt.Println(gogl.GetVersion())

	shaderProgram, err := gogl.NewShader("hellotriangle/shaders/hello.vert", "hellotriangle/shaders/hello.frag")
	if err != nil {
		panic(err)
	}

	vertices := []float32{
		0.5, 0.5, 0.0, 1.0, 1.0,
		0.5, -0.5, 0.0, 1.0, 0.0,
		-0.5, -0.5, 0.0, 0.0, 0.0,
		-0.5, 0.5, 0.0, 0.0, 1.0}

	indices := []uint32 {
		0,1,3, // triangle 1
		1,2,3, // triangle 2
	}

	gogl.GenBindBuffer(gl.ARRAY_BUFFER)
	VAO := gogl.GenBindVertexArray()
	gogl.BufferDataFloat(gl.ARRAY_BUFFER, vertices, gl.STATIC_DRAW)
	gogl.GenBindBuffer(gl.ELEMENT_ARRAY_BUFFER)
	gogl.BufferDataInt(gl.ELEMENT_ARRAY_BUFFER, indices, gl.STATIC_DRAW)

	gl.VertexAttribPointer(0, 3, gl.FLOAT, false, 5*4, nil)
	gl.EnableVertexAttribArray(0)
	gl.VertexAttribPointer(2,2,gl.FLOAT, false,5*4, gl.PtrOffset(2*4))
	gogl.UnbindVertexArray()

	for {
		for event := sdl.PollEvent(); event != nil; event = sdl.PollEvent() {
			switch event.(type) {
			case *sdl.QuitEvent:
				return
			}
		}
		gl.ClearColor(0, 0, 0, 0)
		gl.Clear(gl.COLOR_BUFFER_BIT)

		shaderProgram.Use()
		gogl.BindVertexArray(VAO)
		gl.DrawElements(gl.TRIANGLES, 6, gl.UNSIGNED_INT,gl.PtrOffset(0))

		window.GLSwap()

		shaderProgram.CheckShadersForChanges()
	}
}
