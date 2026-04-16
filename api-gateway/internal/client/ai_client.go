package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"time"
)

type AIClient struct {
	baseURL    string
	httpClient *http.Client
}

func NewAIClient(baseURL string) *AIClient {
	return &AIClient{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: 60 * time.Second,
		},
	}
}

type ParseResumeResponse struct {
	ParsedText string `json:"parsed_text"`
}

type GenerateEmailRequest struct {
	ResumeText     string `json:"resume_text"`
	JobDescription string `json:"job_description"`
	CompanyName    string `json:"company_name"`
	Role           string `json:"role"`
	RecruiterEmail string `json:"recruiter_email"`
	JobLink        string `json:"job_link"`
	Tone           string `json:"tone"`
}

type GenerateEmailResponse struct {
	Subject    string   `json:"subject"`
	Body       string   `json:"body"`
	MatchScore float64  `json:"match_score"`
	KeyPoints  []string `json:"key_points"`
	Reasoning  string   `json:"reasoning"`
}

type SmartApplyResumeCandidate struct {
	ResumeID   string `json:"resume_id"`
	ParsedText string `json:"parsed_text"`
}

type SmartApplyExtractRequest struct {
	RawText string                      `json:"raw_text"`
	Resumes []SmartApplyResumeCandidate `json:"resumes"`
}

type SmartApplyExtractResponse struct {
	CompanyName          string  `json:"company_name"`
	Role                 string  `json:"role"`
	RecruiterEmail       *string `json:"recruiter_email"`
	JobLink              *string `json:"job_link"`
	JobDescription       string  `json:"job_description"`
	SelectedResumeID     string  `json:"selected_resume_id"`
	ExtractionConfidence string  `json:"extraction_confidence"`
}

func (c *AIClient) ParseResume(filePath string, fileBytes []byte) (*ParseResumeResponse, error) {
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	part, err := writer.CreateFormFile("file", filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to create form file: %w", err)
	}

	if _, err := io.Copy(part, bytes.NewReader(fileBytes)); err != nil {
		return nil, fmt.Errorf("failed to copy file data: %w", err)
	}
	writer.Close()

	req, err := http.NewRequest("POST", c.baseURL+"/ai/parse-resume", body)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("AI service request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("AI service returned %d: %s", resp.StatusCode, string(respBody))
	}

	var result ParseResumeResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode AI response: %w", err)
	}

	return &result, nil
}

func (c *AIClient) GenerateEmail(req *GenerateEmailRequest) (*GenerateEmailResponse, error) {
	jsonBody, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	httpReq, err := http.NewRequest("POST", c.baseURL+"/ai/generate-email", bytes.NewReader(jsonBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("AI service request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("AI service returned %d: %s", resp.StatusCode, string(respBody))
	}

	var result GenerateEmailResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode AI response: %w", err)
	}

	return &result, nil
}

func (c *AIClient) SmartApplyExtractAndMatch(req *SmartApplyExtractRequest) (*SmartApplyExtractResponse, error) {
	jsonBody, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	httpReq, err := http.NewRequest("POST", c.baseURL+"/ai/smart-apply/extract-match", bytes.NewReader(jsonBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("AI service request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("AI service returned %d: %s", resp.StatusCode, string(respBody))
	}

	var result SmartApplyExtractResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode AI response: %w", err)
	}

	return &result, nil
}
