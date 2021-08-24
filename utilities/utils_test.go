package utilities

import "testing"

func TestLinearInterpolation(t *testing.T) {
	x1 := 0.0
	x2 := 2.0
	y1 := 0.0
	y2 := 10.0

	x := 1.0

	y := LinearInterpolation(x, x1, x2, y1, y2)

	if y != 5.0 {
		t.Errorf("Expected y == 5, got %v", y)
	}
}
