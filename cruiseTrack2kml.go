// cruiseTrack2kml
package main

import (
	"bufio"
	"errors"
	"fmt"
	"log"
	"math"
	"os"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"time"

	"flag"

	"github.com/BurntSushi/toml"
	"github.com/gershwinlabs/gokml"
	"github.com/jgrelet/cruiseTrack2kml/fileExtractor"
)

// arg var
var (
	echo    *bool
	config  tomlConfig
	tsgFile string
	ctdFile string
	kmlFile string
)

const version string = "cruiseTrack2kml, version 0.21  J.Grelet IRD - US191 IMAGO"

// toml config structure
type tomlConfig struct {
	Cruise  string
	Ship    string
	Windows struct {
		TsgFile string
		CtdFile string
		KmlFile string
	}
	Unix struct {
		TsgFile string
		CtdFile string
		KmlFile string
	}
	CtdPlots      string
	TsgPlots      string
	CtdPrefix     int
	SizePlots     int
	StationNumber bool
	TsgSplit      string
	TsgSkip       int
}

// usefull macro
var p = fmt.Println
var pf = fmt.Printf

// Basic flag declarations are available for string, integer, and boolean options.
func init() {
	var (
		help       *bool
		configFile string
	//	 cruise string
	)

	help = flag.Bool("help", false, "display the help...")
	//	flag.StringVar(&cruise, []string{"cruise"}, "", "cruise name")
	echo = flag.Bool("echo", false, "display source to stdout")
	flag.StringVar(&configFile, "config", "", "use alternate .toml config file")
	flag.Parse()
	if *help {
		flag.PrintDefaults()
		os.Exit(0)
	}

	// print version
	p(version)
	p(time.Now().Format(time.RFC850) + "\n")

	if configFile == "" {
		configFile = "config.toml"
	}
	//  read config file
	if _, err := toml.DecodeFile(configFile, &config); err != nil {
		p(err)
		return
	}

	if runtime.GOOS == "windows" {
		tsgFile = config.Windows.TsgFile
		ctdFile = config.Windows.CtdFile
		kmlFile = config.Windows.KmlFile
	} else {
		tsgFile = config.Unix.TsgFile
		ctdFile = config.Unix.CtdFile
		kmlFile = config.Unix.KmlFile
	}

	pf("Cruise: %s\n", config.Cruise)
	pf("Ship: %s\n", config.Ship)
	pf("CtdPlots: %s\n", config.CtdPlots)
	pf("TsgPlots: %s\n", config.TsgPlots)
	pf("CtdFile: %s\n", ctdFile)
	pf("TsgFile: %s\n", tsgFile)
	pf("KmlFile: %s\n", kmlFile)
}

