package fileExtractor

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
)

// FileExtractOptions contains configurable options for read an ASCII file.
type FileExtractOptions struct {
	filename string
	hdr      []string
	varsList map[string]int
	skip     int // number of line to skip before read data
}

// FileExtracter contains FileExtractOptions object and map data extracted from ASCII file.
type FileExtracter struct {
	options FileExtractOptions
	data    map[string]interface{}
}

// NewFileExtractOptions will create a new FileExtractOptions type with some
// default values.
//   all empty ...
func NewFileExtractOptions() *FileExtractOptions {
	o := &FileExtractOptions{
		filename: "",
		hdr:      []string{},
		varsList: map[string]int{},
		skip:     0,
	}
	return o
}

// SetFilename will set the ASCII file containing data to read and decode
func (o *FileExtractOptions) SetFilename(filename string) *FileExtractOptions {
	o.filename = filename
	return o
}

// Filename will get the ASCII file name (getter)
func (o *FileExtractOptions) Filename() string {
	return o.filename
}

// SetVars will set the parameters and their columns to extract from file
func (o *FileExtractOptions) SetVarsList(split string) *FileExtractOptions {
	// construct map from split
	fields := strings.Split(split, ",")
	for i := 0; i < len(fields); i += 2 {
		if v, err := strconv.Atoi(fields[i+1]); err == nil {
			o.varsList[fields[i]] = v - 1
			o.hdr = append(o.hdr, fields[i])
		} else {
			log.Fatalf("Check the input of SetVars: %v\n", err)
		}
	}
	return o
}

// Filename will get the ASCII file name (getter)
func (o *FileExtractOptions) VarsList() map[string]int {
	return o.varsList
}

// display FileExtractOptions object
func (o FileExtractOptions) String() string {
	return fmt.Sprintf("File: %s\nFields:%s\nVars: %v\n", o.filename, o.hdr, o.varsList)
}

// NewFileExtracter will create a new FileExtracter type with some values from
// configuration (not implemented) or from constructor (setter methods)
func NewFileExtracter(o *FileExtractOptions) *FileExtracter {
	fe := &FileExtracter{}
	fe.options = *o
	fe.data = make(map[string]interface{})
	for _, name := range fe.options.hdr {
		fe.data[name] = []float64{}
	}
	return fe
}

// Read an ASCII file and extract data and save then to map data
func (ext *FileExtracter) Read() {
	fid, err := os.Open(ext.options.filename)
	if err != nil {
		log.Fatal(err)
	}
	defer fid.Close()

	// open bufio for file
	scanner := bufio.NewScanner(fid)

	// read file
	for scanner.Scan() {
		// parse each line to string
		str := scanner.Text()
		values := strings.Fields(str)
		// fill map data
		for key, column := range ext.options.varsList {
			if v, err := strconv.ParseFloat(values[column], 64); err == nil {
				ext.data[key] = append(ext.data[key].([]float64), v)
			}
		}
	}
}

// print the result
func (ext FileExtracter) Print() {
	for key, _ := range ext.options.varsList {
		fmt.Printf("%s: %7.3f\n", key, ext.data[key])
	}
	fmt.Println()
}
