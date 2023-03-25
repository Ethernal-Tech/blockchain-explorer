package utils

import (
	"strconv"

	"github.com/sirupsen/logrus"
)

func ToUint64(str string) uint64 {
	if len(str) == 0 {
		return 0
	}

	var res uint64
	var err error

	if str[0:2] == "0x" {
		if len(str) <= 2 {
			return 0
		}

		res, err = strconv.ParseUint(str[2:], 16, 64)
		if err != nil {
			logrus.Panic("Error converting ", str, " to uint64, err: ", err)
			return 0
		}
	} else {
		res, err = strconv.ParseUint(str, 10, 64)
		if err != nil {
			logrus.Panic("Error converting ", str, " to uint64, err: ", str)
			return 0
		}
	}
	return res
}

func ToUint32(str string) uint32 {
	if len(str) == 0 {
		return 0
	}

	var res32 uint32
	var res64 uint64
	var err error

	if str[0:2] == "0x" {
		if len(str) <= 2 {
			return 0
		}

		res64, err = strconv.ParseUint(str[2:], 16, 32)
		res32 = uint32(res64)
		if err != nil {
			logrus.Panic("Error converting ", str, " to uint32, err: ", err)
			return 0
		}
	} else {
		res64, err = strconv.ParseUint(str, 10, 32)
		res32 = uint32(res64)
		if err != nil {
			logrus.Panic("Error converting ", str, " to uint32, err: ", str)
			return 0
		}
	}
	return res32
}
