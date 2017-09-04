package bootstrap

import (
	"archive/tar"
	"compress/gzip"
	"io"
	"os"
)

func extractFile(filePath string, destPath string) error {
	Debug(CoreExtracting, "Extracting %v to %v", filePath, destPath)
	info, err := os.Stat(filePath)
	if err != nil {
		return err
	}

	file, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	gz, err := gzip.NewReader(NewProgressAwareReader(file, "Extracting", CoreExtracting, info.Size()))
	if err != nil {
		return err
	}

	fileWritten := false
	tarReader := tar.NewReader(gz)
	for true {
		entry, err := tarReader.Next()

		if err == io.EOF {
			break
		}

		if err != nil {
			return err
		}

		name := entry.Name

		switch entry.Typeflag {
		case tar.TypeDir:
			os.Mkdir(name, 0755)
		case tar.TypeReg:
			if fileWritten {

				entryFile, err := os.OpenFile(name, os.O_RDWR|os.O_CREATE, 0755)
				if err != nil {
					return err
				}

				io.Copy(entryFile, tarReader)
				entryFile.Close()
			} else {
				entryFile, err := os.OpenFile(destPath, os.O_RDWR|os.O_CREATE, 0755)
				if err != nil {
					return err
				}

				io.Copy(entryFile, tarReader)
				entryFile.Close()
				fileWritten = true
			}
		default:
			Debug(ProblemExtractingFile, "Unknown type: %c in file %s\n", entry.Typeflag, name)
		}
	}

	return nil
}
