package client_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/we-work-in-the-cloud/nullcloud/terraform-provider-nullcloud/internal/client"
)

func makeVPC(id, name string) client.VPC {
	return client.VPC{ID: id, Name: name, Status: "available", CRN: "crn:nullcloud:vpc:" + id, CreatedAt: time.Now()}
}

func makeSubnet(id, name, vpcID string) client.Subnet {
	return client.Subnet{ID: id, Name: name, Status: "available", VPCID: vpcID, CIDRBlock: "10.0.0.0/24", CreatedAt: time.Now()}
}

func makeInstance(id, name, subnetID string) client.Instance {
	return client.Instance{ID: id, Name: name, Status: "running", SubnetID: subnetID, VPCID: "vpc-1", PrimaryIP: "10.0.0.2", CreatedAt: time.Now()}
}

func TestClient_CreateVPC(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" || r.URL.Path != "/v1/vpcs" {
			t.Errorf("unexpected: %s %s", r.Method, r.URL.Path)
		}
		if r.Header.Get("Authorization") != "tok" {
			t.Error("missing auth header")
		}
		w.WriteHeader(201)
		json.NewEncoder(w).Encode(makeVPC("vpc-1", "my-vpc"))
	}))
	defer srv.Close()

	c := client.New(srv.URL, "tok")
	vpc, err := c.CreateVPC("my-vpc")
	if err != nil || vpc.ID != "vpc-1" {
		t.Fatalf("unexpected: %v %v", vpc, err)
	}
}

func TestClient_GetVPC_Found(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		json.NewEncoder(w).Encode(makeVPC("vpc-1", "my-vpc"))
	}))
	defer srv.Close()

	c := client.New(srv.URL, "tok")
	vpc, found, err := c.GetVPC("vpc-1")
	if err != nil || !found || vpc.ID != "vpc-1" {
		t.Fatalf("unexpected: %v %v %v", vpc, found, err)
	}
}

func TestClient_GetVPC_NotFound(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(404)
		json.NewEncoder(w).Encode(map[string]any{"errors": []map[string]string{{"code": "not_found", "message": "VPC not found"}}})
	}))
	defer srv.Close()

	c := client.New(srv.URL, "tok")
	_, found, err := c.GetVPC("nonexistent")
	if err != nil || found {
		t.Fatalf("expected not found, got: found=%v err=%v", found, err)
	}
}

func TestClient_DeleteVPC(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(204)
	}))
	defer srv.Close()

	c := client.New(srv.URL, "tok")
	if err := c.DeleteVPC("vpc-1"); err != nil {
		t.Fatal(err)
	}
}

func TestClient_CreateSubnet(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" || r.URL.Path != "/v1/subnets" {
			t.Errorf("unexpected: %s %s", r.Method, r.URL.Path)
		}
		var body map[string]any
		json.NewDecoder(r.Body).Decode(&body)
		vpc, _ := body["vpc"].(map[string]any)
		if vpc["id"] != "vpc-1" {
			t.Errorf("expected vpc_id vpc-1, got %v", vpc["id"])
		}
		w.WriteHeader(201)
		json.NewEncoder(w).Encode(makeSubnet("sub-1", "my-subnet", "vpc-1"))
	}))
	defer srv.Close()

	c := client.New(srv.URL, "tok")
	sub, err := c.CreateSubnet("my-subnet", "vpc-1")
	if err != nil || sub.ID != "sub-1" || sub.VPCID != "vpc-1" {
		t.Fatalf("unexpected: %v %v", sub, err)
	}
}

func TestClient_CreateInstance(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(201)
		json.NewEncoder(w).Encode(makeInstance("vsi-1", "my-vsi", "sub-1"))
	}))
	defer srv.Close()

	c := client.New(srv.URL, "tok")
	inst, err := c.CreateInstance("my-vsi", "sub-1", "cx2-2x4", "ubuntu-22")
	if err != nil || inst.ID != "vsi-1" {
		t.Fatalf("unexpected: %v %v", inst, err)
	}
}

func TestClient_ErrorResponse(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(404)
		json.NewEncoder(w).Encode(map[string]any{
			"errors": []map[string]string{{"code": "not_found", "message": "VPC not found"}},
		})
	}))
	defer srv.Close()

	c := client.New(srv.URL, "tok")
	_, err := c.CreateVPC("x")
	if err == nil {
		t.Fatal("expected error")
	}
}
