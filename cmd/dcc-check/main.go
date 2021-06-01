package main

import (
	"fmt"
	"io/ioutil"
	"os"
)

import (
	dcc "github.com/gravestench/dcc/pkg"
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

	_, err = dcc.FromBytes(fileContents)
	if err != nil {
		fmt.Print(err)
		return
	}

	fmt.Println("DCC decode successful")
}

func printUsage() {
	fmt.Printf("Usage:\r\n\t%s path/to/file.dcc", os.Args[0])
}
