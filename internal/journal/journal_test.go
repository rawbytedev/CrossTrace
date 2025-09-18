package journal

import (
	"crosstrace/internal/configs"
	"testing"
)

func NewJournalConfig() *configs.JournalConfig {
	return &configs.JournalConfig{
		CacheSize:   "10MB",
		DBPath:      "dbfolder",
		DBName:      "badgerdb",
		LogSize:     "10MB",
		EncoderName: "rlp",
		HasherName:  "sha256",
	}
}
func TestJournalInsert(t *testing.T) {

}
