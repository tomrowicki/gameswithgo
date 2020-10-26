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

	shaderProgram, err := gogl.CreateProgram("hellotriangle/shaders/hello.vert", "hellotriangle/shaders/hello.frag")
	if err != nil {
		panic(err)
	}

	vertices := []float32{
		-0.5, -0.5, 0.0,
		0.5, -0.5, 0.0,
		0.0, 0.5, 0.0}

	gogl.GenBindBuffer(gl.ARRAY_BUFFER)
	VAO := gogl.GenBindVertexArray()

	gogl.BufferDataFloat(gl.ARRAY_BUFFER, vertices, gl.STATIC_DRAW)

	gl.VertexAttribPointer(0, 3, gl.FLOAT, false, 3*4, nil)
	gl.EnableVertexAttribArray(0)
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

		gogl.UseProgram(shaderProgram)
		gogl.BindVertexArray(VAO)
		gl.DrawArrays(gl.TRIANGLES, 0, 3)

		window.GLSwap()

		gogl.CheckShadersForChanges()
	}
}
