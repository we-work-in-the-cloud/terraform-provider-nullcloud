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
	return client.Subnet{ID: id, Name: name, Status: "available", VPCID: vpcID, Zone: "us-east-1", CIDRBlock: "10.0.0.0/24", CreatedAt: time.Now()}
}

func makeInstance(id, name, subnetID string) client.Instance {
	return client.Instance{ID: id, Name: name, Status: "running", SubnetID: subnetID, VPCID: "vpc-1", PrimaryIP: "10.0.0.2", CreatedAt: time.Now()}
}

func makeLB(id, name string) client.LoadBalancer {
	return client.LoadBalancer{ID: id, Name: name, Status: "active", CRN: "crn:lb:" + id, Protocol: "http", Port: 80, CreatedAt: time.Now()}
}

func makeBucket(id, name string) client.Bucket {
	return client.Bucket{ID: id, Name: name, Status: "available", CRN: "crn:bkt:" + id, Region: "us-east", CreatedAt: time.Now()}
}

func makeDatabase(id, name string) client.Database {
	return client.Database{ID: id, Name: name, Status: "available", CRN: "crn:db:" + id,
		Engine: "postgres", Version: "15", Plan: "small", SubnetIDs: []string{"sub-1"}, CreatedAt: time.Now()}
}

func makeCluster(id, name string) client.KubernetesCluster {
	return client.KubernetesCluster{ID: id, Name: name, Status: "running", CRN: "crn:k8s:" + id,
		Version: "1.30", NodeCount: 2, SubnetIDs: []string{"sub-1"}, CreatedAt: time.Now()}
}

func notFoundResp(w http.ResponseWriter) {
	w.WriteHeader(404)
	json.NewEncoder(w).Encode(map[string]any{"errors": []map[string]string{{"code": "not_found", "message": "not found"}}})
}

func serverErrResp(w http.ResponseWriter) {
	w.WriteHeader(500)
	json.NewEncoder(w).Encode(map[string]any{"errors": []map[string]string{{"code": "internal", "message": "server error"}}})
}

// ---- VPC ----

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
	vpc, err := c.CreateVPC("my-vpc", "us-east")
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
		notFoundResp(w)
	}))
	defer srv.Close()

	c := client.New(srv.URL, "tok")
	_, found, err := c.GetVPC("nonexistent")
	if err != nil || found {
		t.Fatalf("expected not found, got: found=%v err=%v", found, err)
	}
}

func TestClient_GetVPC_ServerError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		serverErrResp(w)
	}))
	defer srv.Close()

	c := client.New(srv.URL, "tok")
	_, found, err := c.GetVPC("vpc-1")
	if err == nil || found {
		t.Fatal("expected error")
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

func TestClient_DeleteVPC_NotFound(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		notFoundResp(w)
	}))
	defer srv.Close()

	c := client.New(srv.URL, "tok")
	if err := c.DeleteVPC("nonexistent"); err != nil {
		t.Fatal("expected no error for 404 delete, got:", err)
	}
}

// ---- Subnet ----

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
		if body["zone"] != "us-east-1" {
			t.Errorf("expected zone us-east-1, got %v", body["zone"])
		}
		w.WriteHeader(201)
		json.NewEncoder(w).Encode(makeSubnet("sub-1", "my-subnet", "vpc-1"))
	}))
	defer srv.Close()

	c := client.New(srv.URL, "tok")
	sub, err := c.CreateSubnet("my-subnet", "vpc-1", "us-east-1")
	if err != nil || sub.ID != "sub-1" || sub.VPCID != "vpc-1" {
		t.Fatalf("unexpected: %v %v", sub, err)
	}
}

func TestClient_GetSubnet_Found(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/subnets/sub-1" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		w.WriteHeader(200)
		json.NewEncoder(w).Encode(makeSubnet("sub-1", "my-subnet", "vpc-1"))
	}))
	defer srv.Close()

	c := client.New(srv.URL, "tok")
	sub, found, err := c.GetSubnet("sub-1")
	if err != nil || !found || sub.ID != "sub-1" {
		t.Fatalf("unexpected: %v %v %v", sub, found, err)
	}
}

func TestClient_GetSubnet_NotFound(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		notFoundResp(w)
	}))
	defer srv.Close()

	c := client.New(srv.URL, "tok")
	_, found, err := c.GetSubnet("nonexistent")
	if err != nil || found {
		t.Fatalf("expected not found: found=%v err=%v", found, err)
	}
}

