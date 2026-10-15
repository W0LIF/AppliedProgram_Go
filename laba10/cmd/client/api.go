package main

import (
	"bytes"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"time"

	"laba10/models"
)

type APIError struct {
	Status  int
	Message string
	Details map[string]string
}

func (e *APIError) Error() string {
	if len(e.Details) > 0 {
		return fmt.Sprintf("%s (%v)", e.Message, e.Details)
	}
	return e.Message
}

type Client struct {
	baseURL string
	http    *http.Client
	token   string
}

func NewClient(baseURL string) *Client {
	cert, err := tls.LoadX509KeyPair("certs/client.crt", "certs/client.key")
	if err != nil {
		log.Fatalf("client cert: %v", err)
	}

	caPEM, err := os.ReadFile("certs/ca.crt")
	if err != nil {
		log.Fatalf("read CA: %v", err)
	}
	pool := x509.NewCertPool()
	if !pool.AppendCertsFromPEM(caPEM) {
		log.Fatal("bad CA cert")
	}

	tlsCfg := &tls.Config{
		Certificates: []tls.Certificate{cert},
		RootCAs:      pool,
		ServerName:   "localhost",
		MinVersion:   tls.VersionTLS12,
	}

	return &Client{
		baseURL: baseURL,
		http: &http.Client{
			Timeout:   10 * time.Second,
			Transport: &http.Transport{TLSClientConfig: tlsCfg},
		},
	}
}

func (c *Client) Token() string     { return c.token }
func (c *Client) SetToken(t string) { c.token = t }

func (c *Client) do(method, path string, body, out any) error {
	var reader io.Reader
	if body != nil {
		b, err := json.Marshal(body)
		if err != nil {
			return err
		}
		reader = bytes.NewReader(b)
	}

	req, err := http.NewRequest(method, c.baseURL+path, reader)
	if err != nil {
		return err
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	if c.token != "" {
		req.Header.Set("Authorization", "Bearer "+c.token)
	}

	resp, err := c.http.Do(req)
	if err != nil {
		return fmt.Errorf("network error: %w", err)
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("read body: %w", err)
	}

	if resp.StatusCode >= 400 {
		apiErr := &APIError{Status: resp.StatusCode}
		var raw struct {
			Error   string            `json:"error"`
			Details map[string]string `json:"details"`
		}
		if json.Unmarshal(data, &raw) == nil && raw.Error != "" {
			apiErr.Message = raw.Error
			apiErr.Details = raw.Details
		} else {
			apiErr.Message = string(data)
		}
		return apiErr
	}

	if out != nil && len(data) > 0 {
		if err := json.Unmarshal(data, out); err != nil {
			return fmt.Errorf("decode response: %w", err)
		}
	}
	return nil
}

func (c *Client) Register(req models.RegisterRequest) (models.AuthResponse, error) {
	var out models.AuthResponse
	err := c.do("POST", "/auth/register", req, &out)
	return out, err
}

func (c *Client) Login(email, password string) (models.AuthResponse, error) {
	var out models.AuthResponse
	err := c.do("POST", "/auth/login",
		models.LoginRequest{Email: email, Password: password}, &out)
	return out, err
}

type ListOptions struct {
	Name   string
	MinAge int
	MaxAge int
	Limit  int
	Offset int
}

func (c *Client) ListUsers(o ListOptions) ([]models.User, error) {
	q := url.Values{}
	if o.Name != "" {
		q.Set("name", o.Name)
	}
	if o.MinAge > 0 {
		q.Set("min_age", strconv.Itoa(o.MinAge))
	}
	if o.MaxAge > 0 {
		q.Set("max_age", strconv.Itoa(o.MaxAge))
	}
	if o.Limit > 0 {
		q.Set("limit", strconv.Itoa(o.Limit))
	}
	if o.Offset > 0 {
		q.Set("offset", strconv.Itoa(o.Offset))
	}

	path := "/users"
	if len(q) > 0 {
		path += "?" + q.Encode()
	}

	var out []models.User
	err := c.do("GET", path, nil, &out)
	return out, err
}

func (c *Client) GetUser(id int) (models.User, error) {
	var out models.User
	err := c.do("GET", fmt.Sprintf("/users/%d", id), nil, &out)
	return out, err
}

func (c *Client) CreateUser(u models.User) (models.User, error) {
	var out models.User
	err := c.do("POST", "/users", u, &out)
	return out, err
}

func (c *Client) UpdateUser(id int, u models.User) (models.User, error) {
	var out models.User
	err := c.do("PUT", fmt.Sprintf("/users/%d", id), u, &out)
	return out, err
}

func (c *Client) DeleteUser(id int) error {
	return c.do("DELETE", fmt.Sprintf("/users/%d", id), nil, nil)
}
