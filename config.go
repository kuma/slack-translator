package main

import "os"

var languageToFlag = map[string]string{
	"en":    "🇺🇸", // English
	"ja":    "🇯🇵", // Japanese
	"fr":    "🇫🇷", // French
	"de":    "🇩🇪", // German
	"es":    "🇪🇸", // Spanish
	"it":    "🇮🇹", // Italian
	"zh":    "🇨🇳", // Chinese
	"ko":    "🇰🇷", // Korean
	"zh-TW": "🇹🇼", // Chinese (Taiwan)
	"zh-CH": "🇨🇳", // Chinese (Chinese)
	"tr":    "🇹🇷", // Turkish
}

var supportedLanguages = map[string]string{
	"en": "English",
	"ja": "Japanese",
	"fr": "French",
	"de": "German",
	"es": "Spanish",
	"it": "Italian",
	"zh": "Chinese",
	"ko": "Korean",
	"tr": "Turkish",
}

var projectID = os.Getenv("PROJECT_ID")
var location = os.Getenv("LOCATION")
var modelName = os.Getenv("MODEL_NAME")
var slackToken = os.Getenv("BOT_TOKEN")
var botName = os.Getenv("BOT_NAME")
var port = os.Getenv("PORT")
