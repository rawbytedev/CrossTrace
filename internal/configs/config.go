package configs

import (
	"fmt"
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
	f, err := os.Open("configs.yaml")
	if err != nil {
		fmt.Errorf("%s", err.Error())
	}
	if f == nil {
		f, err = os.Create("configs.yaml")
		if err != nil {
			fmt.Errorf("%s", err.Error())
		}
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
			SolanaRPC:   "SOLANA_RPC",
			KeypairPath: "Path_TO_KEY_PAIR",
		},
		Minting: MintingConfig{
			CrossmintAPIKey:       "CROSSMINT_API",
			CrossmintBaseURL:      "", // empty for default
			CrossmintCollectionID: "",
			Recipient:             "",
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
	Port int
	MistralAi string
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
