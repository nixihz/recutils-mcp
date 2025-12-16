// recutils package: Encapsulate all recutils database operations
package recutils

import (
	"bytes"
	"context"
	"fmt"
	"io/ioutil"
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

// InsertRecord Insert new record
func (ro *RecordOperation) InsertRecord(ctx context.Context, databaseFile, recordType string, fields map[string]interface{}) (*Result, error) {
	// Build record content
	var recordLines []string
	for fieldName, fieldValue := range fields {
		recordLines = append(recordLines, fmt.Sprintf("%s: %v", fieldName, fieldValue))
	}

	// Check if database file exists or is empty
	fileInfo, err := os.Stat(databaseFile)
	if os.IsNotExist(err) || (err == nil && fileInfo.Size() == 0) {
		// If file does not exist or is empty, create new record set
		// %rec: directive requires blank line separator, fields separated by newlines
		content := fmt.Sprintf("%%rec: %s\n\n%s\n", recordType, strings.Join(recordLines, "\n"))
		err = ioutil.WriteFile(databaseFile, []byte(content), 0644)
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
	} else {
		// If file exists, append new record (using blank line separator)
		file, err := os.OpenFile(databaseFile, os.O_APPEND|os.O_WRONLY, 0644)
		if err != nil {
			return &Result{
				Success: false,
				Output:  "",
				Error:   err.Error(),
			}, fmt.Errorf("failed to open database file: %w", err)
		}
		defer file.Close()

		// Check if file ends with newline
		file.Seek(-1, 2) // Move to second-to-last byte of file
		lastChar := make([]byte, 1)
		file.Read(lastChar)

		// If file does not end with newline, add newline first
		if lastChar[0] != '\n' {
			_, err = file.WriteString("\n")
			if err != nil {
				return &Result{
					Success: false,
					Output:  "",
					Error:   err.Error(),
				}, fmt.Errorf("failed to write newline: %w", err)
			}
		}

		// Add blank line separator and record content (add newline at end of record)
		_, err = file.WriteString("\n" + strings.Join(recordLines, "\n") + "\n")
		if err != nil {
			return &Result{
				Success: false,
				Output:  "",
				Error:   err.Error(),
			}, fmt.Errorf("failed to write record: %w", err)
		}

		return &Result{
			Success: true,
			Output:  "Record inserted successfully",
			Error:   "",
		}, nil
	}
}

// DeleteRecords Delete records
func (ro *RecordOperation) DeleteRecords(ctx context.Context, databaseFile, queryExpression string) (*Result, error) {
	// Backup original file
	backupFile := databaseFile + ".bak"
	originalContent, err := ioutil.ReadFile(databaseFile)
	if err != nil {
		return &Result{
			Success: false,
			Output:  "",
			Error:   err.Error(),
		}, fmt.Errorf("failed to read database file: %w", err)
	}

	err = ioutil.WriteFile(backupFile, originalContent, 0644)
	if err != nil {
		return &Result{
			Success: false,
			Output:  "",
			Error:   err.Error(),
		}, fmt.Errorf("failed to create backup: %w", err)
	}

	// Use recsel to get records to keep
	cmd := []string{"recsel", "-e", fmt.Sprintf("!(%s)", queryExpression), databaseFile}
	result, err := ro.executeRecCommand(ctx, cmd, "")

	if result.Success {
		err = ioutil.WriteFile(databaseFile, []byte(result.Output), 0644)
		if err != nil {
			// Restore backup
			ioutil.WriteFile(backupFile, originalContent, 0644)
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
		ioutil.WriteFile(backupFile, originalContent, 0644)
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
	originalContent, err := ioutil.ReadFile(databaseFile)
	if err != nil {
		return &Result{
			Success: false,
			Output:  "",
			Error:   err.Error(),
		}, fmt.Errorf("failed to read database file: %w", err)
	}

	err = ioutil.WriteFile(backupFile, originalContent, 0644)
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

		// Write back to file
		output := keepResult.Output
		if len(updatedRecords) > 0 {
			if output != "" {
				output += "\n\n"
			}
			output += strings.Join(updatedRecords, "\n") + "\n"
		}

		err = ioutil.WriteFile(databaseFile, []byte(output), 0644)
		if err != nil {
			// Restore backup
			ioutil.WriteFile(backupFile, originalContent, 0644)
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
		ioutil.WriteFile(backupFile, originalContent, 0644)
		os.Remove(backupFile)
		return keepResult, nil
	}
}

// GetDatabaseInfo Get database info
func (ro *RecordOperation) GetDatabaseInfo(ctx context.Context, databaseFile string) (*Result, error) {
	cmd := []string{"recinf", databaseFile}
	return ro.executeRecCommand(ctx, cmd, "")
}
