package main

import (
	"context"
	"flag"
	"log"
	"net"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func main() {

	ctx := context.Background()

	proxyAddress := flag.String("proxy", "", "-proxy some-domain.teleport.sh")
	proxyPorts := flag.String("proxy-ports", "443,3025,3024,3080", "comma-separated list of ports")
	flag.Parse()

	portsStr := strings.Split(*proxyPorts, ",")
	proxyAddresses := make([]string, 0, len(portsStr))

	if *proxyAddress != "" {
		_, err := url.Parse(*proxyAddress)
		if err != nil {
			log.Fatalf("Invalid proxy address %q: %v", *proxyAddress, err)
		}
	}

	for _, port := range portsStr {
		port = strings.TrimSpace(port)
		portInt, err := strconv.Atoi(port)
		if err != nil || portInt < 1 || portInt > 65535 {
			log.Fatalf("Invalid port %q: must be 1–65535", port)
		}
		proxyAddresses = append(proxyAddresses, net.JoinHostPort(*proxyAddress, strconv.Itoa(portInt)))
	}
	config := TeleportConfig{
		ProxyAddresses: proxyAddresses,
	}

	// Create the TeleportClientManager and establish connection
	tcm, err := NewTeleportClientManager(ctx, config)
	if err != nil {
		log.Fatalf("❌ Failed to initialize Teleport Client Manager: %v", err)
	}

	server := mcp.NewServer(&mcp.Implementation{Name: "Teleport MCP Server", Version: "v1.0.0"}, nil)

	mcp.AddTool(server, &mcp.Tool{
		Name:        "getRoles",
		Description: "Retrieves the names of all roles in the Teleport cluster.",
	},
		MakeListTeleportRoles(tcm))

	mcp.AddTool(server, &mcp.Tool{
		Name:        "createRole",
		Description: "Creates a role in the Teleport cluster.",
	},
		MakeCreateRole(tcm))

	mcp.AddTool(server, &mcp.Tool{
		Name:        "createAccessRequest",
		Description: "Creates a New Access Request in Teleport.",
	},
		MakeCreateAccessRequest(tcm))

	handler := mcp.NewStreamableHTTPHandler(
		func(*http.Request) *mcp.Server { return server },
		&mcp.StreamableHTTPOptions{},
	)

	http.HandleFunc("/mcp", handler.ServeHTTP)
	log.Println("MCP Server running at http://localhost:9000/mcp")
	if err := http.ListenAndServe(":9000", nil); err != nil {
		log.Fatal(err)
	}
}
