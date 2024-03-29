import datetime
import json
import os
import subprocess
import time

from subprocess import CompletedProcess
from typing import Any, List, Dict
from pathlib import Path

from google.cloud import firestore
from google.cloud.firestore_v1 import Client, DocumentSnapshot
from sqlalchemy import Engine, create_engine, Connection, text

import qgis_layout_exporter
from event_uploader import upload_event


def run_model(storm_id: str, resolution: int, include_forecasts: bool = False) -> str:
    godin_binary: Path = Path(".", "bin", "godin")
    start_time: float = time.time()
    model_proc: CompletedProcess[Any] = subprocess.run(["./"+str(godin_binary), "-res", str(resolution), storm_id],
                                                       stdout=subprocess.PIPE, stderr=subprocess.STDOUT)
    model_proc.check_returncode()
    end_time: float = time.time()
    model_proc_out: str = str(model_proc.stdout, "utf-8")
    # print(model_proc_out)
    print(f"model {storm_id} took {end_time - start_time}s")
    output_lines: List[str] = model_proc_out.splitlines()
    storm_name: str = [li for li in output_lines if li.startswith("Name:")][0].split(":")[1].strip()

    print(f"model run for {storm_name} finished")
    return storm_name


def create_update_ssg(storm_id: str, storm_name: str, storm_year: int, res: int, file_ts: str, draft: bool = True, ssg_data: Dict = None):
    storm_ssg_name: str = f"{storm_id}_{storm_name.lower()}{storm_year}.md"
    post_path: Path = Path(f"ssg/content/hurricane/{storm_year}/{storm_ssg_name}")

    if not post_path.exists():
        current_env = os.environ
        current_env["HUGO_HURRICANE_RES"] = str(res)
        current_env["HUGO_HURRICANE_TS"] = file_ts
        current_env["HUGO_HURRICANE_ADV_NUM"] = str(ssg_data["AdvNumber"])
        current_env["HUGO_HURRICANE_DISCUSSION"] = ssg_data["Discussion"]
        current_env["HUGO_HURRICANE_SOURCES"] = ";".join(ssg_data["Sources"])
        current_env["HUGO_HURRICANE_STORM_ID"] = storm_id
        current_env["HUGO_HURRICANE_STORM_NAME"] = storm_name.lower()
        current_env["HUGO_HURRICANE_STORM_YEAR"] = str(storm_year)

        hugo_proc: CompletedProcess[Any] = subprocess.run(
            ["hugo", "new", f"hurricane/{storm_year}/{storm_ssg_name}"],
            cwd=f"{os.getcwd()}/ssg",
            stdout=subprocess.PIPE
        )
        hugo_proc.check_returncode()
        print(f"Hugo post created for {storm_id} {ssg_data['AdvNumber']}")

        if not draft:
            lines_out: List[str] = []
            with post_path.open("r") as p:
                post_text: List[str] = p.readlines()
                for line in post_text:
                    if not draft and line.startswith("draft:"):
                        line = f"draft: {str(draft).lower()}\n"

                    lines_out.append(line)

            with post_path.open("w") as p:
                p.writelines(lines_out)
    else:
        lines_out: List[str] = []
        with post_path.open("r") as p:
            post_text: List[str] = p.readlines()
            lines_subbed: List[str] = []
            for line in post_text:
                if line.startswith("resolution:"):
                    line = f"resolution: {res}\n"
                elif line.startswith("hurricane_timestamp:"):
                    line = f"hurricane_timestamp: {file_ts}\n"
                elif line.startswith("adv_number:"):
                    line = f"adv_number: {ssg_data['AdvNumber']}\n"
                elif line.startswith("adv_sources:"):
                    line = f"adv_sources: {';'.join(ssg_data['Sources'])}\n"
                elif line.startswith("last_updated:"):
                    line = f"last_updated: {datetime.datetime.now(tz=datetime.timezone.utc).replace(microsecond=0).isoformat()}\n"
                elif line.startswith("draft:") and not draft:
                    line = f"draft: {str(draft).lower()}\n"
                elif line.startswith("title:"):
                    line = f"title: {storm_name.title()} {storm_year}\n"
                elif line.startswith("storm_name:"):
                    line = f"storm_name: {storm_name.lower()}\n"

                lines_subbed.append(line)

            for line in lines_subbed:
                if str.strip(line) == "## Official Advisory Discussion":
                    lines_out.append(line)
                    lines_out.append(ssg_data["Discussion"])
                    break
                else:
                    lines_out.append(line)

        with post_path.open("w") as p:
            p.writelines(lines_out)


