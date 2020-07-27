package twitch

import "github.com/bwmarrin/discordgo"

type twitchWebHookHandler struct {
	msgChannel  chan *discordgo.MessageSend
	receivedIDs map[string]bool
}

type subRequest struct {
	Callback     string `json:"hub.callback"`
	Mode         string `json:"hub.mode"`
	Topic        string `json:"hub.topic"`
	LeaseSeconds int    `json:"hub.lease_seconds"`
	Secret       string `json:"hub.secret"`
}

// A struct representing twitch response containing fields needed for
// processing in the bot
type streamResponse struct {
	Data []struct {
		ID           string `json:"id"`
		GameID       string `json:"game_id,omitempty"`
		Title        string `json:"title"`
		ViewerCount  int    `json:"viewer_count"`
		ThumbnailURL string `json:"thumbnail_url"`
		Type         string `json:"type"`
	} `json:"data"`
}

type userResponse struct {
	Data []struct {
		ID              string `json:"id"`
		DisplayName     string `json:"display_name,omitempty"`
		ProfileImageURL string `json:"profile_image_url,omitempty"`
	} `json:"data"`
}

type gameResponse struct {
	Data []struct {
		Name string `json:"name"`
	}
}
