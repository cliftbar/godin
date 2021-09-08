from typing import Optional
from qgis.core import QgsApplication

# Export path to QGIS sqlite library if needed
# export DYLD_INSERT_LIBRARIES="/Applications/QGIS.app/Contents/MacOS/lib/libsqlite3.dylib"
# qgis_install_path: str = "/Applications/QGIS.app/Contents/MacOS"  # This needs to be for pycharm installs

qgis_install_path: Optional[str] = None

# Set up application
qgis_app: QgsApplication = QgsApplication([], False)
qgis_app.setPrefixPath(qgis_install_path, True)

qgis_app.initQgis()

qgis_app.exitQgis()

# I get a segfault if I don't do this in the minimal script *shrug*
qgis_app = None
print("finished")
