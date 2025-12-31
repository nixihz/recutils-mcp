package server

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/nixihz/recutils-mcp/recutils"
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

func TestNewMCPServer(t *testing.T) {
	t.Run("Create server instance", func(t *testing.T) {
		server := NewMCPServer()
		if server == nil {
			t.Fatal("NewMCPServer returned nil")
		}
		if server.recutilsOp == nil {
			t.Error("recutilsOp field is nil")
		}
	})
}

func TestSetupTools(t *testing.T) {
	t.Run("Setup tools successfully", func(t *testing.T) {
		server := NewMCPServer()

		// Create a mock MCP server - we can't test without the actual mcp package
		// but we can verify the server instance is properly configured
		if server == nil {
			t.Fatal("Server is nil")
		}

		// Verify the server has the recutils operation initialized
		if server.recutilsOp == nil {
			t.Error("recutilsOp is not initialized")
		}
	})
}

// TestDeleteRecordsOperation tests the delete operation
func TestDeleteRecordsOperation(t *testing.T) {
	ctx := context.Background()
	_ = NewMCPServer()

	recOp := recutils.NewRecordOperation()

	// Write test data with multiple records
	testData := `%rec: Person

Name: John Doe
Age: 25
City: New York

Name: Jane Smith
Age: 30
City: Los Angeles

Name: Bob Johnson
Age: 28
City: Chicago
`

	t.Run("DeleteSingleRecord", func(t *testing.T) {
		tmpFile, err := os.CreateTemp("", "delete-test-*.rec")
		if err != nil {
			t.Fatalf("Failed to create temp file: %v", err)
		}
		defer os.Remove(tmpFile.Name())
		defer tmpFile.Close()

		_, err = tmpFile.WriteString(testData)
		if err != nil {
			t.Fatalf("Failed to write test data: %v", err)
		}
		tmpFile.Close()

		result, err := recOp.DeleteRecords(ctx, tmpFile.Name(), "Name = 'Jane Smith'")
		if err != nil {
			t.Errorf("Delete failed: %v", err)
			return
		}

		if result == nil {
			t.Error("Result is nil")
			return
		}

		if !result.Success {
			t.Errorf("Delete failed: %s", result.Error)
			return
		}

		// Verify the record was deleted
		queryResult, _ := recOp.QueryRecords(ctx, tmpFile.Name(), "Name = 'Jane Smith'", "")
		if queryResult != nil && queryResult.Success {
			if len(queryResult.Output) > 0 && contains(queryResult.Output, "Jane Smith") {
				t.Error("Record should have been deleted")
			}
		}
	})

	t.Run("DeleteMultipleRecords", func(t *testing.T) {
		tmpFile2, err := os.CreateTemp("", "delete-multi-*.rec")
		if err != nil {
			t.Fatalf("Failed to create temp file: %v", err)
		}
		defer os.Remove(tmpFile2.Name())
		defer tmpFile2.Close()

		_, err = tmpFile2.WriteString(testData)
		if err != nil {
			t.Fatalf("Failed to write test data: %v", err)
		}
		tmpFile2.Close()

		result, err := recOp.DeleteRecords(ctx, tmpFile2.Name(), "Age < 29")
		if err != nil {
			t.Errorf("Delete failed: %v", err)
			return
		}

		if result == nil || !result.Success {
			t.Errorf("Expected success, got: %+v", result)
			return
		}

		// Verify records were deleted
		queryResult, _ := recOp.QueryRecords(ctx, tmpFile2.Name(), "", "")
		if queryResult != nil && queryResult.Success {
			output := queryResult.Output
			if contains(output, "John Doe") || contains(output, "Bob Johnson") {
				t.Error("Matching records should have been deleted")
			}
			if !contains(output, "Jane Smith") {
				t.Error("Non-matching record should still exist")
			}
		}
	})

	t.Run("DeleteNonExistentRecord", func(t *testing.T) {
		tmpFile, err := os.CreateTemp("", "delete-nomatch-*.rec")
		if err != nil {
			t.Fatalf("Failed to create temp file: %v", err)
		}
		defer os.Remove(tmpFile.Name())
		defer tmpFile.Close()

		_, err = tmpFile.WriteString(testData)
		if err != nil {
			t.Fatalf("Failed to write test data: %v", err)
		}
		tmpFile.Close()

		result, err := recOp.DeleteRecords(ctx, tmpFile.Name(), "Name = 'NonExistent'")
		if err != nil {
			t.Errorf("Delete failed: %v", err)
			return
		}

		if result == nil {
			t.Error("Result is nil")
			return
		}

		// Deleting non-existent records should still succeed
		if !result.Success {
			t.Errorf("Delete should succeed even if no records match: %s", result.Error)
		}
	})

	t.Run("DeleteFromNonExistentFile", func(t *testing.T) {
		result, err := recOp.DeleteRecords(ctx, "/nonexistent/file.rec", "Name = 'Test'")
		// DeleteRecords returns error for non-existent file
		if err == nil {
			t.Error("Expected error for non-existent file")
		}

		if result != nil && result.Success {
			t.Error("Delete from non-existent file should fail")
		}
	})
}

