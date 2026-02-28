package main

import (
	"context"
	"flag"
	"fmt"
	"math"
	"os"
	"time"

	track "github.com/dladlk/liftoff-auto-drone/track"
	lot_config "github.com/dladlk/liftoff-telemetry/data"
)

//
// ================== Math helpers ==================
//

type Vec3 [3]float64

func (t Vec3) String() string {
	return fmt.Sprintf("[% 5.2f % 5.2f % 5.2f]", t[0], t[1], t[2])
}

type Mat3 [3][3]float64

func (m Mat3) String() string {
	return fmt.Sprintf("[[% 5.2f % 5.2f % 5.2f][% 5.2f % 5.2f % 5.2f][% 5.2f % 5.2f % 5.2f]]", m[0][0], m[0][1], m[0][2], m[1][0], m[1][1], m[1][2], m[2][0], m[2][1], m[2][2])
}

func add(a, b Vec3) Vec3         { return Vec3{a[0] + b[0], a[1] + b[1], a[2] + b[2]} }
func sub(a, b Vec3) Vec3         { return Vec3{a[0] - b[0], a[1] - b[1], a[2] - b[2]} }
func mul(a Vec3, s float64) Vec3 { return Vec3{a[0] * s, a[1] * s, a[2] * s} }
func dot(a, b Vec3) float64      { return a[0]*b[0] + a[1]*b[1] + a[2]*b[2] }
func cross(a, b Vec3) Vec3 {
	return Vec3{
		a[1]*b[2] - a[2]*b[1],
		a[2]*b[0] - a[0]*b[2],
		a[0]*b[1] - a[1]*b[0],
	}
}
func norm(a Vec3) float64 { return math.Sqrt(dot(a, a)) }
func normalize(a Vec3) Vec3 {
	n := norm(a)
	if n < 1e-9 {
		return Vec3{0, 0, 1}
	}
	return Vec3{a[0] / n, a[1] / n, a[2] / n}
}

func clamp(x, lo, hi float64) float64 {
	if x < lo {
		return lo
	}
	if x > hi {
		return hi
	}
	return x
}
func clampVec3(v Vec3, lo, hi float64) Vec3 {
	return Vec3{clamp(v[0], lo, hi), clamp(v[1], lo, hi), clamp(v[2], lo, hi)}
}
func subMat(A, B Mat3) Mat3 {
	return Mat3{
		{A[0][0] - B[0][0], A[0][1] - B[0][1], A[0][2] - B[0][2]},
		{A[1][0] - B[1][0], A[1][1] - B[1][1], A[1][2] - B[1][2]},
		{A[2][0] - B[2][0], A[2][1] - B[2][1], A[2][2] - B[2][2]},
	}
}
func matMul(A, B Mat3) Mat3 {
	var C Mat3
	for i := 0; i < 3; i++ {
		for j := 0; j < 3; j++ {
			C[i][j] = A[i][0]*B[0][j] + A[i][1]*B[1][j] + A[i][2]*B[2][j]
		}
	}
	return C
}
func matT(R Mat3) Mat3 {
	return Mat3{
		{R[0][0], R[1][0], R[2][0]},
		{R[0][1], R[1][1], R[2][1]},
		{R[0][2], R[1][2], R[2][2]},
	}
}

// ============================
//   Rotations (zyx) & quaternions
//
// R = Rz(yaw)*Ry(pitch)*Rx(roll) maps body -> world
//

func EulerToR(roll, pitch, yaw float64) Mat3 {
	cr, sr := math.Cos(roll), math.Sin(roll)
	cp, sp := math.Cos(pitch), math.Sin(pitch)
	cy, sy := math.Cos(yaw), math.Sin(yaw)

	return Mat3{
		{cy * cp, cy*sp*sr - sy*cr, cy*sp*cr + sy*sr},
		{sy * cp, sy*sp*sr + cy*cr, sy*sp*cr - cy*sr},
		{-sp, cp * sr, cp * cr},
	}
}

