package main

type MessageMapping struct {
	ChannelID string `firestore:"channel_id"`
	UserTs    string `firestore:"user_ts"`
	BotTs     string `firestore:"bot_ts"`
}

type ChannelSetting struct {
	Setting []string `firestore:"setting"`
}
