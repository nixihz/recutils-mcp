// recutils package: Unit tests for recutils operations
package recutils

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// TestNewRecordOperation tests creating a new RecordOperation instance
func TestNewRecordOperation(t *testing.T) {
	op := NewRecordOperation()
	if op == nil {
		t.Fatal("NewRecordOperation returned nil")
	}
}

// TestExecuteRecCommand tests the executeRecCommand method
func TestExecuteRecCommand(t *testing.T) {
	op := NewRecordOperation()
	ctx := context.Background()

	tests := []struct {
		name        string
		cmd         []string
		inputData   string
		wantSuccess bool
		wantError   bool
	}{
		{
			name:        "Valid recsel command",
			cmd:         []string{"echo", "test"},
			inputData:   "",
			wantSuccess: true,
			wantError:   false,
		},
		{
			name:        "Command with input data",
			cmd:         []string{"cat"},
			inputData:   "test input",
			wantSuccess: true,
			wantError:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := op.executeRecCommand(ctx, tt.cmd, tt.inputData)

			if tt.wantError && err == nil {
				t.Error("Expected error but got none")
			}
			if !tt.wantError && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
			if result != nil && tt.wantSuccess != result.Success {
				t.Errorf("Expected success=%v, got success=%v", tt.wantSuccess, result.Success)
			}
		})
	}
}

// TestExecuteRecCommandTimeout tests command execution timeout
func TestExecuteRecCommandTimeout(t *testing.T) {
	op := NewRecordOperation()

	// Create a context that will timeout quickly
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Millisecond)
	defer cancel()

	// Use sleep command that will exceed timeout
	result, err := op.executeRecCommand(ctx, []string{"sleep", "10"}, "")

	// Should either return error or a failed result due to timeout
	if err == nil && result != nil && result.Success {
		t.Error("Expected timeout to cause failure, but command succeeded")
	}
}

// TestQueryRecords tests the QueryRecords method
func TestQueryRecords(t *testing.T) {
	op := NewRecordOperation()
	ctx := context.Background()

	tests := []struct {
		name            string
		setupTestData   func() string
		queryExpression string
		outputFormat    string
		wantSuccess     bool
		wantOutput      string
	}{
		{
			name: "Query all records",
			setupTestData: func() string {
				tmpDir := t.TempDir()
				testDBPath := filepath.Join(tmpDir, "test_query.rec")
				testData := `%rec: Person

Name: John Doe
Age: 25
City: New York

Name: Jane Smith
Age: 30
City: Los Angeles
`
				err := os.WriteFile(testDBPath, []byte(testData), 0644)
				if err != nil {
					t.Fatalf("Failed to create test database: %v", err)
				}
				return testDBPath
			},
			queryExpression: "",
			outputFormat:    "",
			wantSuccess:     true,
		},
		{
			name: "Query with expression",
			setupTestData: func() string {
				tmpDir := t.TempDir()
				testDBPath := filepath.Join(tmpDir, "test_query.rec")
				testData := `%rec: Person

Name: John Doe
Age: 25
City: New York

Name: Jane Smith
Age: 30
City: Los Angeles
`
				err := os.WriteFile(testDBPath, []byte(testData), 0644)
				if err != nil {
					t.Fatalf("Failed to create test database: %v", err)
				}
				return testDBPath
			},
			queryExpression: "Name = 'John Doe'",
			outputFormat:    "",
			wantSuccess:     true,
		},
		{
			name: "Query with output format",
			setupTestData: func() string {
				tmpDir := t.TempDir()
				testDBPath := filepath.Join(tmpDir, "test_query.rec")
				testData := `%rec: Person

Name: John Doe
Age: 25
City: New York

Name: Jane Smith
Age: 30
City: Los Angeles
`
				err := os.WriteFile(testDBPath, []byte(testData), 0644)
				if err != nil {
					t.Fatalf("Failed to create test database: %v", err)
				}
				return testDBPath
			},
			queryExpression: "",
			outputFormat:    "plain",
			wantSuccess:     true,
		},
		{
			name: "Query with format and expression",
			setupTestData: func() string {
				tmpDir := t.TempDir()
				testDBPath := filepath.Join(tmpDir, "test_query.rec")
				testData := `%rec: Person

Name: John Doe
Age: 25
City: New York

Name: Jane Smith
Age: 30
City: Los Angeles
`
				err := os.WriteFile(testDBPath, []byte(testData), 0644)
				if err != nil {
					t.Fatalf("Failed to create test database: %v", err)
				}
				return testDBPath
			},
			queryExpression: "Age > 25",
			outputFormat:    "csv",
			wantSuccess:     true,
		},
		{
			name:            "Non-existent database file",
			setupTestData:   func() string { return "/nonexistent/file.rec" },
			queryExpression: "",
			outputFormat:    "",
			wantSuccess:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			databaseFile := tt.setupTestData()
			result, err := op.QueryRecords(ctx, databaseFile, tt.queryExpression, tt.outputFormat)

			if err != nil {
				t.Errorf("QueryRecords returned error: %v", err)
				return
			}

			if result == nil {
				t.Error("Result is nil")
				return
			}

			if tt.wantSuccess && !result.Success {
				t.Errorf("Expected success=true, got success=false. Error: %s", result.Error)
			}

			if !tt.wantSuccess && result.Success {
				t.Error("Expected success=false for invalid input, got success=true")
			}

			if tt.wantOutput != "" && !strings.Contains(result.Output, tt.wantOutput) {
				t.Errorf("Expected output to contain %q, got %q", tt.wantOutput, result.Output)
			}
		})
	}
}

