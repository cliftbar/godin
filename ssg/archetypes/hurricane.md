---
title: "{{ substr .Name 0 (sub (len .Name) 4) | title }} {{ substr .Name (sub (len .Name) 4) }}"
date: {{ .Date }}
draft: true
summary: Hurricane {{ substr .Name 0 (sub (len .Name) 4) | title }} {{ substr .Name (sub (len .Name) 4) }}
disable_share: true
storm_name: {{ substr .Name 0 (sub (len .Name) 4) | lower }}
storm_year: {{ substr .Name (sub (len .Name) 4) }}
resolution: {{ cond (eq (getenv "HUGO_HURRICANE_RES") "") 100 (getenv "HUGO_HURRICANE_RES") }}
hurricane_timestamp: {{ cond (eq (getenv "HUGO_HURRICANE_TS") "") (now.Format "20060102T1504-07") (getenv "HUGO_HURRICANE_TS") }}
---
*DISCLAIMER* This is not official information or modeling, I'm just a dude on the internet.  Please follow all guidance from NOAA and your local officials.

## Latest Windfield Map
![gis_img](https://storage.googleapis.com/godin_hurricane_data/{{<param storm_name>}}{{<param storm_year>}}/latest/{{<param storm_name>}}{{<param storm_year>}}_{{<param resolution>}}x{{<param resolution>}}_{{<param hurricane_timestamp>}}.jpeg)

- as of {{% last_mod_time %}}
- {{<param resolution>}}px per degree
- GWAF 0.9
- No Friction
- default radius of maximum wind is 15kts

## Useful Links
[NOAA Active Cyclones](https://www.nhc.noaa.gov/)


[Tropical Tidbits](https://www.tropicaltidbits.com/storminfo/)

## Latest Data Files
[Download Zip](https://storage.googleapis.com/godin_hurricane_data/{{<param storm_name>}}{{<param storm_year>}}/latest/{{<param storm_name>}}{{<param storm_year>}}_{{<param resolution>}}x{{<param resolution>}}_{{<param hurricane_timestamp>}}.zip)

File List:
- {{<param storm_name>}}{{<param storm_year>}}_100x100.csv
- {{<param storm_name>}}{{<param storm_year>}}_100x100.png
- {{<param storm_name>}}{{<param storm_year>}}_100x100.wld
- {{<param storm_name>}}{{<param storm_year>}}_{{<param resolution>}}x{{<param resolution>}}_{{<param hurricane_timestamp>}}.jpeg

