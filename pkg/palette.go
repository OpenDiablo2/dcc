package pkg

import "image/color"

const numColorsInPalette = 256

func DefaultPalette() color.Palette {
	p := make(color.Palette, numColorsInPalette)

	for idx := range p {
		val := uint8(idx)
		r, g, b, a := val, val, val, uint8(1)
		p[idx] = color.RGBA{r, g, b, a}
	}

	return p
}