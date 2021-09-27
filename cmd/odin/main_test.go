package main

import "testing"

func TestSingleCalc(t *testing.T){
	stormID := "al092021" //Ida 2021

	pixPerDegree := 10
	rMaxDefaultNmi := 15.0
	maxCalcDistNmi := 360.0
	gwaf := 0.9
	includeForecasts := false

	SingleCalc(stormID, pixPerDegree, pixPerDegree, rMaxDefaultNmi, maxCalcDistNmi, gwaf, includeForecasts)
}
