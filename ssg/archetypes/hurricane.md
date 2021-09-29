---
title: {{ cond (eq (getenv "HUGO_HURRICANE_STORM_NAME") "") ( .Name ) (print (getenv "HUGO_HURRICANE_STORM_NAME" | title) " " (getenv "HUGO_HURRICANE_STORM_YEAR")) }}
date: {{ .Date }}
draft: true
summary: Tropical System {{ getenv "HUGO_HURRICANE_STORM_ID" }}
disable_share: true
storm_name: {{ getenv "HUGO_HURRICANE_STORM_NAME" | lower }}
storm_year: {{ substr .Name (sub (len .Name) 4) }}
resolution: {{ cond (eq (getenv "HUGO_HURRICANE_RES") "") 100 (getenv "HUGO_HURRICANE_RES") }}
hurricane_timestamp: {{ cond (eq (getenv "HUGO_HURRICANE_TS") "") (now.Format "20060102T1504-07") (getenv "HUGO_HURRICANE_TS") }}
adv_number: {{ getenv "HUGO_HURRICANE_ADV_NUM" }}
last_updated: {{ now.Format "2006-01-02T15:04:05-07:00" }}
adv_sources: {{ getenv "HUGO_HURRICANE_SOURCES" }}
storm_id: {{ getenv "HUGO_HURRICANE_STORM_ID" }}
---
*DISCLAIMER* This is not official information or modeling, I'm just a dude on the internet.  Please follow all guidance from NOAA and your local officials.

## Windfield Map
![gis_img](https://storage.googleapis.com/godin_hurricane_data/{{<param storm_name>}}{{<param storm_year>}}/latest/{{<param storm_name>}}{{<param storm_year>}}_{{<param resolution>}}x{{<param resolution>}}_{{<param hurricane_timestamp>}}.jpeg)

- as of {{<param last_updated>}}
- {{<param resolution>}}px per degree
- GWAF 0.9
- No Friction
- default radius of maximum wind is 15kts

## Useful Links
- [NOAA Active Cyclones](https://www.nhc.noaa.gov/)
- [Tropical Tidbits](https://www.tropicaltidbits.com/storminfo/)
{{< param_to_list >}}

## Data Files
[Download Zip](https://storage.googleapis.com/godin_hurricane_data/{{<param storm_name>}}{{<param storm_year>}}/latest/{{<param storm_name>}}{{<param storm_year>}}_{{<param resolution>}}x{{<param resolution>}}_{{<param hurricane_timestamp>}}.zip)

File List:
- `{{<param storm_name>}}{{<param storm_year>}}_{{<param resolution>}}x{{<param resolution>}}.csv`
- `{{<param storm_name>}}{{<param storm_year>}}_{{<param resolution>}}x{{<param resolution>}}.png`
- `{{<param storm_name>}}{{<param storm_year>}}_{{<param resolution>}}x{{<param resolution>}}.wld`
- `{{<param storm_name>}}{{<param storm_year>}}_{{<param resolution>}}x{{<param resolution>}}_{{<param hurricane_timestamp>}}.jpeg`

{{ with (getenv "HUGO_HURRICANE_DISCUSSION") }}
## Official Advisory Discussion
{{.}}
{{ end }}