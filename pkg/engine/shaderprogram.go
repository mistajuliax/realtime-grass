package engine

import (
	"fmt"
	"io/ioutil"
	"strings"

	"github.com/go-gl/gl/v4.3-core/gl"
	"github.com/go-gl/mathgl/mgl32"
)

type ShaderProgram struct {
	programHandle uint32
	renderables   []Renderable
}

func MakeProgram(vertexShaderPath, fragmentShaderPath string) (ShaderProgram, error) {
	// loads files
	vertexShaderSource, err := loadFile(vertexShaderPath)
	if err != nil {
		return ShaderProgram{}, fmt.Errorf("Error on: %v\n%v", vertexShaderPath, err)
	}
	fragmentShaderSource, err := loadFile(fragmentShaderPath)
	if err != nil {
		return ShaderProgram{}, fmt.Errorf("Error on: %v\n%v", fragmentShaderPath, err)
	}

	// compile shaders
	vertexShader, err := compileShader(vertexShaderSource, gl.VERTEX_SHADER)
	if err != nil {
		return ShaderProgram{}, fmt.Errorf("Error on: %v\n%v", vertexShaderPath, err)
	}
	fragmentShader, err := compileShader(fragmentShaderSource, gl.FRAGMENT_SHADER)
	if err != nil {
		return ShaderProgram{}, fmt.Errorf("Error on: %v\n%v", fragmentShaderPath, err)
	}

	// create and link program
	program := gl.CreateProgram()
	gl.AttachShader(program, vertexShader)
	gl.AttachShader(program, fragmentShader)
	gl.LinkProgram(program)

	var status int32
	gl.GetProgramiv(program, gl.LINK_STATUS, &status)
	if status == gl.FALSE {
		var logLength int32
		gl.GetProgramiv(program, gl.INFO_LOG_LENGTH, &logLength)

		log := strings.Repeat("\x00", int(logLength+1))
		gl.GetProgramInfoLog(program, logLength, nil, gl.Str(log))

		return ShaderProgram{}, fmt.Errorf("failed to link program: %v", log)
	}

	gl.DetachShader(program, vertexShader)
	gl.DetachShader(program, fragmentShader)
	gl.DeleteShader(vertexShader)
	gl.DeleteShader(fragmentShader)

	shaderProgram := ShaderProgram{program, nil}
	return shaderProgram, nil
}

func MakeGeomProgram(vertexShaderPath, geometryShaderPath, fragmentShaderPath string) (ShaderProgram, error) {
	// loads files
	vertexShaderSource, err := loadFile(vertexShaderPath)
	if err != nil {
		return ShaderProgram{}, fmt.Errorf("Error on: %v\n%v", vertexShaderPath, err)
	}
	geometryShaderSource, err := loadFile(geometryShaderPath)
	if err != nil {
		return ShaderProgram{}, fmt.Errorf("Error on: %v\n%v", geometryShaderPath, err)
	}
	fragmentShaderSource, err := loadFile(fragmentShaderPath)
	if err != nil {
		return ShaderProgram{}, fmt.Errorf("Error on: %v\n%v", fragmentShaderPath, err)
	}

	// compile shaders
	vertexShader, err := compileShader(vertexShaderSource, gl.VERTEX_SHADER)
	if err != nil {
		return ShaderProgram{}, fmt.Errorf("Error on: %v\n%v", vertexShaderPath, err)
	}
	geometryShader, err := compileShader(geometryShaderSource, gl.GEOMETRY_SHADER)
	if err != nil {
		return ShaderProgram{}, fmt.Errorf("Error on: %v\n%v", geometryShaderPath, err)
	}
	fragmentShader, err := compileShader(fragmentShaderSource, gl.FRAGMENT_SHADER)
	if err != nil {
		return ShaderProgram{}, fmt.Errorf("Error on: %v\n%v", fragmentShaderPath, err)
	}

	// create and link program
	program := gl.CreateProgram()
	gl.AttachShader(program, vertexShader)
	gl.AttachShader(program, geometryShader)
	gl.AttachShader(program, fragmentShader)
	gl.LinkProgram(program)

	var status int32
	gl.GetProgramiv(program, gl.LINK_STATUS, &status)
	if status == gl.FALSE {
		var logLength int32
		gl.GetProgramiv(program, gl.INFO_LOG_LENGTH, &logLength)

		log := strings.Repeat("\x00", int(logLength+1))
		gl.GetProgramInfoLog(program, logLength, nil, gl.Str(log))

		return ShaderProgram{}, fmt.Errorf("failed to link program: %v", log)
	}

	gl.DetachShader(program, vertexShader)
	gl.DetachShader(program, geometryShader)
	gl.DetachShader(program, fragmentShader)
	gl.DeleteShader(vertexShader)
	gl.DeleteShader(geometryShader)
	gl.DeleteShader(fragmentShader)

	shaderProgram := ShaderProgram{program, nil}
	return shaderProgram, nil
}