def godin_storm(storm_id: str, resolution: int = 100, include_forecasts: bool = False, ssg_draft: bool = True, ssg_data: Dict = None) -> str:
    print(f"running {storm_id}")
    year: int = int(storm_id[-4:])

    name: str = run_model(storm_id, resolution, include_forecasts)

    raster_start: float = time.time()
    hurricane_base: str = f"{name.upper()}_{year}_{resolution}x{resolution}"
    hurricane_raster: str = qgis_layout_exporter.export_qgis_layout_png(hurricane_base, include_forecasts)

    hurricane_raster_path: Path = Path(hurricane_raster)
    hurricane_raster_ts: str = hurricane_raster_path.stem.split("_")[-1]
    print(f"Raster completed: {time.time() - raster_start}s")

    upload_start: float = time.time()
    upload_event(hurricane_raster_path.name)
    print(f"Upload completed: {time.time() - upload_start}s")

    ssg_start: float = time.time()
    create_update_ssg(storm_id, name, year, resolution, hurricane_raster_ts, ssg_draft, ssg_data)
    print(f"SSG completed: {time.time() - ssg_start}s")

    print(f"godin storm finished for {storm_id} {ssg_data['AdvNumber']}\n")
    return name


def godin_year(year: int = 2021, storm_count: int = 21, resolution: int = 100):

    for i in range(1, storm_count + 1):
        storm: str = f"al{i:02d}{year}"

        godin_storm(storm, resolution, False)

        print(f"{storm} finished, {i} out of {storm_count} for year {year}")


def cloud_run():
    cloud_run_start: float = time.time()

    git_setup_start: float = time.time()
    git_setup()
    print(f"git_setup completed: {time.time() - git_setup_start}s")

    db: Client = firestore.Client(project="godin-324403")

    pending_storms_start: float = time.time()
    pending_storms: List[DocumentSnapshot] = [d for d in db.collection("pending").stream()]
    print(f"pending_storms completed: {time.time() - pending_storms_start}s")

    run_dict: Dict = {}
    for storm in pending_storms:
        storm_dict: Dict = storm.to_dict()
        storm_id: str = storm_dict["StormID"]
        if storm_id in run_dict:
            if storm_dict["AdvNumber"] > run_dict[storm_id]["AdvNumber"]:
                run_dict[storm_id] = storm_dict
        else:
            run_dict[storm_id] = storm_dict

    for storm_id, storm in run_dict.items():
        godin_storm_start: float = time.time()
        godin_storm(storm["StormID"], 100, include_forecasts=True, ssg_draft=False, ssg_data=storm)
        print(f"godin_storm for {storm_id} completed: {time.time() - godin_storm_start}s")

        # pending_delete_start: float = time.time()
        db.collection("pending").document(storm_id).delete()
        # print(f"pending_delete for {storm_id} completed: {time.time() - pending_delete_start}s")

        # git_push_start: float = time.time()
        # git_push([s.to_dict()["Name"] for s in pending_storms])
        git_push([storm_id])
        # print(f"git_push_start completed: {time.time() - git_push_start}s")

    for storm in pending_storms:
        pending_delete_start: float = time.time()
        db.collection("pending").document(storm.id).delete()
        print(f"pending_delete completed: {time.time() - pending_delete_start}s")

    # git_push_start: float = time.time()
    # git_push([s.to_dict()["Name"] for s in pending_storms])
    # print(f"git_push_start completed: {time.time() - git_push_start}s")

    print(f"cloud_run completed: {time.time() - cloud_run_start}s")


