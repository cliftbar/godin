import os
import sys
import csv
from datetime import datetime, timezone
from pathlib import Path
from typing import List, Optional, Tuple

import matplotlib

# Use a non-interactive backend to support headless environments
matplotlib.use("Agg")
import matplotlib.pyplot as plt  # noqa: E402
import matplotlib.patheffects as pe  # noqa: E402
from matplotlib.patches import Patch  # noqa: E402

try:
    import cartopy.crs as ccrs  # type: ignore  # noqa: E402
    import cartopy.feature as cfeature  # type: ignore  # noqa: E402
except Exception as e:  # pragma: no cover
    # Defer import error to runtime usage with a clearer message
    raise RuntimeError(
        "Cartopy is required for cartopy_layout.py. Install cartopy per https://cartopy.readthedocs.io/."
    ) from e


def _parse_hurricane_base(hurricane_file_base: str) -> Tuple[str, int, str]:
    parts: List[str] = hurricane_file_base.split("_")
    if len(parts) < 3:
        raise ValueError(
            f"hurricane_file_base '{hurricane_file_base}' does not match expected pattern '<name><year>_<res>_<timestamp>'"
        )
    name = parts[0]
    # Split trailing digits (year) from name, e.g., ida2021 -> ida + 2021
    # name = "".join([c for c in name_year if not c.isdigit()])
    year_str = parts[1]
    if not year_str:
        raise ValueError(f"Could not parse year from '{year_str}'")
    year = int(year_str)
    resolution = parts[2]
    return name, year, resolution


def _load_track_csv(csv_path: Path) -> Tuple[List[float], List[float], Optional[List[str]], Optional[List[float]]]:
    lons: List[float] = []
    lats: List[float] = []
    types: Optional[List[str]] = []
    speeds: Optional[List[float]] = []

    if not csv_path.exists():
        raise FileNotFoundError(f"Track CSV not found: {csv_path}")

    with csv_path.open(newline="") as f:
        header = [h.strip() for h in f.readline().split(',')]
        reader = csv.DictReader(f, fieldnames=header)
        # Try common column names
        lon_keys = ["lon", "longitude", "lonX", "x", "Lon", "LON"]
        lat_keys = ["lat", "latitude", "latY", "y", "Lat", "LAT"]
        type_keys = ["type", "status", "forecast", "Type", "TYPE", "source"]
        speed_keys = [
            "z",
            "maxWindKts",
            "wind",
            "windspeed",
            "vmax",
            "kt",
            "kts",
            "KT",
            "KTS",
            "mph",
            "MPH",
            "ws",
            "WS",
        ]

        first_row = True
        for row in reader:
            if first_row:
                # Confirm columns available
                lon_key = next((k for k in lon_keys if k in row.keys()), None)
                lat_key = next((k for k in lat_keys if k in row), None)
                type_key = next((k for k in type_keys if k in row), None)
                speed_key = next((k for k in speed_keys if k in row), None)
                if lon_key is None or lat_key is None:
                    raise ValueError(
                        f"CSV missing lon/lat columns. Looked for {lon_keys} and {lat_keys}. Columns present: {list(row.keys())}"
                    )
                # Initialize optional lists only if present
                if type_key is None:
                    types = None
                if speed_key is None:
                    speeds = None
                first_row = False

            # Re-evaluate keys each loop would be redundant; values captured on first row
            lon_val = float(row[lon_key])  # type: ignore[name-defined]
            lat_val = float(row[lat_key])  # type: ignore[name-defined]
            lons.append(lon_val)
            lats.append(lat_val)
            if types is not None:
                types.append(str(row.get(type_key, "")))  # type: ignore[name-defined]
            if speeds is not None:
                try:
                    sp = row.get(speed_key)  # type: ignore[name-defined]
                    speeds.append(float(sp) if sp not in (None, "") else float("nan"))
                except Exception:
                    # If conversion fails, store NaN to keep alignment
                    try:
                        speeds.append(float("nan"))
                    except Exception:
                        pass

    return lons, lats, types, speeds