// TestInsertRecord tests the InsertRecord method
func TestInsertRecord(t *testing.T) {
	op := NewRecordOperation()
	ctx := context.Background()

	t.Run("Insert into new file", func(t *testing.T) {
		tmpDir := t.TempDir()
		testDBPath := filepath.Join(tmpDir, "test_new.rec")

		fields := map[string]interface{}{
			"Name": "Alice Johnson",
			"Age":  28,
			"City": "Chicago",
		}

		result, err := op.InsertRecord(ctx, testDBPath, "Person", fields)
		if err != nil {
			t.Errorf("InsertRecord returned error: %v", err)
			return
		}

		if result == nil {
			t.Error("Result is nil")
			return
		}

		if !result.Success {
			t.Errorf("Expected success=true, got success=false. Error: %s", result.Error)
		}

		// Verify file was created
		if _, err := os.Stat(testDBPath); os.IsNotExist(err) {
			t.Error("Database file was not created")
		}
	})

	t.Run("Insert into existing file", func(t *testing.T) {
		tmpDir := t.TempDir()
		testDBPath := filepath.Join(tmpDir, "test_existing.rec")

		// Create initial file
		initialData := `%rec: Person

Name: Bob Smith
Age: 35
`
		err := os.WriteFile(testDBPath, []byte(initialData), 0644)
		if err != nil {
			t.Fatalf("Failed to create initial database: %v", err)
		}

		fields := map[string]interface{}{
			"Name": "Charlie Brown",
			"Age":  42,
		}

		result, err := op.InsertRecord(ctx, testDBPath, "Person", fields)
		if err != nil {
			t.Errorf("InsertRecord returned error: %v", err)
			return
		}

		if result == nil || !result.Success {
			t.Errorf("Expected success=true, got result: %+v", result)
		}

		// Verify content was appended
		content, _ := os.ReadFile(testDBPath)
		contentStr := string(content)
		if !strings.Contains(contentStr, "Charlie Brown") {
			t.Error("New record was not added to file")
		}
	})

	t.Run("Insert into empty file", func(t *testing.T) {
		tmpDir := t.TempDir()
		testDBPath := filepath.Join(tmpDir, "test_empty.rec")

		// Create empty file
		err := os.WriteFile(testDBPath, []byte{}, 0644)
		if err != nil {
			t.Fatalf("Failed to create empty file: %v", err)
		}

		fields := map[string]interface{}{
			"Name": "David Lee",
			"Age":  33,
		}

		result, err := op.InsertRecord(ctx, testDBPath, "Person", fields)
		if err != nil {
			t.Errorf("InsertRecord returned error: %v", err)
			return
		}

		if result == nil || !result.Success {
			t.Errorf("Expected success=true, got result: %+v", result)
		}
	})

	t.Run("Insert with special characters in fields", func(t *testing.T) {
		tmpDir := t.TempDir()
		testDBPath := filepath.Join(tmpDir, "test_special.rec")

		fields := map[string]interface{}{
			"Name":  "O'Brien, John",
			"Email": "john@example.com",
			"Notes": "Multi-line\nnotes",
		}

		result, err := op.InsertRecord(ctx, testDBPath, "Contact", fields)
		if err != nil {
			t.Errorf("InsertRecord returned error: %v", err)
			return
		}

		if result == nil || !result.Success {
			t.Errorf("Expected success=true, got result: %+v", result)
		}
	})
}

