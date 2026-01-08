// recutils package: Encapsulate all recutils database operations
package recutils

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"
)

// Result Execution result structure
type Result struct {
	Success bool   `json:"success"`
	Output  string `json:"output"`
	Error   string `json:"error"`
}

// RecordOperation recutils operation interface
type RecordOperation struct{}

// NewRecordOperation Create new operation instance
func NewRecordOperation() *RecordOperation {
	return &RecordOperation{}
}

// executeRecCommand Execute recutils command
func (ro *RecordOperation) executeRecCommand(ctx context.Context, cmd []string, inputData string) (*Result, error) {
	var stdout, stderr bytes.Buffer

	command := exec.CommandContext(ctx, cmd[0], cmd[1:]...)
	command.Stdout = &stdout
	command.Stderr = &stderr

	if inputData != "" {
		command.Stdin = strings.NewReader(inputData)
	}

	// Set timeout
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	err := command.Run()
	if err != nil {
		return &Result{
			Success: false,
			Output:  stdout.String(),
			Error:   stderr.String(),
		}, nil
	}

	return &Result{
		Success: true,
		Output:  strings.TrimSpace(stdout.String()),
		Error:   strings.TrimSpace(stderr.String()),
	}, nil
}

// QueryRecords Query records
func (ro *RecordOperation) QueryRecords(ctx context.Context, databaseFile, queryExpression, outputFormat string) (*Result, error) {
	cmd := []string{"recsel"}

	// Add output format (if any)
	if outputFormat != "" {
		cmd = append(cmd, "-t", outputFormat)
	}

	cmd = append(cmd, databaseFile)

	// Add query expression (if any)
	if queryExpression != "" {
		cmd = append(cmd, "-e", queryExpression)
	}

	return ro.executeRecCommand(ctx, cmd, "")
}

// InsertRecord Insert new record using recins command
func (ro *RecordOperation) InsertRecord(ctx context.Context, databaseFile, recordType string, fields map[string]interface{}) (*Result, error) {
	// Build record content for recins
	var recordLines []string
	for fieldName, fieldValue := range fields {
		recordLines = append(recordLines, fmt.Sprintf("%s: %v", fieldName, fieldValue))
	}
	recordContent := strings.Join(recordLines, "\n")

	// Check if database file exists or is empty
	fileInfo, err := os.Stat(databaseFile)
	if os.IsNotExist(err) || (err == nil && fileInfo.Size() == 0) {
		// If file does not exist or is empty, create new record set with %rec: directive
		content := fmt.Sprintf("%%rec: %s\n\n%s\n", recordType, recordContent)
		err = os.WriteFile(databaseFile, []byte(content), 0644)
		if err != nil {
			return &Result{
				Success: false,
				Output:  "",
				Error:   err.Error(),
			}, fmt.Errorf("failed to write database file: %w", err)
		}
		return &Result{
			Success: true,
			Output:  "Record inserted successfully",
			Error:   "",
		}, nil
	} else if err != nil {
		return &Result{
			Success: false,
			Output:  "",
			Error:   err.Error(),
		}, fmt.Errorf("failed to stat database file: %w", err)
	}

	// Use recins to insert record into existing database
	// -t specifies the record type, -r specifies the record content
	cmd := []string{"recins", "-t", recordType, "-r", recordContent, databaseFile}
	result, err := ro.executeRecCommand(ctx, cmd, "")

	if err != nil || !result.Success {
		return &Result{
			Success: false,
			Output:  result.Output,
			Error:   result.Error,
		}, err
	}

	return &Result{
		Success: true,
		Output:  "Record inserted successfully",
		Error:   "",
	}, nil
}

// DeleteRecords Delete records
func (ro *RecordOperation) DeleteRecords(ctx context.Context, databaseFile, queryExpression string) (*Result, error) {
	// Backup original file
	backupFile := databaseFile + ".bak"
	originalContent, err := os.ReadFile(databaseFile)
	if err != nil {
		return &Result{
			Success: false,
			Output:  "",
			Error:   err.Error(),
		}, fmt.Errorf("failed to read database file: %w", err)
	}

	err = os.WriteFile(backupFile, originalContent, 0644)
	if err != nil {
		return &Result{
			Success: false,
			Output:  "",
			Error:   err.Error(),
		}, fmt.Errorf("failed to create backup: %w", err)
	}

	// Extract record type declaration from original file
	recordTypeDecl := extractRecordTypeDeclaration(string(originalContent))

	// Use recsel to get records to keep
	cmd := []string{"recsel", "-e", fmt.Sprintf("!(%s)", queryExpression), databaseFile}
	result, err := ro.executeRecCommand(ctx, cmd, "")

	if result.Success {
		// Rebuild file with record type declaration
		var output string
		if recordTypeDecl != "" {
			output = recordTypeDecl + "\n\n"
		}
		if result.Output != "" {
			output += result.Output + "\n"
		}

		err = os.WriteFile(databaseFile, []byte(output), 0644)
		if err != nil {
			// Restore backup
			os.WriteFile(backupFile, originalContent, 0644)
			return &Result{
				Success: false,
				Output:  "",
				Error:   err.Error(),
			}, fmt.Errorf("failed to write database file: %w", err)
		}

		// Delete backup file
		os.Remove(backupFile)

		return &Result{
			Success: true,
			Output:  fmt.Sprintf("Records matching '%s' deleted successfully", queryExpression),
			Error:   "",
		}, nil
	} else {
		// Restore backup
		os.WriteFile(backupFile, originalContent, 0644)
		os.Remove(backupFile)
		return result, nil
	}
}

