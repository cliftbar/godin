---
title: "Input Parameters"
date: 2021-11-12T10:54:00-08:00
draft: False
tags:
- "30 Minute Post"
- "Modeling"
---

This is just a quick overview of the major parameters and measurements provided by NOAA that get pulled into the model.  There are a few different sources for these parameters, though in the end the sources all have pretty much the same data.  

# Time
- **Timestamp**
  - Data is generally provided in Track datasets, where each set of measurements has a corresponding timestamp.
  - Includes Year, Month, Day, Hour, and often Minute, but hour is the primary resolution
  - measurements are usually at 6 hour intervals, but sometimes interim measurements are made for large or notables storms
  - Models tend to interpolate between time steps for more uniform coverage

# Position and Movement
- **Position**
  - Provided as a Lat/Lng, to a 10th of a degree resolution
  - Range: 90W-90E, 90S-90N
  - Most NOAA products provide this in the format 1234E, 5678S.  So parsing takes dealing with the N/S E/W character and dividing to get the decimal
- **Heading**
  - Direction of Hurricane Movement
  - Range: 0-369 Degrees
  - Generally reported in Compass Degrees (North is 0 degrees)
- **Forward Speed - FSpeed**
  - Forward motion of the hurricane along its heading
  - Range: 0-100 kts
  - Generally reported in Knots (~1.15 mph = 1 kt)

# Intensity
- **Velocity of Max Wind - VMax**
  - Range: 0-250 kts
  - Maximum 1-minute average sustained wind speed in the hurricane
  - Generally reported in Knots (~1.15 mph = 1 kt)
  - This is the major intensity measurement
- **Minimum Sea Level Pressure - MSLP**
  - Also known as **Central Pressure - CP**
  - The minimum barometric pressure in the storm
  - Range: 850-1050 mb
  - Generally reported in millibars
  - This is an alternate intensity measurement, and there are reasonably good conversions between this and Vmax

# Structure
- **Radius of Maximum Wind - RMW, Rmax**
  - The radius, from the center of the hurricane eye, where the maximum wind speed occurs
  - Range: 0-999 nmi
  - Generally measured in Nautical Miles (~1.15 mi = 1 nmi)
  - This basically is the radius of the eye wall, since in a well-formed hurricane that's where the maximum winds are
  - This is one of the more important model parameters, and one of the least accurate unfortunately

# References
- [Best track Specification](https://www.nrlmry.navy.mil/atcf_web/docs/database/new/abdeck.txt)
- [Tropical Cyclone Guidance Project](http://hurricanes.ral.ucar.edu/realtime/)