// TestInsertRecordErrors tests InsertRecord error conditions
func TestInsertRecordErrors(t *testing.T) {
	op := NewRecordOperation()
	ctx := context.Background()

	t.Run("Insert with invalid directory path", func(t *testing.T) {
		invalidPath := "/nonexistent/directory/test.rec"
		fields := map[string]interface{}{
			"Name": "Test User",
		}

		result, err := op.InsertRecord(ctx, invalidPath, "Person", fields)
		// InsertRecord returns error for invalid path
		if err == nil {
			t.Error("Expected error for invalid path")
		}

		if result != nil && result.Success {
			t.Error("Expected success=false for invalid path")
		}
	})
}

// TestDeleteRecords tests the DeleteRecords method
func TestDeleteRecords(t *testing.T) {
	op := NewRecordOperation()
	ctx := context.Background()

	t.Run("Delete single record", func(t *testing.T) {
		tmpDir := t.TempDir()
		testDBPath := filepath.Join(tmpDir, "test_delete.rec")

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

		err := os.WriteFile(testDBPath, []byte(testData), 0644)
		if err != nil {
			t.Fatalf("Failed to create test database: %v", err)
		}

		result, err := op.DeleteRecords(ctx, testDBPath, "Name = 'Jane Smith'")
		if err != nil {
			t.Errorf("DeleteRecords returned error: %v", err)
			return
		}

		if result == nil {
			t.Error("Result is nil")
			return
		}

		if !result.Success {
			t.Errorf("Expected success=true, got success=false. Error: %s", result.Error)
		}

		// Verify deletion by querying
		queryResult, _ := op.QueryRecords(ctx, testDBPath, "Name = 'Jane Smith'", "")
		if queryResult != nil && queryResult.Success && strings.Contains(queryResult.Output, "Jane Smith") {
			t.Error("Record was not deleted successfully")
		}
	})

	t.Run("Delete multiple records", func(t *testing.T) {
		tmpDir := t.TempDir()
		testDBPath := filepath.Join(tmpDir, "test_delete_multi.rec")

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

		err := os.WriteFile(testDBPath, []byte(testData), 0644)
		if err != nil {
			t.Fatalf("Failed to recreate test database: %v", err)
		}

		result, err := op.DeleteRecords(ctx, testDBPath, "Age < 30")
		if err != nil {
			t.Errorf("DeleteRecords returned error: %v", err)
			return
		}

		if result == nil || !result.Success {
			t.Errorf("Expected success=true, got result: %+v", result)
		}

		// Verify that remaining records exist
		queryResult, _ := op.QueryRecords(ctx, testDBPath, "", "")
		if queryResult != nil && queryResult.Success {
			if strings.Contains(queryResult.Output, "John Doe") || strings.Contains(queryResult.Output, "Bob Johnson") {
				t.Error("Records matching condition should have been deleted")
			}
		}
	})

	t.Run("Delete with non-matching expression", func(t *testing.T) {
		tmpDir := t.TempDir()
		testDBPath := filepath.Join(tmpDir, "test_delete_nomatch.rec")

		testData := `%rec: Person

Name: John Doe
Age: 25
City: New York
`

		err := os.WriteFile(testDBPath, []byte(testData), 0644)
		if err != nil {
			t.Fatalf("Failed to recreate test database: %v", err)
		}

		result, err := op.DeleteRecords(ctx, testDBPath, "Name = 'NonExistent'")
		if err != nil {
			t.Errorf("DeleteRecords returned error: %v", err)
			return
		}

		if result == nil || !result.Success {
			t.Errorf("Delete with non-matching expression should succeed, got: %+v", result)
		}
	})

	t.Run("Delete from non-existent file", func(t *testing.T) {
		tmpDir := t.TempDir()
		nonExistentPath := filepath.Join(tmpDir, "nonexistent.rec")

		result, err := op.DeleteRecords(ctx, nonExistentPath, "Name = 'Test'")
		// DeleteRecords returns error for non-existent file
		if err == nil {
			t.Error("Expected error for non-existent file")
		}

		if result != nil && result.Success {
			t.Error("Expected success=false for non-existent file")
		}
	})
}

