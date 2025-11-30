import datetime
import json
import os
import subprocess
import time

from subprocess import CompletedProcess
from typing import Any, List, Dict
from pathlib import Path

from sqlalchemy import Engine, create_engine, Connection, text, CursorResult
from datetime import timezone, datetime

# import qgis_layout_exporter
import cartopy_layout
# from event_uploader import upload_event
from event_uploader_aws import upload_event as upload_event_aws
from jinja2 import Environment


## Git
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
    print("git push begins")
    proc = subprocess.run(["git", "status"], stdout=subprocess.PIPE, stderr=subprocess.STDOUT)
    print(str(proc.stdout, "utf-8"))
    proc = subprocess.run(["git", "add", "ssg/content/*"], stdout=subprocess.PIPE, stderr=subprocess.STDOUT)
    print(str(proc.stdout, "utf-8"))
    # proc = subprocess.run(["git", "status"], stdout=subprocess.PIPE, stderr=subprocess.PIPE)
    # print(str(proc.stdout, "utf-8"))
    # proc = subprocess.run(["git", "commit", "-m", f'"auto build of {", ".join(storms)}"'], stdout=subprocess.PIPE, stderr=subprocess.STDOUT)
    proc = subprocess.run(["git", "commit", "-m", f'"auto build of {", ".join(storms)}"'], stdout=subprocess.PIPE, stderr=subprocess.STDOUT)
    print(str(proc.stdout, "utf-8"))
    proc = subprocess.run(["git", "push", "--set-upstream", "origin", "main"], stdout=subprocess.PIPE, stderr=subprocess.STDOUT)
    print(str(proc.stdout, "utf-8"))
    proc = subprocess.run(["git", "status"], stdout=subprocess.PIPE, stderr=subprocess.STDOUT)
    print(str(proc.stdout, "utf-8"))


# Hugo
def create_update_ssg(storm_id: str, storm_name: str, storm_year: int, res: int, file_ts: str, draft: bool = True, ssg_data: Dict = None):
    storm_ssg_name: str = f"{storm_id}_{storm_name.lower()}{storm_year}.md"
    post_path: Path = Path(f"ssg/content/hurricane/{storm_year}/{storm_ssg_name}")
    ssg_data = {} if ssg_data is None else ssg_data

    if not post_path.exists():
        current_env = os.environ
        current_env["HUGO_HURRICANE_RES"] = str(res)
        current_env["HUGO_HURRICANE_TS"] = file_ts
        current_env["HUGO_HURRICANE_ADV_NUM"] = str(ssg_data.get("adv_number", "-1"))
        current_env["HUGO_HURRICANE_DISCUSSION"] = ssg_data.get("discussion", "No Discussion")
        current_env["HUGO_HURRICANE_SOURCES"] = ";".join(ssg_data.get("sources", ["None"]))
        current_env["HUGO_HURRICANE_STORM_ID"] = storm_id
        current_env["HUGO_HURRICANE_STORM_NAME"] = storm_name.lower()
        current_env["HUGO_HURRICANE_STORM_YEAR"] = str(storm_year)

        hugo_proc: CompletedProcess[Any] = subprocess.run(
            ["hugo", "new", "-v", f"hurricane/{storm_year}/{storm_ssg_name}"],
            cwd=f"{os.getcwd()}/ssg",
            stdout=subprocess.PIPE
        )
        print(f"hugo stdout: {hugo_proc.stdout}")
        print(f"hugo stderr: {hugo_proc.stderr}")
        hugo_proc.check_returncode()
        # print("Hugo post created")

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
                    line = f"adv_number: {ssg_data.get('adv_number', '-1')}\n"
                elif line.startswith("adv_sources:"):
                    line = f"adv_sources: {';'.join(ssg_data.get('sources', ['None']))}\n"
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
                    lines_out.append(ssg_data.get("discussion", "No Discussion"))
                    break
                else:
                    lines_out.append(line)

        with post_path.open("w") as p:
            p.writelines(lines_out)