func rToEuler(R Mat3) (roll, pitch, yaw float64) {
	pitch = -math.Asin(clamp(R[2][0], -1.0, 1.0))
	cp := math.Cos(pitch)
	if cp < 1e-6 {
		// Gimbal lock fallback
		roll = 0
		yaw = math.Atan2(-R[0][1], R[1][1])
		return
	}
	roll = math.Atan2(R[2][1], R[2][2])
	yaw = math.Atan2(R[1][0], R[0][0])
	return
}

// Old name: quatToR
// Defines Quaternion-derived rotation matrix R
// https://en.wikipedia.org/wiki/Quaternions_and_spatial_rotation#Quaternion-derived_rotation_matrix
func quaternionToRotationMatrix(w, x, y, z float64) Mat3 {
	// unit quaternion expected
	xx, yy, zz := x*x, y*y, z*z
	xy, xz, yz := x*y, x*z, y*z
	wx, wy, wz := w*x, w*y, w*z
	return Mat3{
		{1 - 2*(yy+zz), 2 * (xy - wz), 2 * (xz + wy)},
		{2 * (xy + wz), 1 - 2*(xx+zz), 2 * (yz - wx)},
		{2 * (xz - wy), 2 * (yz + wx), 1 - 2*(xx+yy)},
	}
}

// maps a skew-symmetric matrix to a vector to calculate attitude error vector for PD controller
func vee(S Mat3) Vec3 {
	return Vec3{
		(S[2][1] - S[1][2]) * 0.5,
		(S[0][2] - S[2][0]) * 0.5,
		(S[1][0] - S[0][1]) * 0.5,
	}
}

func wrapPi(a float64) float64 {
	for a > math.Pi {
		a -= 2 * math.Pi
	}
	for a < -math.Pi {
		a += 2 * math.Pi
	}
	return a
}

//
// ================== Interfaces ==================
//

// Telemetry: world ENU (m, m/s), attitude as body->world rotation, body rates (rad/s)
type Telemetry struct {
	Index    int
	Time     time.Time
	Position Vec3 // position (world, m) as x East, y North, z Up
	Velocity Vec3 // velocity (world, m/s)
	Rotation Mat3 // rotation (body->world)
	// NOT USED in calculation???
	Omega Vec3 // body rates [p q r], rad/s
}

func (t Telemetry) String() string {
	return fmt.Sprintf("% 5d) %s %s", t.Index, t.Position, t.Velocity)
}

type TelemetryProvider interface {
	ReadTelemetry() (Telemetry, bool, error)
}

// Desired setpoint: position/velocity/acceleration and yaw (heading)
type Setpoint struct {
	PositionDesired        Vec3
	VelocityDesired        Vec3
	AccelerationDesired    Vec3    // desired acceleration, by some reason it was called aff originally
	PsiD                   float64 // desired yaw (rad)
	HasVelocityDesired     bool
	HasAccelerationDesired bool
}
type SetpointProvider interface {
	GetDesiredSetpoint(now time.Time, tel Telemetry) (Setpoint, error)
}

/* ============================
   Controller (ACRO)
   ============================ */

// Joystick actuator in ACRO:
//   - leftVert  : throttle (collective). Un-centered(DLK: REALLY?) (0..+MaxFloat32) by default.
//   - leftHoriz : yaw rate ([-MaxFloat32, +MaxFloat32])
//   - rightVert : pitch rate ([-MaxFloat32, +MaxFloat32])
//   - rightHoriz: roll  rate ([-MaxFloat32, +MaxFloat32])
type JoystickActuator interface {
	// ACRO: LV=throttle [0..+32767] (uncentered) or [-32767..+32767] (centered)
	SendJoystick(joystickPosition JoystickPosition) error
}

type LiftoffTelemetryProvider struct {
	TelemetryListener *track.TelemetryListener
}

