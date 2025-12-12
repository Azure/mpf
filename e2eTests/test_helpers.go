package e2etests

import (
	"encoding/json"
	"io"
	"os"
	"path/filepath"
	"testing"
)

func copyDir(t *testing.T, src, dst string) error {
	t.Helper()

	return filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		relPath, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}

		dstPath := filepath.Join(dst, relPath)

		if info.IsDir() {
			return os.MkdirAll(dstPath, info.Mode())
		}

		srcFile, err := os.Open(path)
		if err != nil {
			return err
		}
		defer srcFile.Close() //nolint:errcheck

		dstFile, err := os.Create(dstPath)
		if err != nil {
			return err
		}
		defer dstFile.Close() //nolint:errcheck

		_, err = io.Copy(dstFile, srcFile)
		return err
	})
}

func updateJSONParamFile(t *testing.T, filePath string, updates map[string]string) error {
	t.Helper()

	file, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer file.Close() //nolint:errcheck

	var data map[string]any
	decoder := json.NewDecoder(file)
	if err := decoder.Decode(&data); err != nil {
		return err
	}

	parameters, ok := data["parameters"].(map[string]any)
	if !ok {
		// Handle case where parameters might be at root or different structure
		// For ARM params, it's usually { "parameters": { ... } }
		// If not found, maybe it's just key-value pairs?
		// Let's assume standard ARM param file structure for now.
		// If not, we might need to adjust.
		// But wait, some param files might be simple key-value?
		// Standard is:
		// {
		//   "$schema": "...",
		//   "contentVersion": "...",
		//   "parameters": {
		//     "paramName": { "value": "..." }
		//   }
		// }
		return nil // Or error?
	}

	for key, value := range updates {
		if param, exists := parameters[key].(map[string]any); exists {
			param["value"] = value
		} else {
			// If param doesn't exist, create it
			parameters[key] = map[string]any{
				"value": value,
			}
		}
	}

	file, err = os.Create(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	return encoder.Encode(data)
}
