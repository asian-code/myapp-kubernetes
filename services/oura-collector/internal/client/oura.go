package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	log "github.com/sirupsen/logrus"
)

const OuraBaseURL = "https://api.ouraring.com/v2/usercollection"

type SleepData struct {
	ID       string `json:"id"`
	Day      string `json:"day"`
	Score    int    `json:"score"`
	Duration int    `json:"duration"`
}

type ActivityData struct {
	ID                string `json:"id"`
	Day               string `json:"day"`
	Score             int    `json:"score"`
	ActiveCalories    int    `json:"active_calories"`
	Steps             int    `json:"steps"`
	MediumActivityMin int    `json:"medium_activity_minutes"`
	HighActivityMin   int    `json:"high_activity_minutes"`
}

type ReadinessData struct {
	ID    string `json:"id"`
	Day   string `json:"day"`
	Score int    `json:"score"`
}

type OuraClient struct {
	apiKey string
	client *http.Client
	logger *log.Entry
}

func New(apiKey string, logger *log.Entry) *OuraClient {
	return &OuraClient{
		apiKey: apiKey,
		client: &http.Client{Timeout: 10 * time.Second},
		logger: logger,
	}
}

func (c *OuraClient) GetSleepData(ctx context.Context, date string) (*SleepData, error) {
	url := fmt.Sprintf("%s/daily_sleep?date=%s", OuraBaseURL, date)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", c.apiKey))

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("oura API returned %d", resp.StatusCode)
	}

	var data SleepData
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return nil, err
	}

	return &data, nil
}

func (c *OuraClient) GetActivityData(ctx context.Context, date string) (*ActivityData, error) {
	url := fmt.Sprintf("%s/daily_activity?date=%s", OuraBaseURL, date)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", c.apiKey))

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("oura API returned %d", resp.StatusCode)
	}

	var data ActivityData
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return nil, err
	}

	return &data, nil
}

func (c *OuraClient) GetReadinessData(ctx context.Context, date string) (*ReadinessData, error) {
	url := fmt.Sprintf("%s/daily_readiness?date=%s", OuraBaseURL, date)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", c.apiKey))

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("oura API returned %d", resp.StatusCode)
	}

	var data ReadinessData
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return nil, err
	}

	return &data, nil
}

func (c *OuraClient) SendToProcessor(ctx context.Context, processorURL string, data interface{}) error {
	jsonData, err := json.Marshal(data)
	if err != nil {
		return err
	}

	req, err := http.NewRequestWithContext(ctx, "POST", processorURL+"/api/v1/ingest", bytes.NewBuffer(jsonData))
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := c.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		return fmt.Errorf("processor returned %d", resp.StatusCode)
	}

	return nil
}
