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
	skipLine int // number of line to skip before read data
}

// FileExtractor contains FileExtractOptions object and map data extracted from ASCII file.
type FileExtractor struct {
	*FileExtractOptions
	data map[string][]float64
	size int
}

// NewFileExtractOptions will create a new FileExtractOptions type with some
// empty default values.
func NewFileExtractOptions() *FileExtractOptions {
	o := &FileExtractOptions{
		filename: "",
		hdr:      []string{},
		varsList: map[string]int{},
		skipLine: 0,
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
			o.varsList[fields[i]] = v
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

// NewFileExtracter will create a new FileExtractor type with some values from
// configuration (not implemented)
func NewFileExtractor(o *FileExtractOptions) *FileExtractor {
	// in this constructor, we use composition (or embedding) vs inheritance
	fe := &FileExtractor{
		FileExtractOptions: o,
		data:               make(map[string][]float64),
		size:               0,
	}
	// initialize map for each key to a slice
	for _, name := range fe.hdr {
		fe.data[name] = []float64{}
	}
	return fe
}

// get the the size of map data
func (fe FileExtractor) Size() int {
	return fe.size
}

// Read an ASCII file and extract data and save then to map data
func (ext *FileExtractor) Read() {
	fid, err := os.Open(ext.filename)
	if err != nil {
		log.Fatal(err)
	}
	defer fid.Close()

	// open bufio for file
	scanner := bufio.NewScanner(fid)

	// skip some lines
	for i := 0; i < ext.skipLine; i++ {
		scanner.Scan()
	}

	// read file
	for scanner.Scan() {
		// parse each line to string
		str := scanner.Text()
		values := strings.Fields(str)
		// fill map data
		for key, column := range ext.varsList {
<<<<<<< HEAD
			if v, err := strconv.ParseFloat(values[column], 64); err == nil {
				// column start at zero
				ext.data[key] = append(ext.data[key], v-1)
=======
			if values[column] == "-1" {
				ext.data[key] = append(ext.data[key], 1e36)
>>>>>>> e001682e4b5a74645d850689d72a7697e52dd11d
			} else {
				if v, err := strconv.ParseFloat(values[column], 64); err == nil {
					ext.data[key] = append(ext.data[key], v)
				} else {
					log.Fatalf("Bad format: %v.  Expected float number, may be skip line", err)
				}
			}
		}
		ext.size += 1
	}
}

// get the the size of map data
func (fe *FileExtractor) Data() map[string][]float64 {
	return fe.data
}

// print the result
func (ext FileExtractor) Print() {
	for key, _ := range ext.varsList {
		fmt.Printf("%s: %7.3f\n", key, ext.data[key])
	}
	fmt.Println()
}
