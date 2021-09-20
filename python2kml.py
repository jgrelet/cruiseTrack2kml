# https://www.sigterritoires.fr/index.php/kml2-comment-creer-des-bulles-ballons-personnalisees/
from netCDF4 import Dataset
import simplekml

file = 'data/amazomix/OS_AMAZOMIX_CTD.nc'

nc = Dataset(file, mode='r')
profiles = nc.variables['PROFILE'][:].tolist()

kml = simplekml.Kml()

for i in range(1, len(profiles)):         
    point = kml.newpoint()
    point.name="St{:05d}".format(profiles[i])
    point.description = "Station: {:05d}".format(profiles[i])
    point.coords=[(nc.variables['LONGITUDE'][i], nc.variables['LATITUDE'][i])]
    
kml.save("amazomix.kml")
print(kml.kml())
nc.close()