---
title: "{{ replace .Name "-" " " | title }}"
date: {{ .Date }}
draft: true
summary: Hurricane {{.Name | title}}
disable_share: true
storm_name: "{{.Name | lower }}"
storm_year: {{now.Format "2006"}}
cloud_filename_part: _100x100_{{now.Format "20060102T1504-07"}}
---
*DISCLAIMER* This is not official information or modeling, I'm just a dude on the internet.  Please follow all guidance from NOAA and your local officials.

## Latest Windfield Map
![gis_img](https://storage.googleapis.com/godin_hurricane_data/{{<param storm_name>}}{{<param storm_year>}}/latest/{{<param storm_name>}}{{<param storm_year>}}{{<param cloud_filename_part>}}.png)

- as of {{% last_mod_time %}}
- 100px per degree
- GWAF 0.9
- No Friction
- default radius of maximum wind is 15kts

## Useful Links
[NOAA Active Cyclones](https://www.nhc.noaa.gov/)


[Tropical Tidbits](https://www.tropicaltidbits.com/storminfo/)

## Latest Data Files
[Download Zip](https://storage.googleapis.com/godin_hurricane_data/{{<param storm_name>}}{{<param storm_year>}}/latest/{{<param storm_name>}}{{<param storm_year>}}{{<param cloud_filename_part>}}.zip)

File List:
- {{<param storm_name>}}{{<param storm_year>}}_100x100.csv
- {{<param storm_name>}}{{<param storm_year>}}_100x100.png
- {{<param storm_name>}}{{<param storm_year>}}_100x100.wld
- {{<param storm_name>}}{{<param storm_year>}}{{<param cloud_filename_part>}}.png

