package main

import (
	"encoding/json"
	"fmt"
	"godin/hurricane"
	"math"
	"reflect"
	"syscall/js"
	"time"
	"unsafe"
)

// Pin buffer to global, so it doesn't get GC'd
var wasmMemoryBuffer []byte

func calculateLandfall(this js.Value, i []js.Value) interface{} {
	start := time.Now()
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

	minValue := i[13].Float()

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

	fmt.Printf("Landfall calculation done in %fs, %d points\n", time.Since(start).Seconds(), len(windfield))
	wasmMemoryBuffer = ToGeoJSONString(windfield, minValue)
	fmt.Printf("GeoJSON Marshal %fs\n", time.Since(start).Seconds())

	buffHeader := (*reflect.SliceHeader)(unsafe.Pointer(&wasmMemoryBuffer))

	retMap := map[string]interface{}{
		"ptr": buffHeader.Data,
		"len":  len(wasmMemoryBuffer),
	}

	return retMap
}

func ToGeoJSONString(c []hurricane.CoordinateValue, minValue float64) (geoBytes []byte) {
	features := make([]map[string]interface{}, 0)
	zeroValPointCoordinates := make([]interface{}, 0)

	for _, p := range c {
		if p.Value == 0 {
			zeroValPointCoordinates = append(zeroValPointCoordinates, []interface{}{p.LonXDeg, p.LatYDeg} )
		} else if minValue < p.Value {
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

	if minValue < 0 {
		features = append(features, zeroMultipoint)
	}

	gObj := map[string]interface{}{
		"type": "FeatureCollection",
		"features": features,
	}

	geoBytes, _ = json.Marshal(gObj)

	return geoBytes
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