func MakeComputeProgram(computeShaderPath string) (ShaderProgram, error) {
	// loads files
	computeShaderSource, err := loadFile(computeShaderPath)
	if err != nil {
		return ShaderProgram{}, fmt.Errorf("Error on: %v\n%v", computeShaderPath, err)
	}

	// compile shaders
	computeShader, err := compileShader(computeShaderSource, gl.COMPUTE_SHADER)
	if err != nil {
		return ShaderProgram{}, fmt.Errorf("Error on: %v\n%v", computeShaderPath, err)
	}

	// create and link program
	program := gl.CreateProgram()
	gl.AttachShader(program, computeShader)
	gl.LinkProgram(program)

	// check status
	var status int32
	gl.GetProgramiv(program, gl.LINK_STATUS, &status)
	if status == gl.FALSE {
		var logLength int32
		gl.GetProgramiv(program, gl.INFO_LOG_LENGTH, &logLength)

		log := strings.Repeat("\x00", int(logLength+1))
		gl.GetProgramInfoLog(program, logLength, nil, gl.Str(log))

		return ShaderProgram{}, fmt.Errorf("failed to link program: %v", log)
	}

	// cleanup shader objects
	gl.DetachShader(program, computeShader)
	gl.DeleteShader(computeShader)

	shaderProgram := ShaderProgram{program, nil}
	return shaderProgram, nil
}

func (shaderProgram *ShaderProgram) AddRenderable(renderable Renderable) {
	renderable.Build(shaderProgram.programHandle)
	shaderProgram.renderables = append(shaderProgram.renderables, renderable)
}
func (ShaderProgram *ShaderProgram) RemoveAllRenderables() {
	// TODO: should renderables be deleted?
	ShaderProgram.renderables = nil
}
func (shaderProgram *ShaderProgram) Render() {
	for _, renderable := range shaderProgram.renderables {
		renderable.Render()
	}
}
func (shaderProgram *ShaderProgram) RenderInstanced(instancecount int32) {
	for _, renderable := range shaderProgram.renderables {
		renderable.RenderInstanced(instancecount)
	}
}
func (ShaderProgram *ShaderProgram) Compute(numgroupsx, numgroupsy, numgroupsz uint32) {
	gl.DispatchCompute(numgroupsx, numgroupsy, numgroupsz)
}
func (shaderProgram *ShaderProgram) Use() {
	gl.UseProgram(shaderProgram.programHandle)
}
func (shaderProgram *ShaderProgram) Delete() {
	gl.DeleteProgram(shaderProgram.programHandle)
	shaderProgram.renderables = nil
}

func (shaderProgram *ShaderProgram) UpdateInt32(uniformName string, i32 int32) {
	location := gl.GetUniformLocation(shaderProgram.programHandle, gl.Str(uniformName+"\x00"))
	if location != -1 {
		gl.Uniform1i(location, i32)
	}
}
func (shaderProgram *ShaderProgram) UpdateFloat32(uniformName string, f32 float32) {
	location := gl.GetUniformLocation(shaderProgram.programHandle, gl.Str(uniformName+"\x00"))
	if location != -1 {
		gl.Uniform1f(location, f32)
	}
}
func (shaderProgram *ShaderProgram) UpdateVec2(uniformName string, vec2 mgl32.Vec2) {
	location := gl.GetUniformLocation(shaderProgram.programHandle, gl.Str(uniformName+"\x00"))
	if location != -1 {
		gl.Uniform2fv(location, 1, &vec2[0])
	}
}
func (shaderProgram *ShaderProgram) UpdateVec3(uniformName string, vec3 mgl32.Vec3) {
	location := gl.GetUniformLocation(shaderProgram.programHandle, gl.Str(uniformName+"\x00"))
	if location != -1 {
		gl.Uniform3fv(location, 1, &vec3[0])
	}
}
func (shaderProgram *ShaderProgram) UpdateMat4(uniformName string, mat mgl32.Mat4) {
	location := gl.GetUniformLocation(shaderProgram.programHandle, gl.Str(uniformName+"\x00"))
	if location != -1 {
		gl.UniformMatrix4fv(location, 1, false, &mat[0])
	}
}

func (shaderProgram *ShaderProgram) GetHandle() uint32 {
	return shaderProgram.programHandle
}

func loadFile(filepath string) (string, error) {
	bytes, err := ioutil.ReadFile(filepath)

	if err != nil {
		return "", err
	}

	bytes = append(bytes, '\000')
	return string(bytes), nil
}
func compileShader(source string, shaderType uint32) (uint32, error) {
	shader := gl.CreateShader(shaderType)

	csources, free := gl.Strs(source)
	gl.ShaderSource(shader, 1, csources, nil)
	gl.CompileShader(shader)
	free()

	err := getGLError(shader, gl.COMPILE_STATUS)
	if err != nil {
		gl.DeleteShader(shader)
		return 0, fmt.Errorf("failed to compile\n'%v'\n%v", source, err)
	}

	return shader, nil
}
func getGLError(shader uint32, statusType int) error {
	var status int32
	gl.GetShaderiv(shader, gl.COMPILE_STATUS, &status)
	if status == gl.FALSE {
		var logLength int32
		gl.GetShaderiv(shader, gl.INFO_LOG_LENGTH, &logLength)
		log := strings.Repeat("\x00", int(logLength))
		gl.GetShaderInfoLog(shader, logLength, nil, gl.Str(log))

		return fmt.Errorf(log)
	}
	return nil
}
