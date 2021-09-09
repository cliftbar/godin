package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"godin/hurricane"
	"godin/hurricane/atcf"
	"image"
	"image/color"
	"image/png"
	"io/ioutil"
	"os"
	"strings"
	"time"
)

func myUsage() {
	fmt.Printf("Usage: %s [flags] stormID\n", os.Args[0])
	flag.PrintDefaults()
}

func main(){
	// stormID := "al082021" //Henri 2021
	// stormID := "al092021" //Ida 2021
	// stormID := "al122005" // katrina 2005
	// stormID := "al182012" // sandy 2012

	//stormID := "al142016" // matthew 2016

	//pixPerDegLatY := 100
	//pixPerDegLonX := 100
	//rMaxDefaultNmi := 15.0
	//maxCalcDistNmi := 360.0
	//gwaf := 0.9
	// includeForecasts := false
	flag.Usage = myUsage
	includeForecasts := flag.Bool("include_forecasts", false, "whether to include forecast track points in the model run")
	pixPerDegree := flag.Int("res", 100, "resolution of the output raster in pixels per degree")
	rMaxDefaultNmi := flag.Float64("default_rmax", 15.0, "radius of max wind to use when not provided by track data")
	maxCalcDistNmi := flag.Float64("max_calc_dist", 360.0, "maximum calculation distance from center of storm")
	gwaf := flag.Float64("gwaf", 0.9, "Gradient Wind Reduction Factor")
	flag.Parse()

	var stormID string
	if 0 < len(flag.Args()) {
		stormID = flag.Arg(0)
	} else {
		fmt.Println("Specify a storm ID")
		os.Exit(2)
	}

	pixPerDegLatY := *pixPerDegree
	pixPerDegLonX := *pixPerDegree

	SingleCalc(stormID, pixPerDegLatY, pixPerDegLonX, *rMaxDefaultNmi, *maxCalcDistNmi, *gwaf, *includeForecasts)
}

func cloudCalc(stormID string){
	// bucket := "godin_hurricane_data"
	rMaxNmiDefault := 15.0
	gwaf := 0.9
	maxCalculationDistance := 360.0
	pxPerDegree := 10
	atcf.FetchATCFBDeckTrack(stormID)
	atcf.FetchATCFForecastTrack(stormID)
	event := atcf.FetchAtcfEvent(stormID, rMaxNmiDefault, gwaf, true)

	ce := event.CalculateEvent(pxPerDegree, pxPerDegree, maxCalculationDistance)

	toRaster(ce)
	trakXYZ := ce.TrackToDelimited(true)

	_ = ioutil.WriteFile(fmt.Sprintf("%s_%dx%d.csv", ce.Info.Name, ce.PixPerDegreeLonX, ce.PixPerDegreeLatY), []byte(trakXYZ), 0644)

	wldText := fmt.Sprintf("%f\n0\n0\n%f\n%d\n%d", 1.0 / float64(ce.PixPerDegreeLonX), -1.0 / float64(ce.PixPerDegreeLatY), ce.Info.Bounds.LonXLeftDeg, ce.Info.Bounds.LatYTopDeg)

	_ = ioutil.WriteFile(fmt.Sprintf("%s_%dx%d.wld", ce.Info.Name, ce.PixPerDegreeLonX, ce.PixPerDegreeLatY), []byte(wldText), 0644)
}

func SingleCalc(stormID string, pixPerDegLatY int, pixPerDegLonX int, rMaxDefaultNmi float64, maxCalcDistNmi float64, gwaf float64, includeForecasts bool) hurricane.CalculatedEvent {
	// stormID := "al082021" //Henri 2021
	//stormID := "al092021" //Ida 2021
	// stormID := "al122005" // katrina 2005
	//stormID := "al182012" // sandy 2012

	event := atcf.FetchAtcfEvent(stormID, rMaxDefaultNmi, gwaf, includeForecasts)

	startTime := time.Now().UTC()
	fmt.Printf("Calc Start time: %s\n", startTime.Format(time.RFC3339))
	ce := event.CalculateEvent(pixPerDegLatY, pixPerDegLonX, maxCalcDistNmi)
	endTime := time.Now().UTC()
	fmt.Printf("Calc End time: %s, Duration: %fs\n", endTime.Format(time.RFC3339), endTime.Sub(startTime).Seconds())

	stormNameYear := fmt.Sprintf("%s%d", strings.ToLower(ce.Info.Name), ce.Info.Year)
	_ = os.MkdirAll(fmt.Sprintf("data/%s/past", stormNameYear), os.ModePerm)
	src := fmt.Sprintf("data/%s", stormNameYear)
	dest := fmt.Sprintf("data/%s/past", stormNameYear)
	dirEntries, _ := os.ReadDir(src)
	for _, entry := range dirEntries {
		if !entry.IsDir() {
			_ = os.Rename(src + "/" + entry.Name(), dest + "/" + entry.Name())
		}
	}

	toRaster(ce)
	trackXYZ := ce.TrackToDelimited(true)

	_ = ioutil.WriteFile(fmt.Sprintf("data/%s/%s_%d_%dx%d.csv", stormNameYear, ce.Info.Name, ce.Info.Year, ce.PixPerDegreeLonX, ce.PixPerDegreeLatY), []byte(trackXYZ), 0644)

	wldText := fmt.Sprintf("%f\n0\n0\n%f\n%d\n%d", 1.0 / float64(ce.PixPerDegreeLonX), -1.0 / float64(ce.PixPerDegreeLatY), ce.Info.Bounds.LonXLeftDeg, ce.Info.Bounds.LatYTopDeg)

	_ = ioutil.WriteFile(fmt.Sprintf("data/%s/%s_%d_%dx%d.wld", stormNameYear, ce.Info.Name, ce.Info.Year, ce.PixPerDegreeLonX, ce.PixPerDegreeLatY), []byte(wldText), 0644)

	fmt.Println("Name: " + ce.Info.Name)

	return ce
}

func main2(){
	fileContents, _ := ioutil.ReadFile("MATTHEW_2016_Sample.json")

	var event hurricane.EventInformation

	_ = json.Unmarshal(fileContents, &event)
	fmt.Printf("Starting calculation %v", time.Now())
	ce := event.CalculateEvent(10, 10, 360)
	fmt.Printf("Finished calculation %v", time.Now())

	toRaster(ce)

	//
	//calcContents, _ := json.Marshal(ce)
	//_ = ioutil.WriteFile("calcdOut.json", calcContents, 0644)
}

func toRaster(ce hurricane.CalculatedEvent) {
	//println(ce.Info.Name)
	stormNameYear := fmt.Sprintf("%s%d", strings.ToLower(ce.Info.Name), ce.Info.Year)

	width := ce.Info.Bounds.GetBlockWidth(ce.PixPerDegreeLonX)
	height := ce.Info.Bounds.GetBlockHeight(ce.PixPerDegreeLatY)
	raster := image.NewGray(image.Rectangle{
		Min: image.Point{X: 0, Y: 0},
		Max: image.Point{X: width, Y: height},
	})
	i := 0
	for y := height; y > 0; y-- {
		for x := 0; x < width; x++ {
			val := uint8(ce.WindField[i].Value)
			raster.SetGray(x, y, color.Gray{Y: val})
			i++
		}
	}

	o, _ := os.Create(fmt.Sprintf("data/%s/%s_%d_%dx%d.png", stormNameYear, ce.Info.Name, ce.Info.Year, ce.PixPerDegreeLonX, ce.PixPerDegreeLatY))

	_ = png.Encode(o, raster)
}
