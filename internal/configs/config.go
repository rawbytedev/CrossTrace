package configs

import (
	"fmt"
	"os"
	"regexp"
	"strconv"
	"time"

	"gopkg.in/yaml.v2"
)

// To remove
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

// Root config loaded at startup
type Configs struct {
	Journal JournalConfig
	Batcher BatcherConfig
	Anchor  AnchorConfig
	Minting MintingConfig
	Server  ServerConfig
}

func LoadConfig(path string) (*Configs, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	var cfg Configs
	dec := yaml.NewDecoder(f)
	if err := dec.Decode(&cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}

func validateConfig(cfg *Configs) error {
	// Add validation logic here
	return nil
}

// Journal-specific knobs
type JournalConfig struct {
	CacheSize   string
	MaxMsgSize  int
	DBPath      string
	DBName      string
	LogSize     string
	EncoderName string
	HasherName  string
	SafeMode    bool
}

// Batcher-specific knobs
// the Main will be in charge of it or a batcher package
// but it still best for main to make the calls
type BatcherConfig struct {
	Depth     int
	MaxLeaves int
	MaxWindow time.Duration
}

// Anchor-specific knobs
// used to anchor onto blockchain
type AnchorConfig struct {
	SolanaRPC   string
	KeypairPath string
}

// Minting-specific knobs
// minting on crossmint
type MintingConfig struct {
	CrossmintAPIKey    string
	CrossmintProjectID string
	CrossmintBaseURL   string
	Recipient          string
}

// Server config
// used by main to start server
type ServerConfig struct {
	Port int
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
