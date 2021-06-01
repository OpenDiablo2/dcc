package pkg

import (
	"errors"
	"fmt"
	"github.com/gravestench/bitstream"
	"image"
)

const streamSizeBits = 20 // num bits for representing a substream

type CompressionFlag = int

const (
	RawPixelCompression CompressionFlag = 1 << iota
	EqualCellsCompression
)

const (
	baseMinx = 100000
	baseMiny = 100000
	baseMaxx = -100000
	baseMaxy = -100000
)

const cellsPerRow = 4

// Direction represents a Direction file.
type Direction struct {
	dcc                        *DCC
	OutSizeCoded               int
	CompressionFlags           int
	Variable0Bits              int
	WidthBits                  int
	HeightBits                 int
	XOffsetBits                int
	YOffsetBits                int
	OptionalDataBits           int
	CodedBytesBits             int
	EqualCellsBitstreamSize    uint32
	PixelMaskBitstreamSize     uint32
	EncodingTypeBitstreamSize  uint32
	RawPixelCodesBitstreamSize uint32
	frames                     []*Frame
	PaletteEntries             [256]byte
	Box                        *image.Rectangle
	Cells                      []*Cell
	PixelData                  []byte
	HorizontalCellCount        int
	VerticalCellCount          int
	PixelBuffer                []PixelBufferEntry
}

func (d *Direction) decode(stream *bitstream.BitStream) (err error) {
	d.frames = make([]*Frame, d.dcc.framesPerDirection)

	if err = d.decodeHeader(stream); err != nil {
		return err
	}

	if err = d.decodeBody(stream); err != nil {
		return err
	}

	return nil
}

func (d *Direction) decodeHeader(stream *bitstream.BitStream) (err error) {
	if val, err := stream.Next(32).Bits().AsUInt32(); err != nil {
		return err
	} else {
		d.OutSizeCoded = int(val)
	}

	if val, err := stream.Next(2).Bits().AsUInt32(); err != nil {
		return err
	} else {
		d.CompressionFlags = int(val)
	}

	if d.Variable0Bits, err = crazyLookup(stream.Next(4).Bits().AsUInt32()); err != nil {
		return err
	}

	if d.WidthBits, err = crazyLookup(stream.Next(4).Bits().AsUInt32()); err != nil {
		return err
	}

	if d.HeightBits, err = crazyLookup(stream.Next(4).Bits().AsUInt32()); err != nil {
		return err
	}

	if d.XOffsetBits, err = crazyLookup(stream.Next(4).Bits().AsUInt32()); err != nil {
		return err
	}

	if d.YOffsetBits, err = crazyLookup(stream.Next(4).Bits().AsUInt32()); err != nil {
		return err
	}

	if d.OptionalDataBits, err = crazyLookup(stream.Next(4).Bits().AsUInt32()); err != nil {
		return err
	}

	if d.CodedBytesBits, err = crazyLookup(stream.Next(4).Bits().AsUInt32()); err != nil {
		return err
	}

	return nil
}

func (d *Direction) decodeBody(stream *bitstream.BitStream) (err error) {
	if err = d.decodeFrameHeaders(stream); err != nil {
		return err
	}

	if d.OptionalDataBits > 0 {
		return errors.New("optional bits in DCC data is not currently supported")
	}

	if err = d.decodeCompressionFlags(stream); err != nil {
		return err
	}

	if err = d.decodePaletteEntries(stream); err != nil {
		return err
	}

	// HERE BE GIANTS:
	// Because of the way this thing mashes bits together, BIT offset matters
	// here. For example, if you are on byte offset 3, bit offset 6, and
	// the EqualCellsBitstreamSize is 20 bytes, then the next bit stream
	// will be located at byte 23, bit offset 6!
	ec := stream.Copy()

	stream.OffsetPosition(int(d.EqualCellsBitstreamSize))

	pm := stream.Copy()

	stream.OffsetBitPosition(int(d.PixelMaskBitstreamSize))

	et := stream.Copy()

	stream.OffsetBitPosition(int(d.EncodingTypeBitstreamSize))

	rpc := stream.Copy()

	stream.OffsetBitPosition(int(d.RawPixelCodesBitstreamSize))

	pcd := stream.Copy()

	d.calculateCells()

	// Fill in the pixel buffer
	if err = d.fillPixelBuffer(pcd, ec, pm, et, rpc); err != nil {
		const fmtErr = "filling pixel buffer, %v"
		return fmt.Errorf(fmtErr, err)
	}

	// Generate the actual frame pixel data
	if err = d.generateFrames(pcd); err != nil {
		const fmtErr = "generating frames, %v"
		return fmt.Errorf(fmtErr, err)
	}

	d.PixelBuffer = nil

	// Verify that everything we expected to read was actually read (sanity check)...
	d.verify(ec, pm, et, rpc)

	stream.OffsetBitPosition(pcd.BitsRead())

	return nil
}

