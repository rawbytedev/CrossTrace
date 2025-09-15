package configs

import (
	"fmt"
	"regexp"
	"strconv"
)

type Config struct {
	DataDir            string
	StoreName          string
	BadgerValueLogSize string
	PebbleCacheSize    string
	JournalCacheSize   string
	Port               int
	MsgSize            string
	logFile            string
}

var sizeRegex = regexp.MustCompile(`^(\d+)([KMGTP]?B)$`)

func ParseSize(sizeStr string) (uint64, error) {
	matches := sizeRegex.FindStringSubmatch(sizeStr)
	if len(matches) != 3 {
		return 0, fmt.Errorf("invalid size format")
	}

	base, err := strconv.ParseUint(matches[1], 10, 64)
	if err != nil {
		return 0, err
	}

	switch matches[2] {
	case "KB":
		return base << 10, nil
	case "MB":
		return base << 20, nil
	case "GB":
		return base << 30, nil
	case "TB":
		return base << 40, nil
	case "PB":
		return base << 50, nil
	default:
		return base, nil
	}
}
