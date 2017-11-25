package image

// From https://github.com/docker/cli/blob/3b8cf20a0c582de8f5e3022a3cbff4204cd6dfbd/cli/command/image/build/context.go
// Licensed under Apache License Version 2.0

import (
	"archive/tar"
	"bytes"
	"crypto/md5"
	"encoding/hex"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/pkg/errors"
)

const DefaultDockerfileName string = "Dockerfile"

// ValidateContextDirectory checks if all the contents of the directory
// can be read and returns an error if some files can't be read
// symlinks which point to non-existing files don't trigger an error
func validateContextDirectory(srcPath string, excludes []string) error {
	contextRoot, err := getContextRoot(srcPath)
	if err != nil {
		return err
	}
	return filepath.Walk(contextRoot, func(filePath string, f os.FileInfo, err error) error {
		if err != nil {
			if os.IsPermission(err) {
				return errors.Errorf("can't stat '%s'", filePath)
			}
			if os.IsNotExist(err) {
				return nil
			}
			return err
		}

		// skip this directory/file if it's not in the path, it won't get added to the context
		if relFilePath, err := filepath.Rel(contextRoot, filePath); err != nil {
			return err
		} else if skip, err := Matches(relFilePath, excludes); err != nil {
			return err
		} else if skip {
			if f.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}

		// skip checking if symlinks point to non-existing files, such symlinks can be useful
		// also skip named pipes, because they hanging on open
		if f.Mode()&(os.ModeSymlink|os.ModeNamedPipe) != 0 {
			return nil
		}

		if !f.IsDir() {
			currentFile, err := os.Open(filePath)
			if err != nil && os.IsPermission(err) {
				return errors.Errorf("no permission to read from '%s'", filePath)
			}
			currentFile.Close()
		}
		return nil
	})
}

// GetContextFromLocalDir uses the given local directory as context for a
// `docker build`. Returns the absolute path to the local context directory,
// the relative path of the dockerfile in that context directory, and a non-nil
// error on success.
func getContextFromLocalDir(localDir, dockerfileName string) (string, string, error) {
	localDir, err := resolveAndValidateContextPath(localDir)
	if err != nil {
		return "", "", err
	}

	// When using a local context directory, and the Dockerfile is specified
	// with the `-f/--file` option then it is considered relative to the
	// current directory and not the context directory.
	if dockerfileName != "" && dockerfileName != "-" {
		if dockerfileName, err = filepath.Abs(dockerfileName); err != nil {
			return "", "", errors.Errorf("unable to get absolute path to Dockerfile: %v", err)
		}
	}

	relDockerfile, err := getDockerfileRelPath(localDir, dockerfileName)
	return localDir, relDockerfile, err
}

// ResolveAndValidateContextPath uses the given context directory for a `docker build`
// and returns the absolute path to the context directory.
func resolveAndValidateContextPath(givenContextDir string) (string, error) {
	absContextDir, err := filepath.Abs(givenContextDir)
	if err != nil {
		return "", errors.Errorf("unable to get absolute context directory of given context directory %q: %v", givenContextDir, err)
	}

	// The context dir might be a symbolic link, so follow it to the actual
	// target directory.
	//
	// FIXME. We use isUNC (always false on non-Windows platforms) to workaround
	// an issue in golang. On Windows, EvalSymLinks does not work on UNC file
	// paths (those starting with \\). This hack means that when using links
	// on UNC paths, they will not be followed.
	if !isUNC(absContextDir) {
		absContextDir, err = filepath.EvalSymlinks(absContextDir)
		if err != nil {
			return "", errors.Errorf("unable to evaluate symlinks in context path: %v", err)
		}
	}

	stat, err := os.Lstat(absContextDir)
	if err != nil {
		return "", errors.Errorf("unable to stat context directory %q: %v", absContextDir, err)
	}

	if !stat.IsDir() {
		return "", errors.Errorf("context must be a directory: %s", absContextDir)
	}
	return absContextDir, err
}

