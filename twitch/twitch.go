// Contains an implementation of a web server communicating with
// twitch using webhooks
package twitch

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/FrNecas/GreyaBot/config"
	"github.com/bwmarrin/discordgo"
	"io"
	"io/ioutil"
	"math/rand"
	"net/http"
	"strconv"
	"strings"
	"time"
)

const handlerPath = "/twitch/streams"
const liveStream = "live"
const thumbnailWidth = "320"
const thumbnailHeight = "180"
const subscribeFor = 86400
const refreshSubAfter = subscribeFor / 4 * 3

func authMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Only POST requests have the signature
		if r.Method == http.MethodPost {
			sig := hmac.New(sha256.New, []byte(config.Config.TwitchSubscribeSecret))
			payload, _ := ioutil.ReadAll(r.Body)
			// We've read the body, need to refresh it for the next middleware/handler
			err := r.Body.Close()
			if err != nil {
				fmt.Println("Failed to close request.Body,", err)
			}
			r.Body = ioutil.NopCloser(bytes.NewBuffer(payload))
			sig.Write(payload)
			sha := hex.EncodeToString(sig.Sum(nil))
			if "sha256="+sha != r.Header.Get("X-Hub-Signature") {
				http.Error(w, "Checksum not matching", http.StatusBadRequest)
			}
		}
		next.ServeHTTP(w, r)
	})
}

func getGameName(gameID string) string {
	if gameID == "" {
		return ""
	}
	client := http.Client{Timeout: time.Second * 2}
	endpoint := fmt.Sprintf("https://api.twitch.tv/helix/games?id=%s", gameID)
	req := createAuthenticatedRequest(http.MethodGet, endpoint, nil)
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("Error occurred when getting game data,", err)
		return ""
	}
	if resp != nil {
		defer resp.Body.Close()
	}
	var gameData gameResponse
	json.NewDecoder(resp.Body).Decode(&gameData)
	return gameData.Data[0].Name
}

func (h *twitchWebHookHandler) createEmbed(data streamResponse, endpoint string) *discordgo.MessageSend {
	streamer, ok := config.Config.EndpointToStreamer[endpoint]
	if !ok {
		return nil
	}
	thumbnail := strings.ReplaceAll(data.Data[0].ThumbnailURL, "{width}", thumbnailWidth)
	embed := &discordgo.MessageEmbed{
		Author: &discordgo.MessageEmbedAuthor{
			Name:    streamer.DisplayName,
			IconURL: streamer.ProfileImageURL,
		},
		Color: 0x6441a5,
		Fields: []*discordgo.MessageEmbedField{
			&discordgo.MessageEmbedField{
				Name:   "Game",
				Value:  getGameName(data.Data[0].GameID),
				Inline: true,
			},
			&discordgo.MessageEmbedField{
				Name:   "Viewers",
				Value:  strconv.Itoa(data.Data[0].ViewerCount),
				Inline: true,
			},
		},
		Image: &discordgo.MessageEmbedImage{
			URL: strings.ReplaceAll(thumbnail, "{height}", thumbnailHeight),
		},
		Thumbnail: &discordgo.MessageEmbedThumbnail{
			URL: streamer.ProfileImageURL,
		},
		Timestamp: time.Now().Format(time.RFC3339),
		Title:     data.Data[0].Title,
	}
	return &discordgo.MessageSend{Content: streamer.Start, Embed: embed}
}

func (h *twitchWebHookHandler) processNotification(r *http.Request) {
	query := r.URL.Query()
	var streamData streamResponse
	err := json.NewDecoder(r.Body).Decode(&streamData)
	if err != nil {
		return
	}
	if endpoint, ok := query["subscription"]; ok {
		if len(endpoint) == 0 {
			return
		}
		streamer, ok := config.Config.EndpointToStreamer[endpoint[0]]
		if !ok {
			return
		}
		if len(streamData.Data) == 0 {
			// Stream offline
			streamer.LastOnline = time.Now()
			msg := streamer.End
			if msg != "" {
				h.msgChannel <- &discordgo.MessageSend{Content: streamer.End}
			}
			return
		}

		// Check that we haven't already received this notification
		if _, ok := h.receivedIDs[streamData.Data[0].ID]; ok {
			return
		}
		h.receivedIDs[streamData.Data[0].ID] = true
		if streamData.Data[0].Type == liveStream {
			// Stream online
			sinceLastStream := time.Now().Sub(streamer.LastOnline)
			if sinceLastStream < time.Duration(config.Config.RestartCoolDown)*time.Minute {
				fmt.Printf("Detected a stream restart for %s\n", streamer.Name)
				return
			}
			if msg := h.createEmbed(streamData, endpoint[0]); msg != nil {
				h.msgChannel <- msg
			}
		}
	}
}

