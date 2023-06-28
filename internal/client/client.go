package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"gihtub.com/lsmoura/hishtory_cc/internal/model"
	"net/http"
	"net/url"
)

type Client struct {
	host       string
	userSecret *UserSecret
	deviceID   string
}

func New(host string, userSecret *UserSecret, deviceID string) *Client {
	return &Client{
		host:       host,
		userSecret: userSecret,
		deviceID:   deviceID,
	}
}

func (c *Client) Register(ctx context.Context) error {
	if c.userSecret.IsEmpty() {
		return fmt.Errorf("user secret not initialized")
	}
	u, err := url.Parse(fmt.Sprintf("%s/api/v1/register", c.host))
	if err != nil {
		return fmt.Errorf("url.Parse: %w", err)
	}
	q := u.Query()
	q.Set("user_id", c.userSecret.UserId())
	q.Set("device_id", c.deviceID)
	u.RawQuery = q.Encode()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
	if err != nil {
		return fmt.Errorf("http.NewRequestWithContext: %w", err)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("http.DefaultClient.Do: %w", err)
	}

	if resp.StatusCode >= 400 {
		return fmt.Errorf("status code: %d", resp.StatusCode)
	}

	return nil
}

func (c *Client) Save(ctx context.Context, entries []*model.HistoryEntry) error {
	if len(entries) == 0 {
		return nil
	}

	if ctx == nil {
		ctx = context.Background()
	}

	u, err := url.Parse(fmt.Sprintf("%s/api/v1/submit", c.host))
	if err != nil {
		return fmt.Errorf("url.Parse: %w", err)
	}
	q := u.Query()
	q.Set("user_id", c.userSecret.UserId())
	q.Set("device_id", c.deviceID)
	u.RawQuery = q.Encode()

	encEntries := make([]*model.EncHistoryEntry, 0, len(entries))
	for _, entry := range entries {
		encEntry, err := Encrypt(c.userSecret.EncryptionKey(), c.userSecret.UserId(), entry)
		if err != nil {
			return fmt.Errorf("cannot encrypt entry: %w", err)
		}
		encEntries = append(encEntries, encEntry)
	}

	reqBody, err := json.Marshal(encEntries)
	if err != nil {
		return fmt.Errorf("json.Marshal: %w", err)
	}
	reqBodyReader := bytes.NewReader(reqBody)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, u.String(), reqBodyReader)
	if err != nil {
		return fmt.Errorf("http.NewRequestWithContext: %w", err)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("http.DefaultClient.Do: %w", err)
	}

	if resp.StatusCode >= 400 {
		return fmt.Errorf("status code: %d", resp.StatusCode)
	}

	return nil
}