def create_ssg_from_jinja(storm_name=None, storm_year=None, is_draft=None, storm_id=None, resolution=None,
                          hurricane_raster_ts=None, adv_num=None, storm_preview_address=None, storm_zip_address=None,
                          sources: list[str] = None, storm_discussion=None):
    """
    Render the generic hurricane markdown from the Jinja template, filling in all
    required variables from environment defaults. This function only returns the
    rendered markdown and does not write any files or trigger side-effects.
    """

    template_path: Path = Path("generic_ssg/_hurricane.md.jinja")
    if not template_path.exists():
        raise FileNotFoundError(f"Jinja template not found at {template_path}")

    template_text: str = template_path.read_text(encoding="utf-8")

    now_iso: str = datetime.now(tz=timezone.utc).replace(microsecond=0).isoformat()
    context: Dict[str, Any] = {
        "title": f"{storm_name.title()} {storm_year}",
        "created_at": hurricane_raster_ts,
        "updated_at": now_iso,
        "draft": is_draft,
        "storm_id": storm_id,
        "storm_name": storm_name,
        "storm_year": storm_year,
        "resolution": resolution,
        "adv_num": adv_num,
        "storm_preview_address": storm_preview_address,
        "storm_zip_address": storm_zip_address,
        "sources": sources or [],
        "storm_discussion": storm_discussion,
    }

    jenv = Environment(autoescape=False, trim_blocks=True, lstrip_blocks=True)
    template = jenv.from_string(template_text)

    render = template.render(**context)

    storm_ssg_name: str = f"{storm_id}_{storm_name.lower()}{storm_year}.md"
    post_path: Path = Path(f"generic_ssg/hurricane/{storm_year}/{storm_ssg_name}")
    post_path.parent.mkdir(parents=True, exist_ok=True)
    with post_path.open("w") as p:
        p.write(render)
    return


# Go Code
def run_model(storm_id: str, resolution: int, include_forecasts: bool = False) -> str:
    godin_binary: Path = Path(".", "bin", "godin")
    start_time: float = time.time()
    model_proc: CompletedProcess[Any] = subprocess.run(["./"+str(godin_binary), "-res", str(resolution), storm_id],
                                                       stdout=subprocess.PIPE, stderr=subprocess.STDOUT)
    print(f"model stdout: {model_proc.stdout}")
    print(f"model stderr: {model_proc.stderr}")
    model_proc.check_returncode()
    end_time: float = time.time()
    model_proc_out: str = str(model_proc.stdout, "utf-8")
    # print(model_proc_out)
    print(f"model {storm_id} took {end_time - start_time}s")
    output_lines: List[str] = model_proc_out.splitlines()
    storm_name: str = [li for li in output_lines if li.startswith("Name:")][0].split(":")[1].strip()

    # print(f"model run for {storm_name} finished")
    return storm_name


# Single Storm Coordination
def godin_storm(storm_id: str, resolution: int = 100, include_forecasts: bool = False, ssg_draft: bool = True, ssg_data: Dict = None, do_uploads: bool = True) -> str:
    ssg_data = ssg_data or {}
    print(f"running {storm_id}")
    year: int = int(storm_id[-4:])

    name: str = run_model(storm_id, resolution, include_forecasts)

    raster_start: float = time.time()
    hurricane_base: str = f"{name.upper()}_{year}_{resolution}x{resolution}"
    # importlib.reload(qgis_layout_exporter)
    # print("qgis reloaded")

    # smol = False
    # if smol:
    #     print("1: " + cartopy_layout.smol_run(hurricane_base, include_forecasts))
    #     print("2: " + cartopy_layout.smol_run(hurricane_base, include_forecasts))
    #     return name
    hurricane_raster: str = cartopy_layout.export_layout_png(hurricane_base, include_forecasts)

    # model_proc: CompletedProcess[Any] = subprocess.run(["python", "./scripts/qgis_layout_exporter.py", hurricane_base, str(include_forecasts)],
    #                                                    stdout=subprocess.PIPE, stderr=subprocess.STDOUT)
    # print(f"model stdout: {model_proc.stdout}")
    # print(f"model stderr: {model_proc.stderr}")
    # model_proc.check_returncode()
    # hurricane_raster: str = str(model_proc.stdout, "utf-8")
    print("qgis done")
    # print(ssg_data)
    # year: int = int(storm_id[-4:])
    # name = ssg_data['name'].lower()
    # data_path: str = f"{os.getcwd()}/data/{name}{year}"
    # hurricane_raster: str = Path(
    #     f"{data_path}/{name}{year}_{resolution}_{ts.isoformat().replace(':', '')}.jpeg"
    # )

    hurricane_raster_path: Path = Path(hurricane_raster)
    hurricane_raster_ts: str = hurricane_raster_path.stem.split("_")[-1]
    hurricane_raster_dt = datetime.strptime(hurricane_raster_ts, "%Y-%m-%dT%H%M%S%z")

    print(f"Raster completed: {time.time() - raster_start}s")

    if do_uploads:
        upload_start: float = time.time()
        # upload_event(hurricane_raster_path.name)
        zip_path, gis_path = upload_event_aws(hurricane_raster_path.name)
        print(f"Upload completed: {time.time() - upload_start}s")

        bucket_url: str = "https://s3w.cliftbar.site/godin-data/"

        ssg_start: float = time.time()
        create_ssg_from_jinja(
            storm_name=name,
            storm_id=storm_id,
            storm_year=year,
            is_draft=ssg_draft,
            resolution=resolution,
            hurricane_raster_ts=hurricane_raster_dt,
            adv_num=ssg_data.get("adv_number", "-1"),
            storm_preview_address=bucket_url + gis_path,
            storm_zip_address=bucket_url + zip_path,
            sources=ssg_data.get("sources", []),
            storm_discussion=ssg_data.get("discussion", "No Discussion")
        )
        # create_update_ssg(storm_id, name, year, resolution, hurricane_raster_ts, ssg_draft, ssg_data)
        print(f"generic SSG completed: {time.time() - ssg_start}s")

    # print("godin storm finished\n")
    return name