// Position2Decimal convert position "DD MM.SS S" to decimal position
func Position2Decimal(pos string) (float64, error) {

	var multiplier float64 = 1
	var value float64

	var regNmeaPos = regexp.MustCompile(`(\d+)\W(\d+.\d+)\s+(\w)`)

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

	var latitude, longitude float64

	// to change the flags on the default logger
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	// create KML header
	k := gokml.NewKML(fmt.Sprintf("%s KML", config.Cruise))
	f := gokml.NewFolder(fmt.Sprintf("%s cruise", config.Cruise),
		"This document was generated with cruiseTrack2kml program")
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

	// define new line
	ls := gokml.NewLineString()

	// read TSG track
	opts := fileExtractor.NewFileExtractOptions().SetFilename(tsgFile)
	opts.SetVarsList(config.TsgSplit)
	opts.SetSkipLine(config.TsgSkip)

	// print options
	p(opts)

	// initialize fileExtractor from options
	ext := fileExtractor.NewFileExtractor(opts)

	// read the file
	ext.Read()

	// display the value
	lats := ext.Data()["LATITUDE"]
	lons := ext.Data()["LONGITUDE"]
	for i := 0; i < ext.Size(); i++ {
		lat := lats[i]
		lon := lons[i]

		// create new point
		np := gokml.NewPoint(lat, lon, 0.0)
		// add point to line
		ls.AddPoint(np)
	}

	// fill description markup with the TSG picture link inside <![CDATA[...]]>
	// All characters enclosed between these two sequences are interpreted as characters
	description := fmt.Sprintf("<![CDATA[\n<img src='%s' width='%d' />]]>",
		config.TsgPlots, config.SizePlots)
	// define block Placemark for line
	placemark := fmt.Sprintf("%s cruise track on R/V %s", config.Cruise, config.Ship)
	pm := gokml.NewPlacemark(placemark, description, ls)
	pm.SetStyle("TrackStyle")
	// add placemark markup to kml file
	f.AddFeature(pm)

	// read CTD position
	fidCtd, err := os.Open(ctdFile)
	if err != nil {
		log.Fatal(err)
	}
	defer fidCtd.Close()

	scannerCtd := bufio.NewScanner(fidCtd)
	i := 1
	var filename string
	var typeCast string

	// TODOS:
	// check the first valid line with a regex
	// add the column label inside .toml file
	for scannerCtd.Scan() {
		str := scannerCtd.Text()
		values := strings.Fields(str)
		//p(values)

		// skip first line
		profile := values[0]
		if profile == config.Cruise {
			continue
		}
		// convert profile with the right format (usually %03d or %05d)
		prfl, _ := strconv.Atoi(profile)
		format := fmt.Sprintf("%%0%1dd", config.CtdPrefix)
		profile = fmt.Sprintf(format, prfl)
		// extract data from station line
		beginDate := values[1]
		beginHour := values[2]
		endDate := values[3]
		endHour := values[4]
		lat := fmt.Sprintf("%s %s", values[5], values[6])
		lon := fmt.Sprintf("%s %s", values[7], values[8])
		pmax := values[9]
		bottomDepth := values[10]
		if len(values) > 11 {
			filename = values[11]
		} else {
			filename = " "
		}
		if len(values) > 12 {
			typeCast = values[12]
		} else {
			typeCast = " "
		}

		// convert position to decimal values
		if latitude, err = Position2Decimal(fmt.Sprintf("%s %s", values[5],
			values[6])); err != nil {
			log.Fatal(err)
		}
		if longitude, err = Position2Decimal(fmt.Sprintf("%s %s", values[7],
			values[8])); err != nil {
			log.Fatal(err)
		}
		// add positions of stations on map
		// create new point for station
		st := gokml.NewPoint(latitude, longitude, 0.0)

		// fill Ascii header from CTD file, use <pre> markup for LF
		header := fmt.Sprintf("\n<pre>Station n° %s  Type: %s  Filename: %s\n"+
			"Begin Date: %s %s  End Date: %s %s\nLatitude: %s  Longitude: %s \n"+
			"Max depth: %s   Bathy: %s</pre>\n",
			profile, typeCast, filename, beginDate, beginHour,
			endDate, endHour, lat, lon, pmax, bottomDepth)

		// fill description markup with the CTD picture link inside <![CDATA[...]]>
		// All characters enclosed between these two sequences are interpreted as characters
		files := fmt.Sprintf(config.CtdPlots, profile)
		description := fmt.Sprintf("%s<![CDATA[\n<img src='%s' width='%d' />]]>",
			header, files, config.SizePlots)

		// add new Placemark markup with station number, description and location (point object)
		//
		var newName string
		if config.StationNumber {
			newName = fmt.Sprintf("%s", profile)
		} else {
			newName = fmt.Sprintf("%d", i)
		}
		pm := gokml.NewPlacemark(newName, description, st)
		pm.SetStyle("ProfileStyle")

		// add placemark markup to kml file
		f.AddFeature(pm)
		i++
	}
	//p(k)

	// display kml content to screen
	if *echo {
		pf("%s", k.Render())
	}

	// open ASCII file for writing result
	fidKml, err := os.Create(kmlFile)
	if err != nil {
		log.Fatal(err)
	}
	defer fidKml.Close()

	// use buffered mode for writing
	fbufKml := bufio.NewWriter(fidKml)
	// write kml to file
	fmt.Fprintln(fbufKml, k.Render())
	fbufKml.Flush()

	// display the filename to screen
	p(kmlFile)
}
