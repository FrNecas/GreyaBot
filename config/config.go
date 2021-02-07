package config

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/FrNecas/GreyaBot/message"
	"io/ioutil"
	"regexp"
	"strings"
	"time"
)

var Config config

type streamerConfig struct {
	Name            string `json:"name"`
	ID              string
	Start           string `json:"start"`
	End             string `json:"end"`
	ProfileImageURL string
	DisplayName     string
	LastOnline      time.Time
}

type OAuthToken struct {
	AccessToken string `json:"access_token"`
	ExpiresIn   int    `json:"expires_in"`
}

type SimpleCommand struct {
	Command  string `json:"command"`
	Reaction string `json:"reaction"`
}

type config struct {
	// Discord API token
	Token   string `json:"token"`
	BotID   string
	AdminID string `json:"admin_id"`
	// Prefix for bot commands
	BotPrefix string `json:"bot_prefix"`
	// Postgresql connection string
	PsqlConnection string `json:"psql_connection"`

	// ID of the channel containing server rules
	RulesChannelID string `json:"rules_channel_id"`
	// ID of the channel where users are greeted
	GreetingsChannelID string `json:"greetings_channel_id"`
	// ID of the channel for twitch notifications
	NotificationChannelID string `json:"notification_channel_id"`

	// ID of the verify role which is granted upon verification
	VerifyRoleID string `json:"verify_role_id"`

	// ID of the message containing rules (that the user has to add emote to)
	RulesMessageID string `json:"rules_message_id"`

	// The greetings message may contain the following macros:
	//    $user which will be replaced with the new user tag
	//    $channel(id) which will be replaced with a link to channel with the id
	GreetingMessage       string `json:"greeting_message"`
	UnknownCommandMessage string `json:"unknown_command_message"`

	// Predefined command reactions. Do not include the BotPrefix in the
	// command name.
	PredefinedCommands []SimpleCommand `json:"predefined_commands"`

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

	TwitchClientID     string `json:"twitch_client_id"`
	TwitchClientSecret string `json:"twitch_client_secret"`
	// Secret used for payload verification
	TwitchSubscribeSecret string `json:"twitch_subscribe_secret"`
	TwitchServerAddress   string `json:"twitch_server_address"`
	TwitchBaseURL         string `json:"twitch_base_url"`
	TwitchOAuth           OAuthToken
	Streamers             []streamerConfig `json:"streamers"`
	EndpointToStreamer    map[string]*streamerConfig
	// Number of minutes after which a stream start is considered a restart and is ignored
	RestartCoolDown int `json:"restart_cool_down"`

	// Voice-related
	MainVoiceID              string `json:"main_voice_id"`
	VoiceCategoryID          string `json:"voice_category_id"`
	NoChannelOwnedMessage    string `json:"no_channel_owned_message"`
	LimitMustBeNumberMessage string `json:"limit_must_be_number_message"`
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

// Preprocesses commands - removes the diacritics in their names
func removeCommandDiacritics() {
	for i, cmd := range Config.PredefinedCommands {
		Config.PredefinedCommands[i].Command = message.RemoveDiacritics(cmd.Command)
	}
}

func ReadConfig() error {
	data, err := ioutil.ReadFile("./config.json")
	Config.EndpointToStreamer = make(map[string]*streamerConfig)
	if err != nil {
		return err
	}
	err = json.Unmarshal(data, &Config)
	if err != nil {
		return err
	}
	prepareBlockList()
	removeCommandDiacritics()
	return nil
}