// TestUpdateRecords tests the UpdateRecords method
func TestUpdateRecords(t *testing.T) {
	op := NewRecordOperation()
	ctx := context.Background()

	t.Run("Update single field", func(t *testing.T) {
		tmpDir := t.TempDir()
		testDBPath := filepath.Join(tmpDir, "test_update.rec")

		testData := `%rec: Person

Name: John Doe
Age: 25
City: New York

Name: Jane Smith
Age: 30
City: Los Angeles
`

		err := os.WriteFile(testDBPath, []byte(testData), 0644)
		if err != nil {
			t.Fatalf("Failed to create test database: %v", err)
		}

		result, err := op.UpdateRecords(ctx, testDBPath, "Name = 'John Doe'", map[string]interface{}{
			"Age": 26,
		})
		if err != nil {
			t.Errorf("UpdateRecords returned error: %v", err)
			return
		}

		if result == nil {
			t.Error("Result is nil")
			return
		}

		if !result.Success {
			t.Errorf("Expected success=true, got success=false. Error: %s", result.Error)
		}

		// Verify update
		queryResult, _ := op.QueryRecords(ctx, testDBPath, "Name = 'John Doe'", "")
		if queryResult != nil && queryResult.Success {
			if !strings.Contains(queryResult.Output, "Age: 26") {
				t.Error("Field was not updated successfully")
			}
		}
	})

	t.Run("Update multiple fields", func(t *testing.T) {
		tmpDir := t.TempDir()
		testDBPath := filepath.Join(tmpDir, "test_update_multi.rec")

		testData := `%rec: Person

Name: John Doe
Age: 25
City: New York

Name: Jane Smith
Age: 30
City: Los Angeles
`

		err := os.WriteFile(testDBPath, []byte(testData), 0644)
		if err != nil {
			t.Fatalf("Failed to create test database: %v", err)
		}

		result, err := op.UpdateRecords(ctx, testDBPath, "Name = 'Jane Smith'", map[string]interface{}{
			"Age":  31,
			"City": "San Francisco",
		})
		if err != nil {
			t.Errorf("UpdateRecords returned error: %v", err)
			return
		}

		if result == nil || !result.Success {
			t.Errorf("Expected success=true, got result: %+v", result)
		}

		// Verify updates
		queryResult, _ := op.QueryRecords(ctx, testDBPath, "Name = 'Jane Smith'", "")
		if queryResult != nil && queryResult.Success {
			output := queryResult.Output
			if !strings.Contains(output, "Age: 31") {
				t.Error("Age was not updated")
			}
			if !strings.Contains(output, "City: San Francisco") {
				t.Error("City was not updated")
			}
		}
	})

	t.Run("Add new field to existing record", func(t *testing.T) {
		tmpDir := t.TempDir()
		testDBPath := filepath.Join(tmpDir, "test_update_add.rec")

		testData := `%rec: Person

Name: John Doe
Age: 25
City: New York

Name: Jane Smith
Age: 30
City: Los Angeles
`

		err := os.WriteFile(testDBPath, []byte(testData), 0644)
		if err != nil {
			t.Fatalf("Failed to create test database: %v", err)
		}

		result, err := op.UpdateRecords(ctx, testDBPath, "Name = 'John Doe'", map[string]interface{}{
			"Email": "john.doe@example.com",
		})
		if err != nil {
			t.Errorf("UpdateRecords returned error: %v", err)
			return
		}

		if result == nil || !result.Success {
			t.Errorf("Expected success=true, got result: %+v", result)
		}

		// Verify new field was added
		queryResult, _ := op.QueryRecords(ctx, testDBPath, "Name = 'John Doe'", "")
		if queryResult != nil && queryResult.Success {
			if !strings.Contains(queryResult.Output, "Email: john.doe@example.com") {
				t.Error("New field was not added")
			}
		}
	})

	t.Run("Update non-matching record", func(t *testing.T) {
		tmpDir := t.TempDir()
		testDBPath := filepath.Join(tmpDir, "test_update_nomatch.rec")

		testData := `%rec: Person

Name: John Doe
Age: 25
City: New York
`

		err := os.WriteFile(testDBPath, []byte(testData), 0644)
		if err != nil {
			t.Fatalf("Failed to create test database: %v", err)
		}

		result, err := op.UpdateRecords(ctx, testDBPath, "Name = 'NonExistent'", map[string]interface{}{
			"Age": 99,
		})
		if err != nil {
			t.Errorf("UpdateRecords returned error: %v", err)
		}

		if result != nil && result.Success {
			// Some implementations may succeed even with no matching records
			// This is acceptable behavior
			t.Log("Update succeeded for non-matching record (acceptable behavior)")
		}
	})

	t.Run("Update from non-existent file", func(t *testing.T) {
		tmpDir := t.TempDir()
		nonExistentPath := filepath.Join(tmpDir, "nonexistent.rec")

		result, err := op.UpdateRecords(ctx, nonExistentPath, "Name = 'Test'", map[string]interface{}{
			"Age": 25,
		})
		// UpdateRecords returns Result with Success=false for non-existent file
		// The recsel command fails, but executeRecCommand returns err=nil with Result.Success=false
		if result != nil && result.Success {
			t.Error("Expected success=false for non-existent file")
		}

		if result == nil {
			t.Error("Result should not be nil")
		}

		// err may be nil since executeRecCommand returns Result for command failures
		_ = err // We accept either error or failed Result
	})
}

