package cloud

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"
)

type Client struct {
	client  *http.Client
	token   string
	baseURL string
}

type ErrorResponse struct {
	Message string `json:"message"`
}

func NewClient(token string) *Client {

	var client = &http.Client{
		Timeout: 30 * time.Second,
	}

	host := os.Getenv("K6CLOUD_HOST")
	if host == "" {
		host = "http://localhost:5000"
	}

	baseURL := fmt.Sprintf("%s/v1", host)

	c := &Client{
		client:  client,
		token:   token,
		baseURL: baseURL,
	}
	return c
}

func (c *Client) NewRequest(method, url string, data interface{}) (*http.Request, error) {
	var buf io.Reader

	if data != nil {
		b, err := json.Marshal(&data)
		if err != nil {
			return nil, err
		}

		buf = bytes.NewBuffer(b)
	}

	return http.NewRequest(method, url, buf)
}

func (c *Client) Do(req *http.Request, v interface{}) error {

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Token %s", c.token))

	resp, err := c.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if v != nil {
		err := json.NewDecoder(resp.Body).Decode(v)
		if err == io.EOF {
			err = nil // Ignore EOF from empty body
		}
	}

	return err

}

type TestRun struct {
	ID   string `json:"id,omitempty"`
	Name string `json:"name"`
}

type CreateTestRunResponse struct {
	ID string `json:"id"`
}

func (c *Client) CreateTestRun(name string) {
	testRun := TestRun{Name: name}

	url := fmt.Sprintf("%s/test-run", c.baseURL)
	req, err := c.NewRequest("POST", url, testRun)
	if err != nil {
	}

	var ctrr = CreateTestRunResponse{}
	err = c.Do(req, ctrr)
	if err != nil {

	}
}

func (c *Client) PushMetric(referenceID string, samples []*Sample) {
	url := fmt.Sprintf("%s/metrics/%s", c.baseURL, referenceID)

	req, err := c.NewRequest("POST", url, samples)
	if err != nil {
		return
	}

	err = c.Do(req, nil)
	if err != nil {
		return
	}
}

type Sample struct {
	Type   string     `json:"type"`
	Metric string     `json:"metric"`
	Data   SampleData `json:"data"`
}

type SampleData struct {
	Time  time.Time         `json:"time"`
	Value float64           `json:"value"`
	Tags  map[string]string `json:"tags,omitempty"`
}