def git_setup():
    proc = subprocess.run(["git", "config", "--global", "init.defaultBranch", "main"], stdout=subprocess.PIPE, stderr=subprocess.STDOUT)
    print(str(proc.stdout, "utf-8"))
    proc = subprocess.run(["git", "config", "--global", "user.email", "cwbarclift@gmail.com"], stdout=subprocess.PIPE, stderr=subprocess.STDOUT)
    print(str(proc.stdout, "utf-8"))
    proc = subprocess.run(["git", "config", "--global", "user.name", "cliftbar"], stdout=subprocess.PIPE, stderr=subprocess.STDOUT)
    print(str(proc.stdout, "utf-8"))
    proc = subprocess.run(["git", "init"], stdout=subprocess.PIPE, stderr=subprocess.STDOUT)
    print(str(proc.stdout, "utf-8"))
    # proc = subprocess.run(["git", "checkout", "origin/main"], stdout=subprocess.PIPE)
    # print(str(proc.stdout, "utf-8"))
    proc = subprocess.run(["git", "checkout", "-b", "main"], stdout=subprocess.PIPE, stderr=subprocess.STDOUT)
    print(str(proc.stdout, "utf-8"))
    proc = subprocess.run(["git", "status"], stdout=subprocess.PIPE, stderr=subprocess.STDOUT)
    print(str(proc.stdout, "utf-8"))
    proc = subprocess.run(["git", "remote", "add", "origin", f"https://cliftbar:{os.getenv('GHT')}@github.com/cliftbar/godin.git"], stdout=subprocess.PIPE, stderr=subprocess.STDOUT)
    print(str(proc.stdout, "utf-8"))

    proc = subprocess.run(["git", "fetch", "--depth", "1", "origin", "main"], stdout=subprocess.PIPE, stderr=subprocess.STDOUT)
    print(str(proc.stdout, "utf-8"))
    proc = subprocess.run(["git", "reset", "--hard", "origin/main"], stdout=subprocess.PIPE, stderr=subprocess.STDOUT)
    print(str(proc.stdout, "utf-8"))
    proc = subprocess.run(["git", "pull", "origin", "main"], stdout=subprocess.PIPE, stderr=subprocess.STDOUT)
    print(str(proc.stdout, "utf-8"))
    proc = subprocess.run(["git", "submodule", "add", "--depth", "1", "https://github.com/theNewDynamic/gohugo-theme-ananke.git themes/ananke"], cwd="ssg/themes", stdout=subprocess.PIPE, stderr=subprocess.STDOUT)
    print(str(proc.stdout, "utf-8"))


def git_push(storms: List[str]):
    print("git push")
    # proc = subprocess.run(["git", "status"], stdout=subprocess.PIPE, stderr=subprocess.STDOUT)
    # print(str(proc.stdout, "utf-8"))
    proc = subprocess.run(["git", "add", "ssg/content/*"], stdout=subprocess.PIPE, stderr=subprocess.STDOUT)
    print(str(proc.stdout, "utf-8"))
    # proc = subprocess.run(["git", "status"], stdout=subprocess.PIPE, stderr=subprocess.PIPE)
    # print(str(proc.stdout, "utf-8"))
    proc = subprocess.run(["git", "commit", "-m", f'"auto build of {", ".join(storms)}"'], stdout=subprocess.PIPE, stderr=subprocess.STDOUT)
    print(str(proc.stdout, "utf-8"))
    proc = subprocess.run(["git", "push", "--set-upstream", "origin", "main"], stdout=subprocess.PIPE, stderr=subprocess.STDOUT)
    print(str(proc.stdout, "utf-8"))
    proc = subprocess.run(["git", "status"], stdout=subprocess.PIPE, stderr=subprocess.STDOUT)
    print(str(proc.stdout, "utf-8"))


def local_run_year(year: int = 2023, doGit: bool = False):
    cloud_run_start: float = time.time()

    if doGit:
        git_setup_start: float = time.time()
        git_setup()
        print(f"git_setup completed: {time.time() - git_setup_start}s")

    godin_year(year, 4)

    if doGit:
        git_push_start: float = time.time()
        git_push([f"Year {year}"])
        print(f"git_push_start completed: {time.time() - git_push_start}s")

    print(f"local run completed: {time.time() - cloud_run_start}s, doGit: {doGit}")


def main():
    # storm: str = "al022020"
    # storm: str = "al122021"
    # storm: str = "al172021":
    # storm: str = "al182021"
    # godin_storm(storm, 10, True)
    # godin_year()
    # cloud_run()
    local_run_year(doGit=True)


if __name__ == "__main__":
    print("start!")
    main()
