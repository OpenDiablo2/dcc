package pkg

import (
	"errors"
	"fmt"
	"image/color"

	"github.com/gravestench/bitstream"
)

const (
	FileSignature byte = 0x74
)

const (
	SanityCheck1 int32 = 1
)

const (
	signatureBits          = 8
	versionBits            = 8
	directionsBits         = 8
	directionOffsetBits    = 32
	framesPerDirectionBits = 32
	sanityCheckBits        = 32
	totalSizeCodedBits     = 32
)

func New() *DCC {
	return (&DCC{}).init()
}

func FromBytes(data []byte) (*DCC, error) {
	return New().FromBytes(data)
}

type DCC struct {
	Version            byte
	TotalSizeCoded     uint32
	numDirections      uint32
	framesPerDirection uint32
	directions         []*Direction
	palette            color.Palette
	stream *bitstream.BitStream
	dirty              bool // when anything is changed this flag is set, causes recalculation
}

func (d *DCC) init() *DCC {
	d.dirty = true

	d.SetPalette(nil)

	return d
}

func (d *DCC) FromBytes(data []byte) (_ *DCC, err error) {
	if !d.dirty {
		d.init()
	}

	stream := bitstream.FromBytes(data...)

	d.stream = stream

	if err = d.Decode(stream); err != nil {
		return nil, err
	}

	return d, nil
}

func (d *DCC) Direction(n int) *Direction {
	if n < 0 || n >= len(d.directions) {
		return nil
	}

	return d.directions[n]
}

func (d *DCC) Directions() []*Direction {
	return append([]*Direction{}, d.directions...)
}

func (d *DCC) Decode(stream *bitstream.BitStream) error {
	if err := d.decodeHeader(stream); err != nil {
		return fmt.Errorf("error decoding dcc header, %v", err)
	}

	if err := d.decodeBody(stream); err != nil {
		return fmt.Errorf("error decoding dcc body, %v", err)
	}

	if err := d.generateImages(); err != nil {
		return fmt.Errorf("error generating frame images, %v", err)
	}

	d.dirty = false

	return nil
}

func (d *DCC) decodeHeader(stream *bitstream.BitStream) (err error) {
	if signature, err := stream.Next(signatureBits).Bits().AsByte(); err != nil {
		return err
	} else if signature != FileSignature {
		const fmtErr = "unexpected file signature %x, expecting %x"
		return fmt.Errorf(fmtErr, signature, FileSignature)
	}

	if d.Version, err = stream.Next(versionBits).Bits().AsByte(); err != nil {
		return err
	}

	if d.numDirections, err = stream.Next(directionsBits).Bits().AsUInt32(); err != nil {
		return err
	} else {
		d.directions = make([]*Direction, d.numDirections)
	}

	if d.framesPerDirection, err = stream.Next(framesPerDirectionBits).Bits().AsUInt32(); err != nil {
		return err
	} else {
		for idx := range d.directions {
			d.directions[idx] = &Direction{}
			d.directions[idx].frames = make([]*Frame, d.framesPerDirection)
		}
	}

	if val, err := stream.Next(sanityCheckBits).Bits().AsInt32(); err != nil {
		return err
	} else if val != SanityCheck1 {
		const fmtErr = "sanity check error, got %x, expecting %x"
		return fmt.Errorf(fmtErr, val, FileSignature)
	}

	if d.TotalSizeCoded, err = stream.Next(totalSizeCodedBits).Bits().AsUInt32(); err != nil {
		return err
	}

	return nil
}

func (d *DCC) decodeBody(stream *bitstream.BitStream) error {
	// decode each direction
	for idx := 0; idx < len(d.directions); idx++ {
		offset, err := stream.Next(directionOffsetBits).Bits().AsUInt32()
		if err != nil {
			return err
		}

		if offset >= uint32(stream.Length()) {
			const fmtErr = "direction offset greater than length of file (%v >= %v)"
			return fmt.Errorf(fmtErr, offset, stream.Length())
		}

		d.directions[idx] = &Direction{dcc: d}

		// the offset we just read is a byte offset within the file data that the direction starts at,
		// so we want to reset the number of bits read and then set the offset here.
		newStream := stream.Copy().SetBitPosition(0).SetPosition(int(offset))

		if err := d.directions[idx].decode(newStream); err != nil {
			const fmtErr = "direction index %d, %v"
			return fmt.Errorf(fmtErr, idx, err)
		}
	}

	return nil
}

func (d *DCC) generateImages() error {

	return nil
}

func (d *DCC) SetPalette(p color.Palette) {
	dst := DefaultPalette()

	for idx := range dst {
		if idx >= len(p) {
			break
		}

		dst[idx] = p[idx]
	}

	d.palette = dst
}

func (d *DCC) Encode() ([]byte, error) {
	return nil, errors.New("not yet implemented")
}

func (d *DCC) Clone() *DCC {
	stream := d.stream.Copy()

	stream.SetPosition(0)
	stream.SetBitPosition(0)
	bytes, _ := stream.Next(stream.Length()).Bytes().AsBytes()

	newDcc, _ := FromBytes(bytes)

	return newDcc
}
