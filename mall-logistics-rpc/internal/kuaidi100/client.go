package kuaidi100

import (
	"context"
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
)

type Config struct {
	Customer        string
	Key             string
	PollEndpoint    string
	WebhookCallback string
	HTTP            *http.Client
}

type Client struct{ cfg Config }

func NewClient(c Config) *Client { return &Client{cfg: c} }

type subscribeParam struct {
	Company    string `json:"company"`
	Number     string `json:"number"`
	Key        string `json:"key"`
	Parameters struct {
		Callbackurl string `json:"callbackurl"`
		Salt        string `json:"salt,omitempty"`
		Resultv2    string `json:"resultv2"`
	} `json:"parameters"`
}

type subscribeResp struct {
	Result     bool   `json:"result"`
	ReturnCode string `json:"returnCode"`
	Message    string `json:"message"`
}

// Subscribe registers a tracking number with kuaidi100.
// carrier is the kuaidi100 company code (e.g. "shunfeng", "jd").
func (c *Client) Subscribe(ctx context.Context, carrier, trackingNo string) error {
	if c.cfg.Customer == "" || c.cfg.Key == "" || c.cfg.PollEndpoint == "" {
		return fmt.Errorf("kuaidi100: missing customer/key/endpoint")
	}
	p := subscribeParam{Company: carrier, Number: trackingNo, Key: c.cfg.Key}
	p.Parameters.Callbackurl = c.cfg.WebhookCallback
	p.Parameters.Resultv2 = "1"
	pbytes, _ := json.Marshal(p)
	form := url.Values{}
	form.Set("schema", "json")
	form.Set("param", string(pbytes))
	req, _ := http.NewRequestWithContext(ctx, http.MethodPost, c.cfg.PollEndpoint,
		strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	resp, err := c.cfg.HTTP.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	var sr subscribeResp
	if err := json.NewDecoder(resp.Body).Decode(&sr); err != nil {
		return err
	}
	if !sr.Result {
		return fmt.Errorf("kuaidi100 subscribe rejected: %s %s", sr.ReturnCode, sr.Message)
	}
	return nil
}

// VerifySign validates webhook signature: sign == upper(md5(param + key)).
func (c *Client) VerifySign(param, sign string) bool {
	h := md5.Sum([]byte(param + c.cfg.Key))
	return strings.EqualFold(hex.EncodeToString(h[:]), sign)
}
