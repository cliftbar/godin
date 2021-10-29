package main

import (
	"encoding/json"
	"fmt"
	"godin/hurricane"
	"math"
	"syscall/js"
)


func calculateLandfall(this js.Value, i []js.Value) interface{} {
	if len(i) < 13 {
		fmt.Println("Not enough arguments")
	} else {
		fmt.Println("Calculate Landfall")
	}

	topLatYDeg := i[0].Int()
	botLatYDeg := i[1].Int()
	leftLonXDeg := i[2].Int()
	rightLonXDeg := i[3].Int()

	positionLatYDeg := i[4].Float()
	positionLonXDeg := i[5].Float()

	maxWindSpeedKts := i[6].Float()
	rMaxNmi := i[7].Float()
	fSpeedKts := i[8].Float()
	headingDeg := i[9].Float()
	gwaf := i[10].Float()

	pixPerDegree := i[11].Int()
	maxCalcDistNmi := i[12].Float()

	bbox := hurricane.BoundingBox{
		LatYTopDeg:    topLatYDeg,
		LonXLeftDeg:   leftLonXDeg,
		LatYBottomDeg: botLatYDeg,
		LonXRightDeg:  rightLonXDeg,
	}

	windfield := hurricane.CalculateEventFrame(
		bbox,
		positionLatYDeg,
		positionLonXDeg,
		maxWindSpeedKts,
		rMaxNmi,
		fSpeedKts,
		headingDeg,
		gwaf,
		pixPerDegree,
		pixPerDegree,
		maxCalcDistNmi,
	)

	fmt.Println("Landfall calculation done")
	converted := ToGeoJSONString(windfield)
	//jsonString, _ := json.MarshalIndent(converted, "", "\t")
	//fmt.Println(string(jsonString))
	//ret := js.ValueOf(converted)
	return converted
}

func ToGeoJSONString(c []hurricane.CoordinateValue) (g string) {
	features := make([]map[string]interface{}, 0)
	zeroValPointCoordinates := make([]interface{}, 0)

	for _, p := range c {
		if p.Value == 0 {
			zeroValPointCoordinates = append(zeroValPointCoordinates, []interface{}{p.LonXDeg, p.LatYDeg} )
		} else {
			f := map[string]interface{}{
				"type": "Feature",
				"geometry": map[string]interface{}{
					"type": "Point",
					"coordinates": []interface{}{p.LonXDeg, p.LatYDeg},
				},
				"properties": map[string]interface{}{
					"zValue": int(math.Round(p.Value)),
				},
			}
			features = append(features, f)
		}

	}

	zeroMultipoint := map[string]interface{}{
		"type": "Feature",
		"geometry": map[string]interface{}{
			"type": "MultiPoint",
			"coordinates": zeroValPointCoordinates,
		},
		"properties": map[string]interface{}{
			"zValue": 0,
		},
	}

	features = append(features, zeroMultipoint)

	gObj := map[string]interface{}{
		"type": "FeatureCollection",
		"features": features,
	}
	gBytes, _ := json.Marshal(gObj)
	g = string(gBytes)

	//g["coordinates"] = coords
	return g
}

func registerCallbacks() {
	js.Global().Set("calculateLandfall", js.FuncOf(calculateLandfall))
}

func main(){
	c := make(chan struct{}, 0)

	println("WASM Go Initialized")
	// register functions
	registerCallbacks()
	<-c
}

