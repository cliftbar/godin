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

//const BUFF_SIZE = 96708022
var buffer []byte
//var bigBuf [BUFF_SIZE]byte
var ptrs = map[string]interface{}{}


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

	//outputPtr := i[13].JSValue()

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
	ToGeoJSONString(windfield, minValue)
	fmt.Printf("Marshal %fs\n", time.Since(start).Seconds())
	//ptr := unsafe.Pointer(&buffer)
	//stringHeader := &reflect.StringHeader{
	//	Data: uintptr(ptr),
	//	Len:  len(buffer),
	//}
	//uintptr(outputPtr)
	//jsonString, _ := json.MarshalIndent(converted, "", "\t")
	//fmt.Println(string(jsonString))
	//ret := js.ValueOf(converted)
	//dst := js.Global().Get("Uint8Array").New(len(converted))
	//js.Global().Set("output", converted)
	//js.CopyBytesToJS(dst, []byte(converted))
	//buffer = []uint8(converted)

	//var testBuf [955973]byte

	//copy(testBuf[:], buffer)
	ptrs["testBuff"] = buffer

	buffHeader := (*reflect.SliceHeader)(unsafe.Pointer(&buffer))

	println(buffHeader.Data)
	println(len(buffer))

	retMap := map[string]interface{}{
		"ptr": buffHeader.Data,
		"len":  len(buffer),
	}

	return retMap //unsafe.Pointer(stringHeader)
}

func InitializeWasmMemory(this js.Value, args []js.Value) interface{} {

	var ptr *[]uint8
	goArrayLen := 10000

	goArray := make([]uint8, goArrayLen)
	ptr = &goArray

	boxedPtr := unsafe.Pointer(ptr)
	boxedPtrMap := map[string]interface{}{
		"internalptr": boxedPtr,
	}
	return js.ValueOf(boxedPtrMap)
}

func ToGeoJSONString(c []hurricane.CoordinateValue, minValue float64)  {
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
	buffer, _ = json.Marshal(gObj)
	//println(len(temp))
	//copy(bigBuf[:], temp)
	fmt.Printf("buffer:\t %v \n", buffer)
	//g = string(gBytes)

	//g["coordinates"] = coords
	//return gBytes
}

func registerCallbacks() {
	js.Global().Set("calculateLandfall", js.FuncOf(calculateLandfall))
	js.Global().Set("getWasmMemoryBufferPointer", js.FuncOf(InitializeWasmMemory))
}

func main(){
	c := make(chan struct{}, 0)

	println("WASM Go Initialized")
	// register functions
	registerCallbacks()
	<-c
}

