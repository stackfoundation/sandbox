package image

// From https://github.com/docker/cli/blob/3b8cf20a0c582de8f5e3022a3cbff4204cd6dfbd/cli/command/image/build.go
// Licensed under BSD 2-clause "Simplified" License

import (
	"io"
	"path/filepath"
)

// BuildOptions Options for building an image
type BuildOptions struct {
	ContextDirectory  string
	DockerfilePath    string
	Dockerignore      string
	ScriptName        string
	DockerfileContent io.Reader
	ScriptContent     io.Reader
}

// BuildImageStream Build the tar stream for the context to send to a docker build
func BuildImageStream(options *BuildOptions) (io.ReadCloser, string, error) {
	dockerignore := options.Dockerignore
	if len(dockerignore) < 1 {
		dockerignore = filepath.Join(options.ContextDirectory, ".dockerignore")
	}

	excludes, err := readDockerignore(dockerignore)
	if err != nil {
		return nil, "", err
	}

	dockerfileTarEntry := ""
	_, contextDirRelativeDockerfilePath, err :=
		getContextFromLocalDir(options.ContextDirectory, options.DockerfilePath)
	if err == nil {
		dockerfileTarEntry, _ = canonicalTarNameForPath(contextDirRelativeDockerfilePath)
	}

	err = validateContextDirectory(options.ContextDirectory, excludes)
	if err != nil {
		return nil, "", err
	}

	excludes = trimBuildFilesFromExcludes(excludes, dockerfileTarEntry, true)

	tarOptions := &TarOptions{
		ExcludePatterns: excludes,
		Compression:     Uncompressed,
	}

	buildContext, err := tarWithOptions(options.ContextDirectory, tarOptions)

	if buildContext != nil && options.DockerfileContent != nil {
		buildContext, dockerfileTarEntry, err =
			addBuildFilesToBuildContext(options.ScriptName, options.DockerfileContent, options.ScriptContent, buildContext)
		if err != nil {
			return nil, "", err
		}
	}

	return buildContext, dockerfileTarEntry, err
}
