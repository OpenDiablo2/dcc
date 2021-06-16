package pkg

import (
	"errors"
	"fmt"
	"image"

	"github.com/gravestench/bitstream"
)

var _ image.PalettedImage = &Frame{}

type Frame struct {
	direction             *Direction
	Box                   image.Rectangle
	Cells                 []Cell
	PixelData             []byte
	Width                 int
	Height                int
	XOffset               int
	YOffset               int
	NumberOfOptionalBytes int
	NumberOfCodedBytes    int
	HorizontalCellCount   int
	VerticalCellCount     int
	FrameIsBottomUp       bool
	valid                 bool
}

func (f *Frame) decodeFrameHeader(stream *bitstream.Reader) (err error) {
	// we dont use var0 width bits
	_, _ = stream.Next(f.direction.Variable0Bits).Bits().AsUInt32()

	width, _ := stream.Next(f.direction.WidthBits).Bits().AsUInt32()
	height, _ := stream.Next(f.direction.HeightBits).Bits().AsUInt32()
	f.Width, f.Height = int(width), int(height)

	f.XOffset, _ = stream.Next(f.direction.XOffsetBits).Bits().AsInt()
	f.YOffset, _ = stream.Next(f.direction.YOffsetBits).Bits().AsInt()

	numOptionBytes, _ := stream.Next(f.direction.OptionalDataBits).Bits().AsUInt32()
	f.NumberOfOptionalBytes = int(numOptionBytes)

	codedBytes, _ := stream.Next(f.direction.CodedBytesBits).Bits().AsUInt32()
	f.NumberOfCodedBytes = int(codedBytes)

	// we will finally use the last returned stream error.
	f.FrameIsBottomUp, err = stream.Next(1).Bits().AsBool()
	if err != nil {
		return fmt.Errorf("stream error, %w", err)
	}

	if f.FrameIsBottomUp {
		return errors.New("bottom up frames are not implemented")
	}

	min := image.Point{X: f.XOffset, Y: f.YOffset - f.Height + 1}
	max := image.Point{X: min.X + f.Width, Y: min.Y + f.Height}
	f.Box = image.Rectangle{Min: min, Max: max}

	f.valid = true

	return nil
}

func (f *Frame) firstCellDimensions() (int, int) {
	// Width, height of the first cell
	w := cellSize - ((f.Box.Min.X - f.direction.Box.Min.X) % cellSize)
	h := cellSize - ((f.Box.Min.Y - f.direction.Box.Min.Y) % cellSize)

	return w, h
}

func (f *Frame) calcCellCounts() {
	const magic2 = 2

	firstW, firstH := f.firstCellDimensions()

	remainderW := f.Width - firstW - 1
	remainderH := f.Height - firstH - 1

	f.HorizontalCellCount = magic2 + (remainderW / cellSize)
	if (remainderW % cellSize) == 0 {
		f.HorizontalCellCount--
	}

	f.VerticalCellCount = magic2 + (remainderH / cellSize)
	if (remainderH % cellSize) == 0 {
		f.VerticalCellCount--
	}

	if f.HorizontalCellCount <= 0 {
		f.HorizontalCellCount = 1
	}

	if f.VerticalCellCount <= 0 {
		f.VerticalCellCount = 1
	}
}

func (f *Frame) calcCellWidths() []int {
	const magic2 = 2

	firstW, _ := f.firstCellDimensions()

	cellWidths := make([]int, f.HorizontalCellCount)
	cellWidths[0] = f.Width

	if f.HorizontalCellCount > 1 {
		cellWidths[0] = firstW
		for i := 1; i < (f.HorizontalCellCount - 1); i++ {
			cellWidths[i] = cellSize
		}

		cellWidths[f.HorizontalCellCount-1] = f.Width - firstW - (cellSize * (f.HorizontalCellCount - magic2))
	}

	return cellWidths
}

func (f *Frame) calcCellHeights() []int {
	const magic2 = 2

	_, firstH := f.firstCellDimensions()

	cellHeights := make([]int, f.VerticalCellCount)
	cellHeights[0] = f.Height

	if f.VerticalCellCount > 1 {
		cellHeights[0] = firstH
		for i := 1; i < (f.VerticalCellCount - 1); i++ {
			cellHeights[i] = cellSize
		}

		cellHeights[f.VerticalCellCount-1] = f.Height - firstH - (cellSize * (f.VerticalCellCount - magic2))
	}

	return cellHeights
}

func (f *Frame) calcCellDimensions() {
	cellWidths, cellHeights := f.calcCellWidths(), f.calcCellHeights()

	frameTopLeft := f.Box.Min
	dirTopLeft := f.direction.Box.Min

	f.Cells = make([]Cell, f.HorizontalCellCount*f.VerticalCellCount)
	pixelY := frameTopLeft.Sub(dirTopLeft).Y

	// cell x,y are cell-coordinates.
	// pixel x,y are the pixel-coordinates (in the frame, not just the cell).
	//
	// basically, using the cell dimensions, we are determining the
	// root (top-left) coordinate of each cell in the frame
	for cellY := 0; cellY < f.VerticalCellCount; cellY++ {
		pixelX := frameTopLeft.Sub(dirTopLeft).X

		for cellX := 0; cellX < f.HorizontalCellCount; cellX++ {
			cell := Cell{
				XOffset: pixelX,
				YOffset: pixelY,
				Width:   cellWidths[cellX],
				Height:  cellHeights[cellY],
			}

			cellIndex := cellX + (cellY * f.HorizontalCellCount)

			f.Cells[cellIndex] = cell

			pixelX += cellWidths[cellX]
		}

		pixelY += cellHeights[cellY]
	}
}

func (f *Frame) recalculateCells() error {
	f.calcCellCounts()

	if f.VerticalCellCount < 1 || f.HorizontalCellCount < 1 {
		return errors.New("cell count cant be less than 1")
	}

	f.calcCellDimensions()

	return nil
}
