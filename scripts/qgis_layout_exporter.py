import gc
import os
import sys

from typing import List
from datetime import datetime, timezone
from pathlib import Path

from qgis.core import (
    QgsApplication,
    QgsProject,
    QgsMapLayerStore,
    QgsRasterLayer,
    QgsCoordinateReferenceSystem,
    QgsVectorLayer,
    QgsLayoutManager,
    QgsLayout,
    QgsLayoutExporter,
    QgsCoordinateTransform,
    QgsLayoutItemLegend,
    QgsLegendModel,
    QgsLayerTree,
    QgsMapLayerLegendUtils,
    QgsLayerTreeLayer
)

from osgeo import gdal

# Suppress error
gdal.PushErrorHandler('CPLQuietErrorHandler')
QgsApplication.setPrefixPath(None, True)
project_instance: QgsProject = QgsProject.instance()

# # Instantiate application
# qgis_install_path = None
# qgis_app: QgsApplication = QgsApplication([], False)
# qgis_app.setPrefixPath(qgis_install_path, True)
# qgis_app.initQgis()
# print("qgis inited")


# Script must use QGIS python interpreter OR use an environment with qgis installed
#   (ex. conda install -c conda-forge qgis)
# Export path to QGIS sqlite library if needed, usually for PyCharm run
# export DYLD_INSERT_LIBRARIES="/Applications/QGIS.app/Contents/MacOS/lib/libsqlite3.dylib"
# windows PYTHONPATH updates
# https://gis.stackexchange.com/questions/40375/fixing-importerror-no-module-named-qgis-core

# Run from project root
def export_qgis_layout_png(hurricane_file_base: str, include_forecasts: bool = False) -> str:
    print("qgis start")
    # Paths
    # qgis_install_path: str = "/Applications/QGIS.app/Contents/MacOS"  # This is for pycharm run
    qgis_install_path = None

    project_path: Path = Path(f"{os.getcwd()}/qgis/godin.qgs")
    style_base_path: Path = Path(f"{os.getcwd()}/qgis")

    # Set up hurricane variables
    hurricane_raster: str = f"{hurricane_file_base}.png"
    hurricane_track: str = f"{hurricane_file_base}.csv"
    hurricane_base_split: List[str] = hurricane_file_base.split("_")
    hurricane_name: str = hurricane_base_split[0].title()
    hurricane_year: int = int(hurricane_base_split[1])
    hurricane_resolution: str = hurricane_base_split[2]

    data_path: str = f"{os.getcwd()}/data/{hurricane_name.lower()}{hurricane_year}"

    # QGIS variables names
    raster_layer_name: str = "raster"
    track_layer_name: str = "track"
    base_map_name: str = "base_map"
    layout_name: str = "godin_layout"
    legend_name: str = "legend"
    print("defined qgis vars")

    # Instantiate application
    QgsApplication.setPrefixPath(qgis_install_path, True)
    qgis_app: QgsApplication = QgsApplication([], False)
    qgis_app.initQgis()
    print("qgis inited")

    # Open project
    # project_instance: QgsProject = QgsProject.instance()
    project_instance.setFileName(str(project_path))
    project_instance.read()
    print("project opened")

    # Set up CRS objects for project
    qcrs: QgsCoordinateReferenceSystem = QgsCoordinateReferenceSystem()
    wgs84 = qcrs.fromEpsgId(4326)
    wgs84pseudo = qcrs.fromEpsgId(3857)
    print("CRS set")

    # Create raster and track layers
    new_raster_layer: QgsRasterLayer = QgsRasterLayer(f"{data_path}/{hurricane_raster}", raster_layer_name)
    new_raster_layer.setCrs(wgs84)
    new_raster_layer.loadNamedStyle(str(style_base_path / Path("hurricane_cat_raster.qml")))
    if not new_raster_layer.isValid():
        print(f"Raster Layer error: {new_raster_layer.error()}")
        exit(2)

    new_delim_layer: QgsVectorLayer = QgsVectorLayer(
        f"file:///{data_path / Path(hurricane_track)}?delimiter=,&xField=lonX&yField=latY",
        track_layer_name,
        "delimitedtext")
    new_delim_layer.setCrs(wgs84)
    new_delim_layer.loadNamedStyle(str(style_base_path / Path("hurricane_track.qml")))
    if not new_delim_layer.isValid():
        print(f"Delim Layer error: {new_delim_layer.error().messageList()}")
        exit(3)

    # Add layers to project
    project_instance.addMapLayer(new_raster_layer)
    project_instance.addMapLayer(new_delim_layer)

    print("Layers created")

    # Set up layout, set extents to a map layer that isn't the base layer
    ls2: QgsMapLayerStore = project_instance.layerStore()

    manager: QgsLayoutManager = project_instance.layoutManager()
    layout: QgsLayout = manager.layoutByName(layout_name)  # name of the layout

    for layer in ls2.children():
        if layer.name() != base_map_name:
            bbox = layer.extent()

            source_crs = wgs84
            dest_crs = wgs84pseudo
            transform = QgsCoordinateTransform(source_crs, dest_crs, project_instance)
            new_bbox = transform.transformBoundingBox(bbox)

            layout.referenceMap().setExtent(new_bbox)
            break

    print("layout set")

    # Edit legend items
    legend: QgsLayoutItemLegend = layout.itemById(legend_name)
    legend.updateLegend()
    legend.setAutoUpdateModel(False)
    legend_model: QgsLegendModel = legend.model()
    legend_group: QgsLayerTree = legend_model.rootGroup()

    for layer in ls2.children():
        if layer.name() == raster_layer_name:
            raster_node: QgsLayerTreeLayer = legend_group.findLayer(layer)
            raster_node.setName(f"{hurricane_name} {hurricane_year} Wind Field")
            QgsMapLayerLegendUtils.setLegendNodeOrder(raster_node, [2, 3, 4, 5, 6, 7])
            legend_model.refreshLayerLegend(raster_node)

        if layer.name() == track_layer_name and not include_forecasts:
            track_node: QgsLayerTreeLayer = legend_group.findLayer(layer)
            track_node.setName(f"{hurricane_name} {hurricane_year} Track")
            QgsMapLayerLegendUtils.setLegendNodeOrder(track_node, [0])
            legend_model.refreshLayerLegend(track_node)

        if layer.name() == base_map_name:
            legend_group.removeLayer(layer)
    print("legend added")

    # Export Layout to PNG
    exporter: QgsLayoutExporter = QgsLayoutExporter(layout)
    img_settings: exporter.ImageExportSettings = exporter.ImageExportSettings()
    img_settings.cropToContents = True
    ts: datetime = datetime.utcnow().replace(tzinfo=timezone.utc)
    ts = ts.replace(microsecond=0, second=0)

    export_filename: Path = Path(
        f"{data_path}/{hurricane_name.lower()}{hurricane_year}_{hurricane_resolution}_{ts.isoformat().replace(':', '')}.jpeg"
    )
    exporter.exportToImage(
        str(export_filename),
        img_settings
    )
    print("exported")

    # Save project to tracking file
    # project_path: str = f"{os.getcwd()}/qgis/godin_save.qgz"
    # project_instance.write(project_path)

    # Close out qgis app
    print("qgis done")
    # project_instance.clear()
    # qgis_app.exitQgis()
    qgis_app.exit()
    # del qgis_app
    # gc.collect()
    print("qgis closed")
    # print(export_filename)
    return str(export_filename)