func TestClient_GetSubnet_ServerError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		serverErrResp(w)
	}))
	defer srv.Close()

	c := client.New(srv.URL, "tok")
	_, found, err := c.GetSubnet("sub-1")
	if err == nil || found {
		t.Fatal("expected error")
	}
}

func TestClient_DeleteSubnet(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "DELETE" || r.URL.Path != "/v1/subnets/sub-1" {
			t.Errorf("unexpected: %s %s", r.Method, r.URL.Path)
		}
		w.WriteHeader(204)
	}))
	defer srv.Close()

	c := client.New(srv.URL, "tok")
	if err := c.DeleteSubnet("sub-1"); err != nil {
		t.Fatal(err)
	}
}

func TestClient_DeleteSubnet_NotFound(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		notFoundResp(w)
	}))
	defer srv.Close()

	c := client.New(srv.URL, "tok")
	if err := c.DeleteSubnet("nonexistent"); err != nil {
		t.Fatal("expected no error for 404 delete, got:", err)
	}
}

// ---- Instance ----

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

func TestClient_GetInstance_Found(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/instances/vsi-1" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		w.WriteHeader(200)
		json.NewEncoder(w).Encode(makeInstance("vsi-1", "my-vsi", "sub-1"))
	}))
	defer srv.Close()

	c := client.New(srv.URL, "tok")
	inst, found, err := c.GetInstance("vsi-1")
	if err != nil || !found || inst.ID != "vsi-1" {
		t.Fatalf("unexpected: %v %v %v", inst, found, err)
	}
}

func TestClient_GetInstance_NotFound(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		notFoundResp(w)
	}))
	defer srv.Close()

	c := client.New(srv.URL, "tok")
	_, found, err := c.GetInstance("nonexistent")
	if err != nil || found {
		t.Fatalf("expected not found: found=%v err=%v", found, err)
	}
}

func TestClient_GetInstance_ServerError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		serverErrResp(w)
	}))
	defer srv.Close()

	c := client.New(srv.URL, "tok")
	_, found, err := c.GetInstance("vsi-1")
	if err == nil || found {
		t.Fatal("expected error")
	}
}

func TestClient_InstanceAction(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" || r.URL.Path != "/v1/instances/vsi-1/actions" {
			t.Errorf("unexpected: %s %s", r.Method, r.URL.Path)
		}
		var body map[string]string
		json.NewDecoder(r.Body).Decode(&body)
		if body["type"] != "stop" {
			t.Errorf("expected action stop, got %q", body["type"])
		}
		inst := makeInstance("vsi-1", "my-vsi", "sub-1")
		inst.Status = "stopped"
		w.WriteHeader(200)
		json.NewEncoder(w).Encode(inst)
	}))
	defer srv.Close()

	c := client.New(srv.URL, "tok")
	inst, err := c.InstanceAction("vsi-1", "stop")
	if err != nil || inst.Status != "stopped" {
		t.Fatalf("unexpected: %v %v", inst, err)
	}
}

func TestClient_InstanceAction_Error(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		serverErrResp(w)
	}))
	defer srv.Close()

	c := client.New(srv.URL, "tok")
	_, err := c.InstanceAction("vsi-1", "stop")
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestClient_DeleteInstance(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "DELETE" || r.URL.Path != "/v1/instances/vsi-1" {
			t.Errorf("unexpected: %s %s", r.Method, r.URL.Path)
		}
		w.WriteHeader(204)
	}))
	defer srv.Close()

	c := client.New(srv.URL, "tok")
	if err := c.DeleteInstance("vsi-1"); err != nil {
		t.Fatal(err)
	}
}

func TestClient_DeleteInstance_NotFound(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		notFoundResp(w)
	}))
	defer srv.Close()

	c := client.New(srv.URL, "tok")
	if err := c.DeleteInstance("nonexistent"); err != nil {
		t.Fatal("expected no error for 404 delete, got:", err)
	}
}

// ---- LoadBalancer ----