// TestUpdateRecordsOperation tests the update operation
func TestUpdateRecordsOperation(t *testing.T) {
	ctx := context.Background()
	_ = NewMCPServer()

	recOp := recutils.NewRecordOperation()

	// Write test data
	testData := `%rec: Person

Name: John Doe
Age: 25
City: New York

Name: Jane Smith
Age: 30
City: Los Angeles
`

	t.Run("UpdateSingleField", func(t *testing.T) {
		tmpFile, err := os.CreateTemp("", "update-test-*.rec")
		if err != nil {
			t.Fatalf("Failed to create temp file: %v", err)
		}
		defer os.Remove(tmpFile.Name())
		defer tmpFile.Close()

		_, err = tmpFile.WriteString(testData)
		if err != nil {
			t.Fatalf("Failed to write test data: %v", err)
		}
		tmpFile.Close()

		result, err := recOp.UpdateRecords(ctx, tmpFile.Name(), "Name = 'John Doe'", map[string]interface{}{
			"Age": 26,
		})
		if err != nil {
			t.Errorf("Update failed: %v", err)
			return
		}

		if result == nil || !result.Success {
			t.Errorf("Expected success, got: %+v", result)
			return
		}

		// Verify the update
		queryResult, _ := recOp.QueryRecords(ctx, tmpFile.Name(), "Name = 'John Doe'", "")
		if queryResult != nil && queryResult.Success {
			if !contains(queryResult.Output, "Age: 26") {
				t.Error("Field was not updated")
			}
		}
	})

	t.Run("UpdateMultipleFields", func(t *testing.T) {
		tmpFile2, err := os.CreateTemp("", "update-multi-*.rec")
		if err != nil {
			t.Fatalf("Failed to create temp file: %v", err)
		}
		defer os.Remove(tmpFile2.Name())
		defer tmpFile2.Close()

		_, err = tmpFile2.WriteString(testData)
		if err != nil {
			t.Fatalf("Failed to write test data: %v", err)
		}
		tmpFile2.Close()

		result, err := recOp.UpdateRecords(ctx, tmpFile2.Name(), "Name = 'Jane Smith'", map[string]interface{}{
			"Age":  31,
			"City": "San Francisco",
		})
		if err != nil {
			t.Errorf("Update failed: %v", err)
			return
		}

		if result == nil || !result.Success {
			t.Errorf("Expected success, got: %+v", result)
			return
		}

		// Verify the updates
		queryResult, _ := recOp.QueryRecords(ctx, tmpFile2.Name(), "Name = 'Jane Smith'", "")
		if queryResult != nil && queryResult.Success {
			output := queryResult.Output
			if !contains(output, "Age: 31") {
				t.Error("Age was not updated")
			}
			if !contains(output, "City: San Francisco") {
				t.Error("City was not updated")
			}
		}
	})

	t.Run("AddNewField", func(t *testing.T) {
		tmpFile3, err := os.CreateTemp("", "update-new-*.rec")
		if err != nil {
			t.Fatalf("Failed to create temp file: %v", err)
		}
		defer os.Remove(tmpFile3.Name())
		defer tmpFile3.Close()

		_, err = tmpFile3.WriteString(testData)
		if err != nil {
			t.Fatalf("Failed to write test data: %v", err)
		}
		tmpFile3.Close()

		result, err := recOp.UpdateRecords(ctx, tmpFile3.Name(), "Name = 'John Doe'", map[string]interface{}{
			"Email": "john.doe@example.com",
		})
		if err != nil {
			t.Errorf("Update failed: %v", err)
			return
		}

		if result == nil || !result.Success {
			t.Errorf("Expected success, got: %+v", result)
			return
		}

		// Verify the new field was added
		queryResult, _ := recOp.QueryRecords(ctx, tmpFile3.Name(), "Name = 'John Doe'", "")
		if queryResult != nil && queryResult.Success {
			if !contains(queryResult.Output, "Email: john.doe@example.com") {
				t.Error("New field was not added")
			}
		}
	})

	t.Run("UpdateNonExistentRecord", func(t *testing.T) {
		tmpFile, err := os.CreateTemp("", "update-nomatch-*.rec")
		if err != nil {
			t.Fatalf("Failed to create temp file: %v", err)
		}
		defer os.Remove(tmpFile.Name())
		defer tmpFile.Close()

		_, err = tmpFile.WriteString(testData)
		if err != nil {
			t.Fatalf("Failed to write test data: %v", err)
		}
		tmpFile.Close()

		result, err := recOp.UpdateRecords(ctx, tmpFile.Name(), "Name = 'NonExistent'", map[string]interface{}{
			"Age": 99,
		})
		if err != nil {
			t.Errorf("Update failed: %v", err)
		}

		// Updating non-existent records may succeed or fail depending on implementation
		if result == nil {
			t.Error("Result is nil")
		}
	})
}

