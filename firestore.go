package main

import (
	"context"
	"log"

	"cloud.google.com/go/firestore"
)

func insertMessageMap(channelID, userTs, botTs string) error {
	ctx := context.Background()

	client, err := firestore.NewClient(ctx, projectID)
	if err != nil {
		log.Fatalf("Firestore: Failed to create client: %v", err)
	}
	defer client.Close()

	m := MessageMapping{
		ChannelID: channelID,
		UserTs:    userTs,
		BotTs:     botTs,
	}

	_, err = client.Collection("messageMappings").Doc(channelID+":"+userTs).Set(ctx, m)
	if err != nil {
		log.Printf("Firestore: An error has occurred - cannot set property: %s", err)
	}

	return err
}

func getMessageMap(channelID, userTs string) (MessageMapping, error) {
	var m MessageMapping
	ctx := context.Background()

	client, err := firestore.NewClient(ctx, projectID)
	if err != nil {
		log.Fatalf("Firestore: Failed to create client: %v", err)
		return m, err
	}
	defer client.Close()

	doc, err := client.Collection("messageMappings").Doc(channelID + ":" + userTs).Get(ctx)
	if err != nil {
		log.Printf("Firestore: mapping not found: %s", err)
		return m, err
	}

	doc.DataTo(&m)
	return m, nil
}

func getSetting(channelID string) (ChannelSetting, error) {
	var m ChannelSetting
	ctx := context.Background()

	client, err := firestore.NewClient(ctx, projectID)
	if err != nil {
		log.Fatalf("Firestore: Failed to create client: %v", err)
	}
	defer client.Close()

	doc, err := client.Collection("channelSettings").Doc(channelID).Get(ctx)
	if err != nil {
		log.Printf("Firestore: setting not found: %s", err)
		return m, err
	}

	doc.DataTo(&m)
	return m, nil
}

func insertSetting(channelID string, setting ChannelSetting) error {

	ctx := context.Background()

	client, err := firestore.NewClient(ctx, projectID)
	if err != nil {
		log.Fatalf("Firestore: Failed to create client: %v", err)
	}
	defer client.Close()

	_, err = client.Collection("channelSettings").Doc(channelID).Set(ctx, setting)
	if err != nil {
		log.Printf("Firestore: An error has occurred - cannot insert setting: %s", err)
	}

	return err
}

func deleteSetting(channelID string) error {
	ctx := context.Background()

	client, err := firestore.NewClient(ctx, projectID)
	if err != nil {
		log.Fatalf("Firestore: Failed to create client: %v", err)
	}
	defer client.Close()

	_, err = client.Collection("channelSettings").Doc(channelID).Delete(ctx)
	if err != nil {
		log.Printf("Firestore: An error has occurred - cannot delete setting: %s", err)
	}

	return err
}