def _compute_extent(lons: List[float], lats: List[float], padding_deg: float = 1.0) -> Tuple[float, float, float, float]:
    if not lons or not lats:
        # Default to a CONUS view if no data
        return (-100.0, -60.0, 15.0, 45.0)
    min_lon, max_lon = min(lons), max(lons)
    min_lat, max_lat = min(lats), max(lats)
    return (
        min_lon - padding_deg,
        max_lon + padding_deg,
        min_lat - padding_deg,
        max_lat + padding_deg,
    )


def _read_world_file(wld_path: Path, img_width: int, img_height: int) -> Optional[Tuple[float, float, float, float]]:
    """
    Parse a world (.wld) file to compute the geographic extent for an image.
    Returns (xmin, xmax, ymin, ymax) if successful and unrotated, otherwise None.
    """
    try:
        with open(wld_path, "r", encoding="utf-8") as f:
            lines = [line.strip() for line in f.readlines() if line.strip()]
        if len(lines) < 6:
            return None
        A = float(lines[0])  # pixel size in X direction
        D = float(lines[1])  # rotation term (about Y)
        B = float(lines[2])  # rotation term (about X)
        E = float(lines[3])  # pixel size in Y direction (typically negative)
        C = float(lines[4])  # X coord of center of upper-left pixel
        F = float(lines[5])  # Y coord of center of upper-left pixel
        # If rotation terms present, skip — imshow extent cannot represent rotation
        if abs(B) > 1e-9 or abs(D) > 1e-9:
            return None
        xmin = C - (A / 2.0)
        ymax = F - (E / 2.0)
        xmax = xmin + A * img_width
        ymin = ymax + E * img_height
        # Ensure ordering left<right, bottom<top
        left, right = (xmin, xmax) if xmin < xmax else (xmax, xmin)
        bottom, top = (ymin, ymax) if ymin < ymax else (ymax, ymin)
        return (left, right, bottom, top)
    except Exception:
        return None


