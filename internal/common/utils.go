package common

import "strconv"

func ParseTgId(str string) (int64, error) {
	value, err := strconv.ParseInt(str, 10, 64)
	if err != nil {
		return 0, err
	}

	return value, nil
}