// UpdateRecords Update records
func (ro *RecordOperation) UpdateRecords(ctx context.Context, databaseFile, queryExpression string, fields map[string]interface{}) (*Result, error) {
	// Get records to update
	queryCmd := []string{"recsel", "-e", queryExpression, databaseFile}
	queryResult, err := ro.executeRecCommand(ctx, queryCmd, "")

	if err != nil || !queryResult.Success {
		return &Result{
			Success: false,
			Output:  "",
			Error:   queryResult.Error,
		}, err
	}

	// Backup original file
	backupFile := databaseFile + ".bak"
	originalContent, err := os.ReadFile(databaseFile)
	if err != nil {
		return &Result{
			Success: false,
			Output:  "",
			Error:   err.Error(),
		}, fmt.Errorf("failed to read database file: %w", err)
	}

	err = os.WriteFile(backupFile, originalContent, 0644)
	if err != nil {
		return &Result{
			Success: false,
			Output:  "",
			Error:   err.Error(),
		}, fmt.Errorf("failed to create backup: %w", err)
	}

	// Get non-matching records
	keepCmd := []string{"recsel", "-e", fmt.Sprintf("!(%s)", queryExpression), databaseFile}
	keepResult, err := ro.executeRecCommand(ctx, keepCmd, "")

	if keepResult.Success {
		// Extract record type declaration from original file
		recordTypeDecl := extractRecordTypeDeclaration(string(originalContent))

		// Build updated records
		updatedRecords := []string{}
		lines := strings.Split(queryResult.Output, "\n")
		currentRecord := []string{}

		for _, line := range lines {
			line = strings.TrimSpace(line)
			if line != "" {
				currentRecord = append(currentRecord, line)
			}
		}

		// Update fields
		for i, line := range currentRecord {
			for fieldName, fieldValue := range fields {
				if strings.HasPrefix(line, fieldName+":") {
					currentRecord[i] = fmt.Sprintf("%s: %v", fieldName, fieldValue)
				} else {
					// Check if field exists
					exists := false
					for _, l := range currentRecord {
						if strings.HasPrefix(l, fieldName+":") {
							exists = true
							break
						}
					}
					if !exists {
						currentRecord = append(currentRecord, fmt.Sprintf("%s: %v", fieldName, fieldValue))
					}
				}
			}
		}

		updatedRecords = currentRecord

		// Write back to file with record type declaration
		var output string
		if recordTypeDecl != "" {
			output = recordTypeDecl + "\n\n"
		}

		// Add non-matching records
		if keepResult.Output != "" {
			output += keepResult.Output
			if len(updatedRecords) > 0 {
				output += "\n\n"
			}
		}

		// Add updated records
		if len(updatedRecords) > 0 {
			output += strings.Join(updatedRecords, "\n") + "\n"
		}

		err = os.WriteFile(databaseFile, []byte(output), 0644)
		if err != nil {
			// Restore backup
			os.WriteFile(backupFile, originalContent, 0644)
			return &Result{
				Success: false,
				Output:  "",
				Error:   err.Error(),
			}, fmt.Errorf("failed to write database file: %w", err)
		}

		// Delete backup file
		os.Remove(backupFile)

		return &Result{
			Success: true,
			Output:  "Records updated successfully",
			Error:   "",
		}, nil
	} else {
		// Restore backup
		os.WriteFile(backupFile, originalContent, 0644)
		os.Remove(backupFile)
		return keepResult, nil
	}
}

// GetDatabaseInfo Get database info
func (ro *RecordOperation) GetDatabaseInfo(ctx context.Context, databaseFile string) (*Result, error) {
	cmd := []string{"recinf", databaseFile}
	return ro.executeRecCommand(ctx, cmd, "")
}

// extractRecordTypeDeclaration extracts the record type declaration (e.g., "%rec: Person") from file content
func extractRecordTypeDeclaration(content string) string {
	lines := strings.Split(content, "\n")
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "%rec:") {
			return trimmed
		}
	}
	return ""
}
