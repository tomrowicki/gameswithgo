package gogl

import (
	"errors"
	"fmt"
	"github.com/go-gl/gl/v3.3-core/gl"
	"io/ioutil"
	"strings"
	"time"
)

type ShaderID uint32
type ProgramID uint32
type BufferID uint32

func GetVersion() string {
	return gl.GoStr(gl.GetString(gl.VERSION))
}

type programInfo struct {
	vertPath string
	fragPath string

	modified time.Time
}

func LoadShader(path string, shaderType uint32) (ShaderID, error) {
	shaderFile, err := ioutil.ReadFile(path)
	if err != nil {
		return 0, err
	}
	shaderFileStr := string(shaderFile)
	shaderId, err := CreateShader(shaderFileStr, shaderType)
	if err != nil {
		return 0, err
	}

	//file, err := os.Stat(path)
	//if err != nil {
	//	panic(err)
	//}
	//modTime := file.ModTime()
	//loadedShaders = append(loadedShaders, programInfo{path,modTime})
	return shaderId, nil
}

func CreateShader(shaderSource string, shaderType uint32) (ShaderID, error) {
	shaderId := gl.CreateShader(shaderType)

	//The error around halfway with shader compiling is due to not adding a null terminator to the shader source strings. You can fix this by adding:
	//+ "\x00"
	//after the final back tick on the shader strings.
	shaderSource = shaderSource + "\x00"
	csource, free := gl.Strs(shaderSource)
	gl.ShaderSource(shaderId, 1, csource, nil)
	free()
	gl.CompileShader(shaderId)
	var status int32
	gl.GetShaderiv(shaderId, gl.COMPILE_STATUS, &status)
	if status == gl.FALSE { // compilation did not succeed...
		var logLength int32
		gl.GetShaderiv(shaderId, gl.INFO_LOG_LENGTH, &logLength)
		log := strings.Repeat("\x00", int(logLength+1))
		gl.GetShaderInfoLog(shaderId, logLength, nil, gl.Str(log))
		fmt.Println("failed to compile shader:\n" + log)
		return 0, errors.New("failed to compile shader: " + log)
	}
	return ShaderID(shaderId), nil
}

func CreateProgram(vertPath string, fragPath string) (ProgramID, error) {
	vert, err := LoadShader(vertPath, gl.VERTEX_SHADER)
	if err != nil {
		return 0, err
	}
	frag, err := LoadShader(fragPath, gl.FRAGMENT_SHADER)
	if err != nil {
		return 0, err
	}
	shaderProgram := gl.CreateProgram()
	gl.AttachShader(shaderProgram, uint32(vert))
	gl.AttachShader(shaderProgram, uint32(frag))
	gl.LinkProgram(shaderProgram)
	var success int32
	gl.GetProgramiv(shaderProgram, gl.LINK_STATUS, &success)

	if success == gl.FALSE {
		var logLength int32
		gl.GetProgramiv(shaderProgram, gl.INFO_LOG_LENGTH, &logLength)
		log := strings.Repeat("\x00", int(logLength+1))
		gl.GetProgramInfoLog(shaderProgram, logLength, nil, gl.Str(log))
		return 0, errors.New("failed to link program:\n" + log)
	}
	gl.DeleteShader(uint32(vert))
	gl.DeleteShader(uint32(frag))

	// TODO
	//file, err := os.Stat(path)
	//if err != nil {
	//	panic(err)
	//}
	//modTime := file.ModTime()
	//loadedShaders = append(loadedShaders, programInfo{path,modTime})

	return ProgramID(shaderProgram), nil
}

func GenBindBuffer(target uint32) BufferID {
	var VBO uint32 // Vertex Buffer Object
	gl.GenBuffers(1, &VBO)
	gl.BindBuffer(target, VBO)
	return BufferID(target)
}

func GenBindVertexArray() BufferID {
	var VAO uint32
	gl.GenVertexArrays(1, &VAO)
	gl.BindVertexArray(VAO)
	return BufferID(VAO)
}

func GenEBO() BufferID {
	var EBO uint32
	gl.GenBuffers(1, &EBO)
	return BufferID(EBO)
}

func BindVertexArray(id BufferID) {
	gl.BindVertexArray(uint32(id))
}

func BufferDataFloat(target uint32, data []float32, usage uint32) {
	gl.BufferData(target, len(data)*4, gl.Ptr(data), usage) // 4, as there are 4 bytes making a float32 value
}

func BufferDataInt(target uint32, data []uint32, usage uint32) {
	gl.BufferData(target, len(data)*4, gl.Ptr(data), usage) // 4, as there are 4 bytes making a float32 value
}

func UnbindVertexArray() {
	gl.BindVertexArray(0)
}

func UseProgram(id ProgramID) {
	gl.UseProgram(uint32(id))
}
