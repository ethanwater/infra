package socket

import (
	"time"
	"errors"
)


func Time(timeFormat string) ([]byte, error) {
	switch timeFormat{
	case "RFCUTC":
		return TimeRFC3339UTC(), nil
	case "RFCLOCAL":
		return TimeRFC3339Local(), nil
	case "UNIXUTC":
		return TimeUnixUTC(), nil
	case "UNIXLOCAL":
		return TimeUnixLocal(), nil
	default:
		return []byte{}, errors.New("invalid time format")
	}
}

func TimeRFC3339UTC() []byte {
	return []byte(time.Now().UTC().Format(time.RFC3339))
}

func TimeRFC3339Local() []byte {
	return []byte(time.Now().Format(time.RFC3339))
}

func TimeUnixUTC() []byte {
	return []byte(time.Now().UTC().Format(time.UnixDate))
}

func TimeUnixLocal() []byte {
	return []byte(time.Now().Format(time.UnixDate))
}


