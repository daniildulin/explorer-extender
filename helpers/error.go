package helpers

import "log"

func CheckErr(err error) {
	if err != nil {
		log.Println(err)
	}
}

func CheckErrBool(ok bool) {
	if !ok {
		log.Println(ok)
	}
}
