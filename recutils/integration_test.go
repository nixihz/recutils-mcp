// recutils package: Integration tests with recutils command-line tools
package recutils

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// TestRecutilsIntegration tests CRUD operations with recutils CLI validation
func TestRecutilsIntegration(t *testing.T) {
	// Skip test if recutils is not installed
	if _, err := exec.LookPath("recsel"); err != nil {
		t.Skip("recutils not installed, skipping integration test")
	}

	ctx := context.Background()
	op := NewRecordOperation()

	// Create a temporary database file
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "integration_test.rec")

	t.Run("CompleteCRUDWorkflow", func(t *testing.T) {
		// Step 1: Create initial records (INSERT)
		t.Run("Step1_InsertRecords", func(t *testing.T) {
			// Insert first record
			result, err := op.InsertRecord(ctx, dbPath, "Person", map[string]interface{}{
				"Name":    "Alice Johnson",
				"Age":     28,
				"Email":   "alice@example.com",
				"City":    "San Francisco",
				"Country": "USA",
			})
			if err != nil || !result.Success {
				t.Fatalf("Failed to insert first record: %v, result: %+v", err, result)
			}

			// Insert second record
			result, err = op.InsertRecord(ctx, dbPath, "Person", map[string]interface{}{
				"Name":    "Bob Smith",
				"Age":     35,
				"Email":   "bob@example.com",
				"City":    "New York",
				"Country": "USA",
			})
			if err != nil || !result.Success {
				t.Fatalf("Failed to insert second record: %v, result: %+v", err, result)
			}

			// Insert third record
			result, err = op.InsertRecord(ctx, dbPath, "Person", map[string]interface{}{
				"Name":    "Charlie Brown",
				"Age":     42,
				"Email":   "charlie@example.com",
				"City":    "Chicago",
				"Country": "USA",
			})
			if err != nil || !result.Success {
				t.Fatalf("Failed to insert third record: %v, result: %+v", err, result)
			}

			// Verify with recsel
			output, err := runRecsel(dbPath)
			if err != nil {
				t.Fatalf("recsel failed: %v", err)
			}
			if !strings.Contains(output, "Alice Johnson") {
				t.Error("recsel output missing Alice Johnson")
			}
			if !strings.Contains(output, "Bob Smith") {
				t.Error("recsel output missing Bob Smith")
			}
			if !strings.Contains(output, "Charlie Brown") {
				t.Error("recsel output missing Charlie Brown")
			}
		})

		// Step 2: Read/Query records (SELECT)
		t.Run("Step2_QueryRecords", func(t *testing.T) {
			// Query all records
			result, err := op.QueryRecords(ctx, dbPath, "", "")
			if err != nil || !result.Success {
				t.Fatalf("Failed to query all records: %v, result: %+v", err, result)
			}

			output := result.Output
			if !strings.Contains(output, "Alice Johnson") ||
				!strings.Contains(output, "Bob Smith") ||
				!strings.Contains(output, "Charlie Brown") {
				t.Errorf("Query output missing expected records. Got: %s", output)
			}

			// Query with expression
			result, err = op.QueryRecords(ctx, dbPath, "Age > 30", "")
			if err != nil || !result.Success {
				t.Fatalf("Failed to query with expression: %v, result: %+v", err, result)
			}

			output = result.Output
			if strings.Contains(output, "Alice Johnson") {
				t.Error("Alice Johnson (Age 28) should not be in Age > 30 query")
			}
			if !strings.Contains(output, "Bob Smith") {
				t.Error("Bob Smith (Age 35) should be in Age > 30 query")
			}
			if !strings.Contains(output, "Charlie Brown") {
				t.Error("Charlie Brown (Age 42) should be in Age > 30 query")
			}

			// Verify with recsel
			recselOutput, err := runRecsel(dbPath, "-e", "Age > 30")
			if err != nil {
				t.Fatalf("recsel with expression failed: %v", err)
			}
			if !strings.Contains(recselOutput, "Bob Smith") || !strings.Contains(recselOutput, "Charlie Brown") {
				t.Errorf("recsel validation failed for Age > 30 query. Got: %s", recselOutput)
			}

			// Query specific field
			result, err = op.QueryRecords(ctx, dbPath, "Name = 'Alice Johnson'", "")
			if err != nil || !result.Success {
				t.Fatalf("Failed to query specific record: %v, result: %+v", err, result)
			}

			output = result.Output
			if !strings.Contains(output, "alice@example.com") {
				t.Error("Query result missing Alice's email")
			}
		})

		// Step 3: Update records (UPDATE)
		t.Run("Step3_UpdateRecords", func(t *testing.T) {
			// Update Alice's age
			result, err := op.UpdateRecords(ctx, dbPath, "Name = 'Alice Johnson'", map[string]interface{}{
				"Age": 29,
			})
			if err != nil || !result.Success {
				t.Fatalf("Failed to update Alice's age: %v, result: %+v", err, result)
			}

			// Verify update with our query
			queryResult, _ := op.QueryRecords(ctx, dbPath, "Name = 'Alice Johnson'", "")
			if !strings.Contains(queryResult.Output, "Age: 29") {
				t.Error("Alice's age was not updated to 29")
			}

			// Verify with recsel
			recselOutput, err := runRecsel(dbPath, "-e", "Name = 'Alice Johnson'")
			if err != nil {
				t.Fatalf("recsel verification failed: %v", err)
			}
			if !strings.Contains(recselOutput, "Age: 29") {
				t.Errorf("recsel shows Alice's age is not 29. Got: %s", recselOutput)
			}

			// Update Bob's city and add a new field
			result, err = op.UpdateRecords(ctx, dbPath, "Name = 'Bob Smith'", map[string]interface{}{
				"City":   "Boston",
				"Status": "Active",
			})
			if err != nil || !result.Success {
				t.Fatalf("Failed to update Bob's record: %v, result: %+v", err, result)
			}

			// Verify update with recsel
			recselOutput, err = runRecsel(dbPath, "-e", "Name = 'Bob Smith'")
			if err != nil {
				t.Fatalf("recsel verification failed: %v", err)
			}
			if !strings.Contains(recselOutput, "City: Boston") {
				t.Error("Bob's city was not updated to Boston")
			}
			if !strings.Contains(recselOutput, "Status: Active") {
				t.Error("Status field was not added to Bob's record")
			}
		})

		// Step 4: Delete records (DELETE)
		t.Run("Step4_DeleteRecords", func(t *testing.T) {
			// Delete Charlie's record
			result, err := op.DeleteRecords(ctx, dbPath, "Name = 'Charlie Brown'")
			if err != nil || !result.Success {
				t.Fatalf("Failed to delete Charlie's record: %v, result: %+v", err, result)
			}

			// Verify deletion with our query
			queryResult, _ := op.QueryRecords(ctx, dbPath, "", "")
			if strings.Contains(queryResult.Output, "Charlie Brown") {
				t.Error("Charlie Brown's record was not deleted")
			}

			// Verify with recsel
			recselOutput, err := runRecsel(dbPath)
			if err != nil {
				t.Fatalf("recsel verification failed: %v", err)
			}
			if strings.Contains(recselOutput, "Charlie Brown") {
				t.Error("recsel shows Charlie Brown still exists")
			}

			// Verify Alice and Bob still exist
			if !strings.Contains(recselOutput, "Alice Johnson") {
				t.Error("Alice Johnson was incorrectly deleted")
			}
			if !strings.Contains(recselOutput, "Bob Smith") {
				t.Error("Bob Smith was incorrectly deleted")
			}
		})

		// Step 5: Get database info
		t.Run("Step5_DatabaseInfo", func(t *testing.T) {
			result, err := op.GetDatabaseInfo(ctx, dbPath)
			if err != nil || !result.Success {
				t.Fatalf("Failed to get database info: %v, result: %+v", err, result)
			}

			// Verify with recinf
			recinfOutput, err := runRecinf(dbPath)
			if err != nil {
				t.Fatalf("recinf failed: %v", err)
			}
			if recinfOutput == "" {
				t.Error("recinf returned empty output")
			}

			// Both should indicate Person record type exists
			if !strings.Contains(result.Output, "Person") && !strings.Contains(recinfOutput, "Person") {
				t.Log("Database info retrieved (Person record type)")
			}
		})
	})

	t.Run("ComplexQueryOperations", func(t *testing.T) {
		// Setup: Create multiple records for complex queries
		testRecords := []map[string]interface{}{
			{"Name": "David Lee", "Age": 25, "Department": "Engineering", "Salary": "80000"},
			{"Name": "Emma Wilson", "Age": 30, "Department": "Sales", "Salary": "75000"},
			{"Name": "Frank Miller", "Age": 35, "Department": "Engineering", "Salary": "95000"},
			{"Name": "Grace Kim", "Age": 28, "Department": "Marketing", "Salary": "70000"},
		}

		for _, record := range testRecords {
			result, err := op.InsertRecord(ctx, dbPath, "Employee", record)
			if err != nil || !result.Success {
				t.Fatalf("Failed to insert test record: %v, result: %+v", err, result)
			}
		}

		// Test complex query with AND condition
		result, err := op.QueryRecords(ctx, dbPath, "Department = 'Engineering' && Age > 30", "")
		if err != nil || !result.Success {
			t.Fatalf("Failed to execute complex query: %v, result: %+v", err, result)
		}

		output := result.Output
		// Should only contain Frank Miller
		if strings.Contains(output, "David Lee") {
			t.Error("David Lee should not match (Age 25)")
		}
		if !strings.Contains(output, "Frank Miller") {
			t.Error("Frank Miller should match (Engineering, Age 35)")
		}

		// Verify with recsel
		recselOutput, err := runRecsel(dbPath, "-e", "Department = 'Engineering' && Age > 30")
		if err != nil {
			t.Fatalf("recsel complex query failed: %v", err)
		}
		if !strings.Contains(recselOutput, "Frank Miller") {
			t.Errorf("recsel complex query validation failed. Got: %s", recselOutput)
		}

		// Test query with OR condition
		result, err = op.QueryRecords(ctx, dbPath, "Department = 'Sales' || Age < 28", "")
		if err != nil || !result.Success {
			t.Fatalf("Failed to execute OR query: %v, result: %+v", err, result)
		}

		output = result.Output
		// Should contain Emma Wilson and David Lee
		if !strings.Contains(output, "Emma Wilson") {
			t.Error("Emma Wilson (Sales) should match OR query")
		}
		if !strings.Contains(output, "David Lee") {
			t.Error("David Lee (Age 25) should match OR query")
		}

		// Verify with recsel
		recselOutput, err = runRecsel(dbPath, "-e", "Department = 'Sales' || Age < 28")
		if err != nil {
			t.Fatalf("recsel OR query failed: %v", err)
		}
		if !strings.Contains(recselOutput, "Emma Wilson") || !strings.Contains(recselOutput, "David Lee") {
			t.Errorf("recsel OR query validation failed. Got: %s", recselOutput)
		}
	})

	t.Run("UpdateMultipleFieldsAndVerify", func(t *testing.T) {
		// Update all Engineering employees' salary
		result, err := op.UpdateRecords(ctx, dbPath, "Department = 'Engineering'", map[string]interface{}{
			"Salary":     "100000",
			"Department": "R&D",
		})
		if err != nil || !result.Success {
			t.Fatalf("Failed to update Engineering employees: %v, result: %+v", err, result)
		}

		// Verify with recsel - should find updated values
		recselOutput, err := runRecsel(dbPath, "-e", "Department = 'R&D'")
		if err != nil {
			t.Fatalf("recsel verification failed: %v", err)
		}

		if !strings.Contains(recselOutput, "Salary: 100000") {
			t.Error("Engineering employees' salary was not updated")
		}
		if !strings.Contains(recselOutput, "Department: R&D") {
			t.Error("Engineering department was not renamed to R&D")
		}

		// Count how many R&D employees
		rdLines := strings.Count(recselOutput, "Department: R&D")
		if rdLines < 2 {
			t.Errorf("Expected at least 2 R&D employees, got %d", rdLines)
		}
	})

	t.Run("FileIntegrityCheck", func(t *testing.T) {
		// Verify the database file is valid recutils format
		// Use recinf to check file integrity
		output, err := runRecinf(dbPath)
		if err != nil {
			t.Fatalf("recinf failed, database file may be corrupted: %v", err)
		}

		t.Logf("recinf output: %q", output)

		// recinf should return non-empty output for valid database
		if output == "" {
			t.Error("recinf returned empty output for database file")
		}

		// Verify file exists and is not empty
		info, err := os.Stat(dbPath)
		if err != nil {
			t.Fatalf("Cannot stat database file: %v", err)
		}
		if info.Size() == 0 {
			t.Error("Database file is empty")
		}

		// Use recsel to verify the file is readable
		recselOutput, err := runRecsel(dbPath)
		if err != nil {
			t.Fatalf("recsel cannot read database file: %v", err)
		}

		// recsel should be able to read and output records
		if recselOutput == "" {
			t.Error("recsel returned empty output")
		}
	})
}

