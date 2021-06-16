package giuwidget

import (
	"fmt"
	"log"

	"github.com/AllenDang/giu"
	"github.com/AllenDang/imgui-go"

	dcclib "github.com/gravestench/dcc/pkg"
)

const (
	imageW, imageH = 32, 32
)

type widget struct {
	id            string
	dcc           *dcclib.DCC
	textureLoader TextureLoader
}

// Create creates a new dcc widget
func Create(state []byte, id string, dcc *dcclib.DCC) giu.Widget {
	result := &widget{
		id:            id,
		dcc:           dcc,
		textureLoader: NewTextureLoader(),
	}

	if giu.Context.GetState(result.getStateID()) == nil && state != nil {
		s := result.getState()
		s.Decode(state)
		result.setState(s)
	}

	return result
}

// Build build a widget
func (p *widget) Build() {
	p.textureLoader.ResumeLoadingTextures()
	p.textureLoader.ProcessTextureLoadRequests()

	viewerState := p.getState()

	imageScale := uint32(viewerState.controls.scale)
	dirIdx := dirLookup(int(viewerState.controls.direction), len(p.dcc.Directions()))
	frameIdx := viewerState.controls.frame

	textureIdx := dirIdx*len(p.dcc.Direction(dirIdx).Frames()) + int(frameIdx)

	if imageScale < 1 {
		imageScale = 1
	}

	err := giu.Context.GetRenderer().SetTextureMagFilter(giu.TextureFilterNearest)
	if err != nil {
		log.Print(err)
	}

	var frameImage *giu.ImageWidget

	if viewerState.textures == nil || len(viewerState.textures) <= int(frameIdx) || viewerState.textures[frameIdx] == nil {
		frameImage = giu.Image(nil).Size(imageW, imageH)
	} else {
		bw := p.dcc.Direction(dirIdx).Box.Dx()
		bh := p.dcc.Direction(dirIdx).Box.Dy()
		w := float32(uint32(bw) * imageScale)
		h := float32(uint32(bh) * imageScale)
		frameImage = giu.Image(viewerState.textures[textureIdx]).Size(w, h)
	}

	numDirections := len(p.dcc.Directions())
	numFrames := len(p.dcc.Direction(0).Frames())

	giu.Layout{
		giu.Label(fmt.Sprintf("Version: %v", p.dcc.Version)),
		giu.Label(fmt.Sprintf("Directions: %v", numDirections)),
		giu.Label(fmt.Sprintf("Frames per Direction: %v", numFrames)),
		giu.Custom(func() {
			imgui.BeginGroup()

			if numDirections > 1 {
				imgui.SliderInt("Direction", &viewerState.controls.direction, 0, int32(numDirections-1))
			}

			if numFrames > 1 {
				imgui.SliderInt("Frames", &viewerState.controls.frame, 0, int32(numFrames-1))
			}

			imgui.SliderInt("Scale", &viewerState.controls.scale, 1, 8)

			imgui.EndGroup()
		}),
		giu.Separator(),
		frameImage,
	}.Build()
}

func dirLookup(dir, numDirs int) int {
	d4 := []int{0, 1, 2, 3}
	d8 := []int{0, 5, 1, 6, 2, 7, 3, 4}
	d16 := []int{0, 9, 5, 10, 1, 11, 6, 12, 2, 13, 7, 14, 3, 15, 4, 8}

	lookup := []int{0}

	switch numDirs {
	case 4:
		lookup = d4
	case 8:
		lookup = d8
	case 16:
		lookup = d16
	default:
		dir = 0
	}

	return lookup[dir]
}