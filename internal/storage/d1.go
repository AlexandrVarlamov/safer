package storage

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

type D1 struct {
	accountID  string
	databaseID string
	token      string
	client     *http.Client
}

func NewD1(accountID, databaseID, token string) *D1 {
	return &D1{
		accountID:  accountID,
		databaseID: databaseID,
		token:      token,
		client:     &http.Client{Timeout: 15 * time.Second},
	}
}

// Query выполняет один SQL-стейтмент с позиционными параметрами
// и возвращает строки ответа.
func (d *D1) Query(ctx context.Context, sql string, params ...any) ([]map[string]any, error) {
	payload, err := json.Marshal(map[string]any{"sql": sql, "params": params})
	if err != nil {
		return nil, fmt.Errorf("marshal query: %w", err)
	}

	url := fmt.Sprintf(
		"https://api.cloudflare.com/client/v4/accounts/%s/d1/database/%s/query",
		d.accountID, d.databaseID)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(payload))
	if err != nil {
		return nil, fmt.Errorf("build request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+d.token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := d.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("d1 request: %w", err)
	}
	defer resp.Body.Close()

	var out struct {
		Success bool `json:"success"`
		Errors  []struct {
			Code    int    `json:"code"`
			Message string `json:"message"`
		} `json:"errors"`
		Result []struct {
			Results []map[string]any `json:"results"`
			Success bool             `json:"success"`
		} `json:"result"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}

	if !out.Success || len(out.Result) == 0 || !out.Result[0].Success {
		msg := "unknown d1 error"
		if len(out.Errors) > 0 {
			msg = fmt.Sprintf("%d: %s", out.Errors[0].Code, out.Errors[0].Message)
		}
		return nil, fmt.Errorf("d1: %s", msg)
	}

	return out.Result[0].Results, nil
}
