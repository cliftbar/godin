---
title: "QGIS Scripting"
date: 2021-09-10T00:00:00+00:00
draft: false
disable_share: true
---

Well, this was a hard one

So, those nice maps on each Hurricane page, I made those in a fantastic open source program called [QGIS](https://www.qgis.org/en/site/). I've been using QGIS on and off for years, and it has consistently been up to the GIS tasks I throw at it.  When I had access to ArcGIS, I'd frequently use QGIS for things ArcGIS couldn't do easily (or that was behind a paid add-on), and when I drew the short straw and lost the ArcGIS seat no one else could tell the difference.  Anyways, enough of the fanboy advertising, but seriously, check it out.  Back to the main story: one of the nice things about QGIS is that it's written in Python (well, and lots of C with Python bindings) so, while a bit cumbersome, you can pretty much `import qgis` (referred to as PyQGIS) and script anything you'd like, which is what I did once I sorted out the process manually in the main program.  Because while I could do all the steps manually in about 5 minutes, who doesn't like automation.

# Environment Setup
### CLI - The Easy Part
This was one of the trickiest parts, largely because I never got the environment to play nicely with PyCharm.  I use [Miniconda](https://docs.conda.io/en/latest/miniconda.html) to manage my Python environments, and some helpful person has set up a PyQGIS package on `conda-forge`.  So, you can follow the standard procedure for setting up a new conda environment and install qgis:

```shell
conda create --name qgis python=3.9
conda install -c conda-forge qgis
```

You can test this out with a minimal python script that just instantiates the PyQGIS app, closes, and exits:

```python
from qgis.core import QgsApplication

# Set up, instantiate, and close application
qgis_app: QgsApplication = QgsApplication([], False)
qgis_app.setPrefixPath(None, True)

qgis_app.initQgis()

qgis_app.exitQgis()

# I get a segfault if I don't do this in the minimal script *shrug*
qgis_app = None
print("finished")
```

and run it, unsurprisingly, with

```shell
python minimal_example.py
```

Great, now you have a python environment with PyQGIS installed and a running script, and it won't work with PyCharm (or VSCode) unless you're smarter than me (let me know if you are).  All you'll get is errors about how qgis.core can't be located.  Couldn't tell you why, but now lets sort out setting up something in the IDE.

### IDE - Kinda Annoying
So, to set things up with an IDE, the trick is to use the Python interpreter that ships with the normal QGIS application.  So the first step is to download and install QGIS using the standard instructions.  Next, you need to get the path for the QGIS install directory; there's many ways to do it, but the way to check the source of truth is to ask QGIS itself.  Open QGIS, start a Python console (under the plugin directory for me) and run

```python
QgsApplication.prefixPath()
```

this will give you the base path for the QGIS (also, note that prefix thing, it'll come back later).  On macOS, this command returned `/Applications/QGIS.app/Contents/MacOS`.  Navigate to that directory, and start hunting for the Python interpreter, I found mine at `/Applications/QGIS.app/Contents/MacOS/bin/python3`.  You can add this interpreter into PyCharm, and then the IDE can start indexing everything as normal.  But, the script still won't run as expected, since the PyQGIS object doesn't know where all its libraries are; this is where that prefix comes in.  In the minimal script, there is a line where the prefix path gets set:

```python
qgis_app.setPrefixPath(None, True)
```

The first `None` argument is the path, and in this case it needs to be that path returned in the QGIS Python console.  This makes the new minimal script

```python
from typing import Optional
from qgis.core import QgsApplication

# Export path to QGIS sqlite library if needed
# export DYLD_INSERT_LIBRARIES="/Applications/QGIS.app/Contents/MacOS/lib/libsqlite3.dylib"
qgis_install_path: str = "/Applications/QGIS.app/Contents/MacOS"


# Set up application
qgis_app: QgsApplication = QgsApplication([], False)
qgis_app.setPrefixPath(qgis_install_path, True)

qgis_app.initQgis()

qgis_app.exitQgis()

# I get a segfault if I don't do this in the minimal script *shrug*
qgis_app = None
print("finished")
```

Another note, there's a line in there about exporting an environment variable.  On my system, the sqlite library wasn't compatible, and I got segfaults with rather cryptic error messages.  That environment variable forces the session to call the QGIS library.

And poof, at the end of all this, you should have a working IDE setup.  This was the actual hard part, where I wasn't sure everything would work.  The next part is poking QGIS until it does what we want.

# Working with PyQGIS
The PyQGIS docs are, we'll say solidly ok.  But, I'd suggest doing as many type hints as possible so the IDE will fill in suggestions and save you time in the docs.  And you'll have to set most of the type hints manually, the way the files are set up doesn't always make the return types obvious to humans or the IDE.

Also, because of the C code, there's going to be segfaults.  On average, its because something wasn't set up or cleaned up properly.  A few segaults I ran into:
- missing the sqlite driver path
- qgis object not fully shutdown
- Invalid layer
- missing files

# Scripting steps
After that, its a matter of figuring out exactly how to script what the goal is.  It helps to set up a base project in the normal QGIS application to start with, it saves time fiddling with things in the script.  The full script for generating maps is [on github](https://github.com/cliftbar/godin/blob/main/scripts/qgis_layout_exporter.py), so I won't go over it here, but the general steps are:
1. Open base project
2. Add new layers (setting styles, setting CRS, etc)
3. Set map extents to layers
4. Open/create a layout and sort out titles, legends, etc
    - this is the most fiddly bit, just keep iterating until it works
5. Export the finished layout to an image

And that's scripting with PyQGIS, well the basics of it.  PyQGIS can actually be embedded in a standalone GUI application, among other things, but thats far more complicated than what I was going for here.  At the end of all this, we have a script that will load the files output from the model code (raster, wld file for raster, and track csv) onto a decently formatted map.
