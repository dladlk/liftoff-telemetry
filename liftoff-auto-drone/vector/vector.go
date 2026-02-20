package vector

import (
	"fmt"
	"math"
	"strings"
)

func VectorPrint(name string, v [4]float32) string {
	return fmt.Sprintf("%s [% .6f % .6f % .6f % .6f]", name, v[0], v[1], v[2], v[3])
}

func VectorPrintTabbed(name string, v [4]float32) string {
	return fmt.Sprintf("%s [% .6f\t% .6f\t% .6f\t% .6f]", name, v[0], v[1], v[2], v[3])
}

func VectorPrintByDecimal(name string, v [4]float32, decimalsAfterComma int) string {
	format := "%s [% .6f % .6f % .6f % .6f]"
	if decimalsAfterComma != 6 {
		format = strings.ReplaceAll(format, "6", fmt.Sprint(decimalsAfterComma))
	}
	return fmt.Sprintf(format, name, v[0], v[1], v[2], v[3])
}

func VectorPrint3(name string, v [3]float32) string {
	return fmt.Sprintf("%s [% 4.1f % 4.1f % 4.1f]", name, v[0], v[1], v[2])
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

// Attention from Datagram - X, Y, Z, W
// Returns Eurle's Angles in Radians as x_phi_roll, y_theta_pitch, z_psi_yaw
// Done by https://en.wikipedia.org/wiki/Conversion_between_quaternions_and_Euler_angles#Quaternion_to_angles_(in_ZYX_sequence)_conversion
func AttentionQuaternionToEulerRadians(Attention [4]float32) [3]float32 {
	x := Attention[0]
	y := Attention[1]
	z := Attention[2]
	w := Attention[3]

	x_phi_roll := atan2(2*(w*x+y*z), 1-2*(x*x+y*y))
	y_theta_pitch := asin(2 * (w*y - x*z))
	z_psi_yaw := atan2(2*(w*z+x*y), 1-2*(y*y+z*z))

	return [3]float32{x_phi_roll, y_theta_pitch, z_psi_yaw}
}

func AttentionQuaternionToEulerDegrees(Attention [4]float32) [3]float32 {
	radians := AttentionQuaternionToEulerRadians(Attention)
	return [3]float32{radianToDegree(radians[0]), radianToDegree(radians[1]), radianToDegree(radians[2])}
}

func radianToDegree(x float32) float32 {
	return x * 180.0 / math.Pi
}

func atan2(x float32, y float32) float32 {
	return float32(math.Atan2(float64(x), float64(y)))
}
func asin(x float32) float32 {
	return float32(math.Asin(float64(x)))
}
