package fileExtractor

import (
	"bufio"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"strconv"
	"strings"
)

// usefull macro
var p = fmt.Println

// use for debug mode
var debugMode = false
var debug io.Writer = ioutil.Discard

// FileExtractOptions contains configurable options for read an ASCII file.
type FileExtractOptions struct {
	filename  string
	hdr       []string
	varsList  map[string]int
	separator string
	skipLine  int // number of line to skip before read data
}

// FileExtractor contains FileExtractOptions object and map data extracted from ASCII file.
type FileExtractor struct {
	*FileExtractOptions
	data map[string][]interface{}
	size int
}

// NewFileExtractOptions will create a new FileExtractOptions type with some
// empty default values.
func NewFileExtractOptions() *FileExtractOptions {
	o := &FileExtractOptions{
		filename:  "",
		hdr:       []string{},
		varsList:  map[string]int{},
		separator: "",
		skipLine:  0,
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

// SetVarsList will set the parameters and their columns to extract from file
func (o *FileExtractOptions) SetVarsList(split string) *FileExtractOptions {
	// create empty map and header list
	m := map[string]int{}
	h := []string{}

	// construct map from split
	fields := strings.Split(split, ",")
	for i := 0; i < len(fields); i += 2 {
		if v, err := strconv.Atoi(fields[i+1]); err == nil {
			m[fields[i]] = v
			h = append(h, fields[i])
		} else {
			log.Fatalf("Check the input of SetVars: %v\n", err)
		}
	}
	// copy map and header list to FileExtractOptions object
	o.varsList = m
	o.hdr = h
	return o
}

// VarsList getter
func (o *FileExtractOptions) VarsList() map[string]int {
	return o.varsList
}

// SetSeparator will override the default separator (space)
func (o *FileExtractOptions) SetSeparator(sep string) *FileExtractOptions {
	o.separator = sep
	return o
}

// SetSkipLine will set to skip header line
func (o *FileExtractOptions) SetSkipLine(line int) *FileExtractOptions {
	o.skipLine = line
	return o
}

// display FileExtractOptions object
func (o FileExtractOptions) String() string {
	return fmt.Sprintf("File: %s\nFields:%s\nVars: %v\nSkipLine: %d\n",
		o.filename, o.hdr, o.varsList, o.skipLine)
}

// NewFileExtractor will create a new FileExtractor type with some values from
// configuration (not implemented)
func NewFileExtractor(o *FileExtractOptions) *FileExtractor {
	if debugMode {
		debug = os.Stdout
	}
	// in this constructor, we use composition (or embedding) vs inheritance
	fe := &FileExtractor{
		FileExtractOptions: o,
		data:               make(map[string][]interface{}),
		size:               0,
	}
	// initialize map for each key to a slice
	/*
		for _, name := range fe.hdr {
			fe.data[name] = []interface{}
		}
	*/
	return fe
}

// Size get the the size of map data
func (fe FileExtractor) Size() int {
	return fe.size
}

// Read an ASCII file and extract data and save then to map data
func (fe *FileExtractor) Read() error {
	fid, err := os.Open(fe.filename)
	if err != nil {
		return err
	}
	defer fid.Close()

	// open bufio for file
	scanner := bufio.NewScanner(fid)

	// skip some lines
	for i := 0; i < fe.skipLine; i++ {
		scanner.Scan()
	}

	// read file
	for scanner.Scan() {
		var values []string

		// parse each line to string
		str := scanner.Text()

		// split the string str with defined separator
		if fe.separator != "" {
			values = strings.Split(str, fe.separator)
		} else {
			// split the string str with one or more space
			values = strings.Fields(str)
		}

		// fill map data
		for key, column := range fe.varsList {
			// slice index start at 0
			ind := column - 1

			if ind < len(values) {
				/*
					if v, err := strconv.ParseFloat(values[ind], 64); err == nil {
						// column start at zero
						fe.data[key] = append(fe.data[key], v)
					}
				*/
				fe.data[key] = append(fe.data[key], values[ind])
			} else {
				continue
			}
		}

		fe.size++

	}
	return nil
}

// Data get the the size of map data
func (fe *FileExtractor) Data() map[string][]interface{} {
	return fe.data
}

// String print the result
func (fe FileExtractor) String() string {
	var s []string
	for key := range fe.varsList {
		s = append(s, fmt.Sprintf("\n%s: %7.3f", key, fe.data[key]))
	}
	return strings.Join(s, "")
}