func (h *twitchWebHookHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		query := r.URL.Query()
		if reason, ok := query["hub.reason"]; ok {
			fmt.Println("Subscription denied,", reason)
			return
		}
		if challenge, ok := query["hub.challenge"]; ok {
			// respond with the challenge token
			w.Header().Set("Content-Type", "text/plain")
			w.WriteHeader(http.StatusOK)
			if len(challenge) == 0 {
				http.Error(w, "No challenge token provided", http.StatusInternalServerError)
				return
			}
			w.Write([]byte(challenge[0]))
		} else {
			http.Error(w, "No hub challenge", http.StatusBadRequest)
		}
	case http.MethodPost:
		w.Write([]byte("OK"))
		h.processNotification(r)
	default:
		http.Error(w, "Invalid method", http.StatusMethodNotAllowed)
	}
}

func getOAuthToken() {
	client := http.Client{Timeout: time.Second * 2}
	req, _ := http.NewRequest(http.MethodPost,
		fmt.Sprintf("https://id.twitch.tv/oauth2/token?client_id=%s&client_secret=%s&grant_type=client_credentials",
			config.Config.TwitchClientID, config.Config.TwitchClientSecret), nil)
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("Error occurred when getting a oauth token,", err)
		return
	}
	if resp != nil {
		defer resp.Body.Close()
	}
	err = json.NewDecoder(resp.Body).Decode(&config.Config.TwitchOAuth)
	var sleep time.Duration
	if err != nil {
		fmt.Println("Couldn't decode json from OAuth request", err)
		sleep = time.Minute
	} else {
		fmt.Printf("Got OAuth token expiring in %d seconds\n", config.Config.TwitchOAuth.ExpiresIn)
		sleep = time.Duration(config.Config.TwitchOAuth.ExpiresIn/4*3) * time.Second
	}
	go func() {
		time.Sleep(sleep)
		getOAuthToken()
	}()
}

func UnsubscribeAll() {
	for endpointID, streamer := range config.Config.EndpointToStreamer {
		subscribe(endpointID, streamer.ID, "unsubscribe")
	}
}

func subscribe(callbackID string, userID string, mode string) {
	callback := fmt.Sprintf("%s%s?subscription=%s", config.Config.TwitchBaseURL,
		handlerPath, callbackID)
	subData := subRequest{
		Callback:     callback,
		Mode:         mode,
		Topic:        fmt.Sprintf("https://api.twitch.tv/helix/streams?user_id=%s", userID),
		LeaseSeconds: subscribeFor,
		Secret:       config.Config.TwitchSubscribeSecret,
	}
	datajson, err := json.Marshal(subData)
	if err != nil {
		fmt.Println("Failed to marshal subscribe data")
		return
	}

	fmt.Printf("%s to %s on %s\n", mode, subData.Topic, callback)
	client := http.Client{Timeout: time.Second * 2}
	reqSub := createAuthenticatedRequest(http.MethodPost, "https://api.twitch.tv/helix/webhooks/hub",
		bytes.NewBuffer(datajson))
	reqSub.Header.Set("Content-Type", "application/json")
	resp, _ := client.Do(reqSub)
	if resp.StatusCode != http.StatusAccepted {
		fmt.Println("Subscription didn't respond with 202")
	}
	if mode == "subscribe" {
		go func() {
			time.Sleep(refreshSubAfter * time.Second)
			subscribe(callbackID, userID, mode)
		}()
	}
}

func getUserData(name string) userResponse {
	client := http.Client{Timeout: time.Second * 2}
	endpoint := fmt.Sprintf("https://api.twitch.tv/helix/users?login=%s", name)
	req := createAuthenticatedRequest(http.MethodGet, endpoint, nil)
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("Error occurred when getting user data,", err)
	}
	if resp != nil {
		defer resp.Body.Close()
	}
	var userData userResponse
	json.NewDecoder(resp.Body).Decode(&userData)

	return userData
}

func createAuthenticatedRequest(method string, target string, body io.Reader) *http.Request {
	req, _ := http.NewRequest(method, target, body)
	req.Header.Add("Client-ID", config.Config.TwitchClientID)
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", config.Config.TwitchOAuth.AccessToken))
	return req
}

func StartServerGoroutine(msgChannel chan *discordgo.MessageSend) {
	getOAuthToken()
	for i, streamer := range config.Config.Streamers {
		data := getUserData(streamer.Name)
		if len(data.Data) == 0 {
			continue
		}
		config.Config.Streamers[i].ID = data.Data[0].ID
		config.Config.Streamers[i].DisplayName = data.Data[0].DisplayName
		config.Config.Streamers[i].ProfileImageURL = data.Data[0].ProfileImageURL
		subNum := 0
		for subNum == 0 {
			newNum := rand.Int()
			if _, ok := config.Config.EndpointToStreamer[string(newNum)]; ok {
				continue
			}
			subNum = newNum
		}
		config.Config.EndpointToStreamer[strconv.Itoa(subNum)] = &config.Config.Streamers[i]
		subscribe(strconv.Itoa(subNum), data.Data[0].ID, "subscribe")
	}
	mux := http.NewServeMux()
	handler := twitchWebHookHandler{msgChannel: msgChannel}
	handler.receivedIDs = make(map[string]bool)
	mux.Handle(handlerPath, authMiddleware(&handler))
	go http.ListenAndServe(config.Config.TwitchServerAddress, mux)
}
