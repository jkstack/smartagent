package utils

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
)

type Version struct {
	data [3]int
}

func ParseVersion(str string) (Version, error) {
	var ret Version
	tmp := strings.SplitN(str, ".", 3)
	if len(tmp) != 3 {
		return ret, errors.New("invalid version")
	}
	n, err := strconv.ParseInt(tmp[0], 10, 64)
	if err != nil {
		return ret, errors.New("invalid major version")
	}
	ret.data[0] = int(n)
	n, err = strconv.ParseInt(tmp[1], 10, 64)
	if err != nil {
		return ret, errors.New("invalid minor version")
	}
	ret.data[1] = int(n)
	n, err = strconv.ParseInt(tmp[2], 10, 64)
	if err != nil {
		return ret, errors.New("invalid patch version")
	}
	ret.data[2] = int(n)
	return ret, nil
}

func (v Version) Greater(v2 Version) bool {
	for i := 0; i < 3; i++ {
		if v.data[i] > v2.data[i] {
			return true
		}
	}
	return false
}

func (v Version) Equal(v2 Version) bool {
	for i := 0; i < 3; i++ {
		if v.data[i] != v2.data[i] {
			return false
		}
	}
	return true
}

func (v Version) String() string {
	return fmt.Sprintf("%d.%d.%d", v.data[0], v.data[1], v.data[2])
}
