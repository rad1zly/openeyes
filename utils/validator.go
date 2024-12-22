package utils

import (
	"regexp"
	"strconv"
)

func IsNIK(input string) bool {
	if len(input) != 16 {
		return false
	}
	_, err := strconv.ParseInt(input, 10, 64)
	return err == nil
}

func IsPhone(input string) bool {
	phoneRegex := regexp.MustCompile(`^[1-9]\d{1,14}$`)
	return phoneRegex.MatchString(input)
}

func IsName(input string) bool {
	nameRegex := regexp.MustCompile(`^[a-zA-Z0-9\s_-]+$`)
	return nameRegex.MatchString(input)
}
