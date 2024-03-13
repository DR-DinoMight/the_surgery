package helpers

import (
	"fmt"
	"regexp"
)

func ProcessEmojis(message string) string {
	cfg := OpenConfigFile()

	// get the message and find all IMG tags and get the src url
	emojiRegex := regexp.MustCompile("<img[^>]*src=[\"']?([^\"^>]*)[\"']?[^>]*>")
	//replace each src with "https://stream.deloughry.co.uk/[EXISTING_URL]"
	processedMessage := emojiRegex.ReplaceAllString(message, fmt.Sprintf("<img src='%s$1' class='emoji'/>", cfg.Stream.URL))

	return processedMessage
}

func AddColourToUserName(userName string, displayColour int) string {
	//map the displayColour to the rgb files found in the config
	cfg := OpenConfigFile()
	//cast displayColour to string for lookup in map
	colourString := fmt.Sprintf("%d", displayColour)
	rgb := cfg.ChatColours[colourString]
	return fmt.Sprintf("<span style='color: rgb(%v, %v, %v);' class='font-bold text-md'>%s</span>", rgb[0], rgb[1], rgb[2], userName)
}
