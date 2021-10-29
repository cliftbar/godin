package hurricane

import (
	"fmt"
	"godin/utilities"
	"math"
	"time"
)

type TrackPointSource string
const (
	Unknown    TrackPointSource = "unknown"
	Best       TrackPointSource = "BEST"
	Forecasted TrackPointSource = "OFCL"
)

type TrackPoint struct {
	Timestamp     time.Time `json:"timestamp"`
	TrackSequence float64 `json:"track_sequence"`

	LatYDeg float64 `json:"lat_y_deg"`
	LonXDeg float64 `json:"lon_x_deg"`

	MaxWindVelocityKts   float64 `json:"max_wind_velocity_kts"`
	MinCentralPressureMb float64 `json:"min_central_pressure_mb"`
	RadiusMaxWindNmi     float64 `json:"radius_max_wind_nmi"`

	CycloneForwardSpeedKts float64 `json:"cyclone_forward_speed_kts"`
	CycloneHeadingDeg      float64 `json:"cyclone_heading_deg"`

	GradientWindAdjustmentFactor float64 `json:"gradient_wind_adjustment_factor"`

	Source TrackPointSource `json:"source"`
}

type BoundingBox struct {
	// Top Left Point
	LatYTopDeg  int `json:"lat_y_top_deg"`
	LonXLeftDeg int `json:"lon_x_left_deg"`

	// Bottom Right
	LatYBottomDeg int `json:"lat_y_bottom_deg"`
	LonXRightDeg  int `json:"lon_x_right_deg"`
}

func (bb BoundingBox) GetBlockHeight(blocksPerDegLatY int) int {
	delta := bb.LatYTopDeg - bb.LatYBottomDeg
	if delta < 0 {delta = -delta}
	return delta * blocksPerDegLatY
}

func (bb BoundingBox) GetBlockWidth(blocksPerDegLonX int) int {
	delta := bb.LonXRightDeg - bb.LonXLeftDeg
	if delta < 0 {delta = -delta}
	return delta * blocksPerDegLonX
}

func (bb BoundingBox) toPoints(blocksPerDegLatY int, blocksPerDegLonX int) []Coordinate {
	precision := 100000
	deltaY := int(math.Abs(float64(bb.LatYTopDeg - bb.LatYBottomDeg))) * precision
	deltaX := int(math.Abs(float64(bb.LonXRightDeg - bb.LonXLeftDeg))) * precision

	blocksY := deltaY * blocksPerDegLatY / precision
	blocksX := deltaX * blocksPerDegLonX / precision

	stepY := int(deltaY / blocksY)
	stepX := int(deltaX / blocksX)

	var points []Coordinate

	lat := int(bb.LatYBottomDeg) * precision
	for y := 0; y < blocksY; y++ {
		lon := int(bb.LonXLeftDeg) * precision
		for x := 0; x < blocksX; x++ {
			points = append(points, Coordinate{float64(lat) / float64(precision), float64(lon) / float64(precision)})
			lon += stepX
		}
		lat += stepY
	}

	return points
}

type EventInformation struct {
	ID string `json:"id"`
	Name string `json:"name"`
	Year int `json:"year"`

	Track []TrackPoint `json:"track"`

	Bounds BoundingBox `json:"bounds"`
}

type Coordinate struct {
	latYDeg float64
	lonXDeg float64
}

type CoordinateValue struct {
	LatYDeg float64 `json:"lat_y_deg"`
	LonXDeg float64 `json:"lon_x_deg"`
	Value   float64 `json:"value"`
}

type CalculatedEvent struct {
	Info                      EventInformation  `json:"info"`
	WindField                 []CoordinateValue `json:"wind_field"`
	MaxCalculationDistanceNmi float64           `json:"max_calculation_distance_nmi"`
	PixPerDegreeLatY          int               `json:"pix_per_degree_lat_y"`
	PixPerDegreeLonX          int               `json:"pix_per_degree_lon_x"`
}


