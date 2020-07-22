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
	Token string `json:"token"`

	RulesChannelID     string `json:"rules_channel_id"`
	GreetingsChannelID string `json:"greetings_channel_id"`

	VerifyRoleID string `json:"verify_role_id"`

	RulesMessageID string `json:"rules_message_id"`

	// The greetings message may contain the following macros:
	//    $user which will be replaced with the new user tag
	//    $channel(id) which will be replaced with a link to channel with the id
	GreetingMessage string `json:"greeting_message"`

	VerifyEmote string `json:"verify_emote"`

	MessageBlockList []string `json:"message_block_list"`
	BlockedRegexExps []*regexp.Regexp
	MaxPaddingBlocked int `json:"max_padding_blocked"`
	WarningMessage string `json:"warning_message"`
}

func prepareBlockList() {
	paddingRegex := fmt.Sprintf(".{0,%d}", Config.MaxPaddingBlocked)
	for _, blockedWord := range Config.MessageBlockList {
		escaped := regexp.QuoteMeta(blockedWord)
		escaped = strings.ToLower(escaped)
		var buffer bytes.Buffer
		for _, character := range escaped {
			buffer.WriteRune(character)
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
