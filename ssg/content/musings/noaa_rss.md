---
title: "NOAA RSS"
date: 2021-09-14T17:49:38-07:00
draft: false
---

## NOAA RSS Feeds
Well, actually they're Atom feeds I guess, but they're basically interchangeable, the library I use handles both automagically, and people know what RSS is.  This is more about parsing the data than a web format, so here's a [wikipedia](https://en.wikipedia.org/wiki/Atom_(Web_standard)) link if you care about that.

So, NOAA publishes a ton of data using their feeds, in a few formats.  Some is historical data, some is re-analyzed data, graphics, GIS, etc etc.  For this project, the most interesting is the [Public Advisories](https://www.nhc.noaa.gov/help/tcp.shtml), which are issued regularly and contain key statistics and forecaster commentary.  The model isn't using the statistics dirctly from th Advisories, but the data is there.  The Discussion section is interesting, and now gets included on hurricane pages as updates happen.  But, the most useful part of the Advisories is as an update event source: NOAA issues Advisories every 6 hours for active storms, and every 3 hours for major storms.

# Reading Atom in Go
Fetching the feed in Go was fairly easy, as was parsing it according to the Atom spec.  [gofeed](https://github.com/mmcdole/gofeed) took all the work out of it, put in a URL, get out a feed object.

```go
import "github.com/mmcdole/gofeed"
fp := gofeed.NewParser()
feed, _ := 	fp.ParseURL("https://www.nhc.noaa.gov/index-at.xml")
```

The URL there will fetch all the "active" feed items as maps, which each have a text area where the notable content goes.  Which NOAA didn't use, they put the text into the description field, odd choice but oh well.  From there, loop through the items, sort out which ones are the Public Advisories, and parse the text data into structs.  Sounds simple, but next comes Regex.

# Regex Parsing in Go
The Advisories follow a format, ish, but the best tool I could come up with to extact the data was muddling through a regex matcher.  And as with all non-trivial regexes, it's a jumbled mess (citation needed).  Behold the hubris!

```go
var publicAdvRegexOfDoom = regexp.MustCompile(`(?s).*(AL[0-9][0-9][0-9][0-9][0-9][0-9])\n(.*)\n.\n.*\n.\n.*SUMMARY.*LOCATION\.\.\.([0-9]?[0-9]?[0-9]?\.[0-9]?[0-9]?)([NS]) ([0-9]?[0-9]?[0-9]?\.[0-9]?[0-9]?)([EW]).*MAXIMUM.*\.\.\.([0-9]?[0-9]?[0-9]?) MPH.*PRESENT.*OR ([0-9]?[0-9]?[0-9]?).*AT ([0-9]?[0-9]?[0-9]?) MPH.*MINIMUM CENTRAL PRESSURE\.\.\.([0-9]?[0-9]?[0-9]?[0-9]?) MB.*DISCUSSION AND OUTLOOK\n(?:[-]*\n)(.*)(?:\n[[:blank:]]\n[[:blank:]]\n).*(?:\n[[:blank:]]\n[[:blank:]]\n).*`)
```

Yes I named it `publicAdvRegexOfDoom` in the actual code, yes I'm sticking to that.  It is not efficent (I assume) nor robust (I know), but it does the thing I need it to.  After using `FindAllStringSubmatch`, each part of the Advisory text that matches where the Regex is in parenthesis gets pulled out in a slice.  Using a less confusing example:

```go
var regex = regexp.MustCompile(`Test ([0-9])([0-9])`)
match [][]string = regex.FindAllStringSubmatch("Test 12", -1)
// match[0][0] is 1
// match[0][1] is 2
```

The parenthesis mark groups that get extracted out when the regex is able to match the text.  The matches are only returned as text, so the code needs to deal with conversion to numeric types, parsing out the Latitude/Longitude format, etc.  I won't put all of that rather boring code here, you can view that [here](https://github.com/cliftbar/godin/blob/main/hurricane/nhc/rss.go).  Its worth noting that Go doesn't support some of the more complex regex features due efficiency concerns, which for me showed up as no support for "match everything except 'x'".

# Final Stucture
The final struct at the end looks like this

```go
type StormFeedInfo struct {
	Name string
	StormID string
	AdvNumber int
	Timestamp time.Time
	LatY float64
	LonX float64
	BearingDeg float64
	ForwardSpeedKts float64
	VMaxKts float64
	MinCpMb float64
	Discussion string
	Graphics []string
	Sources []string
}
```

and this is dumped as a document into two Firestore locations: one in a "pending" collection, and one in a more formal archive location, to be later picked up by the build process.

# Where It Lives
The parser itself is run every 2 hours in Cloud Functions, kicked off by Cloud Scheduler, with PubSub as the glue layer.  Cloud Scheduler is pretty straightforward, cron style format pointed to a PubSub topic, the UI and docs will work for setting that up.  The Cloud Function is triggered by new messages in the topic, but the message content isn't used for anything.  Creating the Cloud Function itself is largely pasting the code into the UI, and making sure the entry point is defined (which is provided by the initial template).  The only gotcha here was making sure to also update the `go.mod` file with any dependancies (gofeed and Firestore in this case).
