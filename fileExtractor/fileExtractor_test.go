// fileExtractor_test
package fileExtractor

import (
	"fmt"
	"log"
	"testing"
	"time"

	"github.com/BurntSushi/toml"
	"github.com/stretchr/testify/assert"
)

// TODOS:
// - read multiples lines
// - define type for variables, default float64

var (
	// configFile string = "config.toml"
	configFile = "test.toml"
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

// test default empty FileExtractor object
func TestEmptyFileExtractor(t *testing.T) {
	assert := assert.New(t)
	opts := NewFileExtractOptions()
	assert.Empty(opts.Filename())
	assert.Empty(opts.VarsList())
	assert.Empty(opts.hdr)
	assert.Equal(opts.skipLine, 0)
}

// fill and test a valid FileExtractor object
func TestValidFileExtractor(t *testing.T) {
	assert := assert.New(t)
	opts := NewFileExtractOptions()
	opts.SetFilename("pirata-fr23_tsg")
	assert.Equal(opts.Filename(), "pirata-fr23_tsg")
	opts.SetVarsList("LATITUDE,3,float64,LONGITUDE,4,float64")
	assert.Equal(opts.VarsList(), map[string]Types{"LATITUDE": {column: 3, types: "float64"}, "LONGITUDE": {column: 4, types: "float64"}})
	assert.Equal(opts.hdr, []string{"LATITUDE", "LONGITUDE"})
	assert.Len(opts.hdr, 2)
	opts.SetSkipLine(2)
	assert.Equal(opts.skipLine, 2)
}

// read and test valid FileExtractor object from config.toml file
func TestFileExtractorFromConfigFile(t *testing.T) {
	var config tomlConfig
	assert := assert.New(t)
	_, err := toml.DecodeFile(configFile, &config)
	if err != nil {
		log.Fatalf("Read %s: %s", configFile, err)
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
	fmt.Println(debug, config)

	// loop over files
	for instrument, file := range config.Files {
		fmt.Fprintf(debug, "Instrument: %s (%s, %s)\n", instrument, file.FileName, file.VarList)
		switch instrument {
		case "ctd":
			assert.Equal(file.FileName, "test/CTD/dfr26001.cnv")
			opts := NewFileExtractOptions()
			opts.SetFilename(file.FileName)
			opts.SetVarsList(file.VarList)
			opts.SetSkipLine(file.SkipLine)
			ext := NewFileExtractor(opts)
			err = ext.Read()
			if err != nil {
				log.Fatalf("NewFileExtractor(opts).Read() for %s: %s", instrument, err)
			}
			//fmt.Println(ext)
			size := ext.Size() - 1
			pres := ext.Data("PRES")
			assert.Equal(2.0, pres[0])    // test the first pressure value
			assert.Equal(7.0, pres[size]) // test the last pressure value
			temp := ext.Data("TEMP")
			assert.Equal(24.7241, temp[0])
			assert.Equal(24.7260, temp[size])
			psal := ext.Data("PSAL")
			assert.Equal(35.7711, psal[0])
			assert.Equal(35.7716, psal[size])

		case "btl":
			assert.Equal(file.FileName, "test/CTD/fr26001.btl")
			opts := NewFileExtractOptions()
			opts.SetFilename(file.FileName)
			opts.SetVarsList(file.VarList)
			opts.SetSkipLine(file.SkipLine)
			/*
				ext := NewFileExtractor(opts)
					err = ext.Read()
					if err != nil {
						log.Fatalf("NewFileExtractor(opts).Read() for %s: %s", instrument, err)
					}
					size := ext.Size() - 1
					btl := ext.Data()["BOTL"]
					assert.Equal(btl[0], 1.0)     // test the first pressure value
					assert.Equal(btl[size], 11.0) // test the last pressure value
					temp := ext.Data()["TE01"]
					assert.Equal(temp[0], 3.5048)
					assert.Equal(temp[size], 3.5048)
					psal := ext.Data()["PSA1"]
					assert.Equal(psal[0], 34.9636)
					assert.Equal(psal[size], 34.9637)
					fmt.Fprintf(debug, ext)
			*/
		case "tsg":
			assert.Equal(file.FileName, "test/TSG/20160308-085453-TS_COLCOR.COLCOR")
			opts := NewFileExtractOptions()
			opts.SetFilename(file.FileName)
			opts.SetVarsList(file.VarList)
			opts.SetSkipLine(file.SkipLine)
			opts.SetSeparator(file.Separator)
			ext := NewFileExtractor(opts)
			err = ext.Read()
			if err != nil {
				log.Fatalf("NewFileExtractor(opts).Read() for %s: %s", instrument, err)
			}
			//fmt.Println(ext)

		case "xbt":
			assert.Equal(file.FileName, "test/XBT/T7_00001.EDF")
			opts := NewFileExtractOptions()
			opts.SetFilename(file.FileName)
			opts.SetVarsList(file.VarList)
			opts.SetSkipLine(file.SkipLine)
			ext := NewFileExtractor(opts)
			err = ext.Read()
			if err != nil {
				log.Fatalf("NewFileExtractor(opts).Read() for %s: %s", instrument, err)
			}
			//fmt.Println(ext)
			size := ext.Size() - 1
			pres := ext.Data("DEPTH")
			//	t := opts.VarsList["DEPTH"].Types
			assert.Equal(0.0, pres[0])    // test the first pressure value
			assert.Equal(5.8, pres[size]) // test the last pressure value
			temp := ext.Data("TEMP")
			assert.Equal(23.32, temp[0])
			assert.Equal(23.23, temp[size])
			svel := ext.Data("SVEL")
			assert.Equal(1530.25, svel[0])
			assert.Equal(1530.10, svel[size])
		}
	}
	assert.Equal(config.Kml.FileName, "pirata-fr26.kml")

}
