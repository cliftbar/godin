import os

from typing import List
from datetime import datetime, timezone

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


# Script must use QGIS python interpreter OR use an environment with qgis installed
#   (ex. conda install -c conda-forge qgis)
# Export path to QGIS sqlite library if needed
# export DYLD_INSERT_LIBRARIES="/Applications/QGIS.app/Contents/MacOS/lib/libsqlite3.dylib"

# Run from project root
def export_qgis_layout_png() -> None:
    # Paths
    qgis_install_path: str = "/Applications/QGIS.app/Contents/MacOS"  # This needs to be edited for other installations
    # project_path: str = "/Users/cameronbarclift/MyFiles/qgis/projects/godin.qgz"
    project_path: str = f"{os.getcwd()}/qgis/godin.qgs"
    data_path: str = f"{os.getcwd()}/data/tmp"
    style_base_path: str = f"{os.getcwd()}/qgis"

    # Set up hurricane variables
    hurricane_file_base: str = "IDA_2021_100x100"
    hurricane_raster: str = f"{hurricane_file_base}.png"
    hurricane_track: str = f"{hurricane_file_base}.csv"
    hurricane_base_split: List[str] = hurricane_file_base.split("_")
    hurricane_name: str = hurricane_base_split[0].title()
    hurricane_year: int = int(hurricane_base_split[1])
    hurricane_resolution: str = hurricane_base_split[2]

    # QGIS variables names
    raster_layer_name: str = "raster"
    track_layer_name: str = "track"
    base_map_name: str = "base_map"
    layout_name: str = "godin_layout"
    legend_name: str = "legend"

    # Instantiate application
    QgsApplication.setPrefixPath(qgis_install_path, True)
    qgis_app: QgsApplication = QgsApplication([], False)
    qgis_app.setPrefixPath(qgis_install_path, True)
    qgis_app.initQgis()

    # Open project
    project_instance: QgsProject = QgsProject.instance()
    project_instance.setFileName(project_path)
    project_instance.read()

    # Set up CRS objects for project
    qcrs: QgsCoordinateReferenceSystem = QgsCoordinateReferenceSystem()
    wgs84 = qcrs.fromEpsgId(4326)
    wgs84pseudo = qcrs.fromEpsgId(3857)

    # Create raster and track layers
    new_raster_layer: QgsRasterLayer = QgsRasterLayer(f"{data_path}/{hurricane_raster}", raster_layer_name)
    new_raster_layer.setCrs(wgs84)
    new_raster_layer.loadNamedStyle(f"{style_base_path}/hurricane_cat_raster.qml")
    print(new_raster_layer.isValid())
    if not new_raster_layer.isValid():
        exit(2)

    new_delim_layer: QgsVectorLayer = QgsVectorLayer(
        f"file://{data_path}/{hurricane_track}?delimiter=,&xField=lonX&yField=latY",
        track_layer_name,
        "delimitedtext")
    new_delim_layer.setCrs(wgs84)
    new_delim_layer.loadNamedStyle(f"{style_base_path}/hurricane_track.qml")
    print(new_delim_layer.isValid())
    if not new_delim_layer.isValid():
        exit(2)

    # Add layers to project
    project_instance.addMapLayer(new_raster_layer)
    project_instance.addMapLayer(new_delim_layer)

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

    # Edit legend items
    legend: QgsLayoutItemLegend = layout.itemById(legend_name)
    legend.updateLegend()
    legend.setAutoUpdateModel(False)
    legend_model: QgsLegendModel = legend.model()
    legend_group: QgsLayerTree = legend_model.rootGroup()

    include_forecasts: bool = False
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

    # Export Layout to PNG
    exporter: QgsLayoutExporter = QgsLayoutExporter(layout)
    img_settings: exporter.ImageExportSettings = exporter.ImageExportSettings()
    img_settings.cropToContents = True
    ts: datetime = datetime.utcnow().replace(tzinfo=timezone.utc)
    ts = ts.replace(microsecond=0, second=0)

    export_filename: str = \
        f"{data_path}/{hurricane_name.lower()}{hurricane_year}_{hurricane_resolution}_{ts.isoformat()}.png"
    exporter.exportToImage(
        export_filename,
        img_settings
    )

    # Save project to tracking file
    # project_path: str = f"{os.getcwd()}/qgis/godin_save.qgz"
    # project_instance.write(project_path)

    # Close out qgis app
    qgis_app.exitQgis()
    print("finished")
    print(export_filename)


if __name__ == "__main__":
    export_qgis_layout_png()
