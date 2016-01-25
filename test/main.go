package main

import (
	"fmt"

	"github.com/jgrelet/cruiseTrack2kml/fileExtractor"
)

// usefull macro
var p = fmt.Println
var pf = fmt.Printf

// example main
func main() {

	// read TSG track
	opts := fileExtractor.NewFileExtractOptions().SetFilename("test.gps")
	opts.SetVars("TIME,1,LATITUDE,2,LONGITUDE,3,TEMP,4")
	p(opts)
	ext := fileExtractor.NewFileExtracter(opts)
	ext.Read()
	ext.Print()

}
