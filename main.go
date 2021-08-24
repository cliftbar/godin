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
	stormID := "al082021"
	atcf.FetchATCFBDeckTrack(stormID)
	atcf.FetchATCFForecastTrack(stormID)
	event := atcf.FetchAtcfEvent(stormID, 15, 0.9)

	ce := event.CalculateEvent(10, 10, 360)

	toRaster(ce)
	trakXYZ := ce.TrackToXYZ(true)

	_ = ioutil.WriteFile(stormID + ".csv", []byte(trakXYZ), 0644)
	fmt.Println(ce.Info.Bounds)
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
	println(ce.Info.Name)

	width := ce.Info.Bounds.GetBlockWidth(ce.PixPerDegreeLonX)
	height := ce.Info.Bounds.GetBlockHeight(ce.PixPerDegreeLatY)
	raster := image.NewGray(image.Rectangle{
		Min: image.Point{X: 0, Y: 0},
		Max: image.Point{X: width, Y: height},
	})
	i := 0
	for y := height; y > 0; y-- {
		for x := 0; x < width; x++ {
		//for x := width; 0 < x; x-- {

			val := uint8(ce.WindField[i].Value)
			raster.SetGray(x, y, color.Gray{Y: val})
			i++
		}
	}

	o, _ := os.Create(ce.Info.Name + ".png")

	_ = png.Encode(o, raster)
}
