---
title: "Hurricane Modeling"
date: 2021-09-07T23:23:52-04:00
draft: true
---

And this is the fun one.

This is the third-ish time I've implemented this model: once in Python, once in Scala, and now in Go (I say "ish" because I've worked with a very similar model in C# and C++, but I don't have access to those anymore).  The model is based on NOAA Technical Report NWS 23, which describes a set of equations for going from a few key parameters (namely Max Wind Velocity, Radius of Maximum Winds, storm position, storm speed, and storm heading).  The nice thing about the model is that it's just math, and (in the realm of modeling) not overly complex math.

# Data Sources / ATCF

### Model Run Steps
#### Calculate the radial decay factor
The radial decay factor scales the maximum wind to the wind speed at the point.  Wind speed drops in a predictable way as distance from the center increases.  Radial decay is calculated using a set of empirical equation derived from the graphs on page 53 and 54 of the pdf. The relationships are based off of the graphs below and created in the Excel file
[here](/get_file/Documentation/Hurricane/NWS23/NWS_23_RadialDecay.xlsx):

**Radial Decay: r < rmax**

![Radial Decay: r < rmax](/get_file/Documentation/Hurricane/NWS23/RadialDecay_Rmax_Inward.PNG)

**Radial Decay: r >= rmax**

![Radial Decay: r < rmax](/get_file/Documentation/Hurricane/NWS23/RadialDecay_Rmax_Outward.PNG)

#### Calculate the asymmetry factor
The asymmetry factor accounts for the forward movement of the storm.  The major components of this factor the heading of the storm, the angle from the center of the storm to the current point, the distance from center, and the radius of maximum winds. The pdf pages are 51, 55, 263, and 281

The inflow angle (phi) is the radial angle that the wind takes compared to the concentric circle intersecting the current point , and is dependent on the radius of max winds and the current radius.
The inflow angle is calculated using an empirical equation from the graph below and created in the Excel
file [here](/get_file/Documentation/Hurricane/NWS23/NOAA_NWS23_Inflow_Calcs.xlsx):

**Inflow Angle, PMH**

![Inflow Angle, PMH](/get_file/Documentation/Hurricane/NWS23/InflowAngle_PMH.PNG)

Using the inflow angle of the current point, inflow angle of the maximum winds, the angle from the center of the storm, and the track heading, a beta angle is calculated as below (Note the 90 shift to convert from a bearing to a cartesian notation, and modulus operators to keep the angle confined to 360 degrees):

```
phi_beta = (phi_radius - phi_max_radius) % 360
bearing_shift = (90 - angle_from_center + track_bearing) % 360
beta = (phi_beta + bearing_shift) % 360
```

The beta angle is the major component of the final asymmetry calculation, as it accounts for the forward speed of the storm and the radial position of the current point, as well as the rotational direction of the wind at the current point.  The final asymmetry equation is below:

```
asymmetry_factor = 1.5 * (forward_speed ^ 0.63) * (to ^ 0.37) * math.cos(math.radians(beta))
```

The **to** factor is a unit conversion, and is 1 for the units used in the model.  The exponents are split to make the units a separate term.

#### Calculate the wind speed
The final wind speed equation is:

```
windspeed = (maximum_gradient_wind * gwaf * radial_decay_factor) + asymmetry_factor
```
#### Output Formats

# What does the model do?
So, what does this model show us?  First, here's some stuff it _doesn't_ have:
- Forecasting: This model **does not** generate a forecast of where a storm will go, or how strong it will be.  That information is input into the model from the NOAA track data and forecasts.
- Validation: OK, this is a bit unfair, the NWS 23 technical report has tons of validation.  But my implementation needs a bunch more validation and testing, and any individual storm could be made better by tweaking various input parameters like the GWAF and R Max.
- Land Friction:  I haven't implemented any land friction, so most wind speeds over land will be overestimated.
- Storm Surge: This is a big one, this model only considers wind, not storm surge.  And these days, in the US, storm surge does way more damage than wind.

Cool, so what the model does output is the wind field of a hurricane at a given timestamp.

# Model Performance
## Max Calc distance
## Pre-calculated tables
## index iteration over for-in style