func (l *LiftoffTelemetryProvider) ReadTelemetry() (Telemetry, bool, error) {
	if !l.TelemetryListener.Running {
		fmt.Printf("Listener not running...\n")
		return Telemetry{}, false, nil
	}
	datagram, index, ok := l.TelemetryListener.LastDatagram()
	if !ok {
		return Telemetry{}, ok, nil
	}
	tel := DatagramToTelemetry(datagram, index)
	return tel, true, nil
}

func DatagramToTelemetry(datagram *lot_config.Datagram, index int) Telemetry {
	R := quaternionToRotationMatrix(float64(datagram.Attitude[3]), float64(datagram.Attitude[0]), float64(datagram.Attitude[1]), float64(datagram.Attitude[2]))
	tel := Telemetry{
		Index:    index,
		Time:     time.Now(),
		Position: liftoffDatagramYZXToVec3(datagram.Position),
		Velocity: liftoffDatagramYZXToVec3(datagram.Velocity),
		Rotation: R,
		// Liftoff exports gyro as (pitch, roll, yaw) per example; map to body rates [p q r] carefully if needed
		Omega: Vec3{float64(datagram.Gyro[1]), float64(datagram.Gyro[0]), float64(datagram.Gyro[2])},
	}
	return tel
}

// Expected Telemetry format: ENU (x East, y North, z Up) - latitude, longtitude, altitude
// Liftoff uses a left-handed, Y-Up coordinate system: the positive x-axis points to the right, the positive y-axis points up, and the positive z-axis points forward.
// So it means that actually we receive coordinates as yzx (pos 0=y, pos 1=z, pos 2=x), so convert to x,y,z vector we should take them as 2,0,1
func liftoffDatagramYZXToVec3(f [3]float32) Vec3 {
	// So actually we recieve it as x,z,y - y,z,x
	return Vec3{float64(f[2]), float64(f[0]), float64(f[1])}
}

//
// ================== Config & State ==================
//

type PosGains struct {
	Kp       Vec3 // position P
	Kv       Vec3 // velocity P (damping)
	Ki       Vec3 // position I
	IntLimit float64
}

type AcroRateLimits struct {
	MaxRollRate  float64 // rad/s
	MaxPitchRate float64 // rad/s
	MaxYawRate   float64 // rad/s
}

// ACRO throttle shaping
type ThrottleMap struct {
	// If Centered=false (typical ACRO), stick ∈ [0,1];  hover is at HoverStick in [0..1]
	// throttleStick = clamp( HoverStick + Slope*(T/(mg)-1), 0, 1 )
	// If Centered=true (spring), stick ∈ [-1,1]; hover is 0
	// throttleStick = clamp( (T/(mg)-1)/CenteredSpan, -1, 1 )
	HoverStick       float64 // where hover sits on [0..1] scale (e.g., 0.5)
	Slope            float64 // linear sensitivity around hover (unitless)
	Centered         bool    // if true, map to [-1,1] centered at hover
	CenteredSpan     float64 // how much relative thrust delta maps to full deflection (e.g., 1.0 → ±100% per ±100% thrust delta)
	Deadzone         float64 // small deadzone around hover or zero
	RateLimitPerTick float64 // max change per tick in stick units
}

type OutputSigns struct {
	InvertRoll  bool // flips right-horizontal
	InvertPitch bool // flips right-vertical
	InvertYaw   bool // flips left-horizontal
}

type ControllerConfig struct {
	Mass float64
	G    float64
	Hz   float64
	Dt   float64

	Pos        PosGains
	R2RateGain Vec3    // gains mapping attitude error -> desired body rates
	YawP       float64 // extra P on yaw heading error to rate (optional)
	Rates      AcroRateLimits
	Throttle   ThrottleMap
	Signs      OutputSigns
}

type ControllerState struct {
	Index                          int
	IntegralPositionalError        Vec3
	lastLV, lastLH, lastRV, lastRH float64
}

