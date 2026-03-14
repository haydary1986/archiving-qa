package services

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/haydary1986/archiving-qa/internal/config"
	"github.com/haydary1986/archiving-qa/internal/models"
)

// AIService provides AI-powered document analysis using multiple providers.
type AIService struct {
	provider string
	apiKey   string
	baseURL  string
	model    string
	client   *http.Client
}

// NewAIService creates a new AIService from the given AI configuration.
func NewAIService(cfg *config.AIConfig) *AIService {
	return &AIService{
		provider: cfg.Provider,
		apiKey:   cfg.APIKey,
		baseURL:  cfg.BaseURL,
		model:    cfg.Model,
		client:   &http.Client{},
	}
}

// AnalyzeDocument sends the extracted document text to an AI provider
// and returns structured analysis results with Arabic field extraction.
func (s *AIService) AnalyzeDocument(ctx context.Context, text string) (*models.AIAnalysisResult, error) {
	prompt := buildAnalysisPrompt(text)

	var responseText string
	var err error

	switch s.provider {
	case "ollama":
		responseText, err = s.callOllama(ctx, prompt)
	case "gemini":
		responseText, err = s.callGemini(ctx, prompt)
	case "deepseek":
		responseText, err = s.callDeepSeek(ctx, prompt)
	default:
		return nil, fmt.Errorf("unsupported AI provider: %s", s.provider)
	}

	if err != nil {
		return nil, fmt.Errorf("AI analysis failed (%s): %w", s.provider, err)
	}

	result, err := parseAIResponse(responseText)
	if err != nil {
		return nil, fmt.Errorf("failed to parse AI response: %w", err)
	}

	return result, nil
}

// buildAnalysisPrompt creates the Arabic prompt for document analysis.
func buildAnalysisPrompt(text string) string {
	return fmt.Sprintf(`أنت مساعد ذكي متخصص في تحليل الوثائق والكتب الرسمية العراقية.
قم بتحليل النص التالي واستخراج المعلومات المطلوبة بصيغة JSON فقط بدون أي نص إضافي.

المعلومات المطلوبة:
1. عنوان_الكتاب: عنوان أو موضوع الكتاب الرسمي
2. الجهة_المصدرة: الجهة أو المؤسسة التي أصدرت الكتاب
3. رقم_العدد: رقم الكتاب أو العدد
4. تاريخ_الكتاب: تاريخ إصدار الكتاب
5. ملخص: ملخص مختصر لمحتوى الكتاب

النص:
%s

أجب بصيغة JSON فقط كالتالي:
{
  "عنوان_الكتاب": "",
  "الجهة_المصدرة": "",
  "رقم_العدد": "",
  "تاريخ_الكتاب": "",
  "ملخص": ""
}`, text)
}

// callOllama sends a request to a local Ollama instance.
func (s *AIService) callOllama(ctx context.Context, prompt string) (string, error) {
	url := strings.TrimRight(s.baseURL, "/") + "/api/generate"

	reqBody := map[string]interface{}{
		"model":  s.model,
		"prompt": prompt,
		"stream": false,
	}

	body, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := s.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to call ollama: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("ollama returned status %d: %s", resp.StatusCode, string(respBody))
	}

	var result struct {
		Response string `json:"response"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", fmt.Errorf("failed to decode ollama response: %w", err)
	}

	return result.Response, nil
}

// callGemini sends a request to the Google Gemini API.
func (s *AIService) callGemini(ctx context.Context, prompt string) (string, error) {
	url := fmt.Sprintf("https://generativelanguage.googleapis.com/v1beta/models/%s:generateContent?key=%s", s.model, s.apiKey)

	reqBody := map[string]interface{}{
		"contents": []map[string]interface{}{
			{
				"parts": []map[string]string{
					{"text": prompt},
				},
			},
		},
		"generationConfig": map[string]interface{}{
			"temperature": 0.1,
			"maxOutputTokens": 1024,
		},
	}

	body, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := s.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to call gemini: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("gemini returned status %d: %s", resp.StatusCode, string(respBody))
	}

	var result struct {
		Candidates []struct {
			Content struct {
				Parts []struct {
					Text string `json:"text"`
				} `json:"parts"`
			} `json:"content"`
		} `json:"candidates"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", fmt.Errorf("failed to decode gemini response: %w", err)
	}

	if len(result.Candidates) == 0 || len(result.Candidates[0].Content.Parts) == 0 {
		return "", fmt.Errorf("gemini returned empty response")
	}

	return result.Candidates[0].Content.Parts[0].Text, nil
}

// callDeepSeek sends a request to a DeepSeek-compatible (OpenAI-compatible) API.
func (s *AIService) callDeepSeek(ctx context.Context, prompt string) (string, error) {
	url := strings.TrimRight(s.baseURL, "/") + "/v1/chat/completions"

	reqBody := map[string]interface{}{
		"model": s.model,
		"messages": []map[string]string{
			{"role": "system", "content": "أنت مساعد متخصص في تحليل الوثائق الرسمية العراقية. أجب دائماً بصيغة JSON فقط."},
			{"role": "user", "content": prompt},
		},
		"temperature": 0.1,
		"max_tokens":  1024,
	}

	body, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+s.apiKey)

	resp, err := s.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to call deepseek: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("deepseek returned status %d: %s", resp.StatusCode, string(respBody))
	}

	var result struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", fmt.Errorf("failed to decode deepseek response: %w", err)
	}

	if len(result.Choices) == 0 {
		return "", fmt.Errorf("deepseek returned empty response")
	}

	return result.Choices[0].Message.Content, nil
}

// parseAIResponse extracts the JSON object from the AI response text
// and unmarshals it into an AIAnalysisResult.
func parseAIResponse(response string) (*models.AIAnalysisResult, error) {
	// Try to find JSON in the response by looking for curly braces
	start := strings.Index(response, "{")
	end := strings.LastIndex(response, "}")

	if start == -1 || end == -1 || end <= start {
		return nil, fmt.Errorf("no JSON object found in AI response")
	}

	jsonStr := response[start : end+1]

	var result models.AIAnalysisResult
	if err := json.Unmarshal([]byte(jsonStr), &result); err != nil {
		return nil, fmt.Errorf("failed to unmarshal AI response JSON: %w", err)
	}

	return &result, nil
}
