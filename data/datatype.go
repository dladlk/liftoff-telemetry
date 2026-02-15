package lot_config

type StreamDataType int

const (
	Timestamp StreamDataType = iota
	Position
	Attitude
	Velocity
	Gyro
	Input
	Battery
	MotorRPM
	Unknown
)