func (d *Direction) decodeFrameHeaders(stream *bitstream.BitStream) error {
	minX := baseMinx
	minY := baseMiny
	maxX := baseMaxx
	maxY := baseMaxy

	for frameIdx := uint32(0); frameIdx < d.dcc.framesPerDirection; frameIdx++ {
		d.frames[frameIdx] = &Frame{
			direction: d,
		}

		if err := d.frames[frameIdx].decodeFrameHeader(stream); err != nil {
			const fmtErr = "frame index %d decode"
			return fmt.Errorf(fmtErr, frameIdx)
		}

		bounds := d.frames[frameIdx].Bounds()

		minX = int(minInt32(int32(bounds.Min.X), int32(minX)))
		minY = int(minInt32(int32(bounds.Min.Y), int32(minY)))
		maxX = int(maxInt32(int32(bounds.Max.X), int32(maxX)))
		maxY = int(maxInt32(int32(bounds.Max.Y), int32(maxY)))
	}

	d.Box = &image.Rectangle{
		image.Point{minX, minY},
		image.Point{maxX, maxY},
	}

	return nil
}

func (d *Direction) decodeCompressionFlags(stream *bitstream.BitStream) (err error) {
	if (d.CompressionFlags & EqualCellsCompression) > 0 {
		d.EqualCellsBitstreamSize, err = stream.Next(streamSizeBits).Bits().AsUInt32()
		if err != nil {
			return err
		}
	}

	d.PixelMaskBitstreamSize, err = stream.Next(streamSizeBits).Bits().AsUInt32()
	if err != nil {
		return err
	}

	if (d.CompressionFlags & RawPixelCompression) > 0 {
		d.EncodingTypeBitstreamSize, err = stream.Next(streamSizeBits).Bits().AsUInt32()
		if err != nil {
			return err
		}

		d.RawPixelCodesBitstreamSize, err = stream.Next(streamSizeBits).Bits().AsUInt32()
		if err != nil {
			return err
		}
	}

	return nil
}

func (d *Direction) decodePaletteEntries(stream *bitstream.BitStream) (err error) {
	for paletteEntryCount, idx := 0, 0; idx < 256; idx++ {
		if valid, err := stream.Next(1).Bits().AsBool(); err != nil {
			return err
		} else if !valid {
			continue
		}

		d.PaletteEntries[paletteEntryCount] = byte(idx)
		paletteEntryCount++
	}

	return nil
}

func (d *Direction) calculateCells() {
	// Calculate the number of vertical and horizontal cells we need
	d.HorizontalCellCount = 1 + (d.Box.Dx()-1)/cellsPerRow
	d.VerticalCellCount = 1 + (d.Box.Dy()-1)/cellsPerRow

	// Calculate the cell widths
	cellWidths := make([]int, d.HorizontalCellCount)
	if d.HorizontalCellCount == 1 {
		cellWidths[0] = d.Box.Dx()
	} else {
		for i := 0; i < d.HorizontalCellCount-1; i++ {
			cellWidths[i] = 4
		}

		// nolint:gomnd // constant
		cellWidths[d.HorizontalCellCount-1] = d.Box.Dx() - (4 * (d.HorizontalCellCount - 1))
	}

	// Calculate the cell heights
	// nolint:gomnd // constant
	cellHeights := make([]int, d.VerticalCellCount)
	if d.VerticalCellCount == 1 {
		cellHeights[0] = d.Box.Dy()
	} else {
		for i := 0; i < d.VerticalCellCount-1; i++ {
			cellHeights[i] = 4
		}

		// nolint:gomnd // constant
		cellHeights[d.VerticalCellCount-1] = d.Box.Dy() - (4 * (d.VerticalCellCount - 1))
	}

	// Set the cell widths and heights in the cell buffer
	d.Cells = make([]*Cell, d.VerticalCellCount*d.HorizontalCellCount)
	yOffset := 0

	for y := 0; y < d.VerticalCellCount; y++ {
		xOffset := 0

		for x := 0; x < d.HorizontalCellCount; x++ {
			d.Cells[x+(y*d.HorizontalCellCount)] = &Cell{
				Width:   cellWidths[x],
				Height:  cellHeights[y],
				XOffset: xOffset,
				YOffset: yOffset,
			}

			xOffset += 4
		}

		yOffset += 4
	}

	// Calculate the cells for each of the frames
	for _, frame := range d.frames {
		frame.recalculateCells()
	}
}

