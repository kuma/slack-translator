package main

import (
	"context"
	"fmt"
	"strings"

	"cloud.google.com/go/translate"
	"cloud.google.com/go/vertexai/genai"
)

func generateContentFromText(systemP, p string) (string, error) {
	ctx := context.Background()
	client, err := genai.NewClient(ctx, projectID, location)
	if err != nil {
		return "", fmt.Errorf("error creating client: %w", err)
	}
	gemini := client.GenerativeModel(modelName)
	prompt := genai.Text(p)
	gemini.SystemInstruction = genai.NewUserContent(genai.Text(systemP))

	resp, err := gemini.GenerateContent(ctx, prompt)
	if err != nil {
		return "", fmt.Errorf("error generating content: %w", err)
	}
	// See the JSON response in
	// https://pkg.go.dev/cloud.google.com/go/vertexai/genai#GenerateContentResponse.
	return fmt.Sprintf("%v", resp.Candidates[0].Content.Parts[0]), nil
}

func detectLanguage(text string) (string, error) {
	ctx := context.Background()
	client, err := translate.NewClient(ctx)
	if err != nil {
		return "", fmt.Errorf("translate.NewClient: %w", err)
	}
	defer client.Close()
	lang, err := client.DetectLanguage(ctx, []string{text})
	if err != nil {
		return "", fmt.Errorf("DetectLanguage: %w", err)
	}
	if len(lang) == 0 || len(lang[0]) == 0 {
		return "", fmt.Errorf("DetectLanguage return value empty")
	}
	return lang[0][0].Language.String(), nil
}

func getFlagEmoji(languageCode string) string {
	languageCode = strings.ToLower(languageCode)
	if flag, exists := languageToFlag[languageCode]; exists {
		return flag
	}
	return "🏳️" // Default flag if language code is not found
}

func isSupportedLanguage(lang string) bool {
	_, exists := supportedLanguages[lang]
	return exists
}
