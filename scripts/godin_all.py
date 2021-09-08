import os
import subprocess

from subprocess import CompletedProcess
from typing import Any, List
from pathlib import Path

from qgis_layout_exporter import export_qgis_layout_png
from event_uploader import upload_event


def run_model(storm_id: str, resolution: int):
    model_proc: CompletedProcess[Any] = subprocess.run(["./godin", "-res", str(resolution), storm_id])
    model_proc.check_returncode()


def create_update_ssg(storm_name: str, storm_year: int, res: int, file_ts: str):
    post_path: Path = Path(f"ssg/content/hurricane/{storm_name.lower()}{storm_year}.md")
    print(post_path)

    if not post_path.exists():
        current_env = os.environ
        current_env["HUGO_HURRICANE_RES"] = str(res)
        current_env["HUGO_HURRICANE_TS"] = file_ts
        hugo_proc: CompletedProcess[Any] = subprocess.run(
            ["hugo", "new", f"hurricane/{storm_name.lower()}{storm_year}.md"],
            cwd=f"{os.getcwd()}/ssg"
        )
        print("Hugo post created")
    else:
        lines_out: List[str] = []
        with post_path.open("r") as p:
            post_text: List[str] = p.readlines()
            for line in post_text:
                if line.startswith("resolution: "):
                    line = f"resolution: {res}\n"
                elif line.startswith("hurricane_timestamp: "):
                    line = f"hurricane_timestamp: {file_ts}\n"
                lines_out.append(line)

        with post_path.open("w") as p:
            p.writelines(lines_out)
        print("Hugo post updated")


if __name__ == "__main__":
    storm: str = "al041992"
    resolution: int = 100
    name: str = "ANDREW"
    year: int = 1992

    run_model(storm, resolution)

    hurricane_base: str = f"{name.upper()}_{year}_{resolution}x{resolution}"
    hurricane_raster: str = export_qgis_layout_png(hurricane_base)

    hurricane_raster_path: Path = Path(hurricane_raster)
    hurricane_raster_ts: str = hurricane_raster_path.stem.split("_")[-1]

    upload_event(hurricane_raster_path.name)

    create_update_ssg(name, year, resolution, hurricane_raster_ts)

    print("finished")