func (s ControllerState) String() string {
	return fmt.Sprintf("IPE %s last:[% 4.1f % 4.1f % 4.1f % 4.1f]", s.IntegralPositionalError, s.lastLV, s.lastLH, s.lastRV, s.lastRH)
}

type Controller struct {
	cfg   ControllerConfig
	state ControllerState
	tprov TelemetryProvider
	sprov SetpointProvider
	act   JoystickActuator
}

func (c *Controller) ResetState(lastLv float64) {
	c.state = ControllerState{lastLV: lastLv}
}

func NewController(cfg ControllerConfig, t TelemetryProvider, s SetpointProvider, a JoystickActuator) *Controller {
	return &Controller{
		cfg:   cfg,
		tprov: t,
		sprov: s,
		act:   a,
	}
}

//
// ================== Control blocks ==================
//

// 1) Outer (position) loop to desired world acceleration
func (c *Controller) positionLoop(position, velocity Vec3, sp Setpoint, dt float64) (acceleration Vec3) {
	velocityDesired := sp.VelocityDesired
	if !sp.HasVelocityDesired {
		velocityDesired = Vec3{0, 0, 0}
	}
	aff := sp.AccelerationDesired
	if !sp.HasAccelerationDesired {
		aff = Vec3{0, 0, 0}
	}

	errorPosition := sub(sp.PositionDesired, position)
	errorVelocity := sub(velocityDesired, velocity)

	// integrate with clamp (anti-windup)
	il := c.cfg.Pos.IntLimit
	c.state.IntegralPositionalError = clampVec3(add(c.state.IntegralPositionalError, mul(errorPosition, dt)), -il, il)

	acceleration = Vec3{
		aff[0] + c.cfg.Pos.Kp[0]*errorPosition[0] + c.cfg.Pos.Kv[0]*errorVelocity[0] + c.cfg.Pos.Ki[0]*c.state.IntegralPositionalError[0],
		aff[1] + c.cfg.Pos.Kp[1]*errorPosition[1] + c.cfg.Pos.Kv[1]*errorVelocity[1] + c.cfg.Pos.Ki[1]*c.state.IntegralPositionalError[1],
		aff[2] + c.cfg.Pos.Kp[2]*errorPosition[2] + c.cfg.Pos.Kv[2]*errorVelocity[2] + c.cfg.Pos.Ki[2]*c.state.IntegralPositionalError[2],
	}
	return
}

// 2) From desired acceleration (a_c) and yaw (psi_d) to desired attitude (Rd) and thrust (T)
func attitudeAndThrustFromAccelYaw(a_c Vec3, g float64, psiD float64, mass float64) (Rd Mat3, thrust float64) {
	gvec := Vec3{0, 0, -g}
	accStar := sub(a_c, gvec)     // a_c - g
	b3d := normalize(accStar)     // desired body z axis (world)
	thrust = mass * norm(accStar) // collective thrust

	// build body x-axis to honor desired yaw
	b1c := Vec3{math.Cos(psiD), math.Sin(psiD), 0}
	b2d := normalize(cross(b3d, b1c))
	if norm(b2d) < 1e-6 {
		// nearly collinear, rotate reference axis by 90°
		b1c = Vec3{math.Cos(psiD + math.Pi/2), math.Sin(psiD + math.Pi/2), 0}
		b2d = normalize(cross(b3d, b1c))
	}
	b1d := cross(b2d, b3d)
	Rd = Mat3{
		{b1d[0], b2d[0], b3d[0]},
		{b1d[1], b2d[1], b3d[1]},
		{b1d[2], b2d[2], b3d[2]},
	}
	return
}

