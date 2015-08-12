// cruiseTrack2kml
package main

import (
	"bufio"
	"errors"
	"fmt"
	"github.com/gershwinlabs/gokml"
	"log"
	"math"
	"os"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"time"
)

const version string = "cruiseTrack2kml, version 0.1  Jgrelet IRD - Cassiopee cruise - R/V Atalante"

var tsg_file_windows = "M:/CASSIOPEE/data-processing/THERMO/cassiopee.gps"
var ctd_file_windows = "M:/CASSIOPEE/data-processing/CTD/cassiopee.ctd"
var kml_file_windows = "M:/CASSIOPEE/data-processing/CTD/tracks/cassiopee.kml"

var tsg_file_unix = "/m/CASSIOPEE/data-processing/THERMO/cassiopee.gps"
var ctd_file_unix = "/m/CASSIOPEE/data-processing/CTD/cassiopee.ctd"
var kml_file_unix = "/m/CASSIOPEE/data-processing/CTD/tracks/cassiopee.kml"

// usefull macro
var p = fmt.Println
var f = fmt.Printf

// convert position "DD MM.SS S" to decimal position
func Position2Decimal(pos string) (float64, error) {

	var multiplier float64 = 1
	var value float64

	var regNmeaPos = regexp.MustCompile(`(\d+)°(\d+.\d+)\s+(\w)`)

	if strings.Contains(pos, "S") || strings.Contains(pos, "W") {
		multiplier = -1.0
	}
	match := regNmeaPos.MatchString(pos)
	if match {
		res := regNmeaPos.FindStringSubmatch(pos)
		deg, _ := strconv.ParseFloat(res[1], 64)
		min, _ := strconv.ParseFloat(res[2], 64)
		tmp := math.Abs(min)
		sec := (tmp - min) * 100.0
		value = (deg + (min+sec/100.0)/60.0) * multiplier
		//fmt.Println("positionDeci:", pos, " -> ", value)
	} else {
		return 1e36, errors.New("positionDeci: failed to decode position")
	}
	return value, nil
}

func main() {

	var tsg_file string
	var ctd_file string
	var kml_file string

	// print version
	fmt.Println(version)
	fmt.Println(time.Now().Format(time.RFC850))

	// create KML header
	k := gokml.NewKML("Cassiopee KML")
	f := gokml.NewFolder("Cassiopee Folder", "This is Cassiopee cruise folder")
	k.AddFeature(f)

	// define new style for station icons
	places := gokml.NewStyle("ProfileStyle", 255, 255, 0, 0)
	// collection of icons Google makes available for Google Earth
	// http://kml4earth.appspot.com/icons.html
	//places.SetIconURL("http://maps.google.com/mapfiles/kml/paddle/wht-circle.png")
	places.SetIconURL("http://maps.google.com/mapfiles/kml/pushpin/red-pushpin.png")
	f.AddFeature(places)

	// define style for line
	track := gokml.NewStyle("TrackStyle", 255, 0, 255, 0)
	f.AddFeature(track)

	// read TSG track
	if runtime.GOOS == "windows" {
		tsg_file = tsg_file_windows
		ctd_file = ctd_file_windows
		kml_file = kml_file_windows
	} else {
		tsg_file = tsg_file_unix
		ctd_file = ctd_file_unix
		kml_file = kml_file_unix
	}
	fid_tsg, err := os.Open(tsg_file)
	if err != nil {
		log.Fatal(err)
	}
	defer fid_tsg.Close()

	// open bufio for tsg
	scanner_tsg := bufio.NewScanner(fid_tsg)

	// define new line
	ls := gokml.NewLineString()

	// read tsg file
	for scanner_tsg.Scan() {
		// parse lat and lon from file, columns 2 and 3
		str := scanner_tsg.Text()
		values := strings.Fields(str)
		//p(values)
		lat, _ := strconv.ParseFloat(values[1], 64)
		lon, _ := strconv.ParseFloat(values[2], 64)

		// create new point
		np := gokml.NewPoint(lat, lon, 0.0)
		// set point to line
		ls.AddPoint(np)
	}

	// define block Placemark for line
	pm := gokml.NewPlacemark("Atalante track", "", ls)
	pm.SetStyle("TrackStyle")
	f.AddFeature(pm)

	// read CTD position
	fid_ctd, err := os.Open(ctd_file)
	if err != nil {
		log.Fatal(err)
	}
	defer fid_ctd.Close()

	var latitude, longitude float64

	scanner_ctd := bufio.NewScanner(fid_ctd)
	var i int = 1

	for scanner_ctd.Scan() {
		str := scanner_ctd.Text()
		values := strings.Fields(str)
		//p(values)
		profile := values[0]
		if profile == "CASSIOPEE" {
			continue
		}
		begin_date := values[1]
		begin_hour := values[2]
		end_date := values[3]
		end_hour := values[4]
		lat := fmt.Sprintf("%s %s", values[5], values[6])
		lon := fmt.Sprintf("%s %s", values[7], values[8])
		pmax := values[9]
		bottom_depth := values[10]
		type_cast := values[11]
		filename := values[12]

		if latitude, err = Position2Decimal(fmt.Sprintf("%s %s", values[5], values[6])); err != nil {
			os.Exit(3)
		}
		fmt.Sprintf("Long: %s %s\n", values[7])
		if longitude, err = Position2Decimal(fmt.Sprintf("%s %s", values[7], values[8])); err != nil {
			os.Exit(4)
		}
		st := gokml.NewPoint(latitude, longitude, 0.0)
		header := fmt.Sprintf("\n<pre>Station %s  Type: %s  Filename: %s\nBegin Date: %s %s  End Date: %s %s\nLatitude: %s  Longitude: %s \nMax depth: %s   Bathy: %s</pre>\n",
			profile, type_cast, filename, begin_date, begin_hour, end_date, end_hour, lat, lon, pmax, bottom_depth)
		description := fmt.Sprintf("%s<![CDATA[\n<img src='http://atalante/cassiopee/data-processing/CTD/plots/downcast/dcsp%s-TS02Dens.jpg' width='700' /><br/&gt;%d<br/> ]]>", header, profile, i)
		pm := gokml.NewPlacemark(fmt.Sprintf("%d", i), description, st)
		pm.SetStyle("ProfileStyle")
		f.AddFeature(pm)
		i++
	}
	//p(k)

	// display kml content to screen
	fmt.Printf("%s", k.Render())

	// open ASCII file for writing result
	fid_kml, err := os.Create(kml_file)
	if err != nil {
		os.Exit(2)
	}
	defer fid_kml.Close()

	// use buffered mode for writing
	fbuf_kml := bufio.NewWriter(fid_kml)
	// write kml to file
	fmt.Fprintln(fbuf_kml, k.Render())
	fbuf_kml.Flush()
}
