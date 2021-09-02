package main

import (
	"encoding/json"
	"fmt"
	"godin/hurricane"
	"godin/hurricane/atcf"
	"image"
	"image/color"
	"image/png"
	"io/ioutil"
	"os"
	"time"
)

func main(){
	// stormID := "al082021" //Henri 2021
	// stormID := "al092021" //Ida 2021
	// stormID := "al122005" // katrina 2005
	stormID := "al182012" // sandy 2012

	pixPerDegLatY := 10
	pixPerDegLonX := 10
	rMaxDefaultNmi := 15.0
	maxCalcDistNmi := 360.0
	gwaf := 0.9

	SingleCalc(stormID, pixPerDegLatY, pixPerDegLonX, rMaxDefaultNmi, maxCalcDistNmi, gwaf)
}

func cloudCalc(stormID string){
	// bucket := "godin_hurricane_data"
	rMaxNmiDefault := 15.0
	gwaf := 0.9
	maxCalculationDistance := 360.0
	pxPerDegree := 10
	atcf.FetchATCFBDeckTrack(stormID)
	atcf.FetchATCFForecastTrack(stormID)
	event := atcf.FetchAtcfEvent(stormID, rMaxNmiDefault, gwaf)

	ce := event.CalculateEvent(pxPerDegree, pxPerDegree, maxCalculationDistance)

	toRaster(ce)
	trakXYZ := ce.TrackToDelimited(true)

	_ = ioutil.WriteFile(fmt.Sprintf("%s_%dx%d.csv", ce.Info.Name, ce.PixPerDegreeLonX, ce.PixPerDegreeLatY), []byte(trakXYZ), 0644)

	wldText := fmt.Sprintf("%f\n0\n0\n%f\n%d\n%d", 1.0 / float64(ce.PixPerDegreeLonX), -1.0 / float64(ce.PixPerDegreeLatY), ce.Info.Bounds.LonXLeftDeg, ce.Info.Bounds.LatYTopDeg)

	_ = ioutil.WriteFile(fmt.Sprintf("%s_%dx%d.wld", ce.Info.Name, ce.PixPerDegreeLonX, ce.PixPerDegreeLatY), []byte(wldText), 0644)
}

func SingleCalc(stormID string, pixPerDegLatY int, pixPerDegLonX int, rMaxDefaultNmi float64, maxCalcDistNmi float64, gwaf float64){
	// stormID := "al082021" //Henri 2021
	//stormID := "al092021" //Ida 2021
	// stormID := "al122005" // katrina 2005
	//stormID := "al182012" // sandy 2012

	event := atcf.FetchAtcfEvent(stormID, rMaxDefaultNmi, gwaf)

	startTime := time.Now().UTC()
	fmt.Printf("Calc Start time: %s\n", startTime.Format(time.RFC3339))
	ce := event.CalculateEvent(pixPerDegLatY, pixPerDegLonX, maxCalcDistNmi)
	endTime := time.Now().UTC()
	fmt.Printf("Calc End time: %s, Duration: %fs\n", endTime.Format(time.RFC3339), endTime.Sub(startTime).Seconds())

	toRaster(ce)
	trackXYZ := ce.TrackToDelimited(true)

	_ = ioutil.WriteFile(fmt.Sprintf("data/tmp/%s_%d_%dx%d.csv", ce.Info.Name, ce.Info.Year, ce.PixPerDegreeLonX, ce.PixPerDegreeLatY), []byte(trackXYZ), 0644)

	wldText := fmt.Sprintf("%f\n0\n0\n%f\n%d\n%d", 1.0 / float64(ce.PixPerDegreeLonX), -1.0 / float64(ce.PixPerDegreeLatY), ce.Info.Bounds.LonXLeftDeg, ce.Info.Bounds.LatYTopDeg)

	_ = ioutil.WriteFile(fmt.Sprintf("data/tmp/%s_%d_%dx%d.wld", ce.Info.Name, ce.Info.Year, ce.PixPerDegreeLonX, ce.PixPerDegreeLatY), []byte(wldText), 0644)
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

	o, _ := os.Create(fmt.Sprintf("data/tmp/%s_%d_%dx%d.png", ce.Info.Name, ce.Info.Year, ce.PixPerDegreeLonX, ce.PixPerDegreeLatY))

	_ = png.Encode(o, raster)
}
