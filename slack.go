package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/slack-go/slack"
	"github.com/slack-go/slack/slackevents"
)

func handleSlackEvents(w http.ResponseWriter, r *http.Request) {
	var api = slack.New(slackToken)

	buf := new(bytes.Buffer)
	buf.ReadFrom(r.Body)
	body := buf.String()

	eventsAPIEvent, err := slackevents.ParseEvent(json.RawMessage(body), slackevents.OptionNoVerifyToken())
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if eventsAPIEvent.Type == slackevents.URLVerification {
		var r *slackevents.ChallengeResponse
		err := json.Unmarshal([]byte(body), &r)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "text")
		w.Write([]byte(r.Challenge))
		return
	}

	if eventsAPIEvent.Type == slackevents.CallbackEvent {
		innerEvent := eventsAPIEvent.InnerEvent
		switch ev := innerEvent.Data.(type) {
		case *slackevents.AppMentionEvent:
			handleAppMentionEvent(api, ev)
		case *slackevents.MessageEvent:
			handleMessageEvent(w, api, ev)
		}
	}
	w.WriteHeader(http.StatusOK)
}

func handleAppMentionEvent(api *slack.Client, event *slackevents.AppMentionEvent) {
	channelID := event.Channel
	threadTimestamp := event.ThreadTimeStamp
	if threadTimestamp == "" {
		threadTimestamp = event.TimeStamp
	}

	response := "You mentioned me!"

	_, _, err := api.PostMessage(
		channelID,
		slack.MsgOptionText(response, false),
		slack.MsgOptionTS(threadTimestamp),
	)
	if err != nil {
		log.Printf("failed posting message: %v", err)
	}
}

func handleMessageEvent(w http.ResponseWriter, api *slack.Client, event *slackevents.MessageEvent) {
	if event.BotID != "" {
		return
	}

	if event.SubType == "message_deleted" || (event.SubType == "message_changed" && event.Message.SubType == "tombstone") {
		if event.PreviousMessage.BotID != "" {
			return
		}
		err := deleteBotMessage(event, api)
		if err != nil {
			log.Printf("failed deleting bot message: %v", err)
		}
		return
	}

	if event.SubType == "message_changed" && event.Message.SubType != "tombstone" {
		if event.PreviousMessage.BotID != "" {
			return
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
		go func() {
			err := modifyBotMessage(event, api)
			if err != nil {
				log.Printf("failed deleting bot message: %v", err)
			}
		}()
		return
	}

	m, err := getMessageMap(event.Channel, event.TimeStamp)
	if err != nil {
		log.Printf("failed getting message map: %v", err)
	}

	if m.UserTs != "" {
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))

	go func() {
		channelID := event.Channel
		threadTimestamp := event.ThreadTimeStamp
		if threadTimestamp == "" {
			threadTimestamp = event.TimeStamp
		}
		userTs := event.TimeStamp

		var response slack.MsgOption
		if event.ChannelType == "im" {
			response = slack.MsgOptionBlocks(
				slack.NewSectionBlock(
					&slack.TextBlockObject{Type: "plain_text", Text: " "},
					[]*slack.TextBlockObject{
						{Type: "plain_text", Text: "Direct message received!"},
					},
					nil,
				),
			)
			_, _, err := api.PostMessage(
				channelID,
				response,
				slack.MsgOptionTS(threadTimestamp),
			)
			if err != nil {
				log.Printf("failed posting message: %v", err)
			}
		} else if strings.Contains(event.Text, botName) {
			if strings.TrimSpace(strings.Trim(event.Text, botName)) == "reset" {
				err := deleteSetting(channelID)
				if err != nil {
					log.Printf("failed deleting channel setting: %v", err)
					return
				}
				response = slack.MsgOptionBlocks(
					slack.NewSectionBlock(
						&slack.TextBlockObject{Type: "plain_text", Text: " "},
						[]*slack.TextBlockObject{
							{Type: "plain_text", Text: "Setting successfully reset."},
						},
						nil,
					),
				)
				_, _, err = api.PostMessage(
					channelID,
					response,
					slack.MsgOptionTS(threadTimestamp),
				)
				if err != nil {
					log.Printf("failed posting message: %v", err)
				}
				return
			}
			langs := strings.Split(strings.Trim(event.Text, botName), ",")
			for i := range langs {
				langs[i] = strings.TrimSpace(langs[i])
				if !isSupportedLanguage(langs[i]) {
					fmt.Printf("%s is not supported\n", langs[i])
					response = slack.MsgOptionBlocks(
						slack.NewSectionBlock(
							nil,
							[]*slack.TextBlockObject{
								{Type: "plain_text", Text: "invalid command or language."},
							},
							nil,
						),
					)
					break
				}
			}
			if response == nil {
				m := ChannelSetting{
					Setting: langs,
				}
				err := insertSetting(channelID, m)
				if err != nil {
					log.Printf("failed setting channel setting: %v", err)
				}
				response = slack.MsgOptionBlocks(
					slack.NewSectionBlock(
						nil,
						[]*slack.TextBlockObject{
							{Type: "plain_text", Text: "good command"},
						},
						nil,
					),
				)
			}
			_, _, err := api.PostMessage(
				channelID,
				response,
				slack.MsgOptionTS(threadTimestamp),
			)
			if err != nil {
				log.Printf("failed posting message: %v", err)
			}
			return
		} else {
			err = insertMessageMap(channelID, userTs, "")
			if err != nil {
				log.Printf("failed inserting message map: %v", err)
			}
			channelSetting, err := getSetting(channelID)
			if err != nil {
				log.Printf("failed getting channel setting: %v", err)
			}

			response, err = createTranslatedMessage(event, channelSetting)
			if err != nil {
				log.Printf("failed creating translated message: %v", err)
				return
			}
			_, botTs, err := api.PostMessage(
				channelID,
				response,
				slack.MsgOptionTS(threadTimestamp),
			)
			if err != nil {
				log.Printf("failed posting message: %v", err)
			}
			fmt.Println(botTs)
			err = insertMessageMap(channelID, userTs, botTs)
			if err != nil {
				log.Printf("failed inserting message map: %v", err)
			}
		}
	}()
}

