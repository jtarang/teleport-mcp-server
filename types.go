package main

import (
	"time"

	"github.com/gravitational/teleport/api/types"
)

// AccessRequestInput defines the input parameters for the handler.
type AccessRequestInput struct {
	RequestConfig AccessRequestConfig
}

// AccessRequestConfig defines the core fields needed to create a new access request.
type AccessRequestConfig struct {
	// Roles is a list of Teleport Role names being requested (e.g., ["editor", "dev-access-role"]).
	Roles []string
	// Reason is a required string explaining why the user needs the access.
	Reason string
	// Resources is a list of resources (e.g., nodes, databases, K8s clusters)
	Resources []types.ResourceID
	// Username Requesters User ID
	Username string
	//Name of the Access Request
	Name string
}

// AccessRequestOutput defines the structured output returned by the handler.
type AccessRequestOutput struct {
	// Message provides a brief summary of the action taken.
	Message string
	// RequestID is the unique ID of the created access request.
	RequestID string
	// Status is the state of the request (e.g., "PENDING").
	Status string
}

// RoleConfig defines the essential configuration for the new Teleport role
type RoleConfig struct {
	Name           string
	RoleConditions types.RoleConditions
	MaxSessionTTL  time.Duration
}

// RoleListResult defines the structure for the roles
type RoleListResult struct {
	Roles []string
	Count int
}

// ListRolesInput is an empty placeholder struct
type ListRolesInput struct{}

// ListDatabasesInput is an empty placeholder struct.
type ListDatabasesInput struct{}

// DatabaseInfo describes a single database registered in the Teleport cluster.
type DatabaseInfo struct {
	Name        string `json:"name"`
	Protocol    string `json:"protocol"`
	URI         string `json:"uri"`
	Description string `json:"description"`
}

// DatabaseListResult defines the structure returned by the listDatabases handler.
type DatabaseListResult struct {
	Databases []DatabaseInfo `json:"databases"`
	Count     int            `json:"count"`
}

// QueryDatabaseInput defines the parameters for running a SQL query through
// Teleport. Traffic is routed via a short-lived `tsh proxy db --tunnel`
// listener, so the caller must already be logged in (tsh login).
type QueryDatabaseInput struct {
	// Database is the Teleport database resource name (see listDatabases).
	Database string `json:"database"`
	// DBUser is the database account to authenticate as (e.g. "postgres").
	DBUser string `json:"db_user"`
	// DBName is the logical database/schema to connect to.
	DBName string `json:"db_name"`
	// Query is the SQL statement to execute.
	Query string `json:"query"`
}

// QueryDatabaseOutput defines the structured result of a SQL query.
type QueryDatabaseOutput struct {
	Columns  []string   `json:"columns"`
	Rows     [][]string `json:"rows"`
	RowCount int        `json:"row_count"`
	Message  string     `json:"message"`
}