// 3) Geometric attitude error (SO(3)) to desired body rates (ACRO targets)
func (c *Controller) desiredBodyRates(R, Rd Mat3, yawCur, yawD float64) (omegaCmd Vec3) {
	Re := matMul(matT(Rd), R)
	eR := vee(subMat(Re, matT(Re))) // attitude error vector in R^3 (approx rotation error)

	// Map attitude error to body rate command (p,q,r)
	omegaCmd = Vec3{
		c.cfg.R2RateGain[0] * eR[0],
		c.cfg.R2RateGain[1] * eR[1],
		c.cfg.R2RateGain[2] * eR[2],
	}

	// Optionally bias yaw rate by heading error explicitly (helps when near hover)
	if c.cfg.YawP > 0 {
		yawErr := wrapPi(yawD - yawCur)
		omegaCmd[2] += c.cfg.YawP * yawErr
	}

	// Clip to rate limits
	omegaCmd[0] = clamp(omegaCmd[0], -c.cfg.Rates.MaxRollRate, c.cfg.Rates.MaxRollRate)
	omegaCmd[1] = clamp(omegaCmd[1], -c.cfg.Rates.MaxPitchRate, c.cfg.Rates.MaxPitchRate)
	omegaCmd[2] = clamp(omegaCmd[2], -c.cfg.Rates.MaxYawRate, c.cfg.Rates.MaxYawRate)
	return
}

//
// ================== ACRO Joystick mapping ==================
//

func signInvert(x float64, inv bool) float64 {
	if inv {
		return -x
	}
	return x
}

func applyDeadzone(x, dz float64) float64 {
	if math.Abs(x) <= dz {
		return 0
	}
	s := math.Copysign(1, x)
	return s * (math.Abs(x) - dz) / (1 - dz)
}
func rateLimit(x, last, maxDelta float64) float64 {
	d := x - last
	if d > maxDelta {
		return last + maxDelta
	}
	if d < -maxDelta {
		return last - maxDelta
	}
	return x
}

const i16Max = math.MaxInt16

func toInt16Signed(norm float64) int16 {
	n := clamp(norm, -1, 1)
	return int16(math.Round(n * i16Max))
}

// Get value in range -32767..32767 (math.MinInt16+1..math.MaxInt16) corresponding to 0..1
func ToInt16Uncentered01(x float64) int16 {
	x = clamp(x, 0, 1)
	return int16(math.MinInt16 + 1 + math.Round(2*x*i16Max))
}

// RH (right-horiz) : roll rate  → normalized by MaxRollRate
// RV (right-vert)  : pitch rate → normalized by MaxPitchRate
// LH (left-horiz)  : yaw rate   → normalized by MaxYawRate
// LV (left-vert)   : throttle   → from thrust T vs hover (mg) using ThrottleMap
//

type JoystickPosition struct {
	LV int16
	LH int16
	RV int16
	RH int16
}

func (j JoystickPosition) String() string {
	return fmt.Sprintf("[%d %d %d %d]", j.LV, j.LH, j.RV, j.RH)
}

