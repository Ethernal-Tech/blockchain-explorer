package utils

import (
	"log"
	"strconv"
)

func ToUint64(str string) uint64 {

	var res uint64
	var err error

	if str[0:2] == "0x" {
		if len(str) <= 2 {
			return 0
		}

		res, err = strconv.ParseUint(str[2:], 16, 64)
		if err != nil {
			log.Printf("Error converting %s to uint64. %s", str, err)
			return 0
		}
	} else {
		res, err = strconv.ParseUint(str, 10, 64)
		if err != nil {
			log.Printf("Error converting %s to uint64. %s", str, err)
			return 0
		}
	}
	return res
}
