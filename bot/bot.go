package bot

import (
	"database/sql"
	"fmt"
	"github.com/FrNecas/GreyaBot/commands"
	"github.com/FrNecas/GreyaBot/config"
	"github.com/FrNecas/GreyaBot/db"
	"github.com/FrNecas/GreyaBot/message"
	"github.com/FrNecas/GreyaBot/twitch"
	"github.com/bwmarrin/discordgo"
	"os"
	"os/signal"
	"strings"
	"syscall"
)

func addHandlers(s *discordgo.Session) {
	// Declare intents and add handlers
	s.Identify.Intents = discordgo.MakeIntent(discordgo.IntentsGuildMembers |
		discordgo.IntentsGuildMessageReactions | discordgo.IntentsGuildMessages |
		discordgo.IntentsGuilds | discordgo.IntentsGuildVoiceStates |
		discordgo.IntentsGuildMembers | discordgo.IntentsGuildPresences)
	s.AddHandler(HandleVerification)
	s.AddHandler(HandleNewMember)
	s.AddHandler(MessageReceived)
	s.AddHandler(BlockUnwantedUpdatedMessages)
	s.AddHandler(VoiceUpdate)
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
	purgeDynamicChannels(session)

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

func MessageReceived(s *discordgo.Session, data *discordgo.MessageCreate) {
	if data.Author.ID == config.Config.BotID {
		return
	}

	if strings.HasPrefix(data.Content, config.Config.BotPrefix) {
		commands.Execute(s, data.Message)
		return
	}

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

func VoiceCountUsers(guild *discordgo.Guild, channel string) int {
	connected := 0
	for _, state := range guild.VoiceStates {
		if state.ChannelID == channel {
			connected++
		}
	}
	return connected
}

func setUpOverwrite(userID string) []*discordgo.PermissionOverwrite {
	var result []*discordgo.PermissionOverwrite
	perm := discordgo.PermissionOverwrite{
		ID:   userID,
		Type: discordgo.PermissionOverwriteTypeMember,
		Allow: discordgo.PermissionVoiceMoveMembers | discordgo.PermissionVoiceMuteMembers |
			discordgo.PermissionVoiceDeafenMembers,
	}
	result = append(result, &perm)
	return result
}

func createNewChannel(s *discordgo.Session, data *discordgo.VoiceStateUpdate, database *sql.DB) {
	// Get user nickname to be set as channel name
	user, err := s.User(data.UserID)
	if err != nil {
		fmt.Println("Failed to retrieve user object creating the channel,", err)
		return
	}
	fmt.Printf("Creating a channel for %s\n", user.Username)
	channelData := discordgo.GuildChannelCreateData{
		Name:                 user.Username,
		Type:                 discordgo.ChannelTypeGuildVoice,
		ParentID:             config.Config.VoiceCategoryID,
		PermissionOverwrites: setUpOverwrite(data.UserID),
	}
	channel, cerr := s.GuildChannelCreateComplex(data.GuildID, channelData)
	if cerr != nil {
		fmt.Println("Failed to set up a new voice channel,", err)
		return
	}
	_, ierr := database.Exec(`INSERT INTO voice_channels(owner_id, channel_id) VALUES ($1, $2)`,
		data.UserID, channel.ID)
	if ierr != nil {
		fmt.Println("Failed to insert a channel into database,", err)
		return
	}
	err = s.GuildMemberMove(channel.GuildID, data.UserID, &channel.ID)
	if err != nil {
		fmt.Println("Failed to move a user to a channel,", err)
	}

}

func isDynamicChannel(database *sql.DB, channelID string) bool {
	rows, qerr := database.Query(`SELECT id FROM voice_channels WHERE channel_id = $1`, channelID)
	if qerr != nil {
		fmt.Println("Error when querying for a voice channel,", qerr)
		return false
	}
	matching := 0
	for rows.Next() {
		matching++
	}
	return matching != 0
}

func removeChannel(s *discordgo.Session, data *discordgo.VoiceStateUpdate, database *sql.DB, guild *discordgo.Guild) {
	// Channel leave, check the previous channel and remove it if it's unused
	if isDynamicChannel(database, data.BeforeUpdate.ChannelID) &&
		VoiceCountUsers(guild, data.BeforeUpdate.ChannelID) == 0 {
		fmt.Printf("Deleting voice channel %s\n", data.BeforeUpdate.ChannelID)
		// Dynamic channel left empty, delete it
		_, err := s.ChannelDelete(data.BeforeUpdate.ChannelID)
		if err != nil {
			fmt.Println("Error when deleting a channel,", err)
		}
		_, err = database.Exec(`DELETE FROM voice_channels WHERE channel_id = $1`, data.BeforeUpdate.ChannelID)
		if err != nil {
			fmt.Println("Error when deleting a channel from a database,", err)
		}
	}
}

func purgeDynamicChannels(s *discordgo.Session) {
	channelUsage := make(map[string]int)
	for _, guild := range s.State.Guilds {
		// Set all voice channels to 0 usages
		channels, err := s.GuildChannels(guild.ID)
		if err != nil {
			fmt.Println("Couldn't get channels of a guild when pruning")
			return
		}
		for _, channel := range channels {
			if channel.Type == discordgo.ChannelTypeGuildVoice {
				channelUsage[channel.ID] = 0
			}
		}
		// Count the number of active voice connections
		for _, state := range guild.VoiceStates {
			channelUsage[state.ChannelID]++
		}
	}
	database, err := db.Connect()
	if err != nil {
		fmt.Println("Error when connecting to database while pruning")
		return
	}
	for channel, usages := range channelUsage {
		if usages == 0 && isDynamicChannel(database, channel) {
			fmt.Printf("Deleting voice channel %s\n", channel)
			_, err := s.ChannelDelete(channel)
			if err != nil {
				fmt.Println("Error when deleting a channel,", err)
			}
			_, err = database.Exec(`DELETE FROM voice_channels WHERE channel_id = $1`, channel)
			if err != nil {
				fmt.Println("Error when deleting a channel from a database,", err)
			}
		}
	}
}

func VoiceUpdate(s *discordgo.Session, data *discordgo.VoiceStateUpdate) {
	database, err := db.Connect()
	if err != nil {
		fmt.Println("Failed to connect to the database,", err)
		return
	}
	guild, err := s.Guild(data.GuildID)
	if err != nil {
		fmt.Println("Failed to obtain the guild object,", err)
		return
	}
	if data.BeforeUpdate != nil && data.BeforeUpdate.ChannelID != "" {
		removeChannel(s, data, database, guild)
	}
	if data.ChannelID == config.Config.MainVoiceID {
		// Create a new channel
		createNewChannel(s, data, database)
	}
}
