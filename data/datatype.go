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

const (
	_float32 int8 = 4
	_byte    int8 = 1
)

func CalculateBlockLength(fields []StreamDataType) int8 {
	var length int8 = 0
	for _, field := range fields {
		var fieldLength int8
		switch field {
		case Timestamp:
			fieldLength = _float32
		case Position, Velocity, Gyro:
			fieldLength = 3 * _float32
		case Attitude, Input:
			fieldLength = 4 * _float32
		case Battery:
			fieldLength = 2 * _float32
		case MotorRPM:
			fieldLength = _byte + 4*_float32
		}
		length += fieldLength
	}
	return length
}

func ParseStreamDataTypeFormats(formatNames []string) []StreamDataType {
	streamFormats := make([]StreamDataType, len(formatNames))
	for i, name := range formatNames {
		var ts StreamDataType
		switch name {
		case "Timestamp":
			ts = Timestamp
		case "Position":
			ts = Position
		case "Attitude":
			ts = Attitude
		case "Velocity":
			ts = Velocity
		case "Gyro":
			ts = Gyro
		case "Input":
			ts = Input
		case "Battery":
			ts = Battery
		case "MotorRPM":
			ts = MotorRPM
		default:
			ts = Unknown
		}
		streamFormats[i] = ts
	}
	return streamFormats
}
