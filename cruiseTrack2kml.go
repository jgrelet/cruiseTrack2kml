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
	xbtFile string
	kmlFile string
)

const version string = "cruiseTrack2kml, version 0.3.1  J.Grelet IRD - US191 IMAGO"

// toml config structure
type tomlConfig struct {
	Cruise  string
	Ship    string
	Windows struct {
		TsgFile string
		CtdFile string
		XbtFile string
	}
	Unix struct {
		TsgFile string
		CtdFile string
		XbtFile string
	}
	CtdPlots      string
	XbtPlots      string
	TsgPlots      string
	CtdPrefix     int
	XbtPrefix     int
	SizePlots     int
	StationNumber bool
	TsgSplit      string
	TsgSkip       int
	CtdSplit      string
	CtdSkip       int
	XbtSplit      string
	XbtSkip       int
}

// usefull macro
var p = fmt.Println
var pf = fmt.Printf

// Basic flag declarations are available for string, integer, and boolean options.
func init() {
	var (
		help       *bool
		configFile string
	)

	help = flag.Bool("help", false, "display the help...")
	echo = flag.Bool("echo", false, "display source to stdout")
	flag.StringVar(&configFile, "config", "", "use alternate .toml config file")
	flag.StringVar(&kmlFile, "output", "", "use alternate  outpout kml file (default is toml Cruise name)")
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
	if kmlFile == "" {
		kmlFile = fmt.Sprintf("%s.kml", config.Cruise)
	}

	if runtime.GOOS == "windows" {
		tsgFile = config.Windows.TsgFile
		ctdFile = config.Windows.CtdFile
		xbtFile = config.Windows.XbtFile
	} else {
		tsgFile = config.Unix.TsgFile
		ctdFile = config.Unix.CtdFile
		xbtFile = config.Unix.XbtFile
	}

	pf("Cruise: %s\n", config.Cruise)
	pf("Ship: %s\n", config.Ship)
	pf("CtdPlots: %s\n", config.CtdPlots)
	pf("TsgPlots: %s\n", config.TsgPlots)
	pf("CtdFile: %s\n", ctdFile)
	pf("XbtFile: %s\n", xbtFile)
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
	const elevation = 0.0

	// to change the flags on the default logger
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	// create KML header
	k := gokml.NewKML(fmt.Sprintf("%s KML", config.Cruise))
	f := gokml.NewFolder(fmt.Sprintf("%s cruise", config.Cruise),
		"This document was generated with cruiseTrack2kml program")
	k.AddFeature(f)

	// define new style for station icons
	// NewStyle(name string, alpha uint8, red uint8, green uint8, blue uint8)
	placeProfiles := gokml.NewStyle("ProfileStyle", 255, 0, 255, 255)
	// collection of icons Google makes available for Google Earth
	// http://kml4earth.appspot.com/icons.html
	// places.SetIconURL("http://maps.google.com/mapfiles/kml/paddle/wht-circle.png")
	placeProfiles.SetIconURL("http://maps.google.com/mapfiles/kml/pushpin/wht-pushpin.png")
	f.AddFeature(placeProfiles)

	// define new style for station icons
	placeStations := gokml.NewStyle("StationStyle", 255, 255, 0, 0)
	// collection of icons Google makes available for Google Earth
	// http://kml4earth.appspot.com/icons.html
	// places.SetIconURL("http://maps.google.com/mapfiles/kml/paddle/wht-circle.png")
	placeStations.SetIconURL("http://maps.google.com/mapfiles/kml/pushpin/wht-pushpin.png")
	f.AddFeature(placeStations)

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
	tsg := fileExtractor.NewFileExtractor(opts)

	// read the file
	if err := tsg.Read(); err != nil {
		log.Fatalln(err)
	}

	// display the value
	lats := tsg.Data("LATITUDE")
	lons := tsg.Data("LONGITUDE")
	for i := 0; i < tsg.Size(); i++ {
		lat := lats[i]
		lon := lons[i]

		// create new point
		//u, _ := strconv.ParseFloat(lat.(string), 64)
		//v, _ := strconv.ParseFloat(lon.(string), 64)
		np := gokml.NewPoint(lat.(float64), lon.(float64), elevation)
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
	opts = fileExtractor.NewFileExtractOptions().SetFilename(ctdFile)
	opts.SetVarsList(config.CtdSplit)
	opts.SetSkipLine(config.CtdSkip)

	// print options
	p(opts)

	// initialize fileExtractor from options
	ctd := fileExtractor.NewFileExtractor(opts)

	// read the file
	if err := ctd.Read(); err != nil {
		log.Fatalln(err)
	}

	// display the value
	profiles := ctd.Data("PRFL")
	latString := ctd.Data("LAT")
	latSign := ctd.Data("LAT_S")
	lonString := ctd.Data("LON")
	lonSign := ctd.Data("LON_S")
	beginDates := ctd.Data("BEGIN_DATE")
	beginTimes := ctd.Data("BEGIN_TIME")
	endDates := ctd.Data("END_DATE")
	endTimes := ctd.Data("END_TIME")
	pmaxs := ctd.Data("PMAX")
	bottomDepths := ctd.Data("BOTTOM_DEPTH")
	profileFormat := fmt.Sprintf("%%0%dd", config.CtdPrefix)
	for i := 0; i < ctd.Size(); i++ {
		profile := profiles[i]
		beginDate := beginDates[i]
		beginHour := beginTimes[i]
		endDate := endDates[i]
		endHour := endTimes[i]
		lat := fmt.Sprintf("%s %s", latString[i].(string), latSign[i].(string))
		lon := fmt.Sprintf("%s %s", lonString[i].(string), lonSign[i].(string))
		pmax := pmaxs[i]
		bottomDepth := bottomDepths[i]
		// convert profile to integer with the rigth Printf format
		theProfile := fmt.Sprintf(profileFormat, profile.(int))
		/*
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
		*/
		filename := " "
		typeCast := " "

		// convert position to decimal values
		var err error
		if latitude, err = Position2Decimal(lat); err != nil {
			log.Fatal(err)
		}
		if longitude, err = Position2Decimal(lon); err != nil {
			log.Fatal(err)
		}
		// add positions of stations on map
		// create new point for station
		st := gokml.NewPoint(latitude, longitude, elevation)

		// fill Ascii header from CTD file, use <pre> markup for LF
		header := fmt.Sprintf("\n<pre>Station: %s Type: %s  Filename: %s\n"+
			"Begin Date: %s %s  End Date: %s %s\nLatitude: %s  Longitude: %s \n"+
			"Max depth: %6.1f   Bathy: %6.1f</pre>\n",
			theProfile, typeCast, filename, beginDate, beginHour,
			endDate, endHour, lat, lon, pmax, bottomDepth)

		// fill description markup with the CTD picture link inside <![CDATA[...]]>
		// All characters enclosed between these two sequences are interpreted as characters
		files := fmt.Sprintf(config.CtdPlots, theProfile)
		description := fmt.Sprintf("%s<![CDATA[\n<img src='%s' width='%d' />]]>",
			header, files, config.SizePlots)

		// add new Placemark markup with station number, description and location (point object)
		var newName string
		if config.StationNumber {
			newName = fmt.Sprintf("%d", profile)
		} else {
			newName = fmt.Sprintf("%d", i)
		}
		pm := gokml.NewPlacemark(newName, description, st)
		pm.SetStyle("StationStyle")

		// add placemark markup to kml file
		f.AddFeature(pm)
	}
	// read XBT positions
	if xbtFile != "none" {
		opts = fileExtractor.NewFileExtractOptions().SetFilename(xbtFile)
		opts.SetVarsList(config.XbtSplit)
		opts.SetSkipLine(config.XbtSkip)

		// print options
		p(opts)

		// initialize fileExtractor from options
		xbt := fileExtractor.NewFileExtractor(opts)

		// read the file
		if err := xbt.Read(); err != nil {
			log.Fatalln(err)
		}

		// display the value
		profiles = xbt.Data("PRFL")
		latString = xbt.Data("LAT")
		latSign = xbt.Data("LAT_S")
		lonString = xbt.Data("LON")
		lonSign = xbt.Data("LON_S")
		beginDates = xbt.Data("BEGIN_DATE")
		beginTimes = xbt.Data("BEGIN_TIME")
		pmaxs = xbt.Data("PMAX")
		typeProbe := xbt.Data("PROBE")
		profileFormat = fmt.Sprintf("%%0%dd", config.XbtPrefix)
		for i := 0; i < xbt.Size(); i++ {
			profile := profiles[i]
			beginDate := beginDates[i]
			beginHour := beginTimes[i]
			lat := fmt.Sprintf("%s %s", latString[i].(string), latSign[i].(string))
			lon := fmt.Sprintf("%s %s", lonString[i].(string), lonSign[i].(string))
			pmax := pmaxs[i]
			// convert profile to integer with the rigth Printf format
			theProfile := fmt.Sprintf(profileFormat, profile)
			theProbe := typeProbe[i]
			filename := " "

			// convert position to decimal values
			var err error
			if latitude, err = Position2Decimal(lat); err != nil {
				log.Fatal(err)
			}
			if longitude, err = Position2Decimal(lon); err != nil {
				log.Fatal(err)
			}
			// add positions of stations on map
			// create new point for station
			st := gokml.NewPoint(latitude, longitude, elevation)

			// fill Ascii header from XBT file, use <pre> markup for LF
			header := fmt.Sprintf("\n<pre>Profile: %s Type: %s  Filename: %s\n"+
				"Begin Date: %s %s\nLatitude: %s  Longitude: %s \n"+
				"Max depth: %6.1f</pre>\n",
				theProfile, theProbe, filename, beginDate, beginHour, lat, lon, pmax)

			// fill description markup with the CTD picture link inside <![CDATA[...]]>
			// All characters enclosed between these two sequences are interpreted as characters
			files := fmt.Sprintf(config.XbtPlots, theProfile)
			description := fmt.Sprintf("%s<![CDATA[\n<img src='%s' width='%d' />]]>",
				header, files, config.SizePlots)

			// add new Placemark markup with station number, description and location (point object)
			var newName string
			if config.StationNumber {
				newName = fmt.Sprintf("%d", profile)
			} else {
				newName = fmt.Sprintf("%d", i)
			}
			pm := gokml.NewPlacemark(newName, description, st)
			pm.SetStyle("ProfileStyle")

			// add placemark markup to kml file
			f.AddFeature(pm)
		}
	}

	// display kml content to screen
	if *echo {
		pf("%s", k.Render())
	}

	// open ASCII file for writing result
	fidKml, err := os.Create(kmlFile)
	p(kmlFile)
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
	pf("TSG mark: %d\n", tsg.Size())
	pf("CTD mark: %d\n", ctd.Size())
	//pf("XBT mark: %d\n", xbt.Size())
}
