---
title: "Maps!"
date: 2021-10-29T18:01:32-07:00
draft: false
---

I played with WebAssembly, and now there's a map!

Well, I hit the "compile to WebAssembly" button on the Go compiler.  Turns out it's not that hard for the basics, which is pretty cool.  Most of the work was put into remembering how the GeoJSON format works, so I could do the string formatting on the Go side.  I'm also taking a bit sending back just a string and parsing the string on the JS side, but it works well enough.  
The other surprisingly nice experience was Mapbox GL.  Last time I did mapping work was with Leaflet a several years ago, which is a great library, but getting reasonabe performance with 10,000+ GeoJSON features was a questionable prospect even with canvas tricks.  This time through, I just chucked the GeoJSON in, and it just kinda works well.

So, putting together Go/WASM and Mapbox GL, and my poor web frontend skills, I put together a page to run the model for a single timestep.  The only real use at the moment is to see what a hurricane wind field looks like at landfall, but that's enough for now.

The [Map]({{< ref "/maps/_index.md" >}}) is set up with default values close to Ida at landfall, so play around and see what the parameters do. 
