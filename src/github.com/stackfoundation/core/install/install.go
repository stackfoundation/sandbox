package bootstrap

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"time"
)

type version struct {
	Version   string            `json:"version"`
	Updated   string            `json:"updated"`
	Checksums map[string]string `json:"checksums"`
	Files     map[string]string `json:"files"`
}

//var PTransport RoundTripper = &http.Transport{Proxy: http.ProxyFromEnvironment}
//client := http.Client{Transport: PTransport}

func downloadFromUrl(url string, dest string) error {
	Debug(CoreDownloading, "Downloading %v to %v", url, dest)

	err := os.MkdirAll(filepath.Dir(dest), os.ModePerm)
	if err != nil {
		return err
	}

	output, err := os.Create(dest)
	if err != nil {
		return err
	}
	defer output.Close()

	response, err := http.Get(url)
	if err != nil {
		return err
	}
	defer response.Body.Close()

	_, err = io.Copy(output,
		NewProgressAwareReader(response.Body, "Downloading", CoreDownloading, response.ContentLength))
	if err != nil {
		return err
	}

	return nil
}

func getVersion(versionUri string) (version, error) {
	response, err := http.Get(versionUri)
	if err != nil {
		Debug(VersionCheckFailed, "Couldn't get version information at %v", versionUri)
		return version{}, err
	}
	defer response.Body.Close()

	var latestVersion version
	err = json.NewDecoder(response.Body).Decode(&latestVersion)
	if err != nil {
		Debug(VersionCheckFailed, "Version information retrieved from %v has errors", versionUri)
		return version{}, err
	}

	return latestVersion, nil
}

type metadata struct {
	Bootstrap version   `json:"bootstrap"`
	Core      version   `json:"core"`
	Checked   time.Time `json:"checked"`
}

func exists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return true, err
}

func readMetadata(metadataPath string) (*metadata, error) {
	metadataReader, err := os.Open(metadataPath)
	if err != nil {
		return nil, err
	}
	defer metadataReader.Close()

	var storedMetadata metadata
	err = json.NewDecoder(metadataReader).Decode(&storedMetadata)
	if err != nil {
		return nil, err
	}

	return &storedMetadata, nil
}

func getMetadata() (*metadata, error) {
	path, err := GetInstallPath()
	if err != nil {
		return nil, err
	}

	metadataPath := filepath.Join(path, CliMetadataFile)
	exists, err := exists(metadataPath)
	if err != nil {
		return nil, err
	}

	if exists {
		return readMetadata(metadataPath)
	} else {
		return nil, nil
	}
}

func saveMetadata(metadata *metadata) error {
	path, err := GetInstallPath()
	if err != nil {
		return err
	}

	metadataPath := filepath.Join(path, CliMetadataFile)
	metadataWriter, err := os.Create(metadataPath)
	if err != nil {
		return err
	}

	err = json.NewEncoder(metadataWriter).Encode(metadata)
	if err != nil {
		return err
	}

	return nil
}

func GetInstallPath() (string, error) {
	path, err := getStackFoundationRoot()
	if err != nil {
		return "", err
	}

	return filepath.Join(path, "cli"), nil
}

func ensureBinaryInstalled(binary string, version version) (bool, string, error) {
	installPath, err := GetInstallPath()
	if err != nil {
		installPath = ""
	}

	extension := ""
	if runtime.GOOS == "windows" {
		extension = ".exe"
	}

	relativeBinaryPath := fmt.Sprintf(binary, version.Version, extension)
	fullBinaryPath := filepath.Join(installPath, relativeBinaryPath)
	exists, err := exists(fullBinaryPath)
	if !exists || err != nil {
		Debug(CoreDownloadRequired, "%v does not exist and must be downloaded", fullBinaryPath)

		platformKey := fmt.Sprintf("%v-%v", runtime.GOOS, runtime.GOARCH)
		if file, exists := version.Files[platformKey]; exists {
			tar := fmt.Sprintf("%v.tar.gz", fullBinaryPath)
			err := downloadFromUrl(BinariesBucket+file, tar)
			if err != nil {
				return false, "", err
			}

			extractFile(tar, fullBinaryPath)
		}

		return true, fullBinaryPath, nil
	}

	return false, fullBinaryPath, nil
}

func EnsureCoreInstalled() (newInstallation bool, installPath string, errorCode error) {
	storedMetadata, err := getMetadata()
	if err != nil {
		return false, "", err
	}

	if storedMetadata == nil || time.Since(storedMetadata.Checked) > time.Duration(12)*time.Hour {
		Debug(CoreVersionCheck, "Checking for latest version of CLI")

		bootstrapVersion, _ := getVersion(BinariesBucket + BootstrapVersion)
		coreVersion, _ := getVersion(BinariesBucket + CoreVersion)

		storedMetadata = &metadata{
			Bootstrap: bootstrapVersion,
			Core:      coreVersion,
			Checked:   time.Now(),
		}

		Debug(VersionCheckFinished, "Latest version of CLI bootstrap is %v and latest core is %v",
			bootstrapVersion.Version, coreVersion.Version)

		saveMetadata(storedMetadata)
	}

	if storedMetadata != nil {
		if len(storedMetadata.Bootstrap.Version) > 0 && storedMetadata.Bootstrap.Version != CliVersion {
			return ensureBinaryInstalled(BootstrapBinary, storedMetadata.Bootstrap)
		}

		return ensureBinaryInstalled(CoreBinary, storedMetadata.Core)
	}

	return false, "", nil
}
