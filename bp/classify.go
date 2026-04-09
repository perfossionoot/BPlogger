package bp

import "image/color"

// Classification holds the display label and color for a BP reading category.
type Classification struct {
	Label string
	Color color.RGBA
}

// Transparent sentinel used when there is no value to classify.
var none = color.RGBA{}

var (
	clsUnknown  = Classification{"", none}
	clsNormal   = Classification{"Normal", color.RGBA{R: 76, G: 175, B: 80, A: 255}}
	clsElevated = Classification{"Elevated", color.RGBA{R: 255, G: 193, B: 7, A: 255}}
	clsStage1   = Classification{"High – Stage 1", color.RGBA{R: 255, G: 152, B: 0, A: 255}}
	clsStage2   = Classification{"High – Stage 2", color.RGBA{R: 220, G: 53, B: 69, A: 255}}
	clsCrisis   = Classification{"Hypertensive Crisis", color.RGBA{R: 136, G: 0, B: 0, A: 255}}
)

// ClassifySystolic returns the BP category for a systolic pressure value (mmHg).
// Ranges follow the 2017 ACC/AHA guidelines.
//
//	< 120        Normal
//	120–129      Elevated
//	130–139      High – Stage 1
//	140–179      High – Stage 2
//	≥ 180        Hypertensive Crisis
func ClassifySystolic(v int) Classification {
	switch {
	case v <= 0:
		return clsUnknown
	case v < 120:
		return clsNormal
	case v < 130:
		return clsElevated
	case v < 140:
		return clsStage1
	case v < 180:
		return clsStage2
	default:
		return clsCrisis
	}
}

// ClassifyDiastolic returns the BP category for a diastolic pressure value (mmHg).
//
//	< 80         Normal
//	80–89        High – Stage 1
//	90–119       High – Stage 2
//	≥ 120        Hypertensive Crisis
func ClassifyDiastolic(v int) Classification {
	switch {
	case v <= 0:
		return clsUnknown
	case v < 80:
		return clsNormal
	case v < 90:
		return clsStage1
	case v < 120:
		return clsStage2
	default:
		return clsCrisis
	}
}