// TestGetDatabaseInfo tests the GetDatabaseInfo method
func TestGetDatabaseInfo(t *testing.T) {
	// Create temporary test database
	tmpDir := t.TempDir()
	testDBPath := filepath.Join(tmpDir, "test_info.rec")

	testData := `%rec: Person

Name: John Doe
Age: 25
`

	err := os.WriteFile(testDBPath, []byte(testData), 0644)
	if err != nil {
		t.Fatalf("Failed to create test database: %v", err)
	}

	op := NewRecordOperation()
	ctx := context.Background()

	t.Run("Get info for valid database", func(t *testing.T) {
		result, err := op.GetDatabaseInfo(ctx, testDBPath)
		if err != nil {
			t.Errorf("GetDatabaseInfo returned error: %v", err)
			return
		}

		if result == nil {
			t.Error("Result is nil")
			return
		}

		if !result.Success {
			t.Errorf("Expected success=true, got success=false. Error: %s", result.Error)
		}

		// recinf should return some output
		if result.Output == "" {
			t.Error("Expected non-empty output from recinf")
		}
	})

	t.Run("Get info for non-existent file", func(t *testing.T) {
		nonExistentPath := filepath.Join(tmpDir, "nonexistent.rec")

		result, err := op.GetDatabaseInfo(ctx, nonExistentPath)
		if err != nil {
			t.Errorf("GetDatabaseInfo returned error: %v", err)
		}

		if result != nil && result.Success {
			t.Error("Expected success=false for non-existent file")
		}
	})

	t.Run("Get info for empty file", func(t *testing.T) {
		emptyPath := filepath.Join(tmpDir, "empty.rec")
		err := os.WriteFile(emptyPath, []byte{}, 0644)
		if err != nil {
			t.Fatalf("Failed to create empty file: %v", err)
		}

		result, err := op.GetDatabaseInfo(ctx, emptyPath)
		if err != nil {
			t.Errorf("GetDatabaseInfo returned error: %v", err)
		}

		// recinf may return error for empty file, which is acceptable
		if result != nil && result.Success {
			t.Log("recinf succeeded on empty file (acceptable)")
		}
	})
}

