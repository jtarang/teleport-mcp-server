package main

import (
	"context"
	"fmt"
	"log"

	"github.com/gravitational/trace"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// MakeListTeleportRoles is a factory function that returns the handler.
func MakeListTeleportRoles(tcm *TeleportClientManager) func(ctx context.Context, req *mcp.CallToolRequest, input ListRolesInput) (*mcp.CallToolResult, RoleListResult, error) {

	return func(ctx context.Context, req *mcp.CallToolRequest, input ListRolesInput) (*mcp.CallToolResult, RoleListResult, error) {
		log.Println("Invoking ListTeleportRoles")

		roles, err := tcm.ListRoles(ctx)
		if err != nil {
			log.Printf("⚠️ Warning: Could not list roles: %v", err)
			return nil, RoleListResult{}, trace.Wrap(err, "failed to list roles")
		}

		log.Printf("📝 Current roles count: %d. Names: %v", len(roles), roles)

		roleList := RoleListResult{
			Roles: roles,
			Count: len(roles),
		}

		return nil, roleList, nil
	}
}

// CreateRoleInput uses a simpler field name for the config.
type CreateRoleInput struct {
	Config RoleConfig `json:"role_config"`
}

// CreateRoleOutput defines the structured output returned by the CreateRole handler.
type CreateRoleOutput struct {
	Message  string `json:"message"`
	RoleName string `json:"role_name"`
	Status   string `json:"status"`
}

// MakeCreateRole is a factory function that returns the handler.
func MakeCreateRole(tcm *TeleportClientManager) func(ctx context.Context, req *mcp.CallToolRequest, input CreateRoleInput) (*mcp.CallToolResult, CreateRoleOutput, error) {

	return func(ctx context.Context, req *mcp.CallToolRequest, input CreateRoleInput) (*mcp.CallToolResult, CreateRoleOutput, error) {
		roleConf := input.Config
		log.Printf("Invoking CreateTeleportRole for: %s", roleConf.Name)

		// Attempt to create the role
		err := tcm.CreateRole(ctx, roleConf)

		// Default output for success
		output := CreateRoleOutput{
			RoleName: roleConf.Name,
			Message:  fmt.Sprintf("Teleport Role '%s' created successfully.", roleConf.Name),
			Status:   "CREATED",
		}

		if err != nil {
			if trace.IsAlreadyExists(err) {
				output.Message = fmt.Sprintf("Role '%s' already exists (VERIFIED).", roleConf.Name)
				output.Status = "VERIFIED"
				// Return success, as the desired state is met
				return nil, output, nil
			}

			// Return actual creation error
			log.Printf("❌ Failed to create role %s: %v", roleConf.Name, err)
			return nil, CreateRoleOutput{}, trace.Wrap(err, "failed to create role")
		}

		return nil, output, nil
	}
}

// MakeCreateAccessRequest is a factory function that returns the handler function.
func MakeCreateAccessRequest(tcm *TeleportClientManager) func(ctx context.Context, req *mcp.CallToolRequest, input AccessRequestInput) (*mcp.CallToolResult, AccessRequestOutput, error) {

	return func(ctx context.Context, req *mcp.CallToolRequest, input AccessRequestInput) (*mcp.CallToolResult, AccessRequestOutput, error) {

		requestConf := input.RequestConfig

		if requestConf.Username == "" || len(requestConf.Roles) == 0 || requestConf.Reason == "" {
			return nil, AccessRequestOutput{}, fmt.Errorf("username, roles, and reason are required for access request")
		}

		log.Printf("Submitting access request for user %s for roles: %v", requestConf.Username, requestConf.Roles)

		// Call the tcm method to create the request
		requestID, err := tcm.CreateAccessRequest(ctx, requestConf)

		if err != nil {
			return nil, AccessRequestOutput{}, trace.Wrap(err, "failed to create access request")
		}

		// Success Output
		output := AccessRequestOutput{
			Message:   fmt.Sprintf("Access request %s submitted for user %s. Waiting for approval.", requestID, requestConf.Username),
			RequestID: requestID,
			Status:    "PENDING",
		}

		return nil, output, nil
	}
}