func createTranslatedMessage(event *slackevents.MessageEvent, setting ChannelSetting) (slack.MsgOption, error) {
	var prompt string
	var systemPrompt string
	if event.SubType == "message_changed" {
		prompt = event.Message.Text
	} else {
		prompt = event.Text
	}

	if setting.Setting == nil {
		systemPrompt = "あなたは言語翻訳のプロです。プロンプトにある文章を、プロンプトで指定した指定した言語に翻訳してください。もし与えられた文章が JP 言語の場合は、 EN 言語に翻訳してください。もし与えられた文章が EN 言語の場合は、 JP 言語に翻訳してください。翻訳した結果のみを表示してください。"

		responseContent, err := generateContentFromText(systemPrompt, prompt)
		if err != nil {
			log.Printf("failed generating content: %v", err)
			return nil, err
		}
		lang, err := detectLanguage(responseContent)
		if err != nil {
			log.Printf("failed detecting language: %v", err)
			return nil, err
		}

		return slack.MsgOptionBlocks(
			slack.NewSectionBlock(
				nil,
				[]*slack.TextBlockObject{
					{Type: "plain_text", Text: getFlagEmoji(lang) + "\n" + responseContent},
				},
				nil,
			),
		), nil
	} else {
		var bo []*slack.TextBlockObject
		promptLang, err := detectLanguage(prompt)
		if err != nil {
			log.Printf("failed detecting language: %v", err)
			return nil, err
		}
		for _, l := range setting.Setting {
			if promptLang == l {
				continue
			}
			systemPrompt = "あなたは言語翻訳のプロです。プロンプトにある文章を、 " + l + " に翻訳してください。翻訳した結果のみを表示してください。"
			responseContent, err := generateContentFromText(systemPrompt, prompt)
			if err != nil {
				log.Printf("failed generating content: %v", err)
				return nil, err
			}
			bo = append(bo, &slack.TextBlockObject{Type: "plain_text", Text: getFlagEmoji(l) + "\n" + responseContent})
		}
		return slack.MsgOptionBlocks(
			slack.NewSectionBlock(
				nil,
				bo,
				nil,
			),
		), nil
	}
}

func deleteBotMessage(event *slackevents.MessageEvent, api *slack.Client) error {
	m, err := getMessageMap(event.Channel, event.PreviousMessage.TimeStamp)
	if err != nil {
		log.Printf("failed getting message map: %v", err)
		return err
	}
	_, _, err = api.DeleteMessage(m.ChannelID, m.BotTs)
	if err != nil {
		log.Printf("failed getting message map: %v", err)
		return err
	}
	return nil
}

func modifyBotMessage(event *slackevents.MessageEvent, api *slack.Client) error {

	fmt.Println("modifying message...")

	m, err := getMessageMap(event.Channel, event.PreviousMessage.TimeStamp)
	if err != nil {
		log.Printf("failed getting message map: %v", err)
		return err
	}

	channelSetting, err := getSetting(event.Channel)
	if err != nil {
		log.Printf("failed getting channel setting: %v", err)
	}

	var response slack.MsgOption

	response, err = createTranslatedMessage(event, channelSetting)
	if err != nil {
		return err
	}

	_, _, _, err = api.UpdateMessage(m.ChannelID, m.BotTs, response)
	if err != nil {
		log.Printf("failed modifying message: %v", err)
		return err
	}
	return nil
}
