// Package scene contains all main entities for rendering and/or interaction with the user.
package scene

import (
	"github.com/go-gl/mathgl/mgl32"

	"github.com/adrianderstroff/realtime-grass/pkg/engine"
)

// Sky renders a skybox around the scene.
type Sky struct {
	shader *engine.ShaderProgram
	skybox *engine.Skybox
	tex    *engine.Texture
}

// MakeSky constructs a Sky given the folder to the cube map textures.
func MakeSky(shaderpath, skypath string) (Sky, error) {
	// make skybox
	cubemappath := skypath + "day/"
	fileending := ".png"
	cubetex, err := engine.MakeCubeMapTexture(
		cubemappath+"left"+fileending,
		cubemappath+"right"+fileending,
		cubemappath+"top"+fileending,
		cubemappath+"bottom"+fileending,
		cubemappath+"front"+fileending,
		cubemappath+"back"+fileending,
	)
	if err != nil {
		panic(err)
	}
	skybox := engine.MakeSkybox(cubetex)

	// make skybox shader
	shader, err := engine.MakeProgram(shaderpath+"skybox/skybox.vert", shaderpath+"skybox/skybox.frag")
	if err != nil {
		panic(err)
	}
	shader.Use()
	shader.AddRenderable(skybox)

	return Sky{
		shader: &shader,
		skybox: &skybox,
		tex:    &cubetex,
	}, nil
}

// Render draws the skybox for the given view and projection matrices.
func (sky *Sky) Render(V, P *mgl32.Mat4) {
	sky.shader.Use()
	sky.shader.UpdateMat4("V", *V)
	sky.shader.UpdateMat4("P", *P)
	sky.shader.Render()
}
