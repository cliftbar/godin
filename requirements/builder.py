import os
import subprocess
import sys
from pathlib import Path

platform: str = sys.platform


def go_build() -> None:
    subprocess.run(["go", "build", "-o", "godin", "cmd/odin/main.go"])
    subprocess.run(["chmod", "+x", "godin"])
    subprocess.run(["go", "build", "-o", "bin/godin", "cmd/odin/main.go"])
    subprocess.run(["chmod", "+x", "bin/godin"])
    subprocess.run(["go", "build", "-o", "bin/rss", "cmd/rss/main.go"])
    subprocess.run(["chmod", "+x", "bin/rss"])


def requirements() -> None:
    conda_env: str = os.getenv("CONDA_DEFAULT_ENV")
    subprocess.run(["conda", "env", "export", "--from-history", "--name", conda_env, "--file", str(Path("requirements", "conda.yml"))])
    # subprocess.run(["conda", "env", "export", "--from-history", "--file", str(Path("requirements", "conda.yml"))])


def main():
    go_build()
    requirements()


if __name__ == "__main__":
    main()