func TestClient_CreateLoadBalancer(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" || r.URL.Path != "/v1/loadbalancers" {
			t.Errorf("unexpected: %s %s", r.Method, r.URL.Path)
		}
		var body map[string]any
		json.NewDecoder(r.Body).Decode(&body)
		if body["name"] != "my-lb" || body["protocol"] != "http" {
			t.Errorf("unexpected body: %v", body)
		}
		lb := makeLB("lb-1", "my-lb")
		w.WriteHeader(201)
		json.NewEncoder(w).Encode(lb)
	}))
	defer srv.Close()

	c := client.New(srv.URL, "tok")
	lb, err := c.CreateLoadBalancer("my-lb", "http", 80, nil)
	if err != nil || lb.ID != "lb-1" {
		t.Fatalf("unexpected: %v %v", lb, err)
	}
}

func TestClient_CreateLoadBalancer_WithTargets(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		lb := makeLB("lb-1", "my-lb")
		lb.Targets = []client.LoadBalancerTarget{{Type: "vsi", ID: "vsi-1"}}
		w.WriteHeader(201)
		json.NewEncoder(w).Encode(lb)
	}))
	defer srv.Close()

	c := client.New(srv.URL, "tok")
	targets := []client.LoadBalancerTarget{{Type: "vsi", ID: "vsi-1"}}
	lb, err := c.CreateLoadBalancer("my-lb", "tcp", 80, targets)
	if err != nil || len(lb.Targets) != 1 || lb.Targets[0].Type != "vsi" {
		t.Fatalf("unexpected: %v %v", lb, err)
	}
}

func TestClient_CreateLoadBalancer_Error(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		serverErrResp(w)
	}))
	defer srv.Close()

	c := client.New(srv.URL, "tok")
	_, err := c.CreateLoadBalancer("lb", "http", 80, nil)
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestClient_GetLoadBalancer_Found(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/loadbalancers/lb-1" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		w.WriteHeader(200)
		json.NewEncoder(w).Encode(makeLB("lb-1", "my-lb"))
	}))
	defer srv.Close()

	c := client.New(srv.URL, "tok")
	lb, found, err := c.GetLoadBalancer("lb-1")
	if err != nil || !found || lb.ID != "lb-1" {
		t.Fatalf("unexpected: %v %v %v", lb, found, err)
	}
}

func TestClient_GetLoadBalancer_NotFound(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		notFoundResp(w)
	}))
	defer srv.Close()

	c := client.New(srv.URL, "tok")
	_, found, err := c.GetLoadBalancer("nonexistent")
	if err != nil || found {
		t.Fatalf("expected not found: found=%v err=%v", found, err)
	}
}

func TestClient_GetLoadBalancer_ServerError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		serverErrResp(w)
	}))
	defer srv.Close()

	c := client.New(srv.URL, "tok")
	_, found, err := c.GetLoadBalancer("lb-1")
	if err == nil || found {
		t.Fatal("expected error")
	}
}

func TestClient_DeleteLoadBalancer(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "DELETE" || r.URL.Path != "/v1/loadbalancers/lb-1" {
			t.Errorf("unexpected: %s %s", r.Method, r.URL.Path)
		}
		w.WriteHeader(204)
	}))
	defer srv.Close()

	c := client.New(srv.URL, "tok")
	if err := c.DeleteLoadBalancer("lb-1"); err != nil {
		t.Fatal(err)
	}
}

func TestClient_DeleteLoadBalancer_NotFound(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		notFoundResp(w)
	}))
	defer srv.Close()

	c := client.New(srv.URL, "tok")
	if err := c.DeleteLoadBalancer("nonexistent"); err != nil {
		t.Fatal("expected no error for 404 delete, got:", err)
	}
}

// ---- Bucket ----

func TestClient_CreateBucket(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" || r.URL.Path != "/v1/buckets" {
			t.Errorf("unexpected: %s %s", r.Method, r.URL.Path)
		}
		var body map[string]any
		json.NewDecoder(r.Body).Decode(&body)
		if body["name"] != "my-bucket" || body["region"] != "eu-west" {
			t.Errorf("unexpected body: %v", body)
		}
		bkt := makeBucket("bkt-1", "my-bucket")
		bkt.Region = "eu-west"
		w.WriteHeader(201)
		json.NewEncoder(w).Encode(bkt)
	}))
	defer srv.Close()

	c := client.New(srv.URL, "tok")
	bkt, err := c.CreateBucket("my-bucket", "eu-west")
	if err != nil || bkt.ID != "bkt-1" || bkt.Region != "eu-west" {
		t.Fatalf("unexpected: %v %v", bkt, err)
	}
}

