// server package: MCP server implementation
package server

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/recutils-mcp/recutils-mcp/recutils"
)

// MCPServer MCP server implementation
type MCPServer struct {
	recutilsOp *recutils.RecordOperation
}

// NewMCPServer Create new MCP server
func NewMCPServer() *MCPServer {
	return &MCPServer{
		recutilsOp: recutils.NewRecordOperation(),
	}
}

// QueryArgs Query parameter structure
type QueryArgs struct {
	DatabaseFile    string `json:"database_file"`
	QueryExpression string `json:"query_expression,omitempty"`
	OutputFormat    string `json:"output_format,omitempty"`
}

// InsertArgs Insert parameter structure
type InsertArgs struct {
	DatabaseFile string                 `json:"database_file"`
	RecordType   string                 `json:"record_type"`
	Fields       map[string]interface{} `json:"fields"`
}

// UpdateArgs Update parameter structure
type UpdateArgs struct {
	DatabaseFile    string                 `json:"database_file"`
	QueryExpression string                 `json:"query_expression"`
	Fields          map[string]interface{} `json:"fields"`
}

// DeleteArgs Delete parameter structure
type DeleteArgs struct {
	DatabaseFile    string `json:"database_file"`
	QueryExpression string `json:"query_expression"`
}

// InfoArgs Info parameter structure
type InfoArgs struct {
	DatabaseFile string `json:"database_file"`
}

// SetupTools Setup MCP tools
func (s *MCPServer) SetupTools(server *mcp.Server) error {
	// Add tool: Query records
	mcp.AddTool(server, &mcp.Tool{
		Name:        "recutils_query",
		Description: "Query records in recutils database",
	}, func(ctx context.Context, req *mcp.CallToolRequest, args QueryArgs) (*mcp.CallToolResult, any, error) {
		result, err := s.recutilsOp.QueryRecords(ctx, args.DatabaseFile, args.QueryExpression, args.OutputFormat)
		if err != nil {
			return &mcp.CallToolResult{
				Content: []mcp.Content{
					&mcp.TextContent{Text: fmt.Sprintf("Error: %v", err)},
				},
			}, nil, nil
		}

		resultJSON, err := json.Marshal(result)
		if err != nil {
			return &mcp.CallToolResult{
				Content: []mcp.Content{
					&mcp.TextContent{Text: fmt.Sprintf("Error marshaling result: %v", err)},
				},
			}, nil, nil
		}

		return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{Text: string(resultJSON)},
			},
		}, nil, nil
	})

	// Add tool: Insert records
	mcp.AddTool(server, &mcp.Tool{
		Name:        "recutils_insert",
		Description: "Insert new record into recutils database",
	}, func(ctx context.Context, req *mcp.CallToolRequest, args InsertArgs) (*mcp.CallToolResult, any, error) {
		result, err := s.recutilsOp.InsertRecord(ctx, args.DatabaseFile, args.RecordType, args.Fields)
		if err != nil {
			return &mcp.CallToolResult{
				Content: []mcp.Content{
					&mcp.TextContent{Text: fmt.Sprintf("Error: %v", err)},
				},
			}, nil, nil
		}

		resultJSON, err := json.Marshal(result)
		if err != nil {
			return &mcp.CallToolResult{
				Content: []mcp.Content{
					&mcp.TextContent{Text: fmt.Sprintf("Error marshaling result: %v", err)},
				},
			}, nil, nil
		}

		return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{Text: string(resultJSON)},
			},
		}, nil, nil
	})

	// Add tool: Update records
	mcp.AddTool(server, &mcp.Tool{
		Name:        "recutils_update",
		Description: "Update records in recutils database",
	}, func(ctx context.Context, req *mcp.CallToolRequest, args UpdateArgs) (*mcp.CallToolResult, any, error) {
		result, err := s.recutilsOp.UpdateRecords(ctx, args.DatabaseFile, args.QueryExpression, args.Fields)
		if err != nil {
			return &mcp.CallToolResult{
				Content: []mcp.Content{
					&mcp.TextContent{Text: fmt.Sprintf("Error: %v", err)},
				},
			}, nil, nil
		}

		resultJSON, err := json.Marshal(result)
		if err != nil {
			return &mcp.CallToolResult{
				Content: []mcp.Content{
					&mcp.TextContent{Text: fmt.Sprintf("Error marshaling result: %v", err)},
				},
			}, nil, nil
		}

		return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{Text: string(resultJSON)},
			},
		}, nil, nil
	})

	// Add tool: Delete records
	mcp.AddTool(server, &mcp.Tool{
		Name:        "recutils_delete",
		Description: "Delete records from recutils database",
	}, func(ctx context.Context, req *mcp.CallToolRequest, args DeleteArgs) (*mcp.CallToolResult, any, error) {
		result, err := s.recutilsOp.DeleteRecords(ctx, args.DatabaseFile, args.QueryExpression)
		if err != nil {
			return &mcp.CallToolResult{
				Content: []mcp.Content{
					&mcp.TextContent{Text: fmt.Sprintf("Error: %v", err)},
				},
			}, nil, nil
		}

		resultJSON, err := json.Marshal(result)
		if err != nil {
			return &mcp.CallToolResult{
				Content: []mcp.Content{
					&mcp.TextContent{Text: fmt.Sprintf("Error marshaling result: %v", err)},
				},
			}, nil, nil
		}

		return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{Text: string(resultJSON)},
			},
		}, nil, nil
	})

	// Add tool: Get database info
	mcp.AddTool(server, &mcp.Tool{
		Name:        "recutils_info",
		Description: "Get recutils database info",
	}, func(ctx context.Context, req *mcp.CallToolRequest, args InfoArgs) (*mcp.CallToolResult, any, error) {
		result, err := s.recutilsOp.GetDatabaseInfo(ctx, args.DatabaseFile)
		if err != nil {
			return &mcp.CallToolResult{
				Content: []mcp.Content{
					&mcp.TextContent{Text: fmt.Sprintf("Error: %v", err)},
				},
			}, nil, nil
		}

		resultJSON, err := json.Marshal(result)
		if err != nil {
			return &mcp.CallToolResult{
				Content: []mcp.Content{
					&mcp.TextContent{Text: fmt.Sprintf("Error marshaling result: %v", err)},
				},
			}, nil, nil
		}

		return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{Text: string(resultJSON)},
			},
		}, nil, nil
	})

	return nil
}

// Run Run MCP server
func (s *MCPServer) Run(ctx context.Context) error {
	// Create server
	server := mcp.NewServer(&mcp.Implementation{
		Name:    "recutils-mcp",
		Version: "1.0.0",
	}, nil)

	// Add tools
	if err := s.SetupTools(server); err != nil {
		return fmt.Errorf("failed to setup tools: %w", err)
	}

	log.Println("Starting Recutils MCP Server...")

	// Create stdio transport and run server
	transport := &mcp.StdioTransport{}
	if err := server.Run(ctx, transport); err != nil {
		return fmt.Errorf("server run failed: %w", err)
	}

	return nil
}
