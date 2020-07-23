// Contains utility functions for processing user messages
package message

import (
	"fmt"
	"github.com/bwmarrin/discordgo"
	"regexp"
	"strings"
)

func IsMaliciousMessage(s string, blockedRegExps []*regexp.Regexp) bool {
	s = strings.ToLower(s)
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
