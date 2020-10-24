package main

import (
	"fmt"
	"github.com/go-gl/gl/v3.3-core/gl"
	"github.com/veandco/go-sdl2/sdl"
	"strings"
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
	version := gl.GoStr(gl.GetString(gl.VERSION))
	fmt.Println(version)

	vertexShaderSource :=
		`#version 330 core
		layout (location = 0) in vec3 aPos;

		void main()
		{
			gl_Position = vec4(aPos.x,aPos.y,aPos.z,1.0);
		}` + "\x00"

	//The error around halfway with shader compiling is due to not adding a null terminator to the shader source strings. You can fix this by adding:
	//+ "\x00"
	//after the final back tick on the shader strings.

	vertexShader := gl.CreateShader(gl.VERTEX_SHADER)
	csource, free := gl.Strs(vertexShaderSource)
	gl.ShaderSource(vertexShader, 1, csource, nil)
	free()
	gl.CompileShader(vertexShader)
	var status int32
	gl.GetShaderiv(vertexShader,gl.COMPILE_STATUS, &status)
	if status == gl.FALSE { // compilation did not succeed...
		var logLength int32
		gl.GetShaderiv(vertexShader,gl.INFO_LOG_LENGTH,&logLength)
		log := strings.Repeat("\x00", int(logLength+1)) // FIXME better way shown in YT comment
		gl.GetShaderInfoLog(vertexShader,logLength, nil,gl.Str(log))
		panic("failed to compile vertex shader:\n" + log)
	}

	fragmentShaderSource :=
		`# version 330 core
		out vec4 FragColor;
		
		void main()
		{
			FragColor = vec4(1.0f,0.5f,0.2f,1.0f);
		}` + "\x00"

	fragmentShader := gl.CreateShader(gl.FRAGMENT_SHADER)
	csource, free = gl.Strs(fragmentShaderSource)
	gl.ShaderSource(fragmentShader, 1, csource, nil)
	free()
	gl.CompileShader(fragmentShader)
	gl.GetShaderiv(fragmentShader,gl.COMPILE_STATUS, &status)
	if status == gl.FALSE { // compilation did not succeed...
		var logLength int32
		gl.GetShaderiv(fragmentShader,gl.INFO_LOG_LENGTH,&logLength)
		log := strings.Repeat("\x00", int(logLength+1))
		gl.GetShaderInfoLog(fragmentShader,logLength, nil,gl.Str(log))
		panic("failed to compile fragment shader:\n" + log)
	}

	shaderProgram := gl.CreateProgram()
	gl.AttachShader(shaderProgram, vertexShader)
	gl.AttachShader(shaderProgram, fragmentShader)
	gl.LinkProgram(shaderProgram)
	var success int32
	gl.GetProgramiv(shaderProgram, gl.LINK_STATUS, &success)
	if success == gl.FALSE {
		var logLength int32
		gl.GetProgramiv(shaderProgram,gl.INFO_LOG_LENGTH,&logLength)
		log := strings.Repeat("\x00", int(logLength+1))
		gl.GetProgramInfoLog(shaderProgram,logLength, nil,gl.Str(log))
		panic("failed to link program:\n" + log)
	}
	gl.DeleteShader(vertexShader)
	gl.DeleteShader(fragmentShader)


	vertices := []float32{
		-0.5, -0.5, 0.0,
		0.5, -0.5, 0.0,
		0.0, 0.5, 0.0}

	var VBO uint32 // Vertex Buffer Object
	gl.GenBuffers(1, &VBO)
	gl.BindBuffer(gl.ARRAY_BUFFER, VBO)
	var VAO uint32
	gl.GenVertexArrays(1, &VAO)
	gl.BindVertexArray(VAO)

	gl.BufferData(gl.ARRAY_BUFFER, len(vertices)*4, gl.Ptr(vertices), gl.STATIC_DRAW) // 4, as there are 4 bytes making a float32 value

	gl.VertexAttribPointer(0,3,gl.FLOAT,false,3*4,nil)
	gl.EnableVertexAttribArray(0)
	gl.BindVertexArray(0)

	for {
		for event := sdl.PollEvent(); event != nil; event = sdl.PollEvent() {
			switch event.(type) {
			case *sdl.QuitEvent:
				return
			}
		}
		gl.ClearColor(0,0,0,0)
		gl.Clear(gl.COLOR_BUFFER_BIT)

		gl.UseProgram(shaderProgram)
		gl.BindVertexArray(VAO)
		gl.DrawArrays(gl.TRIANGLES,0,3)

		window.GLSwap()
	}
}
