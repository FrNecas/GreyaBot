package config

import (
	"encoding/json"
	"io/ioutil"
)

var Config config

type config struct {
	Token string `json:"token"`

	GuildID string `json:"guild_id"`

	RulesChannelID     string `json:"rules_channel_id"`
	GreetingsChannelID string `json:"greetings_channel_id"`

	VerifyRoleID string `json:"verify_role_id"`

	RulesMessageID string `json:"rules_message_id"`
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
	return nil
}