func (c *Controller) toAcroJoystick(tel Telemetry, sp Setpoint, Rd Mat3, T float64) JoystickPosition {
	// current yaw
	_, _, yawCur := rToEuler(tel.Rotation)

	// desired body rates from attitude error
	omegaCmd := c.desiredBodyRates(tel.Rotation, Rd, yawCur, sp.PsiD)

	// normalize to [-1..1]
	rhNorm := omegaCmd[0] / c.cfg.Rates.MaxRollRate
	rvNorm := omegaCmd[1] / c.cfg.Rates.MaxPitchRate
	lhNorm := omegaCmd[2] / c.cfg.Rates.MaxYawRate

	rhNorm = signInvert(rhNorm, c.cfg.Signs.InvertRoll)
	rvNorm = signInvert(rvNorm, c.cfg.Signs.InvertPitch)
	lhNorm = signInvert(lhNorm, c.cfg.Signs.InvertYaw)

	// --------------------
	// throttle mapping
	// --------------------
	tHover := c.cfg.Mass * c.cfg.G
	rel := T / tHover // 1.0 at hover

	dz := clamp(c.cfg.Throttle.Deadzone, 0, 0.2)
	rl := clamp(c.cfg.Throttle.RateLimitPerTick, 0, 1)

	rhNorm = applyDeadzone(clamp(rhNorm, -1, 1), dz)
	rvNorm = applyDeadzone(clamp(rvNorm, -1, 1), dz)
	lhNorm = applyDeadzone(clamp(lhNorm, -1, 1), dz)

	rhNorm = rateLimit(rhNorm, c.state.lastRH, rl)
	rvNorm = rateLimit(rvNorm, c.state.lastRV, rl)
	lhNorm = rateLimit(lhNorm, c.state.lastLH, rl)

	c.state.lastRH = rhNorm
	c.state.lastRV = rvNorm
	c.state.lastLH = lhNorm

	rh := toInt16Signed(rhNorm)
	rv := toInt16Signed(rvNorm)
	lh := toInt16Signed(lhNorm)
	var lv int16

	if c.cfg.Throttle.Centered {
		// Centered throttle in [-1,1], 0 at hover
		span := math.Max(c.cfg.Throttle.CenteredSpan, 1e-6)
		raw := clamp((rel-1.0)/span, -1, 1)
		raw = applyDeadzone(raw, dz)
		raw = rateLimit(raw, c.state.lastLV, rl)
		c.state.lastLV = raw
		lv = toInt16Signed(raw)
	} else {
		// Uncentered throttle in [0,1], HoverStick at rel=1
		raw := c.cfg.Throttle.HoverStick + c.cfg.Throttle.Slope*(rel-1.0)
		raw = clamp(raw, 0, 1)
		raw = rateLimit(raw, c.state.lastLV, rl)
		c.state.lastLV = raw
		lv = ToInt16Uncentered01(raw)
	}
	return JoystickPosition{LV: lv, LH: lh, RV: rv, RH: rh}
}

//
// ================== Run loop ==================
//

func (c *Controller) Run(ctx context.Context) error {
	ticker := time.NewTicker(time.Duration(1e9 / c.cfg.Hz))
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			fmt.Printf("Reset joystick to down position but not full down on finish")
			c.act.SendJoystick(JoystickPosition{math.MinInt16 + 15000, 0, 0, 0})
			time.Sleep(1 * time.Second)

			return ctx.Err()
		case now := <-ticker.C:
			tel, ok, err := c.tprov.ReadTelemetry()
			if !ok {
				continue
			}
			if err != nil {
				fmt.Printf("telemetry read: %v\n", err)
				continue
			}
			sp, err := c.sprov.GetDesiredSetpoint(now, tel)
			if err != nil {
				return fmt.Errorf("setpoint read: %w", err)
			}

			joystickPosition := CalculateJoysticksPosition(c, tel, sp)
			if err := c.act.SendJoystick(joystickPosition); err != nil {
				return fmt.Errorf("joystick send: %w", err)
			}
		}
	}
}

func CalculateJoysticksPosition(c *Controller, tel Telemetry, sp Setpoint) JoystickPosition {
	if showDebugCalculation {
		fmt.Printf("\n************** Cycle %d ***************\n", c.state.Index)
	}

	// 1) Position loop -> desired acceleration

	if showDebugCalculation {
		fmt.Printf("INPUT: telemetry %s , setpoint pos: %s\n", tel, sp.PositionDesired)
	}

	acceleration := c.positionLoop(tel.Position, tel.Velocity, sp, c.cfg.Dt)

	if showDebugCalculation {
		fmt.Printf("1) Acceleration: %s\n", acceleration)
	}

	// 2) Desired attitude from yaw and thrust direction
	Rd, thrust := attitudeAndThrustFromAccelYaw(acceleration, c.cfg.G, sp.PsiD, c.cfg.Mass)

	if showDebugCalculation {
		fmt.Printf("2) Thrust, attitude and state: %f %s %s\n", thrust, Rd, c.state)
	}

	// 3) Map to ACRO sticks (rates + throttle)
	j := c.toAcroJoystick(tel, sp, Rd, thrust)

	if showDebugCalculation {
		fmt.Printf("3) Output joystick: %s\n", j)
	}
	c.state.Index++
	return j
}

