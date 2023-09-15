package hurricane

import "math"

// All numerics are in float64

// deltaPs_hPa: Ps in hPa,
// deltaPsRate: dPcs/dt the intensity change in hPa/hr
// lat_deg: phi absolute value latitiude
// velocityCyclone_ms: vt cyclone translation velocity m/s
func bsFromPressure(deltaPs_hPa float64, deltaPsRate_hPaPerHr float64, lat_deg float64, velocityCyclone_ms float64) float64 {
	xParam := 0.6 * (1 - (deltaPs_hPa / 215))

	return (-0.000044 * deltaPs_hPa * deltaPs_hPa) + (0.01 * deltaPs_hPa) + (0.3 * deltaPsRate_hPaPerHr) -
		(0.014 * lat_deg) + (0.15 * (math.Pow(velocityCyclone_ms, xParam))) + 1.0
}

// vMax_ms: vms, max surface winds (10m) in m/s
// cpMax: central pressure at max wind.  Unit must match periphPressure.  mb or hPa recommended
// periphPressure: peripheral pressure. unit must match cpMax.  mb or hPa recommended
// bs: unitless b term
func bsFromMaxWind(vMax_ms float64, cpMax float64, periphPressure float64) float64 {
	return (vMax_ms * vMax_ms * cpMax * math.E) / (100 * 1)
}

// surfacePressureAtRadius
// pCentral_hPa: pcs
// pPeriph_hPa: pns
// r_km: r, point radius
// rMaxWind_km: rvms, radius of max wind
func surfacePressureAtRadius(r_km float64, rMaxWind_km, pCentral_hPa float64, pPeriph_hPa float64) float64 {
	bScalingParam := 0.9
	deltaP_hPa := pPeriph_hPa - pCentral_hPa
	eTerm := -1.0 * math.Pow((rMaxWind_km/r_km), bScalingParam)
	return pCentral_hPa + (deltaP_hPa * math.Exp(eTerm))
}
