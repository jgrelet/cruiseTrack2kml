// cruiseTrack2kml
package main

import (
	"bufio"
	"errors"
	"fmt"
	"github.com/BurntSushi/toml"
	"github.com/gershwinlabs/gokml"
	flag "github.com/tcnksm/mflag"
	"log"
	"math"
	"os"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"time"
)

// arg var
var (
	config   tomlConfig
	tsg_file string
	ctd_file string
	kml_file string
)

const version string = "cruiseTrack2kml, version 0.2  J.Grelet IRD - US191 IMAGO"

// toml config structure
type tomlConfig struct {
	Cruise  string
	Ship    string
	Windows struct {
		Tsg_file string
		Ctd_file string
		Kml_file string
	}
	Unix struct {
		Tsg_file string
		Ctd_file string
		Kml_file string
	}
	Ctd_plots      string
	Tsg_plots      string
	Ctd_prefix     int
	Size_plots     int
	Station_number bool
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

	help = flag.Bool([]string{"h", "#a", "-help", "#aide", "#-aide"}, false, "display the help")
	//	flag.StringVar(&cruise, []string{"cruise"}, "", "cruise name")
	flag.StringVar(&configFile, []string{"c", "config"}, "", "use alternate .toml config file")
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
		tsg_file = config.Windows.Tsg_file
		ctd_file = config.Windows.Ctd_file
		kml_file = config.Windows.Kml_file
	} else {
		tsg_file = config.Unix.Tsg_file
		ctd_file = config.Unix.Ctd_file
		kml_file = config.Unix.Kml_file
	}

	pf("Cruise: %s\n", config.Cruise)
	pf("Ship: %s\n", config.Ship)
	pf("Ctd_plots: %s\n", config.Ctd_plots)
	pf("Tsg_plots: %s\n", config.Tsg_plots)
	pf("Ctd_file: %s\n", ctd_file)
	pf("Tsg_file: %s\n", tsg_file)
	pf("Kml_file: %s\n", kml_file)
}

// convert position "DD MM.SS S" to decimal position
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
		"This is Cassiopee cruise folder in July-Aug 2015")
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
		// add point to line
		ls.AddPoint(np)
	}

	// fill description markup with the TSG picture link inside <![CDATA[...]]>
	// All characters enclosed between these two sequences are interpreted as characters
	description := fmt.Sprintf("<![CDATA[\n<img src='%s' width='%d' />]]>",
		config.Tsg_plots, config.Size_plots)
	// define block Placemark for line
	placemark := fmt.Sprintf("%s cruise track on R/V %s", config.Cruise, config.Ship)
	pm := gokml.NewPlacemark(placemark, description, ls)
	pm.SetStyle("TrackStyle")
	// add placemark markup to kml file
	f.AddFeature(pm)

	// read CTD position
	fid_ctd, err := os.Open(ctd_file)
	if err != nil {
		log.Fatal(err)
	}
	defer fid_ctd.Close()

	scanner_ctd := bufio.NewScanner(fid_ctd)
	var i int = 1
	var filename string
	var type_cast string

	// TODOS:
	// check the first valid line with a regex
	// add the column label inside .toml file
	for scanner_ctd.Scan() {
		str := scanner_ctd.Text()
		values := strings.Fields(str)
		//p(values)

		// skip first line
		profile := values[0]
		if profile == config.Cruise {
			continue
		}
		// convert profile with the right format (usually %03d or %05d)
		prfl, _ := strconv.Atoi(profile)
		format := fmt.Sprintf("%%0%1dd", config.Ctd_prefix)
		profile = fmt.Sprintf(format, prfl)
		// extract data from station line
		begin_date := values[1]
		begin_hour := values[2]
		end_date := values[3]
		end_hour := values[4]
		lat := fmt.Sprintf("%s %s", values[5], values[6])
		lon := fmt.Sprintf("%s %s", values[7], values[8])
		pmax := values[9]
		bottom_depth := values[10]

		if len(values) > 11 {
			type_cast = values[11]
		} else {
			type_cast = ""
		}
		if len(values) > 12 {
			filename = values[12]
		} else {
			filename = ""
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
			profile, type_cast, filename, begin_date, begin_hour,
			end_date, end_hour, lat, lon, pmax, bottom_depth)

		// fill description markup with the CTD picture link inside <![CDATA[...]]>
		// All characters enclosed between these two sequences are interpreted as characters
		files := fmt.Sprintf(config.Ctd_plots, profile)
		description := fmt.Sprintf("%s<![CDATA[\n<img src='%s' width='%d' />]]>",
			header, files, config.Size_plots)

		// add new Placemark markup with station number, description and location (point object)
		//
		var newName string
		if config.Station_number {
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
	pf("%s", k.Render())

	// open ASCII file for writing result
	fid_kml, err := os.Create(kml_file)
	if err != nil {
		log.Fatal(err)
	}
	defer fid_kml.Close()

	// use buffered mode for writing
	fbuf_kml := bufio.NewWriter(fid_kml)
	// write kml to file
	fmt.Fprintln(fbuf_kml, k.Render())
	fbuf_kml.Flush()

	// display the filename to screen
	p(kml_file)
}
