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

	// initialize options
	//	opts := fileExtractor.NewFileExtractOptions().SetFilename("test.gps")
	//	opts.SetVarsList("TIME,1,LATITUDE,2,LONGITUDE,3,TEMP,4")
	//	opts.SetSkipLine(2)

	// pirata-FR23
	opts := fileExtractor.NewFileExtractOptions().SetFilename("pirata-fr23_tsg")
	opts.SetVarsList("LATITUDE,3,LONGITUDE,4")
	opts.SetSkipLine(2)

	// print options
	p(opts)

	// initialize fileExtractor from options
	ext := fileExtractor.NewFileExtracter(opts)

	// read the file
	ext.Read()

	p(ext.Size())
	//	lats := ext.Data()["LATITUDE"]
	//	for i := 0; i < ext.Size(); i++ {
	//		lat := lats[i]
	//		p(lat)
	//	}
	// display result for test
	ext.Print()

}
