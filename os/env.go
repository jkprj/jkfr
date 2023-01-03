package os

import (
	"os"
	"strconv"
	"strings"
)

func getenv(key string) string {
	value := os.Getenv(key)
	return strings.TrimSpace(value)
}

func GetEnvString(key string, def string) string {
	v := getenv(key)
	if "" == v {
		return def
	}

	return v
}

func GetEnvStrings(key string, sep string, def []string) []string {

	values := []string{}
	tmps := strings.Split(getenv(key), sep)

	for _, str := range tmps {
		tmp := strings.TrimSpace(str)
		if "" != tmp {
			values = append(values, tmp)
		}
	}

	if 0 == len(values) && nil != def {
		return def
	}

	return values
}

func GetEnvInt(key string, def int) int {

	str := getenv(key)
	if "" == str {
		return def
	}

	i, err := strconv.Atoi(getenv(key))
	if nil != err {
		return def
	}

	return i
}

func GetEnvInts(key string, sep string, def []int) []int {
	strs := GetEnvStrings(key, sep, nil)
	values := []int{}

	for _, str := range strs {
		v, err := strconv.Atoi(str)
		if nil == err {
			values = append(values, v)
		}
	}

	if 0 == len(values) && nil != def {
		return def
	}

	return values
}

func GetEnvBool(key string, def bool) bool {

	str := strings.ToLower(getenv(key))

	if "true" == str {
		return true
	} else if "false" == str {
		return false
	}

	return def
}