func (d *Direction) fillPixelBuffer(pcd, ec, pm, et, rp *bitstream.BitStream) (err error) {
	var pixelMaskLookup = []int{0, 1, 1, 2, 1, 2, 2, 3, 1, 2, 2, 3, 2, 3, 3, 4}

	lastPixel := uint32(0)
	maxCellX := 0
	maxCellY := 0

	for _, frame := range d.frames {
		if frame == nil {
			continue
		}

		maxCellX += frame.HorizontalCellCount
		maxCellY += frame.VerticalCellCount
	}

	d.PixelBuffer = make([]PixelBufferEntry, maxCellX*maxCellY)

	for i := 0; i < maxCellX*maxCellY; i++ {
		d.PixelBuffer[i].Frame = -1
		d.PixelBuffer[i].FrameCellIndex = -1
	}

	cellBuffer := make([]*PixelBufferEntry, d.HorizontalCellCount*d.VerticalCellCount)
	frameIndex := -1
	pbIndex := -1

	var pixelMask uint32

	for _, frame := range d.frames {
		frameIndex++

		originCellX := (frame.Box.Min.X - d.Box.Min.X) / cellsPerRow
		originCellY := (frame.Box.Min.Y - d.Box.Min.Y) / cellsPerRow

		for cellY := 0; cellY < frame.VerticalCellCount; cellY++ {
			currentCellY := cellY + originCellY

			for cellX := 0; cellX < frame.HorizontalCellCount; cellX++ {
				currentCell := originCellX + cellX + (currentCellY * d.HorizontalCellCount)
				nextCell := false
				tmp := 0

				if cellBuffer[currentCell] != nil {
					if d.EqualCellsBitstreamSize > 0 {
						val, err := ec.Next(1).Bits().AsUInt32()
						if err != nil {
							const fmtErr = "reading EqualCells bitstream into cell buffer, cell index %v"
							return fmt.Errorf(fmtErr, currentCell)
						}

						tmp = int(val)
					} else {
						tmp = 0
					}

					if tmp == 0 {
						pixelMask, err = pm.Next(4).Bits().AsUInt32() //nolint:gomnd // binary data
						if err != nil {
							const fmtErr = "reading pixel mask into cell buffer, cell index %v"
							return fmt.Errorf(fmtErr, currentCell)
						}
					} else {
						nextCell = true
					}
				} else {
					pixelMask = 0x0F
				}

				if nextCell {
					continue
				}

				// Decode the pixels
				var pixelStack [4]uint32

				lastPixel = 0
				numberOfPixelBits := pixelMaskLookup[pixelMask]
				encodingType := 0

				if (numberOfPixelBits != 0) && (d.EncodingTypeBitstreamSize > 0) {
					val, err := et.Next(1).Bits().AsUInt32()
					if err != nil {
						const fmtErr = "reading encoding type, cell index %v"
						return fmt.Errorf(fmtErr, currentCell)
					}

					encodingType = int(val)
				} else {
					encodingType = 0
				}

				decodedPixel := 0

				for i := 0; i < numberOfPixelBits; i++ {
					if encodingType != 0 {
						if pixelStack[i], err = rp.Next(8).Bits().AsUInt32(); err != nil {
							const fmtErr = "reading into pixel stack, cell index %v"
							return fmt.Errorf(fmtErr, currentCell)
						}
					} else {
						pixelStack[i] = lastPixel
						pixelDisplacement, err := pcd.Next(4).Bits().AsUInt32()
						if err != nil {
							const fmtErr = "reading pixel displacement, cell index %v"
							return fmt.Errorf(fmtErr, currentCell)
						}

						pixelStack[i] += pixelDisplacement
						for pixelDisplacement == 15 {
							pixelDisplacement, err = pcd.Next(4).Bits().AsUInt32()
							if err != nil {
								const fmtErr = "reading pixel displacement, cell index %v"
								return fmt.Errorf(fmtErr, currentCell)
							}

							pixelStack[i] += pixelDisplacement
						}
					}

					if pixelStack[i] == lastPixel {
						pixelStack[i] = 0
						break
					} else {
						lastPixel = pixelStack[i]
						decodedPixel++
					}
				}

				oldEntry := cellBuffer[currentCell]

				pbIndex++

				curIdx := decodedPixel - 1

				for i := 0; i < 4; i++ {
					if (pixelMask & (1 << uint(i))) != 0 {
						if curIdx >= 0 {
							d.PixelBuffer[pbIndex].Value[i] = byte(pixelStack[curIdx])
							curIdx--
						} else {
							d.PixelBuffer[pbIndex].Value[i] = 0
						}
					} else {
						d.PixelBuffer[pbIndex].Value[i] = oldEntry.Value[i]
					}
				}

				cellBuffer[currentCell] = &d.PixelBuffer[pbIndex]
				d.PixelBuffer[pbIndex].Frame = frameIndex
				d.PixelBuffer[pbIndex].FrameCellIndex = cellX + (cellY * frame.HorizontalCellCount)
			}
		}
	}

	// Convert the palette entry index into actual palette entries
	for i := 0; i <= pbIndex; i++ {
		for x := 0; x < 4; x++ {
			d.PixelBuffer[i].Value[x] = d.PaletteEntries[d.PixelBuffer[i].Value[x]]
		}
	}

	return nil
}

