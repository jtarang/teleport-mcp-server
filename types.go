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
