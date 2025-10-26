package main

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"path/filepath"
)

type Conf struct {
	TargetFolder string `json:"target_folder"`
	SourceFile   string `json:"source_file"`
	TargetFile   string `json:"target_file"`
}

func main() {
	// 1. Get configuration from Environment Variables
	key64 := os.Getenv("GITHUB_KEY")
	if key64 == "" {
		log.Fatal("Error: GITHUB_KEY environment variable not set.")
	}

	keyCreateErr := createFileFromBase64("github_key", key64)
	if keyCreateErr != nil {
		log.Fatal(keyCreateErr)
	}

	keyPub64 := os.Getenv("GITHUB_KEY_PUB")
	if keyPub64 == "" {
		log.Fatal("Error: GITHUB_KEY_PUB environment variable not set.")
	}

	keyPubCreateErr := createFileFromBase64("github_key.pub", keyPub64)
	if keyPubCreateErr != nil {
		log.Fatal(keyPubCreateErr)
	}

	repoURL := os.Getenv("REPO_URL")
	if repoURL == "" {
		log.Fatal("Error: REPO_URL environment variable not set.")
	}

	cloneErr := cloneRepo(repoURL)
	if cloneErr != nil {
		log.Fatal(cloneErr)
	}

	configs, err := parseConfigsJSON()
	if err != nil {
		log.Fatal(err)
	}

	for i := 0; i < len(configs); i++ {
		fmt.Println(configs[i].SourceFile)
	}

	copyConfigs(configs)
}

func cloneRepo(repoURL string) error {
	if _, err := os.Stat("./repo"); !os.IsNotExist(err) {
		_ = os.RemoveAll("./repo")
	}

	sshCommand := "ssh -i ./github_key -o StrictHostKeyChecking=no"

	cmd := exec.Command("git", "clone", repoURL, "repo")

	cmd.Env = append(os.Environ(), fmt.Sprintf("GIT_SSH_COMMAND=%s", sshCommand))

	output, err := cmd.CombinedOutput()
	if err != nil {
		// If an error occurs, log the error and the combined output from the command.
		log.Fatalf("Failed to clone repository. Error: %v\nOutput:\n%s", err, string(output))
		return err
	}

	fmt.Printf("Successfully cloned repository!\nOutput:\n%s", string(output))
	return nil
}

func createFileFromBase64(filename string, data string) error {
	filePath := fmt.Sprintf("./%s", filename)
	if _, err := os.Stat(filePath); errors.Is(err, os.ErrNotExist) {
		_ = os.Remove(filePath)
	}

	decodedData, err := base64.StdEncoding.DecodeString(data)
	if err != nil {
		log.Fatalf("Error decoding Base64 string: %v", err)
		return err
	}

	// 2. Write the decoded data (byte slice) to a file
	// os.WriteFile creates the file if it doesn't exist,
	// and truncates it if it does.
	// 0644 is a standard file permission.
	err = os.WriteFile(filename, decodedData, 0644)
	if err != nil {
		log.Fatalf("Error writing file: %v", err)
		return err
	}

	fmt.Printf("File '%s' created successfully from Base64 string.\n", filename)
	return nil
}

func parseConfigsJSON() ([]Conf, error) {
	// Read the entire file
	fileBytes, err := os.ReadFile("./repo/configs.json")
	if err != nil {
		log.Fatalf("Error reading configs.json: %v", err)
		return nil, err
	}

	// Create a slice to hold the configurations
	var configs []Conf

	// Unmarshal (parse) the JSON byte slice into the 'configs' slice
	err = json.Unmarshal(fileBytes, &configs)
	if err != nil {
		log.Fatalf("Error unmarshaling JSON: %v", err)
		return nil, err
	}

	// Success! Print the parsed data to verify
	fmt.Println("Successfully parsed configs:")

	return configs, nil
}

func copyConfigs(configsToCopy []Conf) {
	baseSourceDir := "./repo/configs"
	baseTargetDir := "./data/" // Current directory

	for _, conf := range configsToCopy {
		// Construct the full paths based on the base directories
		fullTargetFolder := filepath.Join(baseTargetDir, conf.TargetFolder)
		fullSourceFile := filepath.Join(baseSourceDir, conf.SourceFile)
		fullTargetFile := filepath.Join(fullTargetFolder, conf.TargetFile)

		fmt.Printf("--- Processing config for: %s ---\n", fullTargetFile)

		// 1. Check if fullTargetFolder exists, if not create it
		// os.Stat returns an error if the path doesn't exist.
		if _, err := os.Stat(fullTargetFolder); os.IsNotExist(err) {
			fmt.Printf("Directory %s does not exist, creating...\n", fullTargetFolder)
			// Create the directory with 0755 permissions (rwxr-xr-x)
			// MkdirAll creates parent directories if needed.
			if err := os.MkdirAll(fullTargetFolder, 0755); err != nil {
				// Use log.Fatalf to print the error and exit(1)
				log.Fatalf("Error: Could not create directory %s: %v\n", fullTargetFolder, err)
			}
			fmt.Printf("Directory %s created successfully.\n", fullTargetFolder)
		} else if err != nil {
			// Handle other potential errors from os.Stat (e.g., permission issues)
			log.Fatalf("Error: Could not check directory %s: %v\n", fullTargetFolder, err)
		} else {
			fmt.Printf("Directory %s already exists.\n", fullTargetFolder)
		}

		// 2. Remove fullTargetFile if it exists
		// os.Remove returns an error if the file doesn't exist, which we can ignore.
		if err := os.Remove(fullTargetFile); err == nil {
			fmt.Printf("Removed existing file: %s\n", fullTargetFile)
		} else if !os.IsNotExist(err) {
			// Report errors other than "file not found"
			// Use log.Printf for non-fatal warnings
			log.Printf("Warning: Could not remove file %s: %v\n", fullTargetFile, err)
			// We might still be able to overwrite, so we don't exit here.
		} else {
			fmt.Printf("File %s does not exist, no removal needed.\n", fullTargetFile)
		}

		// 3. Copy fullSourceFile to fullTargetFile
		fmt.Printf("Attempting to copy %s to %s...\n", fullSourceFile, fullTargetFile)

		// Open the source file for reading
		src, err := os.Open(fullSourceFile)
		if err != nil {
			log.Fatalf("Error: Could not open source file %s: %v\n", fullSourceFile, err)
		}
		// Defer closing the source file. Note: In a long-running app,
		// deferring inside a loop can be risky, but for a simple script
		// that exits, this is fine. It will close when main returns.
		// A more robust way would be to close it explicitly at the end of the loop.
		// Let's do that instead for correctness.

		// Create the destination file for writing (it will truncate if it exists)
		dst, err := os.Create(fullTargetFile)
		if err != nil {
			src.Close() // Close source file on error
			log.Fatalf("Error: Could not create destination file %s: %v\n", fullTargetFile, err)
		}

		// Copy the contents from source to destination
		bytesCopied, err := io.Copy(dst, src)

		// Close files immediately after copy/error
		src.Close()

		if err != nil {
			dst.Close() // Attempt to close dst even on copy error
			log.Fatalf("Error: Could not copy file contents: %v\n", err)
		}

		// Ensure the data is written to stable storage
		if err := dst.Sync(); err != nil {
			log.Printf("Warning: Could not sync file %s: %v\n", fullTargetFile, err)
		}

		// Close destination file after sync
		if err := dst.Close(); err != nil {
			log.Printf("Warning: Could not close destination file %s: %v\n", fullTargetFile, err)
		}

		fmt.Printf("Successfully copied %d bytes to %s.\n", bytesCopied, fullTargetFile)
		fmt.Println("----------------------------------------")
	}
}