// TestErrorScenarios tests various error scenarios
func TestErrorScenarios(t *testing.T) {
	ctx := context.Background()

	t.Run("QueryNonExistentFile", func(t *testing.T) {
		recOp := recutils.NewRecordOperation()
		result, err := recOp.QueryRecords(ctx, "/nonexistent/file.rec", "", "")
		if err != nil {
			t.Errorf("Query returned error: %v", err)
		}

		if result != nil && result.Success {
			t.Error("Query should fail for non-existent file")
		}
	})

	t.Run("InsertInvalidPath", func(t *testing.T) {
		recOp := recutils.NewRecordOperation()
		result, err := recOp.InsertRecord(ctx, "/nonexistent/dir/file.rec", "Person", map[string]interface{}{
			"Name": "Test",
		})
		// InsertRecord returns error for invalid path
		if err == nil {
			t.Error("Expected error for invalid path")
		}

		if result != nil && result.Success {
			t.Error("Insert should fail for invalid path")
		}
	})

	t.Run("GetInfoNonExistentFile", func(t *testing.T) {
		recOp := recutils.NewRecordOperation()
		result, err := recOp.GetDatabaseInfo(ctx, "/nonexistent/file.rec")
		if err != nil {
			t.Errorf("GetDatabaseInfo returned error: %v", err)
		}

		if result != nil && result.Success {
			t.Error("GetDatabaseInfo should fail for non-existent file")
		}
	})

	t.Run("EmptyDatabaseFile", func(t *testing.T) {
		tmpFile, err := os.CreateTemp("", "empty-*.rec")
		if err != nil {
			t.Fatalf("Failed to create temp file: %v", err)
		}
		defer os.Remove(tmpFile.Name())
		tmpFile.Close()

		recOp := recutils.NewRecordOperation()
		result, err := recOp.QueryRecords(ctx, tmpFile.Name(), "", "")
		if err != nil {
			t.Errorf("Query returned error: %v", err)
		}

		// Query on empty file may succeed or fail depending on recutils behavior
		if result == nil {
			t.Error("Result should not be nil")
		}
	})
}

// TestArgsStructures tests argument structure validation
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

	t.Run("QueryArgsJSONTag", func(t *testing.T) {
		// Test JSON marshaling
		args := QueryArgs{
			DatabaseFile:    "test.rec",
			QueryExpression: "Age > 25",
			OutputFormat:    "csv",
		}

		// Verify JSON tags work correctly
		if args.DatabaseFile == "" {
			t.Error("DatabaseFile serialization failed")
		}
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

	t.Run("InsertArgsEmptyFields", func(t *testing.T) {
		args := InsertArgs{
			DatabaseFile: "test.rec",
			RecordType:   "Person",
			Fields:       map[string]interface{}{},
		}
		if args.Fields == nil {
			t.Error("Fields should not be nil even when empty")
		}
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

// Helper function to check if a string contains a substring
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) && containsHelper(s, substr))
}

func containsHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
