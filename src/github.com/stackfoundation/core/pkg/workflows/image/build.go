package image

// From https://github.com/docker/cli/blob/3b8cf20a0c582de8f5e3022a3cbff4204cd6dfbd/cli/command/image/build.go
// Licensed under BSD 2-clause "Simplified" License

import "io"

func BuildImageStream(contextDir, dockerfilePath string, dockerfileContent io.Reader) (io.ReadCloser, string, error) {
        excludes, err := readDockerignore(contextDir)
        if err != nil {
                return nil, "", err
        }

        dockerfileTarEntry := ""
        _, contextDirRelativeDockerfilePath, err := getContextFromLocalDir(contextDir, dockerfilePath)
        if err == nil {
                dockerfileTarEntry, _ = canonicalTarNameForPath(contextDirRelativeDockerfilePath)
        }

        err = validateContextDirectory(contextDir, excludes)
        if err != nil {
                return nil, "", err
        }

        excludes = trimBuildFilesFromExcludes(excludes, dockerfileTarEntry, true)

        tarOptions := &TarOptions{
                ExcludePatterns: excludes,
                Compression:     Uncompressed,
        }

        buildContext, err := tarWithOptions(contextDir, tarOptions)

        if buildContext != nil {
                buildContext, dockerfileTarEntry, err = addDockerfileToBuildContext(dockerfileContent, buildContext)
                if err != nil {
                        return nil, "", err
                }
        }

        return buildContext, dockerfileTarEntry, err
}
