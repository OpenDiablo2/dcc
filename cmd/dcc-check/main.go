package dcc_check

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
)

import (
	dcc "github.com/gravestench/dcc/pkg"
)

func main() {
	srcPath := flag.Arg(1)

	if srcPath == "" {
		printUsage()
		return
	}

	fileContents, err := ioutil.ReadFile(srcPath)
	if err != nil {
		const fmtErr = "could not read file, %v"
		log.Fatal(fmt.Errorf(fmtErr, err))
	}

	_, err = dcc.FromBytes(fileContents)
	if err != nil {
		log.Fatal(err)
	}
}

func printUsage() {
	str := fmt.Sprintf("Usage:\r\n\t%v path/to/file.dcc", flag.Arg(0))

	log.Print(str)
}
