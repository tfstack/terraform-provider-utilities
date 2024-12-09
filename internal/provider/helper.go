package provider

import (
	"io"
	"net/http"
	"os"
	"path/filepath"

	"github.com/pkg/errors"
)

// Function to download a file from URL.
func downloadFile(url, dest string) error {
	// Ensure the destination directory exists
	dir := filepath.Dir(dest)
	if err := os.MkdirAll(dir, os.ModePerm); err != nil {
		return errors.Wrap(err, "failed to create destination directory")
	}

	// Create the file
	out, err := os.Create(dest)
	if err != nil {
		return errors.Wrap(err, "failed to create file")
	}
	defer out.Close()

	// Get the data
	resp, err := http.Get(url)
	if err != nil {
		return errors.Wrap(err, "failed to get URL")
	}
	defer resp.Body.Close()

	// Write the body to file
	_, err = io.Copy(out, resp.Body)
	if err != nil {
		return errors.Wrap(err, "failed to write data to file")
	}

	return nil
}
