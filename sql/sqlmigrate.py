import json
from pathlib import Path
from typing import Any
from datetime import datetime

from sqlalchemy import text, Result, create_engine, Engine, Connection


def migrate() -> None:
    migration_dir: Path = Path("migrations")
    with Path("db.json").open(mode="r") as fi:
        conf: dict[str, Any] = json.load(fi)
        host: str = conf["host"]
        port: int = conf["port"]
        username: str = conf["username"]
        password: str = conf["password"]

    engine: Engine = create_engine(f"postgresql://{username}:{password}@{host}:{port}/postgres")

    with engine.connect() as conn:
        res: Result = conn.execute(text("SELECT 1"))
        print(res.all())

    for fi in migration_dir.glob("*.sql"):
        with fi.open(mode="r") as sql_fi:
            migration: str = sql_fi.read()

            conn: Connection
            with engine.begin() as conn:
                conn.execute(text(migration))
                print(f"{datetime.utcnow().isoformat()}: Completed {fi.name}")


if __name__ == "__main__":
    migrate()
