package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

type Client struct {
	baseURL       string
	token         string
	httpClient    *http.Client
	DefaultRegion string
}

func New(baseURL, token string) *Client {
	return &Client{
		baseURL:       baseURL,
		token:         token,
		httpClient:    &http.Client{Timeout: 30 * time.Second},
		DefaultRegion: "us-east",
	}
}

type VPC struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Status    string    `json:"status"`
	CRN       string    `json:"crn"`
	Region    string    `json:"region"`
	CreatedAt time.Time `json:"created_at"`
}

type Subnet struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Status    string    `json:"status"`
	CRN       string    `json:"crn"`
	VPCID     string    `json:"vpc_id"`
	Zone      string    `json:"zone"`
	CIDRBlock string    `json:"cidr_block"`
	CreatedAt time.Time `json:"created_at"`
}

type Instance struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Status    string    `json:"status"`
	CRN       string    `json:"crn"`
	SubnetID  string    `json:"subnet_id"`
	Profile   string    `json:"profile"`
	Image     string    `json:"image"`
	PrimaryIP string    `json:"primary_ip"`
	CreatedAt time.Time `json:"created_at"`
}

type LoadBalancerTarget struct {
	Type string `json:"type"`
	ID   string `json:"id"`
}

type LoadBalancer struct {
	ID        string               `json:"id"`
	Name      string               `json:"name"`
	Status    string               `json:"status"`
	CRN       string               `json:"crn"`
	Protocol  string               `json:"protocol"`
	Port      int                  `json:"port"`
	Targets   []LoadBalancerTarget `json:"targets"`
	CreatedAt time.Time            `json:"created_at"`
}

type Bucket struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Status    string    `json:"status"`
	CRN       string    `json:"crn"`
	Region    string    `json:"region"`
	CreatedAt time.Time `json:"created_at"`
}

type Database struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Status    string    `json:"status"`
	CRN       string    `json:"crn"`
	Engine    string    `json:"engine"`
	Version   string    `json:"version"`
	Plan      string    `json:"plan"`
	SubnetIDs []string  `json:"subnet_ids"`
	CreatedAt time.Time `json:"created_at"`
	Endpoint  string    `json:"endpoint"`
}

type KubernetesCluster struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Status    string    `json:"status"`
	CRN       string    `json:"crn"`
	Version   string    `json:"version"`
	NodeCount int       `json:"node_count"`
	SubnetIDs []string  `json:"subnet_ids"`
	CreatedAt time.Time `json:"created_at"`
}

type RegionZone struct {
	Name string `json:"name"`
}

type Region struct {
	Name  string       `json:"name"`
	Zones []RegionZone `json:"zones"`
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

// Regions

func (c *Client) ListRegions() ([]Region, error) {
	var result struct {
		Regions []Region `json:"regions"`
	}
	if _, err := c.do("GET", "/v1/regions", nil, &result); err != nil {
		return nil, err
	}
	return result.Regions, nil
}

// VPC

func (c *Client) CreateVPC(name, region string) (*VPC, error) {
	var vpc VPC
	if _, err := c.do("POST", "/v1/vpcs", map[string]string{"name": name, "region": region}, &vpc); err != nil {
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

func (c *Client) CreateSubnet(name, vpcID, zone, cidrBlock string) (*Subnet, error) {
	body := map[string]any{
		"name":       name,
		"vpc":        map[string]string{"id": vpcID},
		"zone":       zone,
		"cidr_block": cidrBlock,
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

// LoadBalancer

func (c *Client) CreateLoadBalancer(name, protocol string, port int, targets []LoadBalancerTarget) (*LoadBalancer, error) {
	body := map[string]any{
		"name":     name,
		"protocol": protocol,
		"port":     port,
		"targets":  targets,
	}
	var lb LoadBalancer
	if _, err := c.do("POST", "/v1/loadbalancers", body, &lb); err != nil {
		return nil, err
	}
	return &lb, nil
}

func (c *Client) GetLoadBalancer(id string) (*LoadBalancer, bool, error) {
	var lb LoadBalancer
	status, err := c.do("GET", "/v1/loadbalancers/"+id, nil, &lb)
	if err != nil {
		if status == 404 {
			return nil, false, nil
		}
		return nil, false, err
	}
	return &lb, true, nil
}

func (c *Client) DeleteLoadBalancer(id string) error {
	status, err := c.do("DELETE", "/v1/loadbalancers/"+id, nil, nil)
	if err != nil && status != 404 {
		return err
	}
	return nil
}

// Bucket

func (c *Client) CreateBucket(name, region string) (*Bucket, error) {
	body := map[string]any{
		"name":   name,
		"region": region,
	}
	var b Bucket
	if _, err := c.do("POST", "/v1/buckets", body, &b); err != nil {
		return nil, err
	}
	return &b, nil
}

func (c *Client) GetBucket(id string) (*Bucket, bool, error) {
	var b Bucket
	status, err := c.do("GET", "/v1/buckets/"+id, nil, &b)
	if err != nil {
		if status == 404 {
			return nil, false, nil
		}
		return nil, false, err
	}
	return &b, true, nil
}

func (c *Client) DeleteBucket(id string) error {
	status, err := c.do("DELETE", "/v1/buckets/"+id, nil, nil)
	if err != nil && status != 404 {
		return err
	}
	return nil
}

// Database

func (c *Client) CreateDatabase(name, engine, version, plan string, subnetIDs []string) (*Database, error) {
	body := map[string]any{
		"name":       name,
		"engine":     engine,
		"version":    version,
		"plan":       plan,
		"subnet_ids": subnetIDs,
	}
	var db Database
	if _, err := c.do("POST", "/v1/databases", body, &db); err != nil {
		return nil, err
	}
	return &db, nil
}

func (c *Client) GetDatabase(id string) (*Database, bool, error) {
	var db Database
	status, err := c.do("GET", "/v1/databases/"+id, nil, &db)
	if err != nil {
		if status == 404 {
			return nil, false, nil
		}
		return nil, false, err
	}
	return &db, true, nil
}

func (c *Client) DeleteDatabase(id string) error {
	status, err := c.do("DELETE", "/v1/databases/"+id, nil, nil)
	if err != nil && status != 404 {
		return err
	}
	return nil
}

// KubernetesCluster

func (c *Client) CreateKubernetesCluster(name, version string, nodeCount int, subnetIDs []string) (*KubernetesCluster, error) {
	body := map[string]any{
		"name":       name,
		"version":    version,
		"node_count": nodeCount,
		"subnet_ids": subnetIDs,
	}
	var cluster KubernetesCluster
	if _, err := c.do("POST", "/v1/clusters", body, &cluster); err != nil {
		return nil, err
	}
	return &cluster, nil
}

func (c *Client) GetKubernetesCluster(id string) (*KubernetesCluster, bool, error) {
	var cluster KubernetesCluster
	status, err := c.do("GET", "/v1/clusters/"+id, nil, &cluster)
	if err != nil {
		if status == 404 {
			return nil, false, nil
		}
		return nil, false, err
	}
	return &cluster, true, nil
}

func (c *Client) DeleteKubernetesCluster(id string) error {
	status, err := c.do("DELETE", "/v1/clusters/"+id, nil, nil)
	if err != nil && status != 404 {
		return err
	}
	return nil
}
