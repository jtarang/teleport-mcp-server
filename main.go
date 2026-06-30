package main

import (
	"context"
	"flag"
	"log"
	"net"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func main() {

	ctx := context.Background()

	proxyAddress := flag.String("proxy", "", "-proxy some-domain.teleport.sh")
	proxyPorts := flag.String("proxy-ports", "443,3025,3024,3080", "comma-separated list of ports")
	transport := flag.String("transport", "stdio", "transport: 'stdio' (for desktop/IDE) or 'http'")
	flag.Parse()

	// Allow the proxy to come from the environment too, which is how the desktop
	// app passes configuration to a stdio server.
	if *proxyAddress == "" {
		*proxyAddress = os.Getenv("TELEPORT_PROXY")
	}

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

	mcp.AddTool(server, &mcp.Tool{
		Name:        "listDatabases",
		Description: "Lists the databases registered in the Teleport cluster (name, protocol, URI).",
	},
		MakeListDatabases(tcm))

	mcp.AddTool(server, &mcp.Tool{
		Name:        "queryDatabase",
		Description: "Runs a SQL query against a Teleport-protected Postgres database via a local 'tsh proxy db' tunnel. Requires db_user, db_name, and the Teleport database name.",
	},
		MakeQueryDatabase(tcm))

	switch *transport {
	case "stdio":
		// Desktop/IDE launch the binary and talk over stdin/stdout. Only protocol
		// bytes may touch stdout; log output goes to stderr, which is safe.
		log.Println("MCP Server running on stdio")
		if err := server.Run(ctx, &mcp.StdioTransport{}); err != nil {
			log.Fatalf("stdio server error: %v", err)
		}
	case "http":
		handler := mcp.NewStreamableHTTPHandler(
			func(*http.Request) *mcp.Server { return server },
			&mcp.StreamableHTTPOptions{},
		)

		http.HandleFunc("/mcp", handler.ServeHTTP)
		log.Println("MCP Server running at http://localhost:9000/mcp")
		if err := http.ListenAndServe(":9000", nil); err != nil {
			log.Fatal(err)
		}
	default:
		log.Fatalf("unknown transport %q: use 'stdio' or 'http'", *transport)
	}
}
