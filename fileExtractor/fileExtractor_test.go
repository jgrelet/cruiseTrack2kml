// fileExtractor_test
package fileExtractor

import (
	"fmt"
	"log"
	"os"
	"strings"
	"testing"
	"text/tabwriter"
	"time"

	"github.com/BurntSushi/toml"
	"github.com/stretchr/testify/assert"
)

var (
	debug = true
	// configFile string = "config.toml"
	configFile string = "cruise.toml"
)

type tomlConfig struct {
	CycleMesure string    `toml:"cruise"`
	Plateforme  string    `toml:"ship"`
	CallSign    string    `toml:"callsign"`
	BeginDate   time.Time `toml:"begin_date"`
	//BeginDate string          `toml:"begin_date"`
	EndDate   time.Time       `toml:"end_date"`
	Institute string          `toml:"institute"`
	Pi        string          `toml:"pi"`
	Creator   string          `toml:"creator"`
	Files     map[string]file `toml:"files"`
	Kml       kml             `toml:"kml"`
}

type file struct {
	FileName  string `toml:"fileName"`
	VarList   string `toml:"varList"`
	Separator string `toml:"separator"`
	PlotNames string `toml:"plotNames,omitempty"`
	Prefix    int    `toml:"prefix,omitempty"`
	PlotSize  int    `toml:"plotSize,omitempty"`
	SkipLine  int    `toml:"skipLine"`
}

type kml struct {
	FileName string `toml:"filename"`
}

func printTypes(md toml.MetaData) {
	tabw := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	for _, key := range md.Keys() {
		fmt.Fprintf(tabw, "%s%s\t%s\n",
			strings.Repeat("    ", len(key)-1), key, md.Type(key...))
	}
	tabw.Flush()
}

func TestFile(t *testing.T) {

	// test default empty configuration object
	opts := NewFileExtractOptions()
	assert := assert.New(t)
	assert.Empty(opts.Filename())
	assert.Empty(opts.VarsList())
	assert.Empty(opts.hdr)
	assert.Equal(opts.skipLine, 0)

	// fill and test the object
	opts.SetFilename("pirata-fr23_tsg")
	assert.Equal(opts.Filename(), "pirata-fr23_tsg")
	opts.SetVarsList("LATITUDE,3,LONGITUDE,4")
	assert.Equal(opts.VarsList(), map[string]int{"LATITUDE": 3, "LONGITUDE": 4})
	assert.Equal(opts.hdr, []string{"LATITUDE", "LONGITUDE"})
	assert.Len(opts.hdr, 2)
	opts.SetSkipLine(2)
	assert.Equal(opts.skipLine, 2)

	// read and test default config.toml file
	var config tomlConfig
	md, err := toml.DecodeFile(configFile, &config)
	if err != nil {
		log.Fatalf("Read config.toml: %s", err)
	}
	assert.Equal(config.CycleMesure, "PIRATA-FR26")
	assert.Equal(config.Plateforme, "THALASSA")
	assert.Equal(config.CallSign, "FNFP")
	assert.Equal(config.BeginDate.Format("01/02/2006"), "03/08/2016")
	//tt, _ := time.Parse("02/01/2006", config.BeginDate)
	//assert.Equal(tt.Format("01/02/2006"), "03/08/2016")
	assert.Equal(config.EndDate.Format("01/02/2006"), "04/14/2016")
	assert.Equal(config.Institute, "IRD")
	assert.Equal(config.Pi, "BOURLES")
	assert.Equal(config.Creator, "Jacques.Grelet_at_ird.fr")

	// display informations only for debugging
	if debug {
		printTypes(md)
		p(config)
	}

	// loop over files
	for instrument, file := range config.Files {
		// pf("Instrument: %s (%s, %s)\n", instrument, file.FileName, file.VarList)
		switch instrument {
		case "ctd":
			assert.Equal(file.FileName, "M:/PIRATA-FR26/data-processing/CTD/data/cnv/dfr26001.cnv")
			opts.SetFilename(file.FileName)
			opts.SetVarsList(file.VarList)
			opts.SetSkipLine(file.SkipLine)
			ext := NewFileExtractor(opts)
			err = ext.Read()
			if err != nil {
				log.Fatalf("NewFileExtractor(opts).Read() for %s: %s", instrument, err)
			}
			size := ext.Size() - 1
			pres := ext.Data()["PRES"]
			assert.Equal(pres[0], 2.0)       // test the first pressure value
			assert.Equal(pres[size], 2023.0) // test the last pressure value
			temp := ext.Data()["TEMP"]
			assert.Equal(temp[0], 24.7241)
			assert.Equal(temp[size], 3.5041)
			psal := ext.Data()["PSAL"]
			assert.Equal(psal[0], 35.7711)
			assert.Equal(psal[size], 34.9640)

		case "btl":
			assert.Equal(file.FileName, "M:/PIRATA-FR26/data-processing/CTD/data/btl/fr26001.btl")
			opts.SetFilename(file.FileName)
			opts.SetVarsList(file.VarList)
			opts.SetSkipLine(file.SkipLine)
			ext := NewFileExtractor(opts)
			err = ext.Read()
			if err != nil {
				log.Fatalf("NewFileExtractor(opts).Read() for %s: %s", instrument, err)
			}
			size := ext.Size() - 1
			btl := ext.Data()["BOTL"]
			assert.Equal(btl[0], 1)     // test the first pressure value
			assert.Equal(btl[size], 11) // test the last pressure value
			temp := ext.Data()["TE01"]
			assert.Equal(temp[0], 3.5048)
			assert.Equal(temp[size], 3.5048)
			psal := ext.Data()["PSA1"]
			assert.Equal(psal[0], 34.9636)
			assert.Equal(psal[size], 34.9637)
		case "tsg":
			assert.Equal(file.FileName, "M:/PIRATA-FR26/data-processing/THERMO/data/*.COLCOR")
		case "xbt":
			assert.Equal(file.FileName, "M:/PIRATA-FR26/data-processing/CELERITE/data/*.EDF")
		}
	}
	assert.Equal(config.Kml.FileName, "pirata-fr26.kml")

}
