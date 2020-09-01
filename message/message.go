// Contains utility functions for processing user messages
package message

import (
	"fmt"
	"github.com/bwmarrin/discordgo"
	"regexp"
	"strings"
)

var diacriticsReplacement = map[rune]rune{
	'ě': 'e',
	'š': 's',
	'č': 'c',
	'ř': 'r',
	'ž': 'z',
	'ý': 'y',
	'á': 'a',
	'í': 'i',
	'é': 'e',
	'ú': 'u',
	'ů': 'u',
	'ó': 'o',
	'ť': 't',
	'ď': 'd',
	'ň': 'n',
}

func IsMaliciousMessage(s string, blockedRegExps []*regexp.Regexp) bool {
	s = RemoveDiacritics(strings.ToLower(s))
	for _, regex := range blockedRegExps {
		if regex.MatchString(s) {
			return true
		}
	}
	return false
}

func FormatWelcomeMessage(welcomeMessage string, data *discordgo.GuildMemberAdd) string {
	// Replace $user with tag of the new member
	userRegex := regexp.MustCompile(`\$user`)
	userTag := fmt.Sprintf("<@%s>", data.User.ID)
	res := userRegex.ReplaceAllString(welcomeMessage, userTag)

	// Replace $channel(x) with a reference to channel x
	channelRegex := regexp.MustCompile(`\$channel\((.+)\)`)
	res = channelRegex.ReplaceAllString(res, "<#$1>")
	return res
}

func RemoveDiacritics(s string) string {
	out := []rune(s)
	for i, char := range out {
		if val, ok := diacriticsReplacement[char]; ok {
			out[i] = val
		}
	}
	return string(out)
}
