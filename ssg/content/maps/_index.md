---
title: Map
layout: blank
draft: false
---
# Storm Calculator
[Standalone Page](/html/storm_calc.html)

### Notes
- Track interpolation goes 12hrs into the past, 24 hours into the future, in a straight line, based of the Forward speed.  It also decreases the wind speed using linear interpolation until its 1/3 the starting speed at landfall
- Double-click the map to set the landfall point
- Track must be created before the animation will run
- Point size and opacity only affects the static layer from "Calculate"

{{< importPartial "mapping_partial.html" >}}