func TestClient_CreateBucket_Error(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		serverErrResp(w)
	}))
	defer srv.Close()

	c := client.New(srv.URL, "tok")
	_, err := c.CreateBucket("bucket", "us-east")
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestClient_GetBucket_Found(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/buckets/bkt-1" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		w.WriteHeader(200)
		json.NewEncoder(w).Encode(makeBucket("bkt-1", "my-bucket"))
	}))
	defer srv.Close()

	c := client.New(srv.URL, "tok")
	bkt, found, err := c.GetBucket("bkt-1")
	if err != nil || !found || bkt.ID != "bkt-1" {
		t.Fatalf("unexpected: %v %v %v", bkt, found, err)
	}
}

func TestClient_GetBucket_NotFound(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		notFoundResp(w)
	}))
	defer srv.Close()

	c := client.New(srv.URL, "tok")
	_, found, err := c.GetBucket("nonexistent")
	if err != nil || found {
		t.Fatalf("expected not found: found=%v err=%v", found, err)
	}
}

func TestClient_GetBucket_ServerError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		serverErrResp(w)
	}))
	defer srv.Close()

	c := client.New(srv.URL, "tok")
	_, found, err := c.GetBucket("bkt-1")
	if err == nil || found {
		t.Fatal("expected error")
	}
}

func TestClient_DeleteBucket(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "DELETE" || r.URL.Path != "/v1/buckets/bkt-1" {
			t.Errorf("unexpected: %s %s", r.Method, r.URL.Path)
		}
		w.WriteHeader(204)
	}))
	defer srv.Close()

	c := client.New(srv.URL, "tok")
	if err := c.DeleteBucket("bkt-1"); err != nil {
		t.Fatal(err)
	}
}

func TestClient_DeleteBucket_NotFound(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		notFoundResp(w)
	}))
	defer srv.Close()

	c := client.New(srv.URL, "tok")
	if err := c.DeleteBucket("nonexistent"); err != nil {
		t.Fatal("expected no error for 404 delete, got:", err)
	}
}

// ---- Database ----

func TestClient_CreateDatabase(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" || r.URL.Path != "/v1/databases" {
			t.Errorf("unexpected: %s %s", r.Method, r.URL.Path)
		}
		var body map[string]any
		json.NewDecoder(r.Body).Decode(&body)
		if body["name"] != "my-db" || body["engine"] != "postgres" {
			t.Errorf("unexpected body: %v", body)
		}
		w.WriteHeader(201)
		json.NewEncoder(w).Encode(makeDatabase("db-1", "my-db"))
	}))
	defer srv.Close()

	c := client.New(srv.URL, "tok")
	db, err := c.CreateDatabase("my-db", "postgres", "15", "small", []string{"sub-1"})
	if err != nil || db.ID != "db-1" || db.Engine != "postgres" {
		t.Fatalf("unexpected: %v %v", db, err)
	}
}

func TestClient_CreateDatabase_Error(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		serverErrResp(w)
	}))
	defer srv.Close()

	c := client.New(srv.URL, "tok")
	_, err := c.CreateDatabase("db", "postgres", "15", "small", []string{"sub-1"})
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestClient_GetDatabase_Found(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/databases/db-1" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		w.WriteHeader(200)
		json.NewEncoder(w).Encode(makeDatabase("db-1", "my-db"))
	}))
	defer srv.Close()

	c := client.New(srv.URL, "tok")
	db, found, err := c.GetDatabase("db-1")
	if err != nil || !found || db.ID != "db-1" {
		t.Fatalf("unexpected: %v %v %v", db, found, err)
	}
	if len(db.SubnetIDs) != 1 {
		t.Fatalf("expected subnet_ids, got %v", db.SubnetIDs)
	}
}

func TestClient_GetDatabase_NotFound(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		notFoundResp(w)
	}))
	defer srv.Close()

	c := client.New(srv.URL, "tok")
	_, found, err := c.GetDatabase("nonexistent")
	if err != nil || found {
		t.Fatalf("expected not found: found=%v err=%v", found, err)
	}
}

func TestClient_GetDatabase_ServerError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		serverErrResp(w)
	}))
	defer srv.Close()

	c := client.New(srv.URL, "tok")
	_, found, err := c.GetDatabase("db-1")
	if err == nil || found {
		t.Fatal("expected error")
	}
}

func TestClient_DeleteDatabase(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "DELETE" || r.URL.Path != "/v1/databases/db-1" {
			t.Errorf("unexpected: %s %s", r.Method, r.URL.Path)
		}
		w.WriteHeader(204)
	}))
	defer srv.Close()

	c := client.New(srv.URL, "tok")
	if err := c.DeleteDatabase("db-1"); err != nil {
		t.Fatal(err)
	}
}

