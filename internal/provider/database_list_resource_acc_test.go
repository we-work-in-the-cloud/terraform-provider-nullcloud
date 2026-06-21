package provider

import (
	"context"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/list"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
	"github.com/we-work-in-the-cloud/nullcloud/terraform-provider-nullcloud/internal/client"
)

func TestDatabaseListResource_Metadata(t *testing.T) {
	r := &DatabaseListResource{}
	req := resource.MetadataRequest{ProviderTypeName: "nullcloud"}
	resp := &resource.MetadataResponse{}

	r.Metadata(context.Background(), req, resp)

	assert.Equal(t, "nullcloud_database", resp.TypeName)
}

func TestDatabaseListResource_ListResourceConfigSchema(t *testing.T) {
	r := &DatabaseListResource{}
	resp := &list.ListResourceSchemaResponse{}

	r.ListResourceConfigSchema(context.Background(), list.ListResourceSchemaRequest{}, resp)

	assert.NotNil(t, resp.Schema)
	assert.Contains(t, resp.Schema.Attributes, "engine")
}

func TestDatabaseListResource_Configure(t *testing.T) {
	mockClient := &client.Client{}
	r := &DatabaseListResource{}
	req := resource.ConfigureRequest{ProviderData: mockClient}
	resp := &resource.ConfigureResponse{}

	r.Configure(context.Background(), req, resp)

	assert.Equal(t, mockClient, r.client)
}

func TestDatabaseListResource_FilterLogic(t *testing.T) {
	databases := []client.Database{
		{ID: "db-1", Name: "db-1", Engine: "postgres", Version: "13", Status: "available", CreatedAt: time.Now()},
		{ID: "db-2", Name: "db-2", Engine: "mysql", Version: "8.0", Status: "available", CreatedAt: time.Now()},
		{ID: "db-3", Name: "db-3", Engine: "postgres", Version: "14", Status: "available", CreatedAt: time.Now()},
	}

	// Test: no filter (all results)
	filtered := filterDatabases(databases, "")
	assert.Equal(t, 3, len(filtered))

	// Test: filter by engine
	filtered = filterDatabases(databases, "postgres")
	assert.Equal(t, 2, len(filtered))

	filtered = filterDatabases(databases, "mysql")
	assert.Equal(t, 1, len(filtered))
}

func filterDatabases(databases []client.Database, engine string) []client.Database {
	var result []client.Database
	for _, db := range databases {
		if engine == "" || db.Engine == engine {
			result = append(result, db)
		}
	}
	return result
}

func TestDatabaseListResource_ConfigModel(t *testing.T) {
	config := databaseListConfig{
		Engine: types.StringValue("postgres"),
	}

	assert.Equal(t, "postgres", config.Engine.ValueString())
}

func TestDatabaseListResource_NewDatabaseListResource(t *testing.T) {
	lr := NewDatabaseListResource()
	assert.NotNil(t, lr)
	assert.IsType(t, &DatabaseListResource{}, lr)
}

func TestDatabaseListResource_ToDatabaseModel(t *testing.T) {
	now := time.Now()
	clientDB := client.Database{
		ID:        "db-1",
		Name:      "my-db",
		Engine:    "postgres",
		Status:    "available",
		CRN:       "crn:db-1",
		CreatedAt: now,
	}

	model := toDatabaseModel(clientDB)

	assert.Equal(t, "db-1", model.ID.ValueString())
	assert.Equal(t, "postgres", model.Engine.ValueString())
	assert.Equal(t, "available", model.Status.ValueString())
}

func TestDatabaseListResource_ShouldIncludeDatabase(t *testing.T) {
	db := client.Database{
		ID:        "db-1",
		Engine:    "postgres",
		Status:    "available",
		CreatedAt: time.Now(),
	}

	config := databaseListConfig{Engine: types.StringValue("postgres")}
	assert.True(t, shouldIncludeDatabase(db, config))

	config = databaseListConfig{Engine: types.StringValue("mysql")}
	assert.False(t, shouldIncludeDatabase(db, config))
}