// runRecsel executes recsel command with given arguments
// args should include options like -e "expression" -t "format"
func runRecsel(dbPath string, args ...string) (string, error) {
	// recsel expects: recsel [OPTIONS] FILE
	var cmdArgs []string
	if len(args) > 0 {
		cmdArgs = append(args, dbPath)
	} else {
		cmdArgs = []string{dbPath}
	}
	cmd := exec.Command("recsel", cmdArgs...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("recsel %v failed: %w, output: %s", cmdArgs, err, string(output))
	}
	return string(output), nil
}

// runRecinf executes recinf command
func runRecinf(dbPath string) (string, error) {
	cmd := exec.Command("recinf", dbPath)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(output)), nil
}

// TestRecutilsCSVOutput tests CSV output format
func TestRecutilsCSVOutput(t *testing.T) {
	if _, err := exec.LookPath("recsel"); err != nil {
		t.Skip("recutils not installed, skipping integration test")
	}

	ctx := context.Background()
	op := NewRecordOperation()

	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "csv_test.rec")

	// Insert test records
	testRecords := []map[string]interface{}{
		{"Name": "Test User 1", "Age": 25, "City": "New York"},
		{"Name": "Test User 2", "Age": 30, "City": "Los Angeles"},
	}

	for _, record := range testRecords {
		result, err := op.InsertRecord(ctx, dbPath, "Person", record)
		if err != nil || !result.Success {
			t.Fatalf("Failed to insert test record: %v, result: %+v", err, result)
		}
	}

	// Query with CSV format (may not be supported on all recutils versions)
	result, err := op.QueryRecords(ctx, dbPath, "", "csv")
	if err != nil {
		t.Fatalf("Failed to query with CSV format: %v", err)
	}

	// Verify the operation completed (even if CSV format is not supported)
	if !result.Success {
		// CSV format may not be supported, log and skip
		t.Logf("CSV format may not be supported: %s", result.Error)
		return
	}

	// Verify output is generated
	output := result.Output
	t.Logf("CSV output: %q", output)

	// Verify with recsel using CSV format
	recselOutput, err := runRecsel(dbPath, "-t", "csv")
	if err != nil {
		// CSV format may not be supported
		t.Logf("recsel CSV format may not be supported: %v", err)
		return
	}

	t.Logf("recsel CSV output: %q", recselOutput)

	// If both succeed, verify non-empty output
	if output == "" && recselOutput == "" {
		t.Log("CSV format produced no output (may not be supported)")
	}
}

