package giuwidget

import (
	"fmt"
	"log"

	"github.com/ianling/giu"
	"github.com/ianling/imgui-go"

	dcclib "github.com/gravestench/dcc/pkg"

	"github.com/OpenDiablo2/HellSpawner/hscommon"
)
const (
	// nolint:gomnd // constant = constant
	maxAlpha = uint8(255)
)

const (
	imageW, imageH = 32, 32
)

type widget struct {
	id            string
	dcc           *dcclib.DCC
	textureLoader hscommon.TextureLoader
}

// Create creates a new dcc widget
func Create(tl hscommon.TextureLoader, state []byte, id string, dcc *dcclib.DCC) giu.Widget {
	result := &widget{
		id:            id,
		dcc:           dcc,
		textureLoader: tl,
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
	viewerState := p.getState()

	imageScale := uint32(viewerState.controls.scale)
	dirIdx := int(viewerState.controls.direction)
	frameIdx := viewerState.controls.frame

	textureIdx := dirIdx*len(p.dcc.Direction(dirIdx).Frames()) + int(frameIdx)

	if imageScale < 1 {
		imageScale = 1
	}

	err := giu.Context.GetRenderer().SetTextureMagFilter(giu.TextureFilterNearest)
	if err != nil {
		log.Print(err)
	}

	var widget *giu.ImageWidget
	if viewerState.textures == nil || len(viewerState.textures) <= int(frameIdx) || viewerState.textures[frameIdx] == nil {
		widget = giu.Image(nil).Size(imageW, imageH)
	} else {
		bw := p.dcc.Direction(dirIdx).Box.Dx()
		bh := p.dcc.Direction(dirIdx).Box.Dy()
		w := float32(uint32(bw) * imageScale)
		h := float32(uint32(bh) * imageScale)
		widget = giu.Image(viewerState.textures[textureIdx]).Size(w, h)
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
		widget,
	}.Build()
}
