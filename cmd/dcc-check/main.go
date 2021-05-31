package dcc_check

import (
	"flag"
	"io/ioutil"
	"log"
)

import (
	dcc "github.com/gravestench/dcc/pkg"
)

func main() {
	srcPath := flag.Arg(1)

	fileContents, err := ioutil.ReadFile(srcPath)
	if err != nil {
		log.Fatal(err)
	}

	_, err = dcc.FromBytes(fileContents)
	if err != nil {
		log.Fatal(err)
	}
}