//
// ================== Demo stubs ==================
//

type MockSetpoint struct {
	PositionDesired Vec3
	PsiD            float64
	Start           time.Time
	moved           bool
}

func (m *MockSetpoint) GetDesiredSetpoint(now time.Time, tel Telemetry) (Setpoint, error) {
	if !m.moved {
		sinceStart := now.Sub(m.Start).Seconds()
		if sinceStart >= 5 {
			m.PositionDesired = Vec3{tel.Position[0], tel.Position[1], tel.Position[2] + 5}
			fmt.Printf("Change setpoint to %s\n", m.PositionDesired)
			m.moved = true
		}
	}
	return Setpoint{
		PositionDesired:        m.PositionDesired,
		VelocityDesired:        Vec3{0, 0, 0}, // stop there
		AccelerationDesired:    Vec3{0, 0, 0}, // no feedforward
		PsiD:                   m.PsiD,
		HasVelocityDesired:     true,
		HasAccelerationDesired: true,
	}, nil
}

type RealJoystick struct {
	client track.JoystickControlClient
}

func (a *RealJoystick) SendJoystick(joystickPosition JoystickPosition) error {
	return a.client.Send(joystickPosition.LV, joystickPosition.LH, joystickPosition.RV, joystickPosition.RH)
}

func NewControllerConfig() ControllerConfig {
	cfg := ControllerConfig{
		Mass: 0.573, // Rotor Riot CL1
		G:    9.81,
		Hz:   10.0, // Set to 100Hz to get each 10 ms, like UDP sent by Liftoff

		Pos: PosGains{
			Kp:       Vec3{1.2, 1.2, 2.5},
			Kv:       Vec3{1.0, 1.0, 1.8},
			Ki:       Vec3{0.05, 0.05, 0.2},
			IntLimit: 2.0,
		},

		// Map attitude error -> rate commands (tune these)
		R2RateGain: Vec3{8.0, 8.0, 4.0}, // [1/s] per rad of error
		YawP:       0.8,                 // extra yaw P (optional)

		Rates: AcroRateLimits{
			MaxRollRate:  4.0, // rad/s (~230 deg/s)
			MaxPitchRate: 4.0,
			MaxYawRate:   3.0,
		},
		Throttle: ThrottleMap{
			HoverStick:       0.305180, // hover mid-stick (uncentered mode). Move started with 0.305180
			Slope:            0.5,      // 50% stick per 100% thrust delta around hover
			Centered:         false,    // We are not centered throttle because 0 is not 0 - it is 50% of max spin of motors... And we should convert [0,1] to [-32000,+32000]
			CenteredSpan:     1.0,      // used only if Centered=true
			Deadzone:         0.03,
			RateLimitPerTick: 0.05,
		},
		Signs: OutputSigns{
			InvertRoll:  false,
			InvertPitch: true, // set true if forward stick should mean nose-down
			InvertYaw:   false,
		},
	}
	cfg.Dt = 1.0 / cfg.Hz
	return cfg
}

const showDebugCalculation = false

