---
title: "Automation!"
date: 2021-09-23T22:38:17-07:00
draft: false
---

Turns out Automation is neat.

After some fairly manic coding, much trial and error, and some grave computing sins (DevOps folks look away), I've set up a Pipeline for automatically generating hurricane pages on this site.  Google Cloud Platform is the MVP here, housing all of the processing pipeline, and (so far) the completely in the free tier.  So now new hurricanes will be automatically added to the site as they're tracked by NOAA, and current storms will reflect (more or less) the latest forecasts.  I'm quite happy to get this into place, I've never gotten this project so far on the automation front.

## Pipeline Flow
- Check for updates from the NOAA Tropical Cyclone RSS feeds every two hours
- When a new Advisory is posted, trigger a build, which consists of the following:
    - Re-run the model with updated information
    - Regenerate the files and maps
    - Create or Update the website page in verison control
- When the build is done, redepoly the website

Over time I'll write about how I managed to get things in place, and any updates to the pipeline.  This page will get updated with links to more specific posts as I write them, so there's a record.

## GCP Pipeline Service List

- Cloud Function
- Cloud Build
- Google Container Registry
- Firestore
- Google Storage
- Google PubSub
- Cloud Scheduler

## Pipeline Cost
- August 2021: $0.00
- September 2021: $0.10
  - Cloud Build ($0.03): I did lots of lazy iterating of the containers in Build
  - Storage ($0.07): I missed that the Container Registry was storing out of date images, so now I'm keeping that cleaned up