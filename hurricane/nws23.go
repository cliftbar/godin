package hurricane

import (
	"godin/utilities"
	"math"
)

/* Constants
SPH: Standard Project Hurricane
PMH: Probable Maximum Hurricane
Pressure profile equation
(p-Cp)/(Pw - Cp) = e ^ (-R/r)
Pw: Peripheral Pressure, pressure at edge of storm, should be a bit below MSLP
Cp: Central Pressure (P0 in paper)
Rmax: Radius of Maximum Winds (R in paper)
Fspeed: Forward speed of hurricane center (T in paper)
Dir: Track direction
Vgx: Maximum Gradient Winds
Rho0: Surface air density
r: distance (radius) from hurricane center
fcorr: Coriolis parameter, dependent on latitude
Vx: Observed maximum 10-m, 10-min winds over open water.  75% to 105% of Vgx.  Standard is 95%
For moving hurricane: Vx = 0.95 * Vgx + (1.5 * T ^ 0.63 * To ^ 0.37 * cos(beta)
Vpt: 10m, 10min winds at a point (V in paper)
*/

const PwSphKpa float64 = 100.8
const PwPmhKpa float64 = 102.0
const PwSphInHg float64 = 29.77
const PwPmhInHg float64 = 30.12
const Rho0Kpa float64 = 101.325 // Mean Sea Level Pressure
const KmToNmi float64 = 0.539957
const MpsToKts float64 = 1.94384
const KpaToInHg float64 = 0.2953
const MbToInHg float64 = 0.02953

// Lookup table for Rmax calculations in Inflow Method
var rMaxInflowFactor = map[int]float64{
	1: math.Pow(1.0, -1.416),
	2: math.Pow(2.0, -1.416),
	3: math.Pow(3.0, -1.416),
	4: math.Pow(4.0, -1.416),
	5: math.Pow(5.0, -1.416),
	6: math.Pow(6.0, -1.416),
	7: math.Pow(7.0, -1.416),
	8: math.Pow(8.0, -1.416),
	9: math.Pow(9.0, -1.416),
	10: math.Pow(10.0, -1.416),
	11: math.Pow(11.0, -1.416),
	12: math.Pow(12.0, -1.416),
	13: math.Pow(13.0, -1.416),
	14: math.Pow(14.0, -1.416),
	15: math.Pow(15.0, -1.416),
	16: math.Pow(16.0, -1.416),
	17: math.Pow(17.0, -1.416),
	18: math.Pow(18.0, -1.416),
	19: math.Pow(19.0, -1.416),
	20: math.Pow(20.0, -1.416),
	30: math.Pow(30.0, -1.416),
	31: math.Pow(31.0, -1.416),
	32: math.Pow(32.0, -1.416),
	33: math.Pow(33.0, -1.416),
	34: math.Pow(34.0, -1.416),
	35: math.Pow(35.0, -1.416),
	36: math.Pow(36.0, -1.416),
	37: math.Pow(37.0, -1.416),
	38: math.Pow(38.0, -1.416),
	39: math.Pow(39.0, -1.416),
	40: math.Pow(40.0, -1.416),
	41: math.Pow(41.0, -1.416),
	42: math.Pow(42.0, -1.416),
	43: math.Pow(43.0, -1.416),
	44: math.Pow(44.0, -1.416),
	45: math.Pow(45.0, -1.416),
	46: math.Pow(46.0, -1.416),
	47: math.Pow(47.0, -1.416),
	48: math.Pow(48.0, -1.416),
	49: math.Pow(49.0, -1.416),
	50: math.Pow(50.0, -1.416),
	51: math.Pow(51.0, -1.416),
	52: math.Pow(52.0, -1.416),
	53: math.Pow(53.0, -1.416),
	54: math.Pow(54.0, -1.416),
	55: math.Pow(55.0, -1.416),
	56: math.Pow(56.0, -1.416),
	57: math.Pow(57.0, -1.416),
	58: math.Pow(58.0, -1.416),
	59: math.Pow(59.0, -1.416),
	60: math.Pow(60.0, -1.416),
	61: math.Pow(61.0, -1.416),
	62: math.Pow(62.0, -1.416),
	63: math.Pow(63.0, -1.416),
	64: math.Pow(64.0, -1.416),
	65: math.Pow(65.0, -1.416),
	66: math.Pow(66.0, -1.416),
	67: math.Pow(67.0, -1.416),
	68: math.Pow(68.0, -1.416),
	69: math.Pow(69.0, -1.416),
	70: math.Pow(70.0, -1.416),
	71: math.Pow(71.0, -1.416),
	72: math.Pow(72.0, -1.416),
	73: math.Pow(73.0, -1.416),
	74: math.Pow(74.0, -1.416),
	75: math.Pow(75.0, -1.416),
	76: math.Pow(76.0, -1.416),
	77: math.Pow(77.0, -1.416),
	78: math.Pow(78.0, -1.416),
	79: math.Pow(79.0, -1.416),
	80: math.Pow(80.0, -1.416),
}

// TODO: confirm that there are two multiplications times 2.0 here
const wCoriolisConstant float64 = 2.0 * 2.0 * math.Pi / 24

// radialDecay Calculates the radial decay factor for a given radius, between 0.0 and 1.0.
// When rMaxNmi < rNmi: NWS 23 pdf page 53, page 27, Figure 2.12, empirical fit.
// When rMaxNmi > rNmi: NWS 23 pdf page 54, page 28, Figure 2.13, empirical fit (logistic regression).
//
// rNmi: Point radius from center of storm in nautical miles
//
// rMaxNmi Radius of maximum winds in nautical miles
//
// return 0 <= radial decay <= 1
func radialDecay(rNmi float64, rMaxNmi float64) (radialDecayFactor float64) {
	ret := 1.0

	if rMaxNmi < rNmi {
		// NWS 23 pdf page 53
		rMaxLog := math.Log(rMaxNmi)
		slope := (-0.051 * rMaxLog) - 0.1757
		intercept := (0.4244 * rMaxLog) + 0.7586
		ret = (slope * math.Log(rNmi)) + intercept
	}
	// Skip this else block as a concession for modeling time series, where everything within the max wind radius is
	//	expected to experience the max wind radius while the storm translates
	// else {

	// NWS 23 pdf page 54
	// ret = 1.01231578 / (1 + math.exp(-8.612066494 * ((r_nmi / float(rmax_nmi)) - 0.678031222)))
	// }

	// clamp radial decay between 0 and 1
	return math.Max(math.Min(ret, 1.0), 0.0)
}

// coriolisFrequency calculates the coriolis factor for a given latitude.
//
// latDeg: latitude in degrees
//
// return coriolis factor in hr^-1
func coriolisFrequency(latDeg float64) (coriolisFreq float64) {
	return wCoriolisConstant * math.Sin(latDeg * utilities.ToRadians)
}

// NWS 23 pdf page 50, page 24, figure 2.10, empirical relationship (linear regression)
// This is for the PMH, We can also improve this relationship
// This is what I thought, but apparently not: (1.0/(Rho0Kpa * math.e)) ** (0.5)
// DEP: lat 24, K 68.1; lat 45, K 65
// SPH: (65-68.1)/(45-24) = -0.147619048
// PMH: (66.2 - 70.1)/(45 - 24) = -0.185714286
func kDensityCoefficient(latDeg float64) (densityCoef float64) {
	return 69.1952184 / (1 + math.Exp(0.20252*(latDeg - 58.72458)))
}

// NWS 23 pdf page 49, page 23, equation 2.2
// gradient wind is the primary output, densityCoefficient and coriolisFreq are for debugging output
func gradientWindAtRadius(pwInHg float64,
	cpInHg float64,
	rNmi float64,
	latDeg float64) (gradientWind float64, densityCoefficient float64, coriolisFreq float64) {

	k := kDensityCoefficient(latDeg)
	f := coriolisFrequency(latDeg)

	gradientWind = (k * math.Pow(pwInHg - cpInHg, 0.5)) - ((rNmi * f) / 2)
	return gradientWind, k, f
}

// Empirical inflow angle calculation of PMH
// NWS 23 pdf page 55
// NOAA_NWS23_Inflow_Calc.xlsx
func inflowAngle(rNmi float64, rMaxNmi float64) (phi float64) {

	rPhiMax := (3.0688 * rMaxNmi) - 2.7151

	if rNmi < rPhiMax {
		// PERF lookup table for rMaxNmi up to 60
		//a := 11.438 * math.Pow(rMaxNmi, -1.416)
		//a := 11.438 * rMaxInflowFactor[int(rMaxNmi)] // Not Good! RMW goes rather higher than 40 for TDs
		a := 11.438 * utilities.FastPow(rMaxNmi, -1.416) // error: +/- 10%, 4 min faster
		//fmt.Printf("rMaxNmi: %f, a: %f, FastPow: %f, error: %f\n", rMaxNmi, a, fastA, (fastA - a) / a * 100)

		b := (1.1453 * rMaxNmi) + 1.4536
		phiMax := 9.7043566358 * math.Log(rMaxNmi) - 2.7295806727
		phi = phiMax / (1 + math.Exp(-1 * a * (rNmi - b)))
	} else {
		rNmiUse := math.Min(rNmi, 130)

		// PERF precalculate rMax Squared
		rMaxNmiSq := rMaxNmi * rMaxNmi
		x1 := (0.0000896902 * rMaxNmiSq) - (0.0036924418 * rMaxNmi) + 0.0072307906
		x2 := (0.000002966 * rMaxNmiSq) - (0.000090532 * rMaxNmi) - 0.0010373287
		x3 := (-0.0000000592 * rMaxNmiSq) + (0.0000019826 * rMaxNmi) - 0.0000020198
		c := (9.7043566341 * math.Log(rMaxNmi)) - 2.7295806689

		phiPrime := rNmiUse-rPhiMax
		phi = (x3 * phiPrime * phiPrime * phiPrime) + (x2 * phiPrime * phiPrime) + (x1 * phiPrime) + c

		if 130 < rNmi && rNmi < 360 { // justification on NWS23 pdf page 287 page 263
			deltaPhi := utilities.LinearInterpolation(rNmi, 130, 360, phi, phi-2)
			phi += deltaPhi
		} else if 360 <= rNmi {
			phi -= 2
		}
	}
	return phi
}

// PERF precalculate tau zero and fspeed calculations
var tauZeroFactor = math.Pow(1.0, 0.37) * 1.5
var fSpeedFactor = map[int]float64{
	 1: math.Pow( 1.0, 0.63),
	 2: math.Pow( 2.0, 0.63),
	 3: math.Pow( 3.0, 0.63),
	 4: math.Pow( 4.0, 0.63),
	 5: math.Pow( 5.0, 0.63),
	 6: math.Pow( 6.0, 0.63),
	 7: math.Pow( 7.0, 0.63),
	 8: math.Pow( 8.0, 0.63),
	 9: math.Pow( 9.0, 0.63),
	10: math.Pow(10.0, 0.63),
	11: math.Pow(11.0, 0.63),
	12: math.Pow(12.0, 0.63),
	13: math.Pow(13.0, 0.63),
	14: math.Pow(14.0, 0.63),
	15: math.Pow(15.0, 0.63),
	16: math.Pow(16.0, 0.63),
	17: math.Pow(17.0, 0.63),
	18: math.Pow(18.0, 0.63),
	19: math.Pow(19.0, 0.63),
	20: math.Pow(20.0, 0.63),
	21: math.Pow(21.0, 0.63),
	22: math.Pow(22.0, 0.63),
	23: math.Pow(23.0, 0.63),
	24: math.Pow(24.0, 0.63),
	25: math.Pow(25.0, 0.63),
	26: math.Pow(26.0, 0.63),
	27: math.Pow(27.0, 0.63),
	28: math.Pow(28.0, 0.63),
	29: math.Pow(29.0, 0.63),
	30: math.Pow(20.0, 0.63),
	31: math.Pow(31.0, 0.63),
	32: math.Pow(32.0, 0.63),
	33: math.Pow(33.0, 0.63),
	34: math.Pow(34.0, 0.63),
	35: math.Pow(35.0, 0.63),
	36: math.Pow(36.0, 0.63),
	37: math.Pow(37.0, 0.63),
	38: math.Pow(38.0, 0.63),
	39: math.Pow(39.0, 0.63),
	40: math.Pow(40.0, 0.63),
}

//var fSpeedFactor float64 = math.Pow(15, 0.63)
// NWS 23 pdf page 51, page 25, equation 2.5
// NWS 23 pdf page 263, page 269
// NWS 23 pdf page 281, page 257
// Factor for a moving hurricane, accounts for effect of forward speed on hurricane winds
// tau zero (to) conversion factors: 1 kt, 0.514791 mps, 1.853248 kph, 1.151556 mph
func asymmetryFactor(fSpeedKts float64,
	rNmi float64,
	rMaxNmi float64,
	bearingFromCenterDeg float64,
	cycloneBearingDeg float64) (asym float64, beta float64) {

	phiR := inflowAngle(rNmi, rMaxNmi)       // need to figure out direction
	phiRmax := inflowAngle(rMaxNmi, rMaxNmi) // need to figure out direction
	phiBeta := math.Mod(phiR-phiRmax, 360)
	bearingShift := math.Mod(90.0 - bearingFromCenterDeg + cycloneBearingDeg, 360)
	beta = math.Mod(phiBeta + bearingShift, 360)

	// PERF precalculate tau zero and fspeed
	//to := 1.0
	//asym = 1.5 * math.Pow(fSpeedKts, 0.63) * math.Pow(to, 0.37) * math.Cos(beta*utilities.ToRadians)
	asym = tauZeroFactor * fSpeedFactor[int(fSpeedKts)] * math.Cos(beta*utilities.ToRadians)
	return asym, beta
}

