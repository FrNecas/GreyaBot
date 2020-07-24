package twitch

type twitchWebHookHandler struct {
	msgChannel  chan string
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
		ID   string `json:"id"`
		Type string `json:"type"`
	} `json:"data"`
}

type userResponse struct {
	Data []struct {
		ID string `json:"id"`
	} `json:"data"`
}