func TestClient_DeleteDatabase_NotFound(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		notFoundResp(w)
	}))
	defer srv.Close()

	c := client.New(srv.URL, "tok")
	if err := c.DeleteDatabase("nonexistent"); err != nil {
		t.Fatal("expected no error for 404 delete, got:", err)
	}
}

// ---- KubernetesCluster ----

func TestClient_CreateKubernetesCluster(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" || r.URL.Path != "/v1/clusters" {
			t.Errorf("unexpected: %s %s", r.Method, r.URL.Path)
		}
		var body map[string]any
		json.NewDecoder(r.Body).Decode(&body)
		if body["name"] != "my-cluster" || body["version"] != "1.30" {
			t.Errorf("unexpected body: %v", body)
		}
		w.WriteHeader(201)
		json.NewEncoder(w).Encode(makeCluster("k8s-1", "my-cluster"))
	}))
	defer srv.Close()

	c := client.New(srv.URL, "tok")
	cluster, err := c.CreateKubernetesCluster("my-cluster", "1.30", 2, []string{"sub-1"})
	if err != nil || cluster.ID != "k8s-1" || cluster.Version != "1.30" {
		t.Fatalf("unexpected: %v %v", cluster, err)
	}
}

func TestClient_CreateKubernetesCluster_Error(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		serverErrResp(w)
	}))
	defer srv.Close()

	c := client.New(srv.URL, "tok")
	_, err := c.CreateKubernetesCluster("cluster", "1.30", 2, []string{"sub-1"})
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestClient_GetKubernetesCluster_Found(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/clusters/k8s-1" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		w.WriteHeader(200)
		json.NewEncoder(w).Encode(makeCluster("k8s-1", "my-cluster"))
	}))
	defer srv.Close()

	c := client.New(srv.URL, "tok")
	cluster, found, err := c.GetKubernetesCluster("k8s-1")
	if err != nil || !found || cluster.ID != "k8s-1" {
		t.Fatalf("unexpected: %v %v %v", cluster, found, err)
	}
	if len(cluster.SubnetIDs) != 1 {
		t.Fatalf("expected subnet_ids, got %v", cluster.SubnetIDs)
	}
}

func TestClient_GetKubernetesCluster_NotFound(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		notFoundResp(w)
	}))
	defer srv.Close()

	c := client.New(srv.URL, "tok")
	_, found, err := c.GetKubernetesCluster("nonexistent")
	if err != nil || found {
		t.Fatalf("expected not found: found=%v err=%v", found, err)
	}
}

func TestClient_GetKubernetesCluster_ServerError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		serverErrResp(w)
	}))
	defer srv.Close()

	c := client.New(srv.URL, "tok")
	_, found, err := c.GetKubernetesCluster("k8s-1")
	if err == nil || found {
		t.Fatal("expected error")
	}
}

func TestClient_DeleteKubernetesCluster(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "DELETE" || r.URL.Path != "/v1/clusters/k8s-1" {
			t.Errorf("unexpected: %s %s", r.Method, r.URL.Path)
		}
		w.WriteHeader(204)
	}))
	defer srv.Close()

	c := client.New(srv.URL, "tok")
	if err := c.DeleteKubernetesCluster("k8s-1"); err != nil {
		t.Fatal(err)
	}
}

func TestClient_DeleteKubernetesCluster_NotFound(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		notFoundResp(w)
	}))
	defer srv.Close()

	c := client.New(srv.URL, "tok")
	if err := c.DeleteKubernetesCluster("nonexistent"); err != nil {
		t.Fatal("expected no error for 404 delete, got:", err)
	}
}

// ---- Error response formats ----

func TestClient_ErrorResponse(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(404)
		json.NewEncoder(w).Encode(map[string]any{
			"errors": []map[string]string{{"code": "not_found", "message": "VPC not found"}},
		})
	}))
	defer srv.Close()

	c := client.New(srv.URL, "tok")
	_, err := c.CreateVPC("x", "us-east")
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestClient_ErrorResponse_NoBody(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(503)
	}))
	defer srv.Close()

	c := client.New(srv.URL, "tok")
	_, err := c.CreateVPC("x", "us-east")
	if err == nil {
		t.Fatal("expected error")
	}
}
