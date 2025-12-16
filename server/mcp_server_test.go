package server

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/recutils-mcp/recutils-mcp/recutils"
)

func TestMCPServer(t *testing.T) {
	ctx := context.Background()
	_ = NewMCPServer()

	// Create temporary test database file
	tmpFile, err := os.CreateTemp("", "test-*.rec")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())
	defer tmpFile.Close()

	// Write test data
	testData := `%rec: Person

Name: John Doe
Age: 25
City: New York
`
	_, err = tmpFile.WriteString(testData)
	if err != nil {
		t.Fatalf("Failed to write test data: %v", err)
	}
	tmpFile.Close()

	// Test recutils operations
	recOp := recutils.NewRecordOperation()

	// Test 1: Query records
	t.Run("QueryRecords", func(t *testing.T) {
		result, err := recOp.QueryRecords(ctx, tmpFile.Name(), "Name = 'John Doe'", "")
		if err != nil {
			t.Errorf("Query failed: %v", err)
			return
		}

		if result == nil {
			t.Error("Result is nil")
			return
		}

		if !result.Success {
			t.Errorf("Query failed: %s", result.Error)
			return
		}

		t.Logf("Query result: %+v", result)
	})

	// Test 2: Get database info
	t.Run("GetDatabaseInfo", func(t *testing.T) {
		result, err := recOp.GetDatabaseInfo(ctx, tmpFile.Name())
		if err != nil {
			t.Errorf("Get info failed: %v", err)
			return
		}

		if result == nil {
			t.Error("Result is nil")
			return
		}

		if !result.Success {
			t.Errorf("Get info failed: %s", result.Error)
			return
		}

		t.Logf("Database info: %+v", result)
	})

	// Test 3: Insert record
	t.Run("InsertRecord", func(t *testing.T) {
		result, err := recOp.InsertRecord(ctx, tmpFile.Name(), "Person", map[string]interface{}{
			"Name": "Bob Johnson",
			"Age":  28,
			"City": "Chicago",
		})
		if err != nil {
			t.Errorf("Insert failed: %v", err)
			return
		}

		if result == nil {
			t.Error("Result is nil")
			return
		}

		if !result.Success {
			t.Errorf("Insert failed: %s", result.Error)
			return
		}

		t.Logf("Insert result: %+v", result)
	})

	// Verify insert result
	t.Run("VerifyInsert", func(t *testing.T) {
		result, err := recOp.QueryRecords(ctx, tmpFile.Name(), "", "")
		if err != nil {
			t.Errorf("Query failed: %v", err)
			return
		}

		if result == nil {
			t.Error("Result is nil")
			return
		}

		if !result.Success {
			t.Errorf("Query failed: %s", result.Error)
			return
		}

		t.Logf("All records: %+v", result)
	})
}

