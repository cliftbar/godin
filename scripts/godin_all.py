import os
import subprocess

from subprocess import CompletedProcess
from time import sleep
from typing import Any, List
from pathlib import Path

import qgis_layout_exporter
from event_uploader import upload_event


def run_model(storm_id: str, resolution: int, include_forecasts: bool = False) -> str:
    print(["./godin", "-res", str(resolution), "-include_forecasts", str(include_forecasts), storm_id])
    model_proc: CompletedProcess[Any] = subprocess.run(["./godin", "-res", str(resolution), storm_id], stdout=subprocess.PIPE)
    model_proc.check_returncode()
    model_proc_out: str = str(model_proc.stdout, "utf-8")
    # print(model_proc_out)
    output_lines: List[str] = model_proc_out.splitlines()
    storm_name: str = [l for l in output_lines if l.startswith("Name:")][0].split(":")[1].strip()

    # print(f"model run for {storm_name} finished")
    return storm_name


def create_update_ssg(storm_name: str, storm_year: int, res: int, file_ts: str, draft: bool = True):
    post_path: Path = Path(f"ssg/content/hurricane/{storm_name.lower()}{storm_year}.md")

    if not post_path.exists():
        current_env = os.environ
        current_env["HUGO_HURRICANE_RES"] = str(res)
        current_env["HUGO_HURRICANE_TS"] = file_ts
        hugo_proc: CompletedProcess[Any] = subprocess.run(
            ["hugo", "new", f"hurricane/{storm_name.lower()}{storm_year}.md"],
            cwd=f"{os.getcwd()}/ssg",
            stdout=subprocess.PIPE
        )
        hugo_proc.check_returncode()
        # print("Hugo post created")
    else:
        lines_out: List[str] = []
        with post_path.open("r") as p:
            post_text: List[str] = p.readlines()
            for line in post_text:
                if line.startswith("resolution:"):
                    line = f"resolution: {res}\n"
                elif line.startswith("hurricane_timestamp:"):
                    line = f"hurricane_timestamp: {file_ts}\n"
                if not draft and line.startswith("draft:"):
                    line = f"draft: {str(draft).lower}\n"

                lines_out.append(line)

        with post_path.open("w") as p:
            p.writelines(lines_out)
        # print("Hugo post updated")


def godin_storm(storm_id: str, resolution: int = 100, include_forecasts: bool = False, ssg_draft: bool = True) -> str:
    year: int = int(storm_id[-4:])

#     name: str = run_model(storm_id, resolution, include_forecasts)
    sleep(1)

    hurricane_base: str = f"{name.upper()}_{year}_{resolution}x{resolution}"
    hurricane_raster: str = qgis_layout_exporter.export_qgis_layout_png(hurricane_base)

    hurricane_raster_path: Path = Path(hurricane_raster)
    hurricane_raster_ts: str = hurricane_raster_path.stem.split("_")[-1]

    upload_event(hurricane_raster_path.name)

    create_update_ssg(name, year, resolution, hurricane_raster_ts, ssg_draft)

    # print("godin storm finished\n")
    return name


def godin_year():
    year: int = 2020
    storm_count: int = 31

    resolution: int = 100
    for i in range(1, storm_count+1):
        storm: str = f"al{i:02d}{year}"

        godin_storm(storm, resolution, False)

        print(f"{storm} finished, {i} out of {storm_count} for year {year}")
        sleep(1)


if __name__ == "__main__":
#     storm: str = "al022020"
    storm: str = "al122021"
    godin_storm(storm, 100, True)
#     godin_year()
