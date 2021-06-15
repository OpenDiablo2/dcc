package main

import (
	"bytes"
	"flag"
	"fmt"
	"image/color"
	"image/png"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"

	dcc "github.com/gravestench/dcc/pkg"
	gpl "github.com/gravestench/gpl/pkg"
)

type options struct {
	dccPath *string
	palPath *string
	pngPath *string
}

func main() {
	var o options

	if showUsage := parseOptions(&o); showUsage {
		flag.Usage()
		return
	}

	//dccBaseName := path.Base(*o.dccPath)
	//dccFileName := fileNameWithoutExt(dccBaseName)

	//palBaseName := path.Base(*o.palPath)
	//palFileName := fileNameWithoutExt(palBaseName)

	data, err := ioutil.ReadFile(*o.dccPath)
	if err != nil {
		const fmtErr = "could not read file, %v"
		fmt.Print(fmt.Errorf(fmtErr, err))

		return
	}

	d, err := dcc.FromBytes(data)
	if err != nil {
		fmt.Println(err)
		return
	}

	outfilePath := *o.pngPath
	numDirections := len(d.Directions())
	framesPerDirection := len(d.Direction(0).Frames())
	hasMultipleImages := numDirections > 1 || framesPerDirection > 1

	if hasMultipleImages {
		noExt := fileNameWithoutExt(outfilePath)
		outfilePath = noExt + "_d%v_f%v.png"
	}

	if *o.palPath != "" {
		palData, err := ioutil.ReadFile(*o.palPath)
		if err != nil {
			fmt.Println(err)
			return
		}

		gplInstance, err := gpl.Decode(bytes.NewBuffer(palData))
		if err != nil {
			fmt.Println("palette is not a GIMP palette file...")
			return
		}

		d.SetPalette(color.Palette(*gplInstance))
	} else {
		d.SetPalette(nil)
	}

	for dirIdx := 0; dirIdx < numDirections; dirIdx++ {
		frames := d.Direction(dirIdx).Frames()

		for frameIdx := range frames {
			outPath := outfilePath
			if hasMultipleImages {
				outPath = fmt.Sprintf(outfilePath, dirIdx, frameIdx)
			}

			f, err := os.Create(outPath)
			if err != nil {
				log.Fatal(err)
			}

			if err := png.Encode(f, frames[frameIdx]); err != nil {
				_ = f.Close()
				log.Fatal(err)
			}

			if err := f.Close(); err != nil {
				log.Fatal(err)
			}
		}
	}
}

func parseOptions(o *options) (terminate bool) {
	o.dccPath = flag.String("dcc", "", "input dcc file (required)")
	o.palPath = flag.String("pal", "", "input pal file (optional)")
	o.pngPath = flag.String("png", "", "path to png file (optional)")

	flag.Parse()

	if *o.dccPath == "" {
		return true
	}

	return false
}

func fileNameWithoutExt(fileName string) string {
	return fileName[:len(fileName)-len(filepath.Ext(fileName))]
}
