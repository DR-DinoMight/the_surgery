package helpers

import (
	"log"
	"os"

	"github.com/pelletier/go-toml"
)

type StreamSiteSetting struct {
	URL         string `toml:"url"`
	EmojiWidth  int    `toml:"emojiWidth"`
	EmojiHeight int    `toml:"emojiHeight"`
}
type StreamSettings struct {
	ChatColours map[string][]int  `toml:"chatColours"`
	Stream      StreamSiteSetting `toml:"stream"`
}

func OpenConfigFile() StreamSettings {
	readFile, err := os.ReadFile("./data/config.toml")
	if err != nil {
		log.Fatal(err)
	}
	var cfg StreamSettings
	err = toml.Unmarshal(readFile, &cfg)
	if err != nil {
		log.Fatal(err)
	}
	return cfg
}

// func GetSettings() StreamSettings {

// 	return StreamSettings{}
// }
