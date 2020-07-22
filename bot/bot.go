package bot

import (
	"fmt"
	"github.com/FrNecas/GreyaBot/config"
	"github.com/bwmarrin/discordgo"
	"os"
	"os/signal"
	"regexp"
	"strings"
	"syscall"
)

func addHandlers(s *discordgo.Session) {
	// Declare intents and add handlers
	s.Identify.Intents = discordgo.MakeIntent(discordgo.IntentsGuildMembers |
		discordgo.IntentsGuildMessageReactions | discordgo.IntentsGuildMessages)
	s.AddHandler(HandleVerification)
	s.AddHandler(HandleNewMember)
	s.AddHandler(BlockUnwantedNewMessages)
	s.AddHandler(BlockUnwantedUpdatedMessages)
}

func RunBot() {
	// Create a new Discord session using the provided bot token.
	session, err := discordgo.New("Bot " + config.Config.Token)
	if err != nil {
		fmt.Println("Error creating Discord session,", err)
		return
	}
	addHandlers(session)

	// Open a websocket connection to Discord and begin listening.
	err = session.Open()
	if err != nil {
		fmt.Println("Error opening connection,", err)
		return
	}

	// Wait here until CTRL-C or other term signal is received.
	fmt.Println("Bot is now running.  Press CTRL-C to exit.")
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)
	<-sc

	err = session.Close()
	if err != nil {
		fmt.Println("Error closing the session,", err)
	}
}

func HandleVerification(s *discordgo.Session, data *discordgo.MessageReactionAdd) {
	if data.MessageID != config.Config.RulesMessageID {
		return
	}
	if data.Emoji.Name == config.Config.VerifyEmote {
		fmt.Printf("Adding verify role to %s\n", data.UserID)
		err := s.GuildMemberRoleAdd(data.GuildID, data.UserID, config.Config.VerifyRoleID)
		if err != nil {
			fmt.Println("Error adding verify role,", err)
		}
	}
	err := s.MessageReactionsRemoveAll(config.Config.RulesChannelID, config.Config.RulesMessageID)
	if err != nil {
		fmt.Println("Error removing reactions from message after verification,", err)
	}
}

func formatWelcomeMessage(data *discordgo.GuildMemberAdd) string {
	userRegex := regexp.MustCompile(`\$user`)
	userTag := fmt.Sprintf("<@%s>", data.User.ID)
	res := userRegex.ReplaceAllString(config.Config.GreetingMessage, userTag)
	channelRegex := regexp.MustCompile(`\$channel\((.+)\)`)
	res = channelRegex.ReplaceAllString(res, "<#$1>")
	return res
}

func HandleNewMember(s *discordgo.Session, data *discordgo.GuildMemberAdd) {
	fmt.Println("A new member joined, sending him a welcome message")
	greeting := formatWelcomeMessage(data)
	_, err := s.ChannelMessageSend(config.Config.GreetingsChannelID, greeting)
	if err != nil {
		fmt.Println("Error sending a welcome message,", err)
	}
}

func isMaliciousMessage(s string) bool {
	s = strings.ToLower(s)
	for _, regex := range config.Config.BlockedRegexExps {
		if regex.MatchString(s) {
			return true
		}
	}
	return false
}

func sendAuthorWarning(s *discordgo.Session, userID string) {
	channel, err := s.UserChannelCreate(userID)
	if err != nil {
		fmt.Println("Failed to create user channel while sending a warning,", err)
	}
	_, errSend := s.ChannelMessageSend(channel.ID, config.Config.WarningMessage)
	if errSend != nil {
		fmt.Printf("Error sending a warning message to %s, %s\n", userID, errSend)
	}
}

func BlockUnwantedNewMessages(s *discordgo.Session, data *discordgo.MessageCreate) {
	if isMaliciousMessage(data.Content) {
		fmt.Println("Removing malicious message with this content,", data.Content)
		err := s.ChannelMessageDelete(data.ChannelID, data.ID)
		if err != nil {
			fmt.Println("Failed to delete a malicious message,", err)
		}
		sendAuthorWarning(s, data.Author.ID)
	}
}

func BlockUnwantedUpdatedMessages(s *discordgo.Session, data *discordgo.MessageUpdate) {
	if isMaliciousMessage(data.Content) {
		fmt.Println("Removing malicious message with the following content,", data.Content)
		err := s.ChannelMessageDelete(data.ChannelID, data.ID)
		if err != nil {
			fmt.Println("Failed to delete a malicious message,", err)
		}
		sendAuthorWarning(s, data.Author.ID)
	}
}