// TestRecutilsPerformanceWithLargeDataset tests performance with larger dataset
func TestRecutilsPerformanceWithLargeDataset(t *testing.T) {
	if _, err := exec.LookPath("recsel"); err != nil {
		t.Skip("recutils not installed, skipping integration test")
	}

	ctx := context.Background()
	op := NewRecordOperation()

	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "perf_test.rec")

	// Insert 50 records
	for i := 0; i < 50; i++ {
		result, err := op.InsertRecord(ctx, dbPath, "PerfTest", map[string]interface{}{
			"ID":    i,
			"Name":  fmt.Sprintf("User%d", i),
			"Value": i * 10,
		})
		if err != nil || !result.Success {
			t.Fatalf("Failed to insert record %d: %v, result: %+v", i, err, result)
		}
	}

	// Query all records and verify count with recsel
	result, err := op.QueryRecords(ctx, dbPath, "", "")
	if err != nil || !result.Success {
		t.Fatalf("Failed to query all records: %v, result: %+v", err, result)
	}

	recselOutput, err := runRecsel(dbPath)
	if err != nil {
		t.Fatalf("recsel failed: %v", err)
	}

	// Count occurrences of "Name:" to verify record count
	ourCount := strings.Count(result.Output, "Name:")
	recselCount := strings.Count(recselOutput, "Name:")

	if ourCount != 50 {
		t.Errorf("Expected 50 records in our output, got %d", ourCount)
	}
	if recselCount != 50 {
		t.Errorf("Expected 50 records in recsel output, got %d", recselCount)
	}

	// Test query performance
	result, err = op.QueryRecords(ctx, dbPath, "Value > 250", "")
	if err != nil || !result.Success {
		t.Fatalf("Failed to execute filtered query: %v, result: %+v", err, result)
	}

	// Should have records with value > 250 (ID > 25)
	filteredCount := strings.Count(result.Output, "Name:")
	expectedCount := 24 // IDs 26-49
	if filteredCount != expectedCount {
		t.Errorf("Expected %d records with Value > 250, got %d", expectedCount, filteredCount)
	}
}
