package main

import (
	"fmt"
	"math"
	"time"
)

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

func detectYawDirection(c *Controller, limitSeconds int) error {
	degLimit := 45
	fmt.Printf("Start to yaw clock-wise until %d degree then counter-wise to -%d and again for %d sec on some minimal altitude to see how rotation matrix is changed to confirm correct mapping of fields\n", degLimit, degLimit, limitSeconds)

	startDetection := time.Now()
	pos := JoystickPosition{}
	moveToStartDetectionAltitude(&pos, c)

	ts, _, _ := c.tprov.ReadTelemetry()
	fmt.Printf("Initial telemetry value: %s\n", ts)

	clockWise := true
	var yawValue float64
	for {
		yawValue = 0.1
		if !clockWise {
			yawValue = -yawValue
		}
		pos.LH = ToInt16Signed(yawValue)
		if err := c.act.SendJoystick(pos); err != nil {
			return fmt.Errorf("joystick send: %w", err)
		}
		time.Sleep(500 * time.Millisecond)
		t, _, _ := c.tprov.ReadTelemetry()
		fmt.Printf("yaw=% 5.2f, LH=% 6d  telemetry: %s \n", yawValue, pos.LH, t)

		degree := float32(RToEulerDegreeVec(t.Rotation)[2])

		if clockWise {
			if degree > float32(degLimit) {
				clockWise = false
			}
		} else {
			if degree < -float32(degLimit) {
				clockWise = true
			}
		}

		if time.Since(startDetection).Seconds() > float64(limitSeconds) {
			break
		}
	}

	// Reset to balanced position
	pos = JoystickPosition{}
	pos.LV = ToInt16Uncentered01(c.cfg.Throttle.HoverStick + 0.04)
	c.act.SendJoystick(pos)
	time.Sleep(1 * time.Second)

	return nil
}

func detectRollDirection(c *Controller, limitSeconds int) error {
	degLimit := 15
	fmt.Printf("Start to roll right until %d degrees then left to -%d and again for %d sec on some altitude to see how rotation matrix is changed to confirm correct mapping of fields\n", degLimit, degLimit, limitSeconds)

	startDetection := time.Now()
	pos := JoystickPosition{}
	moveToStartDetectionAltitude(&pos, c)

	ts, _, _ := c.tprov.ReadTelemetry()
	fmt.Printf("Initial telemetry value: %s\n", ts)

	right := true
	rollSpan := 0.03
	var rollValue float64
	for {
		if right {
			rollValue = rollSpan
		} else {
			rollValue = -rollSpan
		}
		pos.RH = ToInt16Signed(rollValue)
		if err := c.act.SendJoystick(pos); err != nil {
			return fmt.Errorf("joystick send: %w", err)
		}
		time.Sleep(500 * time.Millisecond)
		t, _, _ := c.tprov.ReadTelemetry()

		degree := float32(RToEulerDegreeVec(t.Rotation)[0])
		// IMPORTANT: We receive positive degree for left roll and negative for right roll, change sign for it
		degree = -degree

		fmt.Printf("roll=% 5.2f, RH=% 6d , deg=% 5.1f telemetry: %s \n", rollValue, pos.RH, degree, t)

		if right {
			if degree > float32(degLimit) {
				right = false
			}
		} else {
			if degree < float32(-degLimit) {
				right = true
			}
		}

		if time.Since(startDetection).Seconds() > float64(limitSeconds) {
			break
		}
	}

	stabilizeRollPitchAngle(pos, c)

	return nil
}

func detectPitchDirection(c *Controller, limitSeconds int) error {
	degLimit := 15
	fmt.Printf("Start to pitch forward until %d degrees then back to -%d and again for %d sec on some altitude to see how rotation matrix is changed to confirm correct mapping of fields\n", degLimit, degLimit, limitSeconds)

	startDetection := time.Now()
	pos := JoystickPosition{}
	moveToStartDetectionAltitude(&pos, c)

	ts, _, _ := c.tprov.ReadTelemetry()
	fmt.Printf("Initial telemetry value: %s\n", ts)

	forward := true
	pitchSpan := 0.03
	var pitchValue float64
	for {
		if forward {
			pitchValue = pitchSpan
		} else {
			pitchValue = -pitchSpan
		}
		pos.RV = ToInt16Signed(pitchValue)
		if err := c.act.SendJoystick(pos); err != nil {
			return fmt.Errorf("joystick send: %w", err)
		}
		time.Sleep(500 * time.Millisecond)
		t, _, _ := c.tprov.ReadTelemetry()

		degree := float32(RToEulerDegreeVec(t.Rotation)[1])

		fmt.Printf("roll=% 5.2f, RV=% 6d , deg=% 5.1f telemetry: %s \n", pitchValue, pos.RV, degree, t)

		if forward {
			if degree > float32(degLimit) {
				forward = false
			}
		} else {
			if degree < float32(-degLimit) {
				forward = true
			}
		}

		if time.Since(startDetection).Seconds() > float64(limitSeconds) {
			break
		}
	}

	stabilizeRollPitchAngle(pos, c)

	return nil
}

func moveToStartDetectionAltitude(pos *JoystickPosition, c *Controller) {
	ts, _, _ := c.tprov.ReadTelemetry()
	if ts.Position[2] < 5 {
		// To low, slowly go up
		pos.LV = ToInt16Uncentered01(c.cfg.Throttle.HoverStick + 0.1)
		c.act.SendJoystick(*pos)
		time.Sleep(1 * time.Second)
	}
	// Hover
	pos.LV = ToInt16Uncentered01(c.cfg.Throttle.HoverStick + 0.04)
	c.act.SendJoystick(*pos)
	time.Sleep(1 * time.Second)
}

func stabilizeRollPitchAngle(pos JoystickPosition, c *Controller) {
	pos = JoystickPosition{}
	iterations := 0
	for {
		t, _, _ := c.tprov.ReadTelemetry()

		vec := RToEulerDegreeVec(t.Rotation)

		rollDegree := float32(-vec[0])
		pitchDegree := float32(vec[1])

		if math.Abs(float64(rollDegree)) < 0.01 && math.Abs(float64(pitchDegree)) < 0.01 {
			fmt.Printf("Roll and pitch degree are almost 0, stop after %d iterations: %f %f \n", iterations, rollDegree, pitchDegree)
			break
		}
		pos.LV = ToInt16Uncentered01(c.cfg.Throttle.HoverStick + 0.04)
		rollValue := -rollDegree / 180
		pitchValue := -pitchDegree / 180
		pos.RH = ToInt16Signed(float64(rollValue))
		pos.RV = ToInt16Signed(float64(pitchValue))
		fmt.Printf("Stabilizing, now roll/pitch % 6.3f % 6.3f RH=% 6d RV=% 6d roll=%f, pitch=%f\n", rollDegree, pitchDegree, pos.RH, pos.RV, rollValue, pitchValue)
		c.act.SendJoystick(pos)
		time.Sleep(10 * time.Millisecond)
		iterations++
	}
}
