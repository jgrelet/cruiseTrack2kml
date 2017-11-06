package fileExtractor

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strconv"
	"strings"
)

// usefull macro
var p = fmt.Println
var pf = fmt.Printf

// use for debug mode
var debugMode = true
var debug = ioutil.Discard

// Types contain column number and value type
type Types struct {
	column int
	types  string
}

// FileExtractOptions contains configurable options for read an ASCII file.
type FileExtractOptions struct {
	filename  string
	hdr       []string
	varsList  map[string]Types
	separator string
	skipLine  int // number of line to skip before read data
}

// FileExtractor contains FileExtractOptions object and map data extracted from ASCII file.
type FileExtractor struct {
	*FileExtractOptions
	data map[string][]string
	size int
}

// NewFileExtractOptions will create a new FileExtractOptions type with some
// empty default values.
func NewFileExtractOptions() *FileExtractOptions {
	o := &FileExtractOptions{
		filename:  "",
		hdr:       []string{},
		varsList:  map[string]Types{},
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
	m := map[string]Types{}
	h := []string{}

	// construct map from split
	fields := strings.Split(split, ",")
	if len(fields)%3 != 0 {
		log.Fatalf("Check the pa list: %s, invalid number of parameters, not modulo 3", split)
	}
	for i := 0; i < len(fields); i += 3 {
		if v, err := strconv.Atoi(fields[i+1]); err == nil {
			m[fields[i]] = Types{column: v, types: fields[i+2]}
			//pf("%#v -> %#v\n", h, fields[i])
			h = append(h, fields[i])
		} else {
			log.Fatalf("Check the input of SetVars: [%v]: %v -> %v\n",
				fields[i], fields[i+1], err)
		}
	}
	// copy map and header list to FileExtractOptions object
	o.varsList = m
	o.hdr = h
	return o
}

// VarsList getter
// TODOS: should return map[string]int ?
func (o *FileExtractOptions) VarsList() map[string]Types {
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
		data:               make(map[string][]string),
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
		log.Printf("FileExtractor.Read(): can't open %s, check it !!!\n", fe.filename)
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
		for key, value := range fe.varsList {
			// slice index start at 0
			ind := value.column - 1
			if ind < len(values) {
				fe.data[key] = append(fe.data[key], values[ind])
			} else {
				continue
			}
		}
		fe.size++
	}
	return nil
}

// Data return a slice of type for the var s
func (fe *FileExtractor) Data(s string) []interface{} {
	var data = make([]interface{}, len(fe.data[s]))
	sl := fe.data[s]
	switch v := fe.varsList[s].types; v {
	case "int", "int32", "int64":
		for i := 0; i < len(sl); i++ {
			if v, err := strconv.Atoi(sl[i]); err == nil {
				data[i] = v
			}
		}
	case "float", "float32", "float64":
		for i := 0; i < len(sl); i++ {
			if v, err := strconv.ParseFloat(sl[i], 64); err == nil {
				data[i] = v
			}
		}
	case "string", "char":
		for i := 0; i < len(sl); i++ {
			data[i] = sl[i]
		}
	default:
		log.Fatalf("Invalid type: %v", fe.varsList[s].types)
	}
	return data
}

// String print the result
func (fe FileExtractor) String() string {
	var s []string
	for key := range fe.varsList {
		s = append(s, fmt.Sprintf("\n%s: %s", key, fe.data[key]))
	}
	return strings.Join(s, "")
}
