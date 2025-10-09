package configs

import (
	"fmt"
	"log"
	"math"
	"os"
	"regexp"
	"strconv"
	"time"

	"gopkg.in/yaml.v3"
)

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
func GeneDefault() {
	f, err := os.Create("configs.yaml")
	if err != nil {
		log.Fatalf("%s", err.Error())
	}

	cfg := Configs{
		Journal: JournalConfig{
			CacheSize:   "10MB",
			DBPath:      "dbfolder",
			DBName:      "badgerdb",
			LogSize:     "10MB",
			EncoderName: "rlp",
			HasherName:  "sha256",
			MaxMsgSize:  30,
		},
		Anchor: AnchorConfig{
			SolanaRPC:   "https://api.devnet.solana.com",
			KeypairPath: "Path_TO_KEY_PAIR",
		},
		Minting: MintingConfig{
			CrossmintAPIKey:       "sk_staging_6CJFQGekazgd2bECdmNUF66m7JPD8Ev8JSZerTmSvKX6hAaPUL8jfeRBaaUqVLD1MprP9zgG64AedkkW3xzxe4LiZmWofxwX7KuuxXezvFU4bxBwiGLhkAUnptBZMS8EzFdRx4SrZ6545o1SbHyoS23xz6wNrqvCohx2Q6NwTcjTZx8uwYSm1Zozj3pyNVWzi96qKKFLjZuUQkSvC2DNGzj1",
			CrossmintBaseURL:      "", // empty for default
			CrossmintCollectionID: "cc222c91-a5b9-4bd5-8135-9ba5efc7512b",
			Recipient:             "email:radiationbolt@gmail.com:solana",
		},
		Server: ServerConfig{
			Port: 5555, // port of sse server / port to connect to
		},
	}
	data, err := yaml.Marshal(cfg)
	if err != nil {
		fmt.Printf("%s", err.Error())
	}
	f.Write(data)
}

// implement later
func validateConfig(cfg *Configs) error {
	// Add validation logic here
	return nil
}

// Journal-specific knobs
type JournalConfig struct {
	CacheSize   string // pebbledb cache size
	MaxMsgSize  int    // size of payload
	DBPath      string // path to store database
	DBName      string // database to use
	LogSize     string // logsize for badgerdb
	EncoderName string
	HasherName  string
	SafeMode    bool
}

// Batcher-specific knobs
// the Main will be in charge of it or a batcher package
// but it still best for main to make the calls
// work in progress not needed right now
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
	CrossmintAPIKey       string
	CrossmintCollectionID string
	CrossmintBaseURL      string
	Recipient             string
}

// Server config
// used by main to start server
type ServerConfig struct {
	Port      int
	MistralAi string
}

var sizeRegex = regexp.MustCompile(`^(\d+)([KMGTP]?B)$`)

func SafeUint64ToInt64(val uint64) (int64, error) {
	if val > math.MaxInt64 {
		return 0, fmt.Errorf("unable to convert: Possible Overflow")
	}
	return int64(val), nil
}
func ParseSize(sizeStr string) (int64, error) {
	matches := sizeRegex.FindStringSubmatch(sizeStr)
	if len(matches) != 3 {
		return 0, fmt.Errorf("invalid size format")
	}
	//instead of uint64 use int64
	base, err := strconv.ParseUint(matches[1], 10, 64)
	var intbase int64
	if v, ok := SafeUint64ToInt64(base); ok == nil {
		intbase = v
	} else {
		return 0, ok
	}
	if err != nil {
		return 0, err
	}

	switch matches[2] {
	case "KB":
		return intbase << 10, nil
	case "MB":
		return intbase << 20, nil
	case "GB":
		return intbase << 30, nil
	case "TB":
		return intbase << 40, nil
	case "PB":
		return intbase << 50, nil
	default:
		return intbase, nil
	}
}