func (d *Direction) generateFrames(pcd *bitstream.BitStream) (err error) {
	for _, cell := range d.Cells {
		cell.LastWidth = -1
		cell.LastHeight = -1
	}

	d.PixelData = make([]byte, d.Box.Dx() * d.Box.Dy())

	for frameIndex, frame := range d.frames {
		if err = d.generateFrame(frameIndex, frame, pcd); err != nil {
			const fmtErr = "generating frame with index %v, %v"
			return fmt.Errorf(fmtErr, frameIndex, err)
		}
	}

	d.Cells = nil
	d.PixelData = nil
	d.PixelBuffer = nil

	return nil
}

func (d *Direction) generateFrame(frameIndex int, frame *Frame, pcd *bitstream.BitStream) error {
	pbIdx := 0

	frame.PixelData = make([]byte, d.Box.Dx() * d.Box.Dy())

	for cellIdx, cell := range frame.Cells {
		cellX := cell.XOffset / cellsPerRow
		cellY := cell.YOffset / cellsPerRow
		cellIndex := cellX + (cellY * d.HorizontalCellCount)
		bufferCell := d.Cells[cellIndex]
		pbe := d.PixelBuffer[pbIdx]

		if (pbe.Frame != frameIndex) || (pbe.FrameCellIndex != cellIdx) {
			// This buffer cell has an EqualCell bit set to 1, so copy the frame cell or clear it
			if (cell.Width != bufferCell.LastWidth) || (cell.Height != bufferCell.LastHeight) {
				// Different sizes
				for y := 0; y < cell.Height; y++ {
					for x := 0; x < cell.Width; x++ {
						d.PixelData[x+cell.XOffset+((y+cell.YOffset)*d.Box.Dx())] = 0
					}
				}
			} else {
				// Same sizes
				// Copy the old frame cell into the new position
				for fy := 0; fy < cell.Height; fy++ {
					for fx := 0; fx < cell.Width; fx++ {
						// Frame (buff.lastx, buff.lasty) -> Frame (cell.offx, cell.offy)
						// Cell (0, 0,) ->
						// blit(dir->bmp, dir->bmp, buff_cell->last_x0, buff_cell->last_y0, cell->x0, cell->y0, cell->w, cell->h );
						d.PixelData[fx+cell.XOffset+((fy+cell.YOffset)*d.Box.Dx())] =
							d.PixelData[fx+bufferCell.LastXOffset+((fy+bufferCell.LastYOffset)*d.Box.Dx())]
					}
				}
				// Copy it again into the final frame image
				for fy := 0; fy < cell.Height; fy++ {
					for fx := 0; fx < cell.Width; fx++ {
						// blit(cell->bmp, frm_bmp, 0, 0, cell->x0, cell->y0, cell->w, cell->h );
						frame.PixelData[fx+cell.XOffset+((fy+cell.YOffset)*d.Box.Dx())] = d.PixelData[cell.XOffset+fx+((cell.YOffset+fy)*d.Box.Dx())]
					}
				}
			}
		} else {
			if pbe.Value[0] == pbe.Value[1] {
				// Clear the frame
				for y := 0; y < cell.Height; y++ {
					for x := 0; x < cell.Width; x++ {
						d.PixelData[x+cell.XOffset+((y+cell.YOffset)*d.Box.Dx())] = pbe.Value[0]
					}
				}
			} else {
				// Fill the frame cell with the pixels
				bitsToRead := 1
				if pbe.Value[1] != pbe.Value[2] {
					bitsToRead = 2
				}
				for y := 0; y < cell.Height; y++ {
					for x := 0; x < cell.Width; x++ {
						paletteIndex, err := pcd.Next(bitsToRead).Bits().AsUInt32()
						if err != nil {
							const fmtErr = "reading palette index at coord(%v, %v)"
							return fmt.Errorf(fmtErr, frameIndex, x, y)
						}

						d.PixelData[x+cell.XOffset+((y+cell.YOffset)*d.Box.Dx())] = pbe.Value[paletteIndex]
					}
				}
			}

			// Copy the frame cell into the frame
			for fy := 0; fy < cell.Height; fy++ {
				for fx := 0; fx < cell.Width; fx++ {
					// blit(cell->bmp, frm_bmp, 0, 0, cell->x0, cell->y0, cell->w, cell->h );
					frame.PixelData[fx+cell.XOffset+((fy+cell.YOffset)*d.Box.Dx())] = d.PixelData[fx+cell.XOffset+((fy+cell.YOffset)*d.Box.Dx())]
				}
			}
			pbIdx++
		}

		bufferCell.LastWidth = cell.Width
		bufferCell.LastHeight = cell.Height
		bufferCell.LastXOffset = cell.XOffset
		bufferCell.LastYOffset = cell.YOffset
	}

	// Free up the stuff we no longer need
	frame.Cells = nil

	return nil
}

