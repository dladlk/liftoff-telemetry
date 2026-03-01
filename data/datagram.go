package lot_config

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"log"
	"math"

	"github.com/dladlk/liftoff-telemetry/vector"
)

// UDP Server to get Litfoff Telemtry
// https://steamcommunity.com/sharedfiles/filedetails/?id=3160488434

type Datagram struct {
	Timestamp float32    `desc:"current timestamp of the drone's flight, seconds. This value is reset to zero when the drone is reset"`
	Position  [3]float32 `desc:"world position as a 3D coordinate, meters, where Y - altitude: a left-handed, Y-Up coordinate system: the positive x-axis points to the right, the positive y-axis points up, and the positive z-axis points forward."`
	Attitude  [4]float32 `desc:"world attitude as a quaternion X, Y, Z, W, [-1,1]. Convert to degress with vector.AttentionQuaternionToEulerDegrees"`
	Velocity  [3]float32 `desc:"linear velocity as a 3D vector X, Y, Z in world-space, meters/second. To get velocity in local-space, use Attitude and https://steamcommunity.com/linkfilter/?u=https%3A%2F%2Fmath.stackexchange.com%2Fa%2F3209449 )"`
	Gyro      [3]float32 `desc:"angular velocity rates, represented with three components in the order: pitch, roll and yaw. The unit scale is in degrees/second"`
	Input     [4]float32 `desc:"input at that time, represented with four components in the following order: throttle, yaw, pitch and roll"`
	Battery   [2]float32 `desc:"remaining voltage and charge percentage"`
	Motors    byte       `desc:"number of motors"`
	MotorRPM  []float32  `desc:"rpm per each motor. The sequence of motors for a quadcopter in Liftoff is as follows: left front, right front, left back, right back"`
}

func (d Datagram) DistanceFrom(firstEvent *Datagram) float64 {
	a := firstEvent.Position
	b := d.Position

	return math.Sqrt(math.Pow(float64(a[0]-b[0]), 2) + math.Pow(float64(a[2]-b[2]), 2))
}

func (d *Datagram) ZeroPosition() bool {
	return d.Position[0] == 0 && d.Position[1] == 0 && d.Position[2] == 0
}

func (cur *Datagram) ParseDatagram(reader *bytes.Reader, fields *[]StreamDataType) {
	order := binary.LittleEndian

	for _, dataType := range *fields {
		switch dataType {
		case Timestamp:
			if err := binary.Read(reader, order, &cur.Timestamp); err != nil {
				log.Fatalf("Failed to read Timestamp as float: %s\n", err)
			}
		case Position:
			if err := binary.Read(reader, order, &cur.Position); err != nil {
				log.Fatalf("Failed to read Position as float[3]: %s\n", err)
			}
		case Attitude:
			if err := binary.Read(reader, order, &cur.Attitude); err != nil {
				log.Fatalf("Failed to read Attitude as float[4]: %s\n", err)
			}
		case Velocity:
			if err := binary.Read(reader, order, &cur.Velocity); err != nil {
				log.Fatalf("Failed to read Velocity as float[3]: %s\n", err)
			}
		case Gyro:
			if err := binary.Read(reader, order, &cur.Gyro); err != nil {
				log.Fatalf("Failed to read Gyro as float[3]: %s\n", err)
			}
		case Input:
			if err := binary.Read(reader, order, &cur.Input); err != nil {
				log.Fatalf("Failed to read Input as float[4]: %s\n", err)
			}
		case Battery:
			if err := binary.Read(reader, order, &cur.Battery); err != nil {
				log.Fatalf("Failed to read Battery as float[2]: %s\n", err)
			}
		case MotorRPM:
			if err := binary.Read(reader, order, &cur.Motors); err != nil {
				log.Fatalf("Failed to read Motors as byte: %s\n", err)
			}
			cur.MotorRPM = make([]float32, cur.Motors)
			if err := binary.Read(reader, order, &cur.MotorRPM); err != nil {
				log.Fatalf("Failed to read MotorRPM as float[%d]: %s\n", cur.Motors, err)
			}
		}
	}
}

func (d Datagram) String() string {
	return fmt.Sprintf("Ts=%6.3f\t%s\t%s\t%s\t%s\t%s\t%s", d.Timestamp,
		print3Big("Pos", d.Position),
		printAttitude("Att", d.Attitude),
		printVelocity("Vel", d.Velocity, d.Attitude),
		print3Big("Gyr", d.Gyro),
		print4Small("Inp", d.Input),
		printRPM("RPM", d.Motors, d.MotorRPM),
	)
}

func print3Big(name string, v [3]float32) string {
	return fmt.Sprintf("%s=[% 6.1f % 6.1f % 6.1f]", name, v[0], v[1], v[2])
}

func printVelocity(name string, v [3]float32, attitude [4]float32) string {
	return fmt.Sprintf("%s %s", print3Small(name, v), print3Small("VelLocal", vector.VelocityWorldSpaceToLocalSpace(v, attitude)))
}

func print3Small(name string, v [3]float32) string {
	return fmt.Sprintf("%s=[% 4.3f % 4.3f % 4.3f]", name, v[0], v[1], v[2])
}

func printAttitude(name string, v [4]float32) string {
	eulerDeg := AttitudeQuaternionToEulerDegrees(v)
	return fmt.Sprintf("%s %s", print4Small(name, v), print3Big("Deg", eulerDeg))
}

func print4Small(name string, v [4]float32) string {
	return fmt.Sprintf("%s=[% 4.3f % 4.3f % 4.3f % 4.3f]", name, v[0], v[1], v[2], v[3])
}

func printRPM(name string, len byte, v []float32) string {
	if len >= 4 {
		return fmt.Sprintf("%s=[% 4.1f % 4.1f % 4.1f % 4.1f]", name, v[0], v[1], v[2], v[3])
	}
	return "Non-4 rotors"
}

func AttitudeQuaternionToEulerDegrees(Attention [4]float32) [3]float32 {
	radians := AttitudeQuaternionToEulerRadians(Attention)
	return [3]float32{radianToDegree(radians[0]), radianToDegree(radians[1]), radianToDegree(radians[2])}
}

// Attitude from Datagram - X, Y, Z, W
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

func radianToDegree(x float32) float32 {
	return x * 180.0 / math.Pi
}

func atan2(x float32, y float32) float32 {
	return float32(math.Atan2(float64(x), float64(y)))
}
func asin(x float32) float32 {
	return float32(math.Asin(float64(x)))
}
