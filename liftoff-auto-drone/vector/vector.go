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

func VectorPrint3Short(name string, v [3]float32) string {
	return fmt.Sprintf("%s [% 4.1f % 4.1f % 4.1f]", name, v[0], v[1], v[2])
}

func VectorPrint3Long(name string, v [3]float32) string {
	return fmt.Sprintf("%s [% 7.6f % 7.6f % 7.6f]", name, v[0], v[1], v[2])
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
func AttitudeQuaternionToEulerRadians(Attitude [4]float32) [3]float32 {
	x := Attitude[0]
	y := Attitude[1]
	z := Attitude[2]
	w := Attitude[3]

	x_phi_roll := atan2(2*(w*x+y*z), 1-2*(x*x+y*y))
	y_theta_pitch := asin(2 * (w*y - x*z))
	z_psi_yaw := atan2(2*(w*z+x*y), 1-2*(y*y+z*z))

	return [3]float32{x_phi_roll, y_theta_pitch, z_psi_yaw}
}

func AttitudeQuaternionToEulerDegrees(Attention [4]float32) [3]float32 {
	radians := AttitudeQuaternionToEulerRadians(Attention)
	return [3]float32{radianToDegree(radians[0]), radianToDegree(radians[1]), radianToDegree(radians[2])}
}

// Represents quaternion as 4 real numbers a, b, c, d which stay for a + bi + cj + dk
// See definition in https://en.wikipedia.org/wiki/Quaternion
type Quaternion struct {
	a, b, c, d float32
}

func (q Quaternion) ToAttitude() [4]float32 {
	return [4]float32{q.b, q.c, q.d, q.a}
}

// Conjugated quaternion - https://en.wikipedia.org/wiki/Quaternion#Conjugation,_the_norm,_and_reciprocal
func (q Quaternion) ToConjugated() Quaternion {
	return Quaternion{q.a, -q.b, -q.c, -q.d}
}

// Attitude form for quaternion - x,y,z,w - corresponds to b,c,d,a in Quaternion definition from wiki
func AttitudeToQuaternion(att []float32) Quaternion {
	return Quaternion{a: att[3], b: att[0], c: att[1], d: att[2]}
}

func EulerAnglesToQuaternion(x_phi_roll, y_theta_pitch, z_psi_yaw float64) Quaternion {
	c1 := math.Cos(x_phi_roll / 2)
	s1 := math.Sin(x_phi_roll / 2)
	c2 := math.Cos(y_theta_pitch / 2)
	s2 := math.Sin(y_theta_pitch / 2)
	c3 := math.Cos(z_psi_yaw / 2)
	s3 := math.Sin(z_psi_yaw / 2)

	q := Quaternion{}
	q.a = float32(c1*c2*c3 + s1*s2*s3)
	q.b = float32(s1*c2*c3 - c1*s2*s3)
	q.c = float32(c1*s2*c3 + s1*c2*s3)
	q.d = float32(c1*c2*s3 - s1*s2*c3)

	return q
}

// Hamiltonian product of x and y (see https://en.wikipedia.org/wiki/Quaternion#Hamilton_product)
func QuaternionMultiplication(n1, n2 Quaternion) Quaternion {
	return Quaternion{
		a: n1.a*n2.a - n1.b*n2.b - n1.c*n2.c - n1.d*n2.d,
		b: n1.a*n2.b + n1.b*n2.a + n1.c*n2.d - n1.d*n2.c,
		c: n1.a*n2.c - n1.b*n2.d + n1.c*n2.a + n1.d*n2.b,
		d: n1.a*n2.d + n1.b*n2.c - n1.c*n2.b + n1.d*n2.a,
	}
}

// Converts linear velocity as a 3D vector X, Y, Z in world-space, meters/second to local-space.
// To get velocity in local-space, use Attitude and https://steamcommunity.com/linkfilter/?u=https%3A%2F%2Fmath.stackexchange.com%2Fa%2F3209449 )
// Looks like the person who wrote it had around 4,5 seconds to make the documentation, so did not bother to explain what exactly should be done.
// Let's suppose that conversion is done by 2 consequent quaternion multiplications of:
// Attitude (q) on Velocity (v, real part a set to 0),
// and again on conjugated Attitude q∗
// Result formula:
// v′=qvq∗=(q0,q1,q2,q3)(0,v1,v2,v3)(q0,−q1,−q2,−q3)=(w,x,y,z)
func VelocityWorldSpaceToLocalSpace(velocity [3]float32, attitude [4]float32) [3]float32 {
	q := AttitudeToQuaternion(attitude[:])
	v := Quaternion{a: 0, b: velocity[0], c: velocity[1], d: velocity[2]}
	qconj := q.ToConjugated()

	qr := QuaternionMultiplication(QuaternionMultiplication(q, v), qconj)

	// If you have done everything correctly, you should get w=0 and x,y,z as linear functions of v1,v2,v3

	if math.Abs(float64(qr.a)) > 0.000001 {
		fmt.Printf("ATTENTION! Conversion of velocity in world space to local space is expected to return real part of Quarternion as 0, but got % 8.6f. Input values: velocity: %+v, attitude: %+v\n", qr.a, velocity, attitude)
	}

	return [3]float32{qr.b, qr.c, qr.d}
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
