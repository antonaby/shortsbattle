package common

import (
	"errors"
	"strconv"
)

func ParseTgId(str string) (int64, error) {
	value, err := strconv.ParseInt(str, 10, 64)
	if err != nil {
		return 0, err
	}

	return value, nil
}

func IsServErr(err error) (ServiceError, bool) {
	var sErr ServiceError
	if errors.As(err, &sErr) {
		return sErr, true
	}

	return sErr, false
}
