package utilities

import (
	"math"
	"time"
)

const ToRadians float64 = math.Pi / 180.0
const ToDegrees float64 =  180.0 / math.Pi

func LinearInterpolation(x float64, x1 float64, x2 float64, y1 float64, y2 float64) float64 {
	return (((y2 - y1) / (x2 - x1)) * (x - x1)) + y1
}

// CalcBearingNorthZero Calculate the simple bearing (pythagorean angles) from reference to location.
func CalcBearingNorthZero(latRefDeg float64,
						  lonRefDeg float64,
						  latLocDeg float64,
						  lonLocDeg float64) (bearingDeg float64) {
	lonDelta := lonLocDeg - lonRefDeg
	latDelta := latLocDeg - latRefDeg

	angleDeg := math.Atan2(lonDelta, latDelta) * ToDegrees
	// PERF integer math instrad of float
	//return float64(int(angleDeg+360.0) % 360)
	return math.Mod(angleDeg + 360.0, 360.0)
}


// HaversineDegreesToMeters Haversine equation for finding the distance between two lat-lon points in meters.
// reference: http://www.movable-type.co.uk/scripts/latlong.html, http://stackoverflow.com/questions/4102520/how-to-transform-a-distance-from-degrees-to-metres
func HaversineDegreesToMeters(latYDegRef float64, lonXDegRef float64, latYDeg float64, lonXDeg float64) (distanceM float64){
	//r := 6371000.0
	deltaLat := (latYDeg - latYDegRef) * ToRadians
	deltaLon := (lonXDeg - lonXDegRef) * ToRadians

	dLatSin := math.Sin(deltaLat / 2)
	dLonSin := math.Sin(deltaLon / 2)

	a := (dLatSin * dLatSin) +
		(math.Cos(latYDegRef * ToRadians) * math.Cos(latYDeg * ToRadians) *
			 (dLonSin * dLonSin))
	//c := 2.0 * math.Atan2(math.Sqrt(a), math.Sqrt(1 - a))
	//return r * c
	// PERF - precalculate constants
	return 12742000 * math.Atan2(math.Sqrt(a), math.Sqrt(1 - a))
}

func CalculateSpeedMps(latYStart float64, lonXStart float64, tsStart time.Time, latYEnd float64, lonXEnd float64, tsEnd time.Time) (speedKts float64) {
	distanceMeters := HaversineDegreesToMeters(latYStart, lonXStart, latYEnd, lonXEnd)
	durationSec := tsEnd.Sub(tsStart).Seconds()

	return distanceMeters / durationSec
}

func FastDistanceDegSq(latYDegRef float64, lonXDegRef float64, latYDeg float64, lonXDeg float64) (distanceDegSq float64) {
	y := latYDeg - latYDegRef
	x := lonXDeg - lonXDegRef

	return  (x * x) + (y * y)

}

// public static double pow(final double a, final double b) {
//		final long tmp = Double.doubleToLongBits(a);
//		final long tmp2 = (long)(b * (tmp - 4606921280493453312L)) + 4606921280493453312L;
//		return Double.longBitsToDouble(tmp2);
// }
// https://martin.ankerl.com/2007/10/04/optimized-pow-approximation-for-java-and-c-c/

func FastPow(a float64, b float64) float64 {
	tmp := math.Float64bits(a)
	tmp2 := uint64(b * float64(tmp - 4606921280493453312)) + 4606921280493453312
	return math.Float64frombits(tmp2)
}