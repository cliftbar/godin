---
title: "About"
date: 2021-08-30T16:08:08-04:00
draft: false
---

This site is a place for me to post Hurricane model runs from a program I've developed in my free time.  The code is an implementaiton of the NWS 23 model developed by NOAA.  See my [Github](https://github.com/cliftbar) for various versions I've made over the years.

Oh, it's called "Odin" because the previous name (FlapyDisaster) was silly and I had an idea for an icon, nothing too deep there.

### Tech
Site:
- Hugo as a Static Site Generator
- Github Pages for site hosting
- Github Actions for deployment
- Google Cloud Storage for data file hosting
- Namecheap for the domain name
- Google Analytics and ~~Google AdSense~~ (it would seem I'm not yet worthy of AdSense)

Model:
- Model code written in Golang (Python and Scala version also exist)
- The code is an implementation of the [NOAA NWS 23 Model](https://repository.library.noaa.gov/view/noaa/6948)
- Land Friction is currently not considered
- High resolution model runs (100 pixels per degree) take ~3min singlethreaded

Other
- Maps generated using QGIS/PyQGIS
- General scripting done in Python 3

### DISCLAIMER
Nothing on this site is official information or verified modeling, I'm just a dude on the internet.  Please follow all guidance from NOAA and your local officials.