func CalculateEventFrame(bbox BoundingBox,
						 latYDeg float64,
						 lonXDeg float64,
						 maxWindVelocityKts float64,
						 radiusMaxWindNmi float64,
						 cycloneForwardSpeedKts float64,
						 cycloneHeadingDeg float64,
						 gradientWindAdjustmentFactor float64,
						 pixPerDegLatY int,
						 pixPerDegLonX int,
						 maxCalculationDistanceNmi float64) (windField []CoordinateValue) {
	gridPoints := bbox.toPoints(pixPerDegLatY, pixPerDegLonX)

	maxDistDegApproxDeg := maxCalculationDistanceNmi / 60 // convert nmi to degrees
	maxDistDegApproxDegSq := maxDistDegApproxDeg * maxDistDegApproxDeg

	// PERF use pointers and basic iterators instead of range and assignment to dodge duffcopy
	for i := 0; i < len(gridPoints); i++ {
		var c = &gridPoints[i]
		maxWindSpeedAtCoordinate := 0.0

		// PERF use a less accurate, simpler max distance check - this is huge
		checkDistDegSq := utilities.FastDistanceDegSq(latYDeg, lonXDeg, c.latYDeg, c.lonXDeg) // get the square of the distance to check against
		if checkDistDegSq < maxDistDegApproxDegSq {

			distanceToCenterNmi := utilities.HaversineDegreesToMeters(latYDeg, lonXDeg, c.latYDeg, c.lonXDeg) / 1000.0 * 0.539957 // convert to nautical miles

			bearingFromCenter := utilities.CalcBearingNorthZero(latYDeg, lonXDeg, c.latYDeg, c.lonXDeg)

			windSpeed, _ := CalcWindSpeedFrame(maxWindVelocityKts, distanceToCenterNmi, radiusMaxWindNmi, cycloneForwardSpeedKts, bearingFromCenter, cycloneHeadingDeg, gradientWindAdjustmentFactor)
			maxWindSpeedAtCoordinate = math.Max(maxWindSpeedAtCoordinate, windSpeed)
		}

		windField = append(windField, CoordinateValue{
			LatYDeg: c.latYDeg,
			LonXDeg: c.lonXDeg,
			Value:   maxWindSpeedAtCoordinate,
		})
	}

	return windField
}

func (ei EventInformation) CalculateEvent(pixPerDegLatY int, pixPerDegLonX int, maxCalculationDistanceNmi float64) (event CalculatedEvent) {
	var windField []CoordinateValue

	gridPoints := ei.Bounds.toPoints(pixPerDegLatY, pixPerDegLonX)

	maxDistDegApproxDeg := maxCalculationDistanceNmi / 60 // convert nmi to degrees
	maxDistDegApproxDegSq := maxDistDegApproxDeg * maxDistDegApproxDeg

	// PERF use pointers and basic iterators instead of range and assignment to dodge duffcopy
	for i := 0; i < len(gridPoints); i++ {
		var c = &gridPoints[i]
		maxWindSpeedAtCoordinate := 0.0
		for tpj := 0; tpj < len(ei.Track); tpj++ {
			var tp = &ei.Track[tpj]
			// PERF use a less accurate, simpler max distance check - this is huge
			checkDistDegSq := utilities.FastDistanceDegSq(tp.LatYDeg, tp.LonXDeg, c.latYDeg, c.lonXDeg) // get the square of the distance to check against
			if checkDistDegSq < maxDistDegApproxDegSq {

				distanceToCenterNmi := utilities.HaversineDegreesToMeters(tp.LatYDeg, tp.LonXDeg, c.latYDeg, c.lonXDeg) / 1000.0 * 0.539957 // convert to nautical miles

				bearingFromCenter := utilities.CalcBearingNorthZero(tp.LatYDeg, tp.LonXDeg, c.latYDeg, c.lonXDeg)

				windSpeed, _ := CalcWindSpeed(tp.MaxWindVelocityKts, distanceToCenterNmi, tp.RadiusMaxWindNmi, tp.CycloneForwardSpeedKts, bearingFromCenter, tp.CycloneHeadingDeg, tp.GradientWindAdjustmentFactor)
				maxWindSpeedAtCoordinate = math.Max(maxWindSpeedAtCoordinate, windSpeed)
			}
		}

		windField = append(windField, CoordinateValue{
			LatYDeg: c.latYDeg,
			LonXDeg: c.lonXDeg,
			Value:   maxWindSpeedAtCoordinate,
		})
	}

	event = CalculatedEvent{
		Info:                      ei,
		WindField:                 windField,
		MaxCalculationDistanceNmi: maxCalculationDistanceNmi,
		PixPerDegreeLatY:          pixPerDegLatY,
		PixPerDegreeLonX:          pixPerDegLonX,
	}

	return event
}

func (ce CalculatedEvent) TrackToDelimited(header bool) string {
	outString := ""
	if header {
		outString = "ts, lonX, latY, maxWindKts, headingDeg, rMax, source, fSpeedKts\n"
	}
	for _, row := range ce.Info.Track {
		rowString := fmt.Sprintf("%s, %f, %f, %f, %f, %f, %s, %f\n", row.Timestamp.Format(time.RFC3339), row.LonXDeg, row.LatYDeg, row.MaxWindVelocityKts, row.CycloneHeadingDeg, row.RadiusMaxWindNmi, row.Source, row.CycloneForwardSpeedKts)
		outString = outString + rowString
	}
	return outString
}