// getDockerfileRelPath returns the dockerfile path relative to the context
// directory
func getDockerfileRelPath(absContextDir, givenDockerfile string) (string, error) {
	var err error

	if givenDockerfile == "-" {
		return givenDockerfile, nil
	}

	absDockerfile := givenDockerfile
	if absDockerfile == "" {
		// No -f/--file was specified so use the default relative to the
		// context directory.
		absDockerfile = filepath.Join(absContextDir, DefaultDockerfileName)

		// Just to be nice ;-) look for 'dockerfile' too but only
		// use it if we found it, otherwise ignore this check
		if _, err = os.Lstat(absDockerfile); os.IsNotExist(err) {
			altPath := filepath.Join(absContextDir, strings.ToLower(DefaultDockerfileName))
			if _, err = os.Lstat(altPath); err == nil {
				absDockerfile = altPath
			}
		}
	}

	// If not already an absolute path, the Dockerfile path should be joined to
	// the base directory.
	if !filepath.IsAbs(absDockerfile) {
		absDockerfile = filepath.Join(absContextDir, absDockerfile)
	}

	// Evaluate symlinks in the path to the Dockerfile too.
	//
	// FIXME. We use isUNC (always false on non-Windows platforms) to workaround
	// an issue in golang. On Windows, EvalSymLinks does not work on UNC file
	// paths (those starting with \\). This hack means that when using links
	// on UNC paths, they will not be followed.
	if !isUNC(absDockerfile) {
		absDockerfile, err = filepath.EvalSymlinks(absDockerfile)
		if err != nil {
			return "", errors.Errorf("unable to evaluate symlinks in Dockerfile path: %v", err)

		}
	}

	if _, err := os.Lstat(absDockerfile); err != nil {
		if os.IsNotExist(err) {
			return "", errors.Errorf("Cannot locate Dockerfile: %q", absDockerfile)
		}
		return "", errors.Errorf("unable to stat Dockerfile: %v", err)
	}

	relDockerfile, err := filepath.Rel(absContextDir, absDockerfile)
	if err != nil {
		return "", errors.Errorf("unable to get relative Dockerfile path: %v", err)
	}

	if strings.HasPrefix(relDockerfile, ".."+string(filepath.Separator)) {
		return "", errors.Errorf("the Dockerfile (%s) must be within the build context", givenDockerfile)
	}

	return relDockerfile, nil
}

// isUNC returns true if the path is UNC (one starting \\). It always returns
// false on Linux.
func isUNC(path string) bool {
	return runtime.GOOS == "windows" && strings.HasPrefix(path, `\\`)
}

func generateDockerfileName(content []byte) string {
	hash := md5.New()
	hash.Write(content)
	return ".dockerfile." + hex.EncodeToString(hash.Sum(nil))
}

func addBuildFilesToBuildContext(
	scriptName string,
	dockerfileContent, scriptContent io.Reader,
	buildContext io.ReadCloser) (io.ReadCloser, string, error) {

	dockerfile, err := ioutil.ReadAll(dockerfileContent)
	if err != nil {
		return nil, "", err
	}

	script, err := ioutil.ReadAll(scriptContent)
	if err != nil {
		return nil, "", err
	}

	now := time.Now()

	dockerfileHeader := &tar.Header{
		Mode:       0600,
		Uid:        0,
		Gid:        0,
		ModTime:    now,
		Typeflag:   tar.TypeReg,
		AccessTime: now,
		ChangeTime: now,
	}

	scriptHeader := &tar.Header{
		Mode:       0600,
		Uid:        0,
		Gid:        0,
		ModTime:    now,
		Typeflag:   tar.TypeReg,
		AccessTime: now,
		ChangeTime: now,
	}

	dockerfileName := generateDockerfileName(dockerfile)

	buildContext = replaceFileTarWrapper(buildContext, map[string]TarModifierFunc{
		// Add the dockerfile with a random filename
		dockerfileName: func(_ string, h *tar.Header, content io.Reader) (*tar.Header, []byte, error) {
			return dockerfileHeader, dockerfile, nil
		},
		// Add the script with a random filename
		scriptName: func(_ string, h *tar.Header, content io.Reader) (*tar.Header, []byte, error) {
			return scriptHeader, script, nil
		},
		// Update .dockerignore to include the random filename
		".dockerignore": func(_ string, h *tar.Header, content io.Reader) (*tar.Header, []byte, error) {
			if h == nil {
				h = dockerfileHeader
			}

			b := &bytes.Buffer{}
			if content != nil {
				if _, err := b.ReadFrom(content); err != nil {
					return nil, nil, err
				}
			} else {
				b.WriteString(".dockerignore")
			}
			b.WriteString("\n" + dockerfileName + "\n")
			return h, b.Bytes(), nil
		},
	})

	return buildContext, dockerfileName, nil
}
