// fileExtractor_test
package fileExtractor

import (
	"fmt"
	"testing"

	"github.com/BurntSushi/toml"
	"github.com/stretchr/testify/assert"
)

type tomlConfig struct {
	Cruise string
	Ship   string
	Files  map[string]file
	Kml    kml
}

type file struct {
	FileName  string
	VarList   string
	PlotNames string
	Prefix    int
	PlotSize  int
	SkipLine  int
}

type kml struct {
	FileName string
}

// usefull macro
var p = fmt.Println
var pf = fmt.Printf

func TestFile(t *testing.T) {

	opts := NewFileExtractOptions()
	assert := assert.New(t)
	assert.Empty(opts.Filename())
	assert.Empty(opts.VarsList())
	assert.Empty(opts.hdr)
	assert.Equal(opts.skipLine, 0)

	opts.SetFilename("pirata-fr23_tsg")
	assert.Equal(opts.Filename(), "pirata-fr23_tsg")
	opts.SetVarsList("LATITUDE,3,LONGITUDE,4")
	assert.Equal(opts.VarsList(), map[string]int{"LATITUDE": 3, "LONGITUDE": 4})
	assert.Equal(opts.hdr, []string{"LATITUDE", "LONGITUDE"})
	assert.Len(opts.hdr, 2)
	opts.SetSkipLine(2)
	assert.Equal(opts.skipLine, 2)

	// read toml file
	var config tomlConfig
	_, err := toml.DecodeFile("config.toml", &config)
	assert.Nil(err)
	assert.Equal(config.Cruise, "PIRATA-FR26")
	p(config.Files)
	for instrument, file := range config.Files {
		pf("Instrument: %s (%s, %s)\n", instrument, file.FileName, file.VarList)
	}
}
