package bot

import (
	"fmt"
	"github.com/FrNecas/GreyaBot/config"
	"github.com/FrNecas/GreyaBot/message"
	"github.com/FrNecas/GreyaBot/twitch"
	"github.com/bwmarrin/discordgo"
	"os"
	"os/signal"
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

func RunBot(msgChannel chan *discordgo.MessageSend) {
	// Create a new Discord session using the provided bot token.
	session, err := discordgo.New("Bot " + config.Config.Token)
	if err != nil {
		fmt.Println("Error creating Discord session,", err)
		return
	}
	addHandlers(session)
	u, errUser := session.User("@me")
	if errUser != nil {
		fmt.Println("Error getting BotID,", err)
		return
	}
	config.Config.BotID = u.ID

	// Open a websocket connection to Discord and begin listening.
	err = session.Open()
	if err != nil {
		fmt.Println("Error opening connection,", err)
		return
	}

	fmt.Println("Bot is now running.  Press CTRL-C to exit.")
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)
	for n := 1; n > 0; {
		select {
		case msg := <-msgChannel:
			_, err = session.ChannelMessageSendComplex(config.Config.NotificationChannelID, msg)
			if err != nil {
				fmt.Println("Error sending a notification,", err)
			}
		case <-sc:
			n--
		}
	}

	twitch.UnsubscribeAll()
	err = session.Close()
	if err != nil {
		fmt.Println("Error closing the session,", err)
	}
}

func HandleVerification(s *discordgo.Session, data *discordgo.MessageReactionAdd) {
	if data.MessageID != config.Config.RulesMessageID || data.UserID == config.Config.BotID {
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
	} else {
		err = s.MessageReactionAdd(config.Config.RulesChannelID, config.Config.RulesMessageID,
			config.Config.VerifyEmote)
		if err != nil {
			fmt.Println("Error adding verify reaction to rules message,", err)
		}
	}
}

func HandleNewMember(s *discordgo.Session, data *discordgo.GuildMemberAdd) {
	fmt.Println("A new member joined, sending him a welcome message")
	greeting := message.FormatWelcomeMessage(config.Config.GreetingMessage, data)
	_, err := s.ChannelMessageSend(config.Config.GreetingsChannelID, greeting)
	if err != nil {
		fmt.Println("Error sending a welcome message,", err)
	}
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
	if message.IsMaliciousMessage(data.Content, config.Config.BlockedRegexExps) {
		fmt.Println("Removing malicious message with this content,", data.Content)
		err := s.ChannelMessageDelete(data.ChannelID, data.ID)
		if err != nil {
			fmt.Println("Failed to delete a malicious message,", err)
		}
		sendAuthorWarning(s, data.Author.ID)
	}
}

func BlockUnwantedUpdatedMessages(s *discordgo.Session, data *discordgo.MessageUpdate) {
	if message.IsMaliciousMessage(data.Content, config.Config.BlockedRegexExps) {
		fmt.Println("Removing malicious message with the following content,", data.Content)
		err := s.ChannelMessageDelete(data.ChannelID, data.ID)
		if err != nil {
			fmt.Println("Failed to delete a malicious message,", err)
		}
		sendAuthorWarning(s, data.Author.ID)
	}
}
