package main

import (
	"fmt"
	"io/ioutil"
	"os"

	"github.com/OpenDiablo2/HellSpawner/hscommon"
	"github.com/ianling/giu"

	lib "github.com/gravestench/dcc/pkg"
	"github.com/gravestench/dcc/pkg/giuwidget"
)

const (
	title = "dcc viewer"
	defaultWidth = 256
	defaultHeight = 256
	windowFlags = giu.MasterWindowFlagsFloating & giu.MasterWindowFlagsNotResizable
)

func main() {
	if len(os.Args) < 2 {
		printUsage()
		return
	}

	srcPath := os.Args[1]

	fileContents, err := ioutil.ReadFile(srcPath)
	if err != nil {
		const fmtErr = "could not read file, %v"
		fmt.Print(fmt.Errorf(fmtErr, err))

		return
	}

	dcc, err := lib.FromBytes(fileContents)
	if err != nil {
		fmt.Print(err)
		return
	}

	window := giu.NewMasterWindow(title, defaultWidth, defaultHeight, windowFlags, nil)

	tl := hscommon.NewTextureLoader()

	widget := giuwidget.Create(tl, nil, "dccviewer", dcc)

	window.Run(func() {
		tl.ResumeLoadingTextures()
		tl.ProcessTextureLoadRequests()
		giu.SingleWindow("dcc viewer").Layout(widget)
	})
}

func printUsage() {
	fmt.Printf("Usage:\r\n\t%s path/to/file.lib", os.Args[0])
}

func render() {

}