func (d *Direction) verify(
	equalCellsBitstream,
	pixelMaskBitstream,
	encodingTypeBitstream,
	rawPixelCodesBitstream *bitstream.BitStream,
) error {
	steps := []struct{
		name string
		stream *bitstream.BitStream
		expectedBitsRead int
	}{
		{"EqualCells", equalCellsBitstream, int(d.EqualCellsBitstreamSize)},
		{"PixelMask", pixelMaskBitstream, int(d.PixelMaskBitstreamSize)},
		{"EncodingType", encodingTypeBitstream, int(d.EncodingTypeBitstreamSize)},
		{"RawPixelCodes", rawPixelCodesBitstream, int(d.RawPixelCodesBitstreamSize)},
	}

	for idx := range steps {
		actual := steps[idx].stream.BitsRead()
		expected := steps[idx].expectedBitsRead

		if actual != expected {
			const fmtErr = "verifying %v bitstream, read %v bits but expected to read %v bits"
			return fmt.Errorf(fmtErr, steps[idx].name, actual, expected)
		}
	}

	return nil
}

func (d *Direction) Bounds() image.Rectangle {
	return d.Box.Bounds()
}

func (d *Direction) Frame(n int) *Frame {
	if n >= len(d.frames) || n < 0 {
		return nil
	}

	return d.frames[n]
}

func (d *Direction) Frames() []*Frame {
	if d.frames == nil {
		return nil
	}

	return append([]*Frame{}, d.frames...)
}

func crazyLookup(idx uint32, err error) (int, error) {
	if err != nil {
		return 0, err
	}

	// nolint:gomnd // constant
	var crazyBitTable = []byte{0, 1, 2, 4, 6, 8, 10, 12, 14, 16, 20, 24, 26, 28, 30, 32}

	return int(crazyBitTable[idx]), err
}

func minInt32(a, b int32) int32 {
	if a < b {
		return a
	}

	return b
}

func maxInt32(a, b int32) int32 {
	if a > b {
		return a
	}

	return b
}
