package gogl

import (
	"fmt"
	"os"
	"time"
)

type Shader struct {
	id               ProgramID
	vertexPath       string
	fragmentPath     string
	vertexModified   time.Time
	fragmentModified time.Time
}

func NewShader(vertexPath, fragmentPath string) (*Shader, error) {
	id, err := CreateProgram(vertexPath, fragmentPath)
	if err != nil {
		return nil, err
	}
	result := &Shader{id, vertexPath,fragmentPath,
		getModifiedTime(vertexPath), getModifiedTime(fragmentPath)}
	return  result, nil
}

func (shader *Shader) Use()  {
	UseProgram(shader.id)
}

func getModifiedTime(filePath string) time.Time {
	file, err := os.Stat(filePath)
	if err != nil {
		panic(err)
	}
	return file.ModTime()
}

func (shader *Shader) CheckShadersForChanges() {
		vertexModTime := getModifiedTime(shader.vertexPath)
		fragModTime := getModifiedTime(shader.fragmentPath)
		// check if greater than?
		if !vertexModTime.Equal(shader.vertexModified) || !fragModTime.Equal(shader.fragmentModified) {
			id, err := CreateProgram(shader.vertexPath, shader.fragmentPath)
			if err != nil {
				fmt.Println(err)
			} else {
				//gl.DeleteProgram(uint32(shader.id)) // prevents from changing shader go during runtime
				shader.id = id
			}
		}
}
