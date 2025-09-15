package configs

import (
	"fmt"
	"regexp"
	"strconv"
)

type Config struct {
	DataDir            string // path to store database
	StoreName          string // database to use
	BadgerValueLogSize string // logsize for badgerdb
	PebbleCacheSize    string // pebbledb cache size
	JournalCacheSize   string // JournalCacheSize max amount of Item in cache
	Port               int    // Port to serve
	MsgSize            string // size of payload
	LogFile            string // location of configs

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