# Multi Storm Coordination
def godin_year(year: int = 2025, resolution: int = 10, do_git: bool = False):
    storm_count: int = 12

    if do_git:
        git_setup_start: float = time.time()
        git_setup()
        print(f"git_setup completed: {time.time() - git_setup_start}s")

    storms: list = []
    for i in range(1, storm_count + 1):
        storm: str = f"al{i:02d}{year}"
        storms.append(storm)

        godin_storm(storm, resolution, include_forecasts=True, ssg_draft=False, do_uploads=True)

        print(f"{storm} finished, {i} out of {storm_count} for year {year}")

    if do_git:
        git_push_start: float = time.time()
        git_push(storms)
        print(f"git_push_start completed: {time.time() - git_push_start}s")


def nhc_adv_rss():
    rss_binary: Path = Path(".", "bin", "rss")

    start_time: float = time.time()
    rss_proc: CompletedProcess[Any] = subprocess.run(["./"+str(rss_binary)], stdout=subprocess.PIPE, stderr=subprocess.STDOUT)
    rss_proc_out: str = str(rss_proc.stdout, "utf-8")
    print(rss_proc_out)
    rss_proc.check_returncode()
    end_time: float = time.time()

    print(f"rss took {end_time - start_time}s")


def generate_from_adv(do_rss: bool = True, do_git: bool = False, do_uploads: bool = True, pending_only: bool = True):
    cloud_run_start: float = time.time()

    if do_rss:
        nhc_adv_rss()

    with Path("sql/db.json").open(mode="r") as fi:
        conf: dict[str, Any] = json.load(fi)
        host: str = conf["host"]
        port: int = conf["port"]
        username: str = conf["username"]
        password: str = conf["password"]

    engine: Engine = create_engine(f"postgresql://{username}:{password}@{host}:{port}/postgres")

    pending_storms_start: float = time.time()
    conn: Connection
    with engine.begin() as conn:
        pending_only_check: str = ""
        if pending_only:
            pending_only_check = "NOT processed AND"
        res: CursorResult[Any] = conn.execute(text(f"WITH g AS (SELECT *, ROW_NUMBER() OVER(PARTITION BY storm_id ORDER BY adv_num DESC) AS rn FROM odin.nhc_rss) SELECT * FROM g WHERE {pending_only_check} rn = 1 --LIMIT 1"))
        pending_storms = {r["storm_id"]: r["parsed"] for r in res.mappings().all()}
    print(f"pending_storms completed: {time.time() - pending_storms_start}s")

    if do_git:
        git_setup_start: float = time.time()
        git_setup()
        print(f"git_setup completed: {time.time() - git_setup_start}s")

    for storm_id, storm in pending_storms.items():
        godin_storm_start: float = time.time()
        godin_storm(storm["storm_id"], 100, include_forecasts=True, ssg_draft=False, ssg_data=storm, do_uploads=do_uploads)
        print(f"godin_storm for {storm_id} completed: {time.time() - godin_storm_start}s")

    if do_git:
        git_push_start: float = time.time()
        git_push([v["name"] for k, v in pending_storms.items()])
        print(f"git_push_start completed: {time.time() - git_push_start}s")

    for storm_id, storm in pending_storms.items():
        with engine.begin() as conn:
            conn.execute(text(f"UPDATE odin.nhc_rss SET processed = True WHERE storm_id = '{storm['storm_id']}' AND adv_num = {storm['adv_number']}"))
            print(f"pending_storms completed: {time.time() - pending_storms_start}s")

    print(f"adv gen completed: {time.time() - cloud_run_start}s")


def main():
    # storm: str = "al022020"
    # storm: str = "al122021"
    # storm: str = "al172021":
    # storm: str = "al182021"
    # godin_storm(storm, 10, True)
    # godin_year()
    godin_year(2024, 100, False)
    godin_year(2025, 100, False)
    generate_from_adv(do_rss = True, do_git=False, do_uploads=True, pending_only=False)
    # godin_year(False)


if __name__ == "__main__":
    main()
