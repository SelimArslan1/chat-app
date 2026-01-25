package websocket

type ClientEvent struct {
	Type      string `json:"type"`
	ChannelID string `json:"channel_id"`
	Content   string `json:"content"`
	ImageURL  string `json:"image_url,omitempty"`
}

type ServerEvent struct {
	Type    string      `json:"type"`
	Payload interface{} `json:"payload"`
}