def smol_run(hurricane_file_base: str, include_forecasts: bool = False):
    print("qgis start")
    # Paths
    # qgis_install_path: str = "/Applications/QGIS.app/Contents/MacOS"  # This is for pycharm run
    qgis_install_path = None

    project_path: Path = Path(f"{os.getcwd()}/qgis/godin.qgs")
    style_base_path: Path = Path(f"{os.getcwd()}/qgis")

    # Set up hurricane variables
    hurricane_raster: str = f"{hurricane_file_base}.png"
    hurricane_track: str = f"{hurricane_file_base}.csv"
    hurricane_base_split: List[str] = hurricane_file_base.split("_")
    hurricane_name: str = hurricane_base_split[0].title()
    hurricane_year: int = int(hurricane_base_split[1])
    hurricane_resolution: str = hurricane_base_split[2]

    data_path: str = f"{os.getcwd()}/data/{hurricane_name.lower()}{hurricane_year}"

    # QGIS variables names
    raster_layer_name: str = "raster"
    track_layer_name: str = "track"
    base_map_name: str = "base_map"
    layout_name: str = "godin_layout"
    legend_name: str = "legend"
    print("defined qgis vars")

    # supply path to qgis install location
    # QgsApplication.setPrefixPath(None, True)

    # create a reference to the QgsApplication, setting the
    # second argument to False disables the GUI
    qgs = QgsApplication([], False)

    # load providers
    qgs.initQgis()

    print("qgis inited")

    # Open project
    project_instance.setFileName(str(project_path))
    project_instance.read()
    print("project opened")

    print("I ran smol qgis")

    # qgs.exitQgis()
    qgs.exit()
    return "smol"


if __name__ == "__main__":
    hurricane_base: str = sys.argv[1]
    include_forecasts: bool = bool(sys.argv[2])
    # hurricane_base: str = "LARRY_2021_100x100"
    export_qgis_layout_png(hurricane_base, include_forecasts)
