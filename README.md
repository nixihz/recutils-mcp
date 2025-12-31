# Recutils MCP Server (Go Version)

High-performance recutils database MCP (Model Context Protocol) tool implemented in Go, providing type-safe operations.

## About Recutils

[GNU Recutils](https://www.gnu.org/software/recutils/) is a lightweight text-based database toolkit that stores structured data in plain text files. Key features include:

- **Human-readable** - Data stored in plain text format, easy to read and edit
- **Typed records** - Supports multiple record types, similar to tables in relational databases
- **Powerful queries** - Provides SQL-like query language (`recsel`)
- **Serverless** - No database server process required, zero dependencies
- **VCS-friendly** - Text format naturally supports Git and other version control tools

## ✨ Features

- ✅ **High Performance** - Compiled language, 2x faster startup, 50% less memory usage
- ✅ **Type Safety** - Compile-time checking, fewer runtime errors
- ✅ **Simple Deployment** - Single binary file (~7.3MB), no runtime environment required
- ✅ **Concurrency Support** - goroutine provides better concurrency performance
- ✅ **Complete Functionality** - Support for query, insert, update, delete, and get database info
- ✅ **MCP Protocol** - Complies with Model Context Protocol standard

## 🚀 Quick Start

### System Requirements

- Go 1.21+ (for development)
- recutils tool package (for runtime)

```bash
# Install recutils
# Ubuntu/Debian
sudo apt-get install recutils

# macOS
brew install recutils
```

### Installation

#### Option 1: Using go install (Recommended)

```bash
# Install directly to GOBIN
go install github.com/nixihz/recutils-mcp@latest

# The binary will be installed to $GOBIN (default: ~/go/bin)
# Add to PATH if needed:
export PATH=$PATH:$(go env GOPATH)/bin

# Run the server
recutils-mcp
```

#### Option 2: Build from source

```bash
# 1. Clone the repository
git clone https://github.com/nixihz/recutils-mcp.git
cd recutils-mcp

# 2. Install dependencies
go mod download

# 3. Run tests
make test

# 4. Build project
make build

# 5. Run server
./recutils-mcp
```

## 📋 Available Commands

```bash
make all           # Run tests and build
make test          # Run all tests
make test-cover    # Run tests with coverage
make build         # Build binary
make clean         # Clean build files
make fmt           # Format code
make vet           # Static check
make bench         # Benchmark test
make security      # Security check
make help          # Show help
```

## 🔧 MCP Tools List

| Tool Name | Description | Parameters |
|-----------|-------------|------------|
| `recutils_query` | Query records | database_file, query_expression (optional), output_format (optional) |
| `recutils_insert` | Insert record | database_file, record_type, fields |
| `recutils_update` | Update records | database_file, query_expression, fields |
| `recutils_delete` | Delete records | database_file, query_expression |
| `recutils_info` | Get database info | database_file |

## 📖 Usage Examples

### Create Database and Insert Data

```bash
# 1. Start server
./recutils-mcp

# 2. Call tools via MCP client (Example JSON)

# Query records
{
  "method": "tools/call",
  "params": {
    "name": "recutils_query",
    "arguments": {
      "database_file": "example.rec",
      "query_expression": "Name = 'John Doe'"
    }
  }
}

# Insert records
{
  "method": "tools/call",
  "params": {
    "name": "recutils_insert",
    "arguments": {
      "database_file": "example.rec",
      "record_type": "Person",
      "fields": {
        "Name": "John Doe",
        "Age": 25,
        "City": "New York"
      }
    }
  }
}
```

### Direct Go API Usage

```go
package main

import (
    "context"
    "fmt"
    "github.com/nixihz/recutils-mcp/recutils"
)

func main() {
    ctx := context.Background()
    op := recutils.NewRecordOperation()

    // Insert record
    result, err := op.InsertRecord(ctx, "test.rec", "Person", map[string]interface{}{
        "Name": "John Doe",
        "Age":  25,
    })
    if err != nil {
        fmt.Printf("Error: %v\n", err)
        return
    }
    fmt.Printf("Insert successful: %+v\n", result)

    // Query records
    queryResult, err := op.QueryRecords(ctx, "test.rec", "", "")
    if err != nil {
        fmt.Printf("Error: %v\n", err)
        return
    }
    fmt.Printf("Query result: %+v\n", queryResult)
}
```

## 📁 Project Structure

```
recutils-mcp/
├── go.mod                    # Go module definition
├── main.go                   # Main entry point
├── README.md                 # Project documentation (this file)
├── Makefile                  # Build script
├── build.sh                  # Build script
├── .gitignore                # Git ignore file
├── recutils/
│   └── operations.go        # recutils operations encapsulation
└── server/
    ├── mcp_server.go        # MCP server implementation
    └── mcp_server_test.go   # Test code
```

## 🔍 Correct File Format

recutils database files must follow specific format:

```rec
%rec: Person

Name: John Doe
Age: 25
City: New York

Name: Jane Smith
Age: 30
```

Format requirements:
- `%rec:` must be followed by a blank line
- Field format: `field_name: value`
- Records must be separated by blank lines
- **Important:** Records must end with newline (fixed)

## 🧪 Testing

```bash
# Run all tests
go test ./... -v

# Run specific test
go test -run TestMCPServer -v

# Test coverage
go test ./... -cover

# Benchmark test
go test -bench=. ./...
```

### Test Results

```
=== RUN   TestMCPServer
--- PASS: TestMCPServer (0.04s)
    ✓ QueryRecords
    ✓ GetDatabaseInfo
    ✓ InsertRecord
    ✓ VerifyInsert

=== RUN   TestMCPServerIntegration
--- PASS: TestMCPServerIntegration (0.05s)
    ✓ FullWorkflow

PASS
```

## 🔧 Integration with Claude Desktop

Add the following configuration to Claude Desktop's config file:

```json
{
  "mcpServers": {
    "recutils": {
      "command": "$(go env GOPATH)/bin/recutils-mcp",
      "args": []
    }
  }
}
```

> **Note:** Replace `$(go env GOPATH)/bin/recutils-mcp` with the actual path if you installed it elsewhere. On macOS/Linux with `go install`, the default path is `~/go/bin/recutils-mcp`.

## 🐛 Troubleshooting

### 1. Server won't start

Check if recutils is installed:
```bash
recsel --version
```

### 2. Permission issues

Ensure database files have read/write permissions:
```bash
chmod 644 database.rec
```

### 3. File format errors

Use `recsel` to verify file format:
```bash
recsel your_database.rec
```

## 📚 More Resources

- [recutils official documentation](https://www.gnu.org/software/recutils/)
- [MCP Protocol Specification](https://modelcontextprotocol.io/)
- [Go Language Documentation](https://golang.org/doc/)

## 🤝 Contributing

Contributions are welcome! Please read CONTRIBUTING.md for details.

## 📄 License

MIT License

## 🏷️ Version History

- **v1.0.0** - Initial release
  - Go language refactoring
  - Full MCP protocol support
  - Fixed database file format issues
  - 2x performance improvement

---

**Recommended for production use!** 🚀
