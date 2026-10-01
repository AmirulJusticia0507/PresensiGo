package ai

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"
)

type Client struct {
	baseURL string
	http    *http.Client
}

type VerifyResult struct {
	Verified       bool    `json:"verified"`
	Similarity     float64 `json:"similarity"`
	Threshold      float64 `json:"threshold"`
	LivenessPassed bool    `json:"liveness_passed"`
}

func NewClient(baseURL string, timeout time.Duration) *Client {
	return &Client{
		baseURL: strings.TrimRight(baseURL, "/"),
		http:    &http.Client{Timeout: timeout},
	}
}

func (c *Client) Enroll(ctx context.Context, images []string) ([]float32, error) {
	var response struct {
		Embedding []float32 `json:"embedding"`
	}
	if err := c.post(ctx, "/v1/faces/enroll", map[string]any{"images": images}, &response); err != nil {
		return nil, err
	}
	if len(response.Embedding) != 512 {
		return nil, errors.New("AI service returned an invalid face embedding")
	}
	return response.Embedding, nil
}

func (c *Client) Verify(ctx context.Context, image string, embedding []float32, challenge string, threshold float64) (*VerifyResult, error) {
	var response VerifyResult
	err := c.post(ctx, "/v1/faces/verify", map[string]any{
		"image":              image,
		"enrolled_embedding": embedding,
		"challenge":          challenge,
		"threshold":          threshold,
	}, &response)
	return &response, err
}

func (c *Client) post(ctx context.Context, path string, payload any, target any) error {
	body, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	for attempt := 0; attempt < 2; attempt++ {
		req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+path, bytes.NewReader(body))
		if err != nil {
			return err
		}
		req.Header.Set("Content-Type", "application/json")

		resp, requestErr := c.http.Do(req)
		if requestErr != nil {
			if attempt == 0 {
				select {
				case <-ctx.Done():
					return fmt.Errorf("AI service unavailable: %w", ctx.Err())
				case <-time.After(500 * time.Millisecond):
					continue
				}
			}
			return fmt.Errorf("AI service unavailable: %w", requestErr)
		}

		if resp.StatusCode >= 500 && attempt == 0 {
			resp.Body.Close()
			time.Sleep(500 * time.Millisecond)
			continue
		}
		if resp.StatusCode < 200 || resp.StatusCode >= 300 {
			var problem struct {
				Detail string `json:"detail"`
			}
			_ = json.NewDecoder(resp.Body).Decode(&problem)
			resp.Body.Close()
			if problem.Detail == "" {
				problem.Detail = "face processing failed"
			}
			return fmt.Errorf("AI service rejected image: %s", problem.Detail)
		}
		decodeErr := json.NewDecoder(resp.Body).Decode(target)
		resp.Body.Close()
		if decodeErr != nil {
			return fmt.Errorf("invalid AI service response: %w", decodeErr)
		}
		return nil
	}
	return errors.New("AI service unavailable")
}
