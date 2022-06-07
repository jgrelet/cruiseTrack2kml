# cruiseTrack2kml [![Build Status](https://travis-ci.com/jgrelet/cruiseTrack2kml.svg?branch=master)](https://app.travis-ci.com/github/jgrelet/cruiseTrack2kml)

The cruiseTrack2kml program is used for rendering oceanographic data, Seabird CTD, XBT profiles and TSG (thermosalinograph) plots to Google Earth Keyhole Markup Language (KML) files

see example kml file for PIRATA cruise FR26:

[PIRATA-26.kml](http://www.brest.ird.fr/pirata/images/cruise_tracks/pirata-fr26.kml)

The makefile works only with Linux and Windows git bash

## update go packages

'''bash
go get -u ./...
go mod tidy
'''
