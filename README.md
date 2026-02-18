# 🚀 Teleport MCP Server
---

## ▶️ Running the Server

```bash
git clone https://github.com/jtarang/teleport-mcp-server.git
go run . --proxy your-proxy-address.teleport.sh
```

Server runs at:

```
http://localhost:9000
```

---

# 🧩 Extending the API

You can extend this API server further:

- Approve / deny access requests
- List databases, servers, apps, kubernetes clusters
- Manage Teleport users

Teleport API docs:  
`https://pkg.go.dev/github.com/gravitational/teleport/api/client`
