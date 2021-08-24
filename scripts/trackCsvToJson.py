import csv
import json
import pathlib
from datetime import datetime
from datetime import timezone
filename = "../MATTHEW_2016_Sample.txt"

file = pathlib.Path(filename)

with file.open() as fi:
    fi_csv = csv.DictReader(fi, delimiter="\t")
    
    print(fi_csv.fieldnames)

    # fi_csv_rows = [r for r in fi_csv]

    clean_rows = []
    for row in fi_csv:

        ts_parts = row["timestamp"].split("-")
        # print(row)
        # print(ts_parts)
        ts = datetime(int(ts_parts[0]), int(ts_parts[1]), int(ts_parts[2]), int(ts_parts[3]), int(ts_parts[4]), tzinfo=timezone.utc)
        # print(row)
        r = {
            "timestamp": ts.isoformat(),
            "track_sequence": int(row["sequence"]),
            "lat_y_deg": float(row["lat_y"]),
            "lon_x_deg": float(row["lon_x"]),
            "max_wind_velocity_kts": float(row["max_wind_kts"]),
            "min_central_pressure_mb": float(row["min_cp_mb"]) if row["min_cp_mb"] else None,
            "radius_max_wind_nmi": float(row["rmax_nmi"]),
            "cyclone_forward_speed_kts": float(row["fspeed_kts"]),
            "cyclone_heading_deg": float(row["heading"]),
            "gradient_wind_adjustment_factor": float(row["gwaf"])
        }
        clean_rows.append(r)
    print(clean_rows)

with open(f"MATTHEW_2016_Sample.json", mode="w") as outFi:
    outFi.write(json.dumps(clean_rows))

