package hurricane

import (
	"fmt"
	"testing"
)

func TestBoundingBoxToPoints(t *testing.T) {
	testBox := BoundingBox{
		LatYTopDeg:    10,
		LonXLeftDeg:   -10,
		LatYBottomDeg: -10,
		LonXRightDeg:  10,
	}

	points := testBox.toPoints(1, 1)
	expectedLength := 20 * 1 * 20 * 1
	if len(points) != expectedLength {
		t.Errorf("Expected length == %v, got %v", expectedLength, len(points))
	}

	fmt.Println("TestBoundingBoxToPoints: Pass")
}

func TestBoundingBoxToPoints180(t *testing.T) {
	testBox := BoundingBox{
		LatYTopDeg:    100,
		LonXLeftDeg:   80,
		LatYBottomDeg: 80,
		LonXRightDeg:  100,
	}

	points := testBox.toPoints(1, 1)
	expectedLength := 20 * 1 * 20 * 1
	if len(points) != expectedLength {
		t.Errorf("Expected length == %v, got %v", expectedLength, len(points))
	}

	fmt.Println("TestBoundingBoxToPoints180: Pass")
}