// CalcWindSpeed Calculate the wind speed at a given point from parameters.  This function does the least computational
// work of the CalcWindSpeed functions
// gradientWindAdjustmentFactor suggested default: 0.9
func CalcWindSpeed(cycloneVelocityMaxKts float64,
	rNmi float64,
	rMaxNmi float64,
	fSpeedKts float64,
	bearingFromCenterDeg float64,
	cycloneHeadingDegDeg float64,
	gradientWindAdjustmentFactor float64) (windSpeedKts float64, asym float64) {

	// Step 1: Calculate the Radial Decay
	radialDecayFactor := radialDecay(rNmi, rMaxNmi)

	// Step 2: Calculate the Asymmetry Factor
	asym, _ = asymmetryFactor(fSpeedKts, rNmi, rMaxNmi, bearingFromCenterDeg, cycloneHeadingDegDeg)

	// apply all factors and return wind speed at point
	windSpeedKts = (cycloneVelocityMaxKts * gradientWindAdjustmentFactor * radialDecayFactor) + asym

	return windSpeedKts, asym
}

// CalcWindSpeedFromPressure Calculate the wind speed at a given point from parameters
// gradientWindAdjustmentFactor suggested default: 0.9
// pwKpa suggested default: PwPmhKpa
func CalcWindSpeedFromPressure(cpMb float64,
	pwKpa float64,
	latDeg float64,
	rNmi float64,
	rMaxNmi float64,
	fSpeedKts float64,
	bearingFromCenterDeg float64,
	cycloneHeadingDeg float64,
	gradientWindAdjustmentFactor float64) (windSpeedKts float64, asym float64) {
	cpInHg := cpMb * MbToInHg
	pwInHg := pwKpa * KpaToInHg

	// Calculate Maximum Gradient Wind Speed, 10m-10min Average
	vgx, _, _ := gradientWindAtRadius(pwInHg, cpInHg, rMaxNmi, latDeg)

	return CalcWindSpeed(vgx, rNmi, rMaxNmi, fSpeedKts, bearingFromCenterDeg, cycloneHeadingDeg, gradientWindAdjustmentFactor)
}

func CalcWindSpeedFromPressureXY(pointLatYDeg float64,
	pointLonXDeg float64,
	eyeLatYDeg float64,
	eyeLonXDeg float64,
	cpMb float64,
	pwKpa float64,
	latDeg float64,
	rNmi float64,
	rMaxNmi float64,
	fSpeedKts float64,
	cycloneHeadingDeg float64,
	gradientWindAdjustmentFactor float64) (windSpeedKts float64, asym float64) {
	bearingFromCenterDeg := utilities.CalcBearingNorthZero(eyeLatYDeg, eyeLonXDeg, pointLatYDeg, pointLonXDeg)

	return CalcWindSpeedFromPressure(cpMb, pwKpa, latDeg, rNmi, rMaxNmi, fSpeedKts, bearingFromCenterDeg, cycloneHeadingDeg, gradientWindAdjustmentFactor)
}

func CalcWindSpeedXY(pointLatYDeg float64,
	pointLonXDeg float64,
	eyeLatYDeg float64,
	eyeLonXDeg float64,
	cycloneVelocityMaxKts float64,
	rNmi float64,
	rMaxNmi float64,
	fSpeedKts float64,
	cycloneHeadingDeg float64,
	gradientWindAdjustmentFactor float64) (windSpeedKts float64, asym float64) {

	bearingFromCenterDeg := utilities.CalcBearingNorthZero(eyeLatYDeg, eyeLonXDeg, pointLatYDeg, pointLonXDeg)

	return CalcWindSpeed(cycloneVelocityMaxKts, rNmi, rMaxNmi, fSpeedKts, bearingFromCenterDeg, cycloneHeadingDeg, gradientWindAdjustmentFactor)
}