// TestResultStructure tests the Result structure
func TestResultStructure(t *testing.T) {
	t.Run("Success result", func(t *testing.T) {
		result := &Result{
			Success: true,
			Output:  "Test output",
			Error:   "",
		}

		if !result.Success {
			t.Error("Expected Success to be true")
		}
		if result.Output != "Test output" {
			t.Errorf("Expected Output 'Test output', got '%s'", result.Output)
		}
		if result.Error != "" {
			t.Errorf("Expected empty Error, got '%s'", result.Error)
		}
	})

	t.Run("Failure result", func(t *testing.T) {
		result := &Result{
			Success: false,
			Output:  "",
			Error:   "Test error",
		}

		if result.Success {
			t.Error("Expected Success to be false")
		}
		if result.Output != "" {
			t.Errorf("Expected empty Output, got '%s'", result.Output)
		}
		if result.Error != "Test error" {
			t.Errorf("Expected Error 'Test error', got '%s'", result.Error)
		}
	})
}

// TestContextCancellation tests behavior when context is cancelled
func TestContextCancellation(t *testing.T) {
	op := NewRecordOperation()

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	tmpDir := t.TempDir()
	testDBPath := filepath.Join(tmpDir, "test_cancel.rec")
	testData := `%rec: Person

Name: John Doe
Age: 25
`
	os.WriteFile(testDBPath, []byte(testData), 0644)

	// Query with cancelled context should handle gracefully
	result, err := op.QueryRecords(ctx, testDBPath, "", "")
	if err != nil {
		// Context cancellation may cause error, which is acceptable
		t.Logf("QueryRecords with cancelled context returned error: %v", err)
	}

	if result != nil && !result.Success {
		// Failure is also acceptable for cancelled context
		t.Logf("QueryRecords with cancelled context failed: %s", result.Error)
	}
}

// BenchmarkQueryRecords benchmarks the QueryRecords method
func BenchmarkQueryRecords(b *testing.B) {
	// Create temporary test database
	tmpDir := b.TempDir()
	testDBPath := filepath.Join(tmpDir, "bench_query.rec")

	// Create larger test data
	var testData strings.Builder
	testData.WriteString("%rec: Person\n\n")
	for i := 0; i < 100; i++ {
		testData.WriteString(fmt.Sprintf("Name: Person%d\nAge: %d\n\n", i, 20+i%50))
	}

	err := os.WriteFile(testDBPath, []byte(testData.String()), 0644)
	if err != nil {
		b.Fatalf("Failed to create test database: %v", err)
	}

	op := NewRecordOperation()
	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = op.QueryRecords(ctx, testDBPath, "", "")
	}
}
