package main

import (
	"bytes"
	"flag"
	"fmt"
	"image/color"
	"io/ioutil"

	"github.com/AllenDang/giu"

	gpl "github.com/gravestench/gpl/pkg"

	dccLib "github.com/gravestench/dcc/pkg"
	dccWidget "github.com/gravestench/dcc/pkg/giuwidget"
)

const (
	title         = "dcc viewer"
	defaultWidth  = 256
	defaultHeight = 256
	windowFlags   = giu.MasterWindowFlagsFloating & giu.MasterWindowFlagsNotResizable
)

func main() {
	var o options

	if showUsage := parseOptions(&o); showUsage {
		flag.Usage()
		return
	}

	fileContents, err := ioutil.ReadFile(*o.dccPath)
	if err != nil {
		const fmtErr = "could not read file, %w"

		fmt.Print(fmt.Errorf(fmtErr, err))

		return
	}

	dcc, err := dccLib.FromBytes(fileContents)
	if err != nil {
		fmt.Print(err)
		return
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

		dcc.SetPalette(color.Palette(*gplInstance))
	} else {
		dcc.SetPalette(nil)
	}

	window := giu.NewMasterWindow(title, defaultWidth, defaultHeight, windowFlags, nil)

	widget := dccWidget.Create(nil, "dccviewer", dcc)

	window.Run(func() {
		giu.SingleWindow("dcc viewer").Layout(widget)
	})
}

type options struct {
	dccPath *string
	palPath *string
	pngPath *string
}

func parseOptions(o *options) (terminate bool) {
	o.dccPath = flag.String("dcc", "", "input dcc file (required)")
	o.palPath = flag.String("pal", "", "input pal file (optional)")
	o.pngPath = flag.String("png", "", "path to png file (optional)")

	flag.Parse()

	return *o.dccPath == ""
}
