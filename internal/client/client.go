package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

type Client struct {
	baseURL    string
	token      string
	httpClient *http.Client
}

func New(baseURL, token string) *Client {
	return &Client{
		baseURL:    baseURL,
		token:      token,
		httpClient: &http.Client{Timeout: 30 * time.Second},
	}
}

type VPC struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Status    string    `json:"status"`
	CRN       string    `json:"crn"`
	CreatedAt time.Time `json:"created_at"`
}

type Subnet struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Status    string    `json:"status"`
	CRN       string    `json:"crn"`
	VPCID     string    `json:"vpc_id"`
	CIDRBlock string    `json:"cidr_block"`
	CreatedAt time.Time `json:"created_at"`
}

type Instance struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Status    string    `json:"status"`
	CRN       string    `json:"crn"`
	SubnetID  string    `json:"subnet_id"`
	VPCID     string    `json:"vpc_id"`
	Profile   string    `json:"profile"`
	Image     string    `json:"image"`
	PrimaryIP string    `json:"primary_ip"`
	CreatedAt time.Time `json:"created_at"`
}

type apiErr struct {
	Errors []struct {
		Code    string `json:"code"`
		Message string `json:"message"`
	} `json:"errors"`
}

func (c *Client) do(method, path string, body, result any) (int, error) {
	var bodyReader *bytes.Reader
	if body != nil {
		b, err := json.Marshal(body)
		if err != nil {
			return 0, err
		}
		bodyReader = bytes.NewReader(b)
	} else {
		bodyReader = bytes.NewReader(nil)
	}

	req, err := http.NewRequest(method, c.baseURL+path, bodyReader)
	if err != nil {
		return 0, err
	}
	req.Header.Set("Authorization", c.token)
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		var errs apiErr
		json.NewDecoder(resp.Body).Decode(&errs)
		if len(errs.Errors) > 0 {
			return resp.StatusCode, fmt.Errorf("%s: %s", errs.Errors[0].Code, errs.Errors[0].Message)
		}
		return resp.StatusCode, fmt.Errorf("HTTP %d", resp.StatusCode)
	}

	if result != nil {
		if err := json.NewDecoder(resp.Body).Decode(result); err != nil {
			return resp.StatusCode, err
		}
	}
	return resp.StatusCode, nil
}

// VPC

func (c *Client) CreateVPC(name string) (*VPC, error) {
	var vpc VPC
	if _, err := c.do("POST", "/v1/vpcs", map[string]string{"name": name}, &vpc); err != nil {
		return nil, err
	}
	return &vpc, nil
}

func (c *Client) GetVPC(id string) (*VPC, bool, error) {
	var vpc VPC
	status, err := c.do("GET", "/v1/vpcs/"+id, nil, &vpc)
	if err != nil {
		if status == 404 {
			return nil, false, nil
		}
		return nil, false, err
	}
	return &vpc, true, nil
}

func (c *Client) DeleteVPC(id string) error {
	status, err := c.do("DELETE", "/v1/vpcs/"+id, nil, nil)
	if err != nil && status != 404 {
		return err
	}
	return nil
}

// Subnet

func (c *Client) CreateSubnet(name, vpcID string) (*Subnet, error) {
	body := map[string]any{
		"name": name,
		"vpc":  map[string]string{"id": vpcID},
	}
	var sub Subnet
	if _, err := c.do("POST", "/v1/subnets", body, &sub); err != nil {
		return nil, err
	}
	return &sub, nil
}

func (c *Client) GetSubnet(id string) (*Subnet, bool, error) {
	var sub Subnet
	status, err := c.do("GET", "/v1/subnets/"+id, nil, &sub)
	if err != nil {
		if status == 404 {
			return nil, false, nil
		}
		return nil, false, err
	}
	return &sub, true, nil
}

func (c *Client) DeleteSubnet(id string) error {
	status, err := c.do("DELETE", "/v1/subnets/"+id, nil, nil)
	if err != nil && status != 404 {
		return err
	}
	return nil
}

// Instance

func (c *Client) CreateInstance(name, subnetID, profile, image string) (*Instance, error) {
	body := map[string]any{
		"name":    name,
		"subnet":  map[string]string{"id": subnetID},
		"profile": map[string]string{"name": profile},
		"image":   map[string]string{"id": image},
	}
	var inst Instance
	if _, err := c.do("POST", "/v1/instances", body, &inst); err != nil {
		return nil, err
	}
	return &inst, nil
}

func (c *Client) GetInstance(id string) (*Instance, bool, error) {
	var inst Instance
	status, err := c.do("GET", "/v1/instances/"+id, nil, &inst)
	if err != nil {
		if status == 404 {
			return nil, false, nil
		}
		return nil, false, err
	}
	return &inst, true, nil
}

func (c *Client) InstanceAction(id, action string) (*Instance, error) {
	var inst Instance
	if _, err := c.do("POST", "/v1/instances/"+id+"/actions", map[string]string{"type": action}, &inst); err != nil {
		return nil, err
	}
	return &inst, nil
}

func (c *Client) DeleteInstance(id string) error {
	status, err := c.do("DELETE", "/v1/instances/"+id, nil, nil)
	if err != nil && status != 404 {
		return err
	}
	return nil
}
