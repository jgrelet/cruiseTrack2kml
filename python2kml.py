# https://www.sigterritoires.fr/index.php/kml2-comment-creer-des-bulles-ballons-personnalisees/
# https://simplekml.readthedocs.io/en/latest/geometries.html
from netCDF4 import Dataset
import simplekml
from configparser import ConfigParser
import argparse
import toml
import logging

elevation = 0

def processArgs():
    parser = argparse.ArgumentParser(
        description="This program reads several ASCII or NetCDF files and creates a KML file \
            for Google Earth to display the ship's route and station positions. \
            Each element is clickable to display surface temperature/salinity data, \
            vertical profiles of CTD, LADCP and XBT stations.",

        usage='\npython python2kml.py\n'
        'python python2kml.py -c config.toml\n'
        'python python2kml.py -d\n'
        'python python2kml.py -h\n'
        '\n',
        epilog='J. Grelet IRD US191 - Sep 2021 / Nov 2021')
    parser.add_argument('-d', '--debug', help='display debug informations',
                        action='store_true')
    #parser.add_argument('-h', '--help', help='display help informations',
    #                    action='store_true')
    parser.add_argument('-c', '--config', help="toml configuration file, (default: %(default)s)",
                        default='config.toml')
    return parser

if __name__ == "__main__":
    '''
    usage:
    > python python2kml.py 
    > python python2kml.py -c config.toml
    '''
    # recover and process line arguments
    parser = processArgs()
    args = parser.parse_args()

    # set looging mode if debug
    if args.debug:
        logging.basicConfig(
            format='%(levelname)s:%(message)s', level=logging.DEBUG)

    # read config Toml file and get the physical parameter list (Roscop code) for the specified instrument
    cfg = toml.load(args.config)
    print(cfg)

    ctdfile = 'data/amazomix/OS_AMAZOMIX_CTD.nc'
    xbtfile = 'data/amazomix/OS_AMAZOMIX_XBT.nc'
    tsgfile = 'data/amazomix/OS_AMAZOMIX_TSG.nc'
    kml_file = "amazomix.kml"

    ctd = Dataset(ctdfile, mode='r')
    xbt = Dataset(xbtfile, mode='r')
    tsg = Dataset(tsgfile, mode='r')

    # CTD
    profiles = ctd.variables['PROFILE'][:].tolist()
    ctd_url = "http://www.brest.ird.fr/us191/cruises/amazomix/CTD/AMAZOMIX-{:05d}_CTD.png"

    kml = simplekml.Kml()
    style = simplekml.Style()
    style.iconstyle.color = simplekml.Color.red  # Make the icon red
    style.iconstyle.scale = 1
    style.iconstyle.icon.href = 'http://maps.google.com/mapfiles/kml/pushpin/wht-pushpin.png'
    #style.iconstyle.icon.href = 'http://maps.google.com/mapfiles/kml/shapes/placemark_circle.png'

    # plot CTD station icons red
    for i in range(0, len(profiles)):    
        url = ctd_url.format(profiles[i]) 
        cdata = '<![CDATA[\n<img src={} width={:d} />]]>'.format(url, 700)     
        point = kml.newpoint()
        point.name="ctd {:05d}".format(profiles[i])
        point.description = "CTD Station: {:05d}\n{}".format(profiles[i], cdata)
        point.coords=[(ctd.variables['LONGITUDE'][i], ctd.variables['LATITUDE'][i], elevation)]
        point.altitudemode = simplekml.AltitudeMode.relativetoground 
        point.style = style
    print("CTD: {} stations".format(len(profiles)))

    # XBT
    profiles = xbt.variables['PROFILE'][:].tolist()
    xbt_url = "http://www.brest.ird.fr/us191/cruises/amazomix/XBT/AMAZOMIX-{:05d}_XBT.png"

    # plot XBT profiles icons green
    style = simplekml.Style()
    style.iconstyle.color = simplekml.Color.azure  # Make the icon green

    for i in range(0, len(profiles)):    
        url = xbt_url.format(profiles[i]) 
        cdata = '<![CDATA[\n<img src={} width={:d} />]]>'.format(url, 700)     
        point = kml.newpoint()
        point.name="xbt {:03d}".format(profiles[i])
        point.description = "XBT Profile: {:05d}\n{}".format(profiles[i], cdata)
        point.coords=[(xbt.variables['LONGITUDE'][i], xbt.variables['LATITUDE'][i], elevation)]
        point.altitudemode = simplekml.AltitudeMode.relativetoground 
        point.style = style
    print("XBT: {} profiles".format(len(profiles)))

    # TSG
    data = tsg.variables['TIME'][:].tolist()
    tsg_url = "http://www.brest.ird.fr/us191/cruises/amazomix/TSG/AMAZOMIX_TSG_COLCOR_SCATTER.png"

    # plot TSG data as lineString in blue
    style = simplekml.Style()
    style.linestyle.color = simplekml.Color.blue  # Make the line blue
    style.linestyle.width = 3
    cdata = '<![CDATA[\n<img src={} width={:d} />]]>'.format(tsg_url, 500)     
    ls = kml.newlinestring()
    ls.name="TSG - SSTP/SSPS"
    ls.description = cdata
    #ls.altitudemode = simplekml.AltitudeMode.relativetoground 
    ls.style = style

    for i in range(0, len(data)):    
        ls.coords.addcoordinates([(tsg.variables['LONGITUDE'][i], 
            tsg.variables['LATITUDE'][i])])
    print("TSG: {} data".format(len(data)))

    kml.save(kml_file)
    print("File {} saved".format(kml_file))
    #print(kml.kml())
    ctd.close()
    xbt.close()
    tsg.close()