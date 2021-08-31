package atcf

import (
	"compress/gzip"
	"godin/hurricane"
	"godin/utilities"
	"io/ioutil"
	"math"
	"net/http"
	"strconv"
	"strings"
	"time"
)

// Data File Format: https://www.nrlmry.navy.mil/atcf_web/docs/database/new/abdeck.txt
// Index https://ftp.nhc.noaa.gov/atcf/


type UrlAtcf string

const (
	index UrlAtcf = "https://ftp.nhc.noaa.gov/atcf/"
	bdeck UrlAtcf = "https://ftp.nhc.noaa.gov/atcf/btk/b{{.}}.dat"
	bdeckArchive UrlAtcf = "https://ftp.nhc.noaa.gov/atcf/archive/{{.}}/b{{.}}.dat.gz"
	forecasts UrlAtcf = "https://ftp.nhc.noaa.gov/atcf/fst/{{.}}.fst"
)

func FetchATCFBDeckData(stormID string) (data string) {
	url := strings.Replace(string(bdeck), "{{.}}", stormID, 1)
	resp, _ := http.Get(url)

	if resp.StatusCode == http.StatusOK {
		respData, _ := ioutil.ReadAll(resp.Body)
		data = string(respData)
	} else {
		client := new(http.Client)

		archiveUrl := strings.Replace(string(bdeckArchive), "{{.}}", stormID[4:], 1)
		archiveUrl = strings.Replace(string(archiveUrl), "{{.}}", stormID, 1)
		// print(archiveUrl)

		request, _ := http.NewRequest("GET", archiveUrl, nil)
		request.Header.Add("Accept-Encoding", "gzip")

		response, _ := client.Do(request)

		reader, _ := gzip.NewReader(response.Body)

		readerData, _ := ioutil.ReadAll(reader)
		data = string(readerData)
	}
	//fmt.Printf(data)

	//AtcfParser(data, 15)

	return data
}

func FetchATCFBDeckTrack(stormID string) (track []hurricane.TrackPoint) {
	_ = FetchATCFBDeckData(stormID)

	var t []hurricane.TrackPoint
	return t
}

func FetchATCFForecastData(stormID string) (data string) {
	url := strings.Replace(string(forecasts), "{{.}}", stormID, 1)
	resp, _ := http.Get(url)

	if resp.StatusCode == http.StatusOK {
		respData, _ := ioutil.ReadAll(resp.Body)
		data = string(respData)
		//fmt.Printf(data)
	}

	//AtcfParser(data, 15)

	return data
}

func FetchATCFForecastTrack(stormID string) (track []hurricane.TrackPoint) {
	_ = FetchATCFForecastData(stormID)

	var t []hurricane.TrackPoint
	return t
}

func FetchAtcfEvent(stormID string, rMaxNmi float64, gwaf float64) hurricane.EventInformation {
	stormYear, _ := strconv.Atoi(stormID[4:])
	btrackData := FetchATCFBDeckData(stormID)
	ftrackData := FetchATCFForecastData(stormID)

	ftrackPoints := make([]atcfTrackPoint, 0)
	btrackPoints := AtcfParser(btrackData, rMaxNmi)
	if ftrackData != "" {
		ftrackPoints = AtcfParser(ftrackData, rMaxNmi)
	}

	// entireTrackPointsUnfiltered := append(btrackPoints, ftrackPoints...)

	entireTrackPointsUnfiltered := btrackPoints

	for _, tp := range ftrackPoints {
		if tp.Timestamp.After(btrackPoints[len(btrackPoints) - 1].Timestamp){
			entireTrackPointsUnfiltered =
				append(entireTrackPointsUnfiltered, tp)
		}
	}

	// filter for duplicate timestamps
	checkMap := map[time.Time]bool{}
	var entireTrackPoints []atcfTrackPoint
	for _, tp := range entireTrackPointsUnfiltered {
		if _, ok := checkMap[tp.Timestamp]; ok {
			continue
		} else {
			checkMap[tp.Timestamp] = true
			entireTrackPoints = append(entireTrackPoints, tp)
		}
	}
	
	var updatedTrack []hurricane.TrackPoint
	// forward speed: defined as speed based on previous point.  initial speed is 0

	// bbox Values
	bboxMaxLatY := -9999.9
	bboxMinLatY := 9999.9
	bboxMaxLonX := -9999.9
	bboxMinLonX := 9999.9
	bboxOffset := 5.0

	for i := 0; i < len(entireTrackPoints); i++ {
		curr := entireTrackPoints[i]

		bboxMaxLatY = math.Max(curr.LatYDeg + bboxOffset, bboxMaxLatY)
		bboxMinLatY = math.Min(curr.LatYDeg - bboxOffset, bboxMinLatY)
		bboxMaxLonX = math.Max(curr.LonXDeg + bboxOffset, bboxMaxLonX)
		bboxMinLonX = math.Min(curr.LonXDeg - bboxOffset, bboxMinLonX)

		fspeedKts := 0.0
		if i != 0 {
			prev := entireTrackPoints[i - 1]
			fspeedKts = utilities.CalculateSpeedMps(prev.LatYDeg, prev.LonXDeg, prev.Timestamp, curr.LatYDeg, curr.LonXDeg, curr.Timestamp) * 1.94384519992989  // m/s to kts
		}

		// cyclone heading: defined as heading based on next point.  final heading is the same as the heading before it
		var headingDeg float64
		if i == len(entireTrackPoints) - 1 {
			headingDeg = updatedTrack[len(updatedTrack) - 1].CycloneHeadingDeg
		} else {
			next := entireTrackPoints[i + 1]
			headingDeg = utilities.CalcBearingNorthZero(curr.LatYDeg, curr.LonXDeg, next.LatYDeg, next.LonXDeg)
		}

		tp := hurricane.TrackPoint {
			Timestamp:                    curr.Timestamp,
			TrackSequence:                float64(i),
			LatYDeg:                      curr.LatYDeg,
			LonXDeg:                      curr.LonXDeg,
			MaxWindVelocityKts:           curr.MaxWindVelocityKts,
			MinCentralPressureMb:         curr.MinCentralPressureMb,
			RadiusMaxWindNmi:             curr.RMaxNmi,
			CycloneForwardSpeedKts:       fspeedKts,
			CycloneHeadingDeg:            headingDeg,
			GradientWindAdjustmentFactor: gwaf,
			Source:                       curr.Source,
		}

		updatedTrack = append(updatedTrack, tp)
	}

	// Time interpolate Track
	var interpolatedTrack []hurricane.TrackPoint
	interpolationStep := 30 * time.Minute
	for i := 0; i < len(updatedTrack) - 1; i++ {
		curr := updatedTrack[i]
		interpolatedTrack = append(interpolatedTrack, curr)
		if i == len(updatedTrack){
			continue
		}

		next := updatedTrack[i + 1]

		iSeq := curr.TrackSequence
		for t := curr.Timestamp.Add(interpolationStep); t.Before(next.Timestamp); t = t.Add(interpolationStep) {
			iLast := interpolatedTrack[len(interpolatedTrack) - 1]
			iLatYDeg := utilities.LinearInterpolation(float64(t.Unix()), float64(curr.Timestamp.Unix()), float64(next.Timestamp.Unix()), curr.LatYDeg, next.LatYDeg)
			iLonXDeg := utilities.LinearInterpolation(float64(t.Unix()), float64(curr.Timestamp.Unix()), float64(next.Timestamp.Unix()), curr.LonXDeg, next.LonXDeg)
			iBearingDeg := utilities.CalcBearingNorthZero(iLast.LatYDeg, iLast.LonXDeg, iLatYDeg, iLonXDeg)
			iSeq += 0.1
			itp := hurricane.TrackPoint{
				Timestamp:                    t,
				TrackSequence:                iSeq,
				LatYDeg:                      iLatYDeg,
				LonXDeg:                      iLonXDeg,
				MaxWindVelocityKts:           utilities.LinearInterpolation(float64(t.Unix()), float64(curr.Timestamp.Unix()), float64(next.Timestamp.Unix()), curr.MaxWindVelocityKts, next.MaxWindVelocityKts),
				MinCentralPressureMb:         utilities.LinearInterpolation(float64(t.Unix()), float64(curr.Timestamp.Unix()), float64(next.Timestamp.Unix()), curr.MinCentralPressureMb, next.MinCentralPressureMb),
				RadiusMaxWindNmi:             utilities.LinearInterpolation(float64(t.Unix()), float64(curr.Timestamp.Unix()), float64(next.Timestamp.Unix()), curr.RadiusMaxWindNmi, next.RadiusMaxWindNmi),
				CycloneForwardSpeedKts:       utilities.LinearInterpolation(float64(t.Unix()), float64(curr.Timestamp.Unix()), float64(next.Timestamp.Unix()), curr.CycloneForwardSpeedKts, next.CycloneForwardSpeedKts),
				CycloneHeadingDeg:            iBearingDeg,
				GradientWindAdjustmentFactor: utilities.LinearInterpolation(float64(t.Unix()), float64(curr.Timestamp.Unix()), float64(next.Timestamp.Unix()), curr.GradientWindAdjustmentFactor, next.GradientWindAdjustmentFactor),
				Source:                       "INTRP",
			}
			interpolatedTrack = append(interpolatedTrack, itp)
		}
	}

	interpolatedTrack = append(interpolatedTrack, updatedTrack[len(updatedTrack) - 1])

	stormName := stormID
	for i := len(btrackPoints) - 1; 0 < i; i-- {
		if btrackPoints[i].Name != "" {
			stormName = btrackPoints[i].Name
			break
		}
	}

	event := hurricane.EventInformation{
		ID:     stormID,
		Name:   stormName,
		Year: stormYear,
		Track:  interpolatedTrack,
		Bounds: hurricane.BoundingBox{
			LatYTopDeg:    int(math.Round(bboxMaxLatY)),
			LonXLeftDeg:   int(math.Round(bboxMinLonX)),
			LatYBottomDeg: int(math.Round(bboxMinLatY)),
			LonXRightDeg:  int(math.Round(bboxMaxLonX)),
		},
	}

	return event
}

func AtcfParser(data string, rMaxDefault float64) (tps []atcfTrackPoint){
	// All trimming will happen on demand, but we get rid of a trailing newline
	dataRows := strings.Split(strings.TrimSpace(data), "\n")

	for _, row := range dataRows {
		parsedRow := atcfRowParser(row, rMaxDefault)
		//fmt.Printf("%v\n", parsedRow)
		tps = append(tps, parsedRow)
	}

	return tps
}

func atcfRowParser(row string, rMaxDefault float64) (trackPoint atcfTrackPoint) {

	values := strings.Split(row, ",")

	// basin := strings.TrimSpace(values[0])
	// cycloneNumber, _ := strconv.Atoi(strings.TrimSpace(values[1]))

	// get source early, to determine whether to store minutes or not
	techVal := strings.TrimSpace(values[4])
	source := hurricane.Unknown
	if techVal == string(hurricane.Best) {
		source = hurricane.Best
	} else if techVal == string(hurricane.Forecasted) {
		source = hurricane.Forecasted
	}

	year, _ := strconv.Atoi(strings.TrimSpace(values[2])[:4])
	month, _ := strconv.Atoi(strings.TrimSpace(values[2])[4:6])
	day, _ := strconv.Atoi(strings.TrimSpace(values[2])[6:8])
	hour, _ := strconv.Atoi(strings.TrimSpace(values[2])[8:])

	//minuteVal := strings.TrimSpace(values[3])
	minute := 0
	//if minuteVal != "" && source == hurricane.Best {
	//	minute, _ = strconv.Atoi(minuteVal)
	//}

	timestamp := time.Date(year, time.Month(month), day, hour, minute, 0, 0, time.UTC)

	tau, tauErr := strconv.Atoi(strings.TrimSpace(values[5]))
	if tauErr == nil && 0 < tau {
		timestamp = timestamp.Add(time.Duration(tau) * time.Hour)
	}

	latVal := strings.TrimSpace(values[6])
	latYDegDec, _ := strconv.Atoi(latVal[0:len(latVal) - 1])
	latYDeg := float64(latYDegDec) / 10.0
	latNS := latVal[len(latVal) - 1:]
	if latNS == "S" {
		latYDeg = -latYDeg
	}

	lonVal := strings.TrimSpace(values[7])
	lonXDegDec, _ := strconv.Atoi(lonVal[0:len(lonVal) - 1])
	lonXDeg := float64(lonXDegDec) / 10.0
	lonEW := lonVal[len(lonVal) - 1:]
	if lonEW == "W" {
		lonXDeg = -lonXDeg
	}

	vmaxKts, _ := strconv.ParseFloat(strings.TrimSpace(values[8]), 64)
	mslpMb, _ := strconv.ParseFloat(strings.TrimSpace(values[9]), 64)
	if mslpMb < 600 {
		mslpMb = -1
	}

	rMax := rMaxDefault
	if 19 < len(values) {
		rMaxValue, rMaxErr := strconv.ParseFloat(strings.TrimSpace(values[19]), 64)
		if rMaxErr == nil && 0.0 < rMaxValue {
			rMax = rMaxValue
		}
	}

	name := ""
	if 27 < len(values) {
		name = strings.TrimSpace(values[27])
	}

	atp := atcfTrackPoint {
		Timestamp:                    timestamp,
		LatYDeg:                      latYDeg,
		LonXDeg:                      lonXDeg,
		MaxWindVelocityKts:           vmaxKts,
		MinCentralPressureMb:         mslpMb,
		Source:                       source,
		Name:						  name,
		RMaxNmi:					  rMax,
	}

	return atp
}

type atcfTrackPoint struct {
	Timestamp time.Time
	LatYDeg float64
	LonXDeg float64
	MaxWindVelocityKts float64
	MinCentralPressureMb float64
	Source hurricane.TrackPointSource
	Name string
	RMaxNmi float64
}