def export_layout_png(hurricane_file_base: str, include_forecasts: bool = False) -> str:
    """
    Cartopy-based reimplementation of the QGIS layout exporter.

    Inputs/outputs mirror the QGIS script conventions:
    - Input base name: '<storm><year>_<resolution>_<timestamp>' (without extension)
    - Reads track from: ./data/{storm}/{hurricane_file_base}.csv
    - Optionally overlays raster if present: ./data/{storm}/{hurricane_file_base}.png (best-effort)
    - Writes JPEG to: ./data/{storm}/{storm}{year}_{resolution}_{utc_ts}.jpeg

    Returns the path to the exported JPEG.
    """
    name, year, resolution = _parse_hurricane_base(hurricane_file_base)
    # storm_dir = Path(os.getcwd()) / f"{name.lower()}{year}"
    storm_dir = Path(os.getcwd()) / "data" / f"{name.lower()}{year}"

    track_csv = storm_dir / f"{hurricane_file_base}.csv"
    raster_png = storm_dir / f"{hurricane_file_base}.png"

    lons, lats, types, speeds = _load_track_csv(track_csv)

    # Apply source-based filtering per requirements:
    # - Always exclude rows with source "interp"/"intrp" (case-insensitive)
    # - Rows with source "OFCL" are forecasts; include only if include_forecasts=True
    if types is not None:
        filtered_lons: List[float] = []
        filtered_lats: List[float] = []
        filtered_types: List[str] = []
        filtered_speeds: Optional[List[float]] = [] if speeds is not None else None
        for i, (lo, la, ty) in enumerate(zip(lons, lats, types)):
            src = str(ty).strip().lower()
            # Normalize common interpolation labels
            if src in ("interp", "intrp"):
                continue
            if src == "ofcl" and not include_forecasts:
                continue
            filtered_lons.append(lo)
            filtered_lats.append(la)
            filtered_types.append(ty)
            if filtered_speeds is not None and speeds is not None:
                filtered_speeds.append(speeds[i])
        # Only replace if we actually performed filtering to avoid breaking alignment when no source column
        lons, lats, types, speeds = filtered_lons, filtered_lats, filtered_types, filtered_speeds

    # Compute dynamic extent from track; add some padding
    west, east, south, north = _compute_extent(lons, lats, padding_deg=2.5)

    # Matplotlib/Cartopy setup
    proj = ccrs.PlateCarree()
    fig = plt.figure(figsize=(10, 7), dpi=150)
    ax = plt.axes(projection=proj)

    # Background features
    ax.set_extent([west, east, south, north], crs=proj)
    ax.add_feature(cfeature.LAND.with_scale("50m"), facecolor="#f7f7f5")
    ax.add_feature(cfeature.OCEAN.with_scale("50m"), facecolor="#e6f2ff")
    ax.add_feature(cfeature.COASTLINE.with_scale("50m"), linewidth=0.6, edgecolor="#2c3e50")
    ax.add_feature(cfeature.BORDERS.with_scale("50m"), linewidth=0.3, edgecolor="#7f8c8d")
    try:
        ax.add_feature(cfeature.STATES.with_scale("50m"), linewidth=0.2, edgecolor="#95a5a6")
    except Exception:
        # STATES may not be available globally
        pass

    # Prepare to capture raster styling for legend/colorbar
    raster_cmap = None
    raster_norm = None

    # Optional raster overlay (best effort)
    if raster_png.exists():
        try:
            import numpy as np  # local import to avoid hard dep if not used
            import matplotlib.colors as mcolors

            def _build_hurricane_cmap_and_norm() -> Tuple[matplotlib.colors.Colormap, matplotlib.colors.Normalize]:
                # Hurricane category anchor points (kts) and colors from QGIS style
                # <=34 transparent; then green→yellow→orange→red
                anchors = [
                    (34.0, "#b7e2a8"),  # TS start (make under transparent)
                    (64.0, "#e7f5b7"),  # Cat 1
                    (83.0, "#fee8a4"),  # Cat 2
                    (96.0, "#fdba6e"),  # Cat 3
                    (113.0, "#ed6e43"), # Cat 4
                    (137.0, "#d7191c"), # Cat 5+
                ]
                vmin = anchors[0][0]
                vmax = anchors[-1][0]
                xs = [ (v - vmin) / (vmax - vmin) for v, _ in anchors ]
                cs = [ c for _, c in anchors ]
                cmap = mcolors.LinearSegmentedColormap.from_list("hurr_cat", list(zip(xs, cs)))
                # Everything under 34 kts is transparent
                cmap.set_under((0, 0, 0, 0))
                norm = mcolors.Normalize(vmin=vmin, vmax=vmax)
                return cmap, norm

            def _prepare_speed_field(img_arr: np.ndarray) -> np.ndarray:
                # Convert loaded image to a 2D array of windspeed (kts). Use ONLY the red channel; do NOT rescale.
                arr = img_arr
                if arr.ndim == 3:  # RGB or RGBA -> use red channel only
                    # matplotlib.pyplot.imread returns floats 0-1 for PNG/JPEG; we assume values already encode kts appropriately.
                    # Select the red channel (index 0). If alpha is present, it's ignored.
                    arr = arr[..., 0]
                return arr

            from PIL import Image as _PIL_Image
            with _PIL_Image.open(str(raster_png)) as _im:
                # Force single-band grayscale; do not convert to RGB
                im_gray = _im.convert("L")
            img = np.array(im_gray)
            speed_field = _prepare_speed_field(img)
            cmap, norm = _build_hurricane_cmap_and_norm()
            raster_cmap = cmap
            raster_norm = norm

            # Determine extent: prefer world file if present and unrotated
            wld_path = raster_png.with_suffix(".wld")
            world_extent: Optional[Tuple[float, float, float, float]] = None
            try:
                w, h = im_gray.size  # width, height in pixels
            except Exception:
                w, h = (img.shape[1], img.shape[0]) if hasattr(img, "shape") else (None, None)
            if w is not None and h is not None and wld_path.exists():
                world_extent = _read_world_file(wld_path, w, h)

            extent_to_use: Tuple[float, float, float, float] = world_extent if world_extent is not None else (west, east, south, north)

            ax.imshow(
                speed_field,
                origin="upper",
                extent=list(extent_to_use),
                transform=proj,
                cmap=cmap,
                norm=norm,
                interpolation="nearest",
            )
        except Exception:
            # If reading fails, skip silently (we still draw the track)
            pass

    # Plot track as a thin black line, and style points by source
    ax.plot(
        lons,
        lats,
        color="#000000",
        linewidth=0.8,
        transform=proj,
        label=f"Track",
    )

    # Points: BEST = blue, OFCL = pink
    if types is not None and len(types) == len(lons):
        best_lons: List[float] = []
        best_lats: List[float] = []
        ofcl_lons: List[float] = []
        ofcl_lats: List[float] = []
        # other_lons: List[float] = []
        # other_lats: List[float] = []
        for lo, la, ty in zip(lons, lats, types):
            src = (ty or "").strip().upper()
            if src == "BEST":
                best_lons.append(lo)
                best_lats.append(la)
            elif src == "OFCL":
                ofcl_lons.append(lo)
                ofcl_lats.append(la)
            # else:
            #     other_lons.append(lo)
            #     other_lats.append(la)
        if best_lons:
            ax.scatter(best_lons, best_lats, s=14, color="#2980b9", transform=proj, zorder=3, label="Observed")
        if ofcl_lons:
            ax.scatter(ofcl_lons, ofcl_lats, s=14, color="#ff69b4", transform=proj, zorder=3, label="Forecasted")
        # if other_lons:
        #     ax.scatter(other_lons, other_lats, s=12, color="#7f8c8d", transform=proj, zorder=3, label="Other")
    else:
        # Fallback if source info unavailable
        ax.scatter(lons, lats, s=12, color="#2980b9", transform=proj, zorder=3)

    # Annotate windspeed labels if available (decimated to reduce clutter)
    try:
        if speeds is not None and len(speeds) == len(lons):
            n = len(speeds)
            max_labels = 12
            step = max(1, n // max_labels)
            for i in range(0, n, step):
                sp = speeds[i]
                if sp is None:
                    continue
                try:
                    # Skip NaN values
                    if sp != sp:  # NaN check without importing math
                        continue
                except Exception:
                    continue
                label = f"{sp:.0f}"
                ax.text(
                    lons[i],
                    lats[i],
                    label,
                    transform=proj,
                    fontsize=7,
                    color="#2c3e50",
                    ha="left",
                    va="bottom",
                    zorder=5,
                    path_effects=[pe.withStroke(linewidth=2, foreground="white", alpha=0.85)],
                )
    except Exception:
        # Avoid failing the whole render due to labeling
        pass

    # Forecast points: if requested and type info available, render differently
    if include_forecasts and types is not None:
        f_lons: List[float] = []
        f_lats: List[float] = []
        for lo, la, ty in zip(lons, lats, types):
            if ty and str(ty).lower().startswith("f"):
                f_lons.append(lo)
                f_lats.append(la)
        if f_lons and f_lats:
            ax.scatter(f_lons, f_lats, s=28, marker="x", color="#2980b9", transform=proj, zorder=4, label="Forecast")

    # Title removed per requirement: no title on output image
    # Legends will be composed after the colorbar is created to include its labels
    try:
        pass
    except Exception:
        pass

    # Add raster colorbar (legend) as a compact inset in the top-right
    try:
        if raster_cmap is not None and raster_norm is not None:
            # import matplotlib.cm as cm
            # # Create an inset axes in the top-right corner of the plot
            # try:
            #     cax = ax.inset_axes((0.1, 0.1, 0.26, 0.02))  # [x, y, width, height] in axes fraction
            #     coords = ax.transAxes.inverted().transform(cax.get_tightbbox())
            #     ax.add_patch(plt.Rectangle(coords[0]+0.03, w, h, fc=(1, 1, 1, 0.8),
            #                                transform=ax.transAxes, zorder=2))
            #
            # except Exception:
            #     cax = None
            # sm = cm.ScalarMappable(cmap=raster_cmap, norm=raster_norm)
            # sm.set_array([])
            # if cax is not None:
            #     cbar = plt.colorbar(sm, cax=cax, orientation="horizontal")
            # else:
            #     # Fallback: place on the right side, small fraction
            #     cbar = plt.colorbar(sm, ax=ax, orientation="horizontal", fraction=0.03, pad=0.02)
            # Category ticks and smaller labels to keep compact
            ticks = [34, 64, 83, 96, 113, 137]
            # cbar.set_ticks(ticks)
            # Add hurricane categories to the colorbar tick labels
            cbar_labels = [
                "TS    (34)",      # ≤34 kts (transparent/under-range)
                "Cat 1 (64)",
                "Cat 2 (83)",
                "Cat 3 (96)",
                "Cat 4 (113)",
                "Cat 5 (137)",
            ]
            # cbar.set_ticklabels(cbar_labels)
            # cbar.ax.tick_params(labelsize=6, labelrotation=90)
            # cbar.set_label("Wind (kts)", fontsize=6)

            # Also add these colorbar labels into the main legend using colored rectangles
            try:
                # Prepare representative colors at midpoints of each category range
                mids = [
                    (34 + 64) / 2.0,
                    (64 + 83) / 2.0,
                    (83 + 96) / 2.0,
                    (96 + 113) / 2.0,
                    (113 + 137) / 2.0,
                    140.0,  # slightly above max for Cat 5+
                ]
                cat_patches = [
                    Patch(facecolor=raster_cmap(raster_norm(v)), edgecolor="none", label=lab)
                    for v, lab in zip(mids, cbar_labels)
                ]
                # Gather existing handles/labels from plotted data
                h_existing, l_existing = ax.get_legend_handles_labels()
                # Combine with category patches
                h_all = h_existing + cat_patches
                l_all = l_existing + cbar_labels
                ax.legend(handles=h_all, labels=l_all, loc="upper right", fontsize=6, framealpha=0.9, title="Wind Speed Knots", title_fontsize=6)
            except Exception:
                # If anything fails, fall back to a simple legend
                try:
                    ax.legend(loc="upper right", fontsize=6, framealpha=0.9)
                except Exception:
                    pass
    except Exception:
        pass

    # Output filename mirrors QGIS exporter time stamp convention
    ts: datetime = datetime.now(tz=timezone.utc)
    ts = ts.replace(microsecond=0, second=0)
    out_path = storm_dir / f"{name.lower()}{year}_{resolution}_{ts.isoformat().replace(':', '')}.jpeg"
    out_path.parent.mkdir(parents=True, exist_ok=True)

    plt.savefig(out_path, bbox_inches="tight", facecolor="white", pad_inches=0)
    plt.close(fig)

    return str(out_path)


def smol_run(hurricane_file_base: str, include_forecasts: bool = False) -> str:
    """Small wrapper to mirror the QGIS module helper and enable quick CLI runs."""
    return export_layout_png(hurricane_file_base, include_forecasts)


if __name__ == "__main__":
    if len(sys.argv) < 2:
        print("Usage: python scripts/cartopy_layout.py <hurricane_file_base> [--include-forecasts]")
        sys.exit(2)
    base = sys.argv[1]
    include = "--include-forecasts" in sys.argv[2:]
    out = export_layout_png(base, include_forecasts=include)
    print(out)
