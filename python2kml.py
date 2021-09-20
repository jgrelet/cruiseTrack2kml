# https://www.sigterritoires.fr/index.php/kml2-comment-creer-des-bulles-ballons-personnalisees/
from netCDF4 import Dataset
import simplekml
file = 'data/amazomix/OS_AMAZOMIX_CTD.nc'
kml = simplekml.Kml()
nc = Dataset(file, mode='r')
profiles = nc.variables['PROFILE'][:].tolist()

for i in range(1, len(profiles)):         
    pnt = kml.newpoint(name="St{:05d}".format(profiles[i]), coords=[(nc.variables['LONGITUDE'][i],
        nc.variables['LATITUDE'][i])])  # lon, lat, optional height
    pnt.description = "Station: {:05d}".format(profiles[i])
kml.save("amazomix.kml")
print(kml.kml())
nc.close()