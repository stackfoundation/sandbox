package metadata

import (
	"encoding/json"
	"os"
	"path/filepath"
	"time"

	"github.com/stackfoundation/install"
	"github.com/stackfoundation/io"
)

const cliMetadataFile = "cli.json"
const modificationWarning = "Used internally by the Sandbox CLI. Please do not modify this JSON file."

// Metadata CLI metadata
type Metadata struct {
	Description string    `json:"description"`
	Bootstrap   Version   `json:"bootstrap"`
	Core        Version   `json:"core"`
	Checked     time.Time `json:"checked"`
	Driver      string    `json:"driver"`
}

// Version Version of CLI component
type Version struct {
	Version  string `json:"version"`
	Updated  string `json:"updated"`
	Checksum string `json:"checksum"`
	Local    string `json:"local"`
	Remote   string `json:"remote"`
}

// WithVersions Create new metadata with the specified bootstrap and core component versions
func WithVersions(bootstrapVersion Version, coreVersion Version) *Metadata {
	return &Metadata{
		Description: modificationWarning,
		Bootstrap:   bootstrapVersion,
		Core:        coreVersion,
		Checked:     time.Now(),
	}
}

// GetMetadata Get the metadata from the standard location, if any exists
func GetMetadata() (*Metadata, error) {
	path, err := install.GetInstallPath()
	if err != nil {
		return nil, err
	}

	metadataPath := filepath.Join(path, cliMetadataFile)
	exists, err := io.Exists(metadataPath)
	if err != nil {
		return nil, err
	}

	if exists {
		return readMetadata(metadataPath)
	}

	return nil, nil
}

// NeedsUpdate Does the specified metadata need to be updated?
func (metadata *Metadata) NeedsUpdate() bool {
	return metadata == nil ||
		len(metadata.Bootstrap.Version) == 0 ||
		len(metadata.Core.Version) == 0 ||
		time.Since(metadata.Checked) > time.Duration(12)*time.Hour
}

func readMetadata(metadataPath string) (*Metadata, error) {
	metadataReader, err := os.Open(metadataPath)
	if err != nil {
		return nil, err
	}
	defer metadataReader.Close()

	var storedMetadata Metadata
	err = json.NewDecoder(metadataReader).Decode(&storedMetadata)
	if err != nil {
		return nil, err
	}

	return &storedMetadata, nil
}

// SaveMetadata Save the given metadata in the standard location
func SaveMetadata(metadata *Metadata) error {
	path, err := install.GetInstallPath()
	if err != nil {
		return err
	}

	metadataPath := filepath.Join(path, cliMetadataFile)
	metadataWriter, err := os.Create(metadataPath)
	if err != nil {
		return err
	}
	defer metadataWriter.Close()

	err = json.NewEncoder(metadataWriter).Encode(metadata)
	if err != nil {
		return err
	}

	return nil
}