func TestMCPServerIntegration(t *testing.T) {
	// Set timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	_ = NewMCPServer()

	// Create temporary test database file
	tmpFile, err := os.CreateTemp("", "integration-*.rec")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())
	defer tmpFile.Close()

	// Get recutils operation instance
	recOp := recutils.NewRecordOperation()

	// Full workflow test
	t.Run("FullWorkflow", func(t *testing.T) {
		// 1. Insert initial record
		insertResult, err := recOp.InsertRecord(ctx, tmpFile.Name(), "Person", map[string]interface{}{
			"Name": "Jane Smith",
			"Age":  30,
			"City": "Los Angeles",
		})
		if err != nil {
			t.Fatalf("Insert failed: %v", err)
		}
		if insertResult == nil || !insertResult.Success {
			t.Fatalf("Insert failed: %v", insertResult)
		}
		t.Logf("Insert result: %+v", insertResult)

		// 2. Query record
		queryResult, err := recOp.QueryRecords(ctx, tmpFile.Name(), "Name = 'Jane Smith'", "")
		if err != nil {
			t.Fatalf("Query failed: %v", err)
		}
		if queryResult == nil || !queryResult.Success {
			t.Fatalf("Query failed: %v", queryResult)
		}
		t.Logf("Query result: %+v", queryResult)

		// 3. Update record
		updateResult, err := recOp.UpdateRecords(ctx, tmpFile.Name(), "Name = 'Jane Smith'", map[string]interface{}{
			"Age": 31,
		})
		if err != nil {
			t.Fatalf("Update failed: %v", err)
		}
		if updateResult == nil || !updateResult.Success {
			t.Fatalf("Update failed: %v", updateResult)
		}
		t.Logf("Update result: %+v", updateResult)

		// 4. Verify update
		verifyResult, err := recOp.QueryRecords(ctx, tmpFile.Name(), "Name = 'Jane Smith'", "")
		if err != nil {
			t.Fatalf("Verify query failed: %v", err)
		}
		if verifyResult == nil || !verifyResult.Success {
			t.Fatalf("Verify query failed: %v", verifyResult)
		}
		t.Logf("Verify result: %+v", verifyResult)

		// 5. Get database info
		infoResult, err := recOp.GetDatabaseInfo(ctx, tmpFile.Name())
		if err != nil {
			t.Fatalf("Get info failed: %v", err)
		}
		if infoResult == nil || !infoResult.Success {
			t.Fatalf("Get info failed: %v", infoResult)
		}
		t.Logf("Database info: %+v", infoResult)
	})
}

func TestArgsStructures(t *testing.T) {
	t.Run("QueryArgs", func(t *testing.T) {
		args := QueryArgs{
			DatabaseFile:    "test.rec",
			QueryExpression: "Name = 'Test'",
			OutputFormat:    "plain",
		}
		if args.DatabaseFile == "" {
			t.Error("DatabaseFile should not be empty")
		}
		if args.QueryExpression == "" {
			t.Log("QueryExpression is optional, can be empty")
		}
		t.Logf("QueryArgs: %+v", args)
	})

	t.Run("InsertArgs", func(t *testing.T) {
		args := InsertArgs{
			DatabaseFile: "test.rec",
			RecordType:   "Person",
			Fields: map[string]interface{}{
				"Name": "Test",
			},
		}
		if args.DatabaseFile == "" {
			t.Error("DatabaseFile should not be empty")
		}
		if args.RecordType == "" {
			t.Error("RecordType should not be empty")
		}
		if args.Fields == nil {
			t.Error("Fields should not be nil")
		}
		t.Logf("InsertArgs: %+v", args)
	})

	t.Run("UpdateArgs", func(t *testing.T) {
		args := UpdateArgs{
			DatabaseFile:    "test.rec",
			QueryExpression: "Name = 'Test'",
			Fields: map[string]interface{}{
				"Age": 30,
			},
		}
		if args.DatabaseFile == "" {
			t.Error("DatabaseFile should not be empty")
		}
		if args.QueryExpression == "" {
			t.Error("QueryExpression should not be empty")
		}
		if args.Fields == nil {
			t.Error("Fields should not be nil")
		}
		t.Logf("UpdateArgs: %+v", args)
	})

	t.Run("DeleteArgs", func(t *testing.T) {
		args := DeleteArgs{
			DatabaseFile:    "test.rec",
			QueryExpression: "Name = 'Test'",
		}
		if args.DatabaseFile == "" {
			t.Error("DatabaseFile should not be empty")
		}
		if args.QueryExpression == "" {
			t.Error("QueryExpression should not be empty")
		}
		t.Logf("DeleteArgs: %+v", args)
	})

	t.Run("InfoArgs", func(t *testing.T) {
		args := InfoArgs{
			DatabaseFile: "test.rec",
		}
		if args.DatabaseFile == "" {
			t.Error("DatabaseFile should not be empty")
		}
		t.Logf("InfoArgs: %+v", args)
	})
}
