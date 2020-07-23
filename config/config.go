package config

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"regexp"
	"strings"
)

var Config config

type config struct {
	// Discord API token
	Token string `json:"token"`

	// ID of the channel containing server rules
	RulesChannelID string `json:"rules_channel_id"`
	// ID of the channel where users are greeted
	GreetingsChannelID string `json:"greetings_channel_id"`

	// ID of the verify role which is granted upon verification
	VerifyRoleID string `json:"verify_role_id"`

	// ID of the message containing rules (that the user has to add emote to)
	RulesMessageID string `json:"rules_message_id"`

	// The greetings message may contain the following macros:
	//    $user which will be replaced with the new user tag
	//    $channel(id) which will be replaced with a link to channel with the id
	GreetingMessage string `json:"greeting_message"`

	// The emote used for verification
	VerifyEmote string `json:"verify_emote"`

	// List of blocked expressions
	MessageBlockList []string `json:"message_block_list"`
	BlockedRegexExps []*regexp.Regexp
	// Maximum number of extra characters between each character of blocked words
	// that will be blocked
	MaxPaddingBlocked int `json:"max_padding_blocked"`
	// A message that is sent to a user violating rules
	WarningMessage string `json:"warning_message"`
}

// Prepares RegExps for blocking malicious messages
func prepareBlockList() {
	paddingRegex := fmt.Sprintf(".{0,%d}", Config.MaxPaddingBlocked)
	for _, blockedWord := range Config.MessageBlockList {
		escaped := regexp.QuoteMeta(blockedWord)
		escaped = strings.ToLower(escaped)
		var buffer bytes.Buffer
		for _, character := range escaped {
			buffer.WriteRune(character)
			// Some characters may be escaped, we don't want to separate the escaped
			// character and the backslash by padding
			if character != '\\' {
				buffer.WriteString(paddingRegex)
			}
		}
		Config.BlockedRegexExps = append(Config.BlockedRegexExps, regexp.MustCompile(buffer.String()))
	}
}

func ReadConfig() error {
	data, err := ioutil.ReadFile("./config.json")
	if err != nil {
		return err
	}
	err = json.Unmarshal(data, &Config)
	if err != nil {
		return err
	}
	prepareBlockList()
	return nil
}
