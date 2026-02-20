package vector

import (
	"fmt"
	"math"
	"strings"
)

func VectorPrint(name string, v [4]float32) string {
	return fmt.Sprintf("%s [% .6f % .6f % .6f % .6f]", name, v[0], v[1], v[2], v[3])
}

func VectorPrintByDecimal(name string, v [4]float32, decimalsAfterComma int) string {
	format := "%s [% .6f % .6f % .6f % .6f]"
	if decimalsAfterComma != 6 {
		format = strings.ReplaceAll(format, "6", fmt.Sprint(decimalsAfterComma))
	}
	return fmt.Sprintf(format, name, v[0], v[1], v[2], v[3])
}

func VectorPrintGyro(name string, v [3]float32) string {
	return fmt.Sprintf("%s [% .6f % .6f % .6f]", name, v[0], v[1], v[2])
}

func VectorDiff(v1 [4]float32, v2 [4]float32) [4]float32 {
	diff := [4]float32{}
	for i := range v1 {
		diff[i] = v1[i] - v2[i]
	}
	return diff
}
func VectorZero(v []float32) bool {
	var absSum float64
	for _, vv := range v {
		absSum += math.Abs(float64(vv))
	}
	return absSum < 0.000001
}