func main() {
	flag.Usage = func() {
		fmt.Printf("Usage: %s [OPTIONS]\n", os.Args[0])
		flag.PrintDefaults()
	}
	doHelp := flag.Bool("help", false, "Prints help")
	doDebug := flag.Bool("debug", false, "Log debug messages")
	joystickServerAddress := flag.String("address", "127.0.0.1:9002", "Address of local UDP server to send joystick commands to")

	flag.Parse()
	if *doHelp {
		flag.Usage()
		os.Exit(1)
	}

	joy := &RealJoystick{}
	err := joy.client.Start(*joystickServerAddress, *doDebug)
	if err != nil {
		fmt.Printf("Cannot connect to joystick server on address %s: %v", *joystickServerAddress, err)
		os.Exit(1)
	}
	fmt.Printf("Connected to joystick control service on address %s\n", *joystickServerAddress)

	cfg := NewControllerConfig()

	tel := &LiftoffTelemetryProvider{
		TelemetryListener: &track.TelemetryListener{},
	}
	tel.TelemetryListener.Toggle()
	var startPosition Vec3
	for {
		dt, _, ok := tel.TelemetryListener.LastDatagram()
		if ok {
			startPosition = liftoffDatagramYZXToVec3(dt.Position)
			break
		}
		time.Sleep(100 * time.Millisecond)
	}

// When we run in debug, compilation takes enough time to switch to liftoff so no need to wait. When run built version - set sleepSecondsBeforeStart to 3 seconds
	sleepSecondsBeforeStart := 0
	desiredFront := 0
	desiredRight := 0
	desiredAltitude := 5
	maxTimeSeconds := 10
	calculateThrottleHover := false

	desiredIncrement := Vec3{float64(desiredFront), float64(desiredRight), float64(desiredAltitude)} // move from first successful telemetry
	positionDesired := add(startPosition, desiredIncrement)

	fmt.Printf("Init pos %v\n", startPosition)
	fmt.Printf("Goal pos %v\n", positionDesired)

	sp := &MockSetpoint{
		PositionDesired: positionDesired,
		PsiD:            0 * math.Pi / 180.0, // face 0 degree yaw
	}

	ctrl := NewController(cfg, tel, sp, joy)
	// To speedup initial movement, let's start with HoverStick value as Last value - so we do not waist 5 cycles on slowly growing throttle
	ctrl.state.lastLV = cfg.Throttle.HoverStick

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Before starting, sleep 3 seconds to switch to LiftOff
	fmt.Printf("Before navigation, sleep %d seconds to switch to Liftoff program\n", sleepSecondsBeforeStart)
	time.Sleep(time.Duration(sleepSecondsBeforeStart) * time.Second)

	// Before navigation, start props by setting throttle to min value for half second
	fmt.Println("Starting engine")
	joy.SendJoystick(JoystickPosition{math.MinInt16, 0, 0, 0})
	time.Sleep(500 * time.Millisecond)

	if calculateThrottleHover {
		fmt.Println("Throttle up until start moving and stay to define hover value")
		detectHoverThrottle(ctrl)
		fmt.Printf("Start navigation with throttle hover %f\n", ctrl.cfg.Throttle.HoverStick)
	}

	sp.Start = time.Now()

	// Stop after some time
	go func() {
		time.Sleep(time.Duration(maxTimeSeconds) * time.Second)
		fmt.Printf("Stop controller after %d seconds", maxTimeSeconds)
		cancel()
	}()
	if err := ctrl.Run(ctx); err != nil && err != context.Canceled {
		fmt.Println("controller stopped with error:", err)
	}
}

func detectHoverThrottle(c *Controller) error {
	throttle := int16(math.MinInt16)
	pos := JoystickPosition{}

	for {
		pos.LV = throttle
		if err := c.act.SendJoystick(pos); err != nil {
			return fmt.Errorf("joystick send: %w", err)
		}
		time.Sleep(10 * time.Millisecond)
		t, ok, _ := c.tprov.ReadTelemetry()
		if ok {
			if t.Position[2] > 0.01 {
				break
			}
		}
		throttle += 50
	}

	//time.Sleep(10 * time.Second)

	throttleHover := float64(throttle-math.MinInt16) / float64(math.MaxInt16-math.MinInt16)
	fmt.Printf("Started to move with throttle %d, set it to c.cfg.Throttle.HoverStick as %f\n", throttle, throttleHover)
	c.cfg.Throttle.HoverStick = throttleHover
	return nil
}
