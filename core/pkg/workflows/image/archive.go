package image

// Modified from https://github.com/moby/moby/blob/1009e6a40b295187e038b67e184e9c0384d95538/pkg/archive/archive.go
// Licensed under the Apache License Version 2.0

import (
        "archive/tar"
        "bufio"
        "bytes"
        "compress/gzip"
        "fmt"
        "io"
        "io/ioutil"
        "os"
        "os/exec"
        "path/filepath"
        "runtime"
        "strings"
        "syscall"

        "github.com/docker/docker/pkg/system"
        "github.com/Sirupsen/logrus"
)

type (
        // Compression is the state represents if compressed or not.
        Compression int
        // WhiteoutFormat is the format of whiteouts unpacked
        WhiteoutFormat int

        // TarOptions wraps the tar options.
        TarOptions struct {
                IncludeFiles         []string
                ExcludePatterns      []string
                Compression          Compression
                NoLchown             bool
                UIDMaps              []IDMap
                GIDMaps              []IDMap
                ChownOpts            *IDPair
                IncludeSourceDir     bool
                // WhiteoutFormat is the expected on disk format for whiteout files.
                // This format will be converted to the standard format on pack
                // and from the standard format on unpack.
                WhiteoutFormat       WhiteoutFormat
                // When unpacking, specifies whether overwriting a directory with a
                // non-directory is allowed and vice versa.
                NoOverwriteDirNonDir bool
                // For each include when creating an archive, the included name will be
                // replaced with the matching name from this map.
                RebaseNames          map[string]string
                InUserNS             bool
        }
)

// Archiver allows the reuse of most utility functions of this package
// with a pluggable Untar function. Also, to facilitate the passing of
// specific id mappings for untar, an archiver can be created with maps
// which will then be passed to Untar operations
type Archiver struct {
        Untar      func(io.Reader, string, *TarOptions) error
        IDMappings *IDMappings
}

// breakoutError is used to differentiate errors related to breaking out
// When testing archive breakout in the unit tests, this error is expected
// in order for the test to pass.
type breakoutError error

const (
        // Uncompressed represents the uncompressed.
        Uncompressed Compression = iota
        // Bzip2 is bzip2 compression algorithm.
        Bzip2
        // Gzip is gzip compression algorithm.
        Gzip
        // Xz is xz compression algorithm.
        Xz
)

const (
        // AUFSWhiteoutFormat is the default format for whiteouts
        AUFSWhiteoutFormat WhiteoutFormat = iota
        // OverlayWhiteoutFormat formats whiteout according to the overlay
        // standard.
        OverlayWhiteoutFormat
)

const (
        modeISDIR = 040000  // Directory
        modeISFIFO = 010000  // FIFO
        modeISREG = 0100000 // Regular file
        modeISLNK = 0120000 // Symbolic link
        modeISBLK = 060000  // Block special file
        modeISCHR = 020000  // Character special file
        modeISSOCK = 0140000 // Socket
)

// CompressStream compresses the dest with specified compression algorithm.
func compressStream(dest io.Writer, compression Compression) (io.WriteCloser, error) {
        p := BufioWriter32KPool
        buf := p.Get(dest)
        switch compression {
        case Uncompressed:
                writeBufWrapper := p.NewWriteCloserWrapper(buf, buf)
                return writeBufWrapper, nil
        case Gzip:
                gzWriter := gzip.NewWriter(dest)
                writeBufWrapper := p.NewWriteCloserWrapper(buf, gzWriter)
                return writeBufWrapper, nil
        case Bzip2, Xz:
                // archive/bzip2 does not support writing, and there is no xz support at all
                // However, this is not a problem as docker only currently generates gzipped tars
                return nil, fmt.Errorf("Unsupported compression format %s", (&compression).Extension())
        default:
                return nil, fmt.Errorf("Unsupported compression format %s", (&compression).Extension())
        }
}

// TarModifierFunc is a function that can be passed to ReplaceFileTarWrapper to
// modify the contents or header of an entry in the archive. If the file already
// exists in the archive the TarModifierFunc will be called with the Header and
// a reader which will return the files content. If the file does not exist both
// header and content will be nil.
type TarModifierFunc func(path string, header *tar.Header, content io.Reader) (*tar.Header, []byte, error)

// ReplaceFileTarWrapper converts inputTarStream to a new tar stream. Files in the
// tar stream are modified if they match any of the keys in mods.
func replaceFileTarWrapper(inputTarStream io.ReadCloser, mods map[string]TarModifierFunc) io.ReadCloser {
        pipeReader, pipeWriter := io.Pipe()

        go func() {
                tarReader := tar.NewReader(inputTarStream)
                tarWriter := tar.NewWriter(pipeWriter)
                defer inputTarStream.Close()
                defer tarWriter.Close()

                modify := func(name string, original *tar.Header, modifier TarModifierFunc, tarReader io.Reader) error {
                        header, data, err := modifier(name, original, tarReader)
                        switch {
                        case err != nil:
                                return err
                        case header == nil:
                                return nil
                        }

                        header.Name = name
                        header.Size = int64(len(data))
                        if err := tarWriter.WriteHeader(header); err != nil {
                                return err
                        }
                        if len(data) != 0 {
                                if _, err := tarWriter.Write(data); err != nil {
                                        return err
                                }
                        }
                        return nil
                }

                var err error
                var originalHeader *tar.Header
                for {
                        originalHeader, err = tarReader.Next()
                        if err == io.EOF {
                                break
                        }
                        if err != nil {
                                pipeWriter.CloseWithError(err)
                                return
                        }

                        modifier, ok := mods[originalHeader.Name]
                        if !ok {
                                // No modifiers for this file, copy the header and data
                                if err := tarWriter.WriteHeader(originalHeader); err != nil {
                                        pipeWriter.CloseWithError(err)
                                        return
                                }
                                if _, err := Copy(tarWriter, tarReader); err != nil {
                                        pipeWriter.CloseWithError(err)
                                        return
                                }
                                continue
                        }
                        delete(mods, originalHeader.Name)

                        if err := modify(originalHeader.Name, originalHeader, modifier, tarReader); err != nil {
                                pipeWriter.CloseWithError(err)
                                return
                        }
                }

                // Apply the modifiers that haven't matched any files in the archive
                for name, modifier := range mods {
                        if err := modify(name, nil, modifier, nil); err != nil {
                                pipeWriter.CloseWithError(err)
                                return
                        }
                }

                pipeWriter.Close()

        }()
        return pipeReader
}

// Extension returns the extension of a file that uses the specified compression algorithm.
func (compression *Compression) Extension() string {
        switch *compression {
        case Uncompressed:
                return "tar"
        case Bzip2:
                return "tar.bz2"
        case Gzip:
                return "tar.gz"
        case Xz:
                return "tar.xz"
        }
        return ""
}

// FileInfoHeader creates a populated Header from fi.
// Compared to archive pkg this function fills in more information.
// Also, regardless of Go version, this function fills file type bits (e.g. hdr.Mode |= modeISDIR),
// which have been deleted since Go 1.9 archive/tar.
func FileInfoHeader(name string, fi os.FileInfo, link string) (*tar.Header, error) {
        hdr, err := tar.FileInfoHeader(fi, link)
        if err != nil {
                return nil, err
        }
        hdr.Mode = fillGo18FileTypeBits(int64(chmodTarEntry(os.FileMode(hdr.Mode))), fi)
        name, err = canonicalTarName(name, fi.IsDir())
        if err != nil {
                return nil, fmt.Errorf("tar: cannot canonicalize path: %v", err)
        }
        hdr.Name = name
        if err := setHeaderForSpecialDevice(hdr, name, fi.Sys()); err != nil {
                return nil, err
        }
        return hdr, nil
}

// fillGo18FileTypeBits fills type bits which have been removed on Go 1.9 archive/tar
// https://github.com/golang/go/commit/66b5a2f
func fillGo18FileTypeBits(mode int64, fi os.FileInfo) int64 {
        fm := fi.Mode()
        switch {
        case fm.IsRegular():
                mode |= modeISREG
        case fi.IsDir():
                mode |= modeISDIR
        case fm & os.ModeSymlink != 0:
                mode |= modeISLNK
        case fm & os.ModeDevice != 0:
                if fm & os.ModeCharDevice != 0 {
                        mode |= modeISCHR
                } else {
                        mode |= modeISBLK
                }
        case fm & os.ModeNamedPipe != 0:
                mode |= modeISFIFO
        case fm & os.ModeSocket != 0:
                mode |= modeISSOCK
        }
        return mode
}

// ReadSecurityXattrToTarHeader reads security.capability xattr from filesystem
// to a tar header
func ReadSecurityXattrToTarHeader(path string, hdr *tar.Header) error {
        capability, _ := system.Lgetxattr(path, "security.capability")
        if capability != nil {
                hdr.Xattrs = make(map[string]string)
                hdr.Xattrs["security.capability"] = string(capability)
        }
        return nil
}

type tarWhiteoutConverter interface {
        ConvertWrite(*tar.Header, string, os.FileInfo) (*tar.Header, error)
        ConvertRead(*tar.Header, string) (bool, error)
}

type tarAppender struct {
        TarWriter         *tar.Writer
        Buffer            *bufio.Writer

        // for hardlink mapping
        SeenFiles         map[uint64]string
        IDMappings        *IDMappings

        // For packing and unpacking whiteout files in the
        // non standard format. The whiteout files defined
        // by the AUFS standard are used as the tar whiteout
        // standard.
        WhiteoutConverter tarWhiteoutConverter
}

func newTarAppender(idMapping *IDMappings, writer io.Writer) *tarAppender {
        return &tarAppender{
                SeenFiles:  make(map[uint64]string),
                TarWriter:  tar.NewWriter(writer),
                Buffer:     BufioWriter32KPool.Get(nil),
                IDMappings: idMapping,
        }
}

// canonicalTarName provides a platform-independent and consistent posix-style
//path for files and directories to be archived regardless of the platform.
func canonicalTarName(name string, isDir bool) (string, error) {
        name, err := canonicalTarNameForPath(name)
        if err != nil {
                return "", err
        }

        // suffix with '/' for directories
        if isDir && !strings.HasSuffix(name, "/") {
                name += "/"
        }
        return name, nil
}

// addTarFile adds to the tar archive a file from `path` as `name`
func (ta *tarAppender) addTarFile(path, name string) error {
        fi, err := os.Lstat(path)
        if err != nil {
                return err
        }

        var link string
        if fi.Mode() & os.ModeSymlink != 0 {
                var err error
                link, err = os.Readlink(path)
                if err != nil {
                        return err
                }
        }

        hdr, err := FileInfoHeader(name, fi, link)
        if err != nil {
                return err
        }
        if err := ReadSecurityXattrToTarHeader(path, hdr); err != nil {
                return err
        }

        // if it's not a directory and has more than 1 link,
        // it's hard linked, so set the type flag accordingly
        if !fi.IsDir() && hasHardlinks(fi) {
                inode, err := getInodeFromStat(fi.Sys())
                if err != nil {
                        return err
                }
                // a link should have a name that it links too
                // and that linked name should be first in the tar archive
                if oldpath, ok := ta.SeenFiles[inode]; ok {
                        hdr.Typeflag = tar.TypeLink
                        hdr.Linkname = oldpath
                        hdr.Size = 0 // This Must be here for the writer math to add up!
                } else {
                        ta.SeenFiles[inode] = name
                }
        }

        //handle re-mapping container ID mappings back to host ID mappings before
        //writing tar headers/files. We skip whiteout files because they were written
        //by the kernel and already have proper ownership relative to the host
        if !strings.HasPrefix(filepath.Base(hdr.Name), WhiteoutPrefix) && !ta.IDMappings.Empty() {
                fileIDPair, err := getFileUIDGID(fi.Sys())
                if err != nil {
                        return err
                }
                hdr.Uid, hdr.Gid, err = ta.IDMappings.ToContainer(fileIDPair)
                if err != nil {
                        return err
                }
        }

        if ta.WhiteoutConverter != nil {
                wo, err := ta.WhiteoutConverter.ConvertWrite(hdr, path, fi)
                if err != nil {
                        return err
                }

                // If a new whiteout file exists, write original hdr, then
                // replace hdr with wo to be written after. Whiteouts should
                // always be written after the original. Note the original
                // hdr may have been updated to be a whiteout with returning
                // a whiteout header
                if wo != nil {
                        if err := ta.TarWriter.WriteHeader(hdr); err != nil {
                                return err
                        }
                        if hdr.Typeflag == tar.TypeReg && hdr.Size > 0 {
                                return fmt.Errorf("tar: cannot use whiteout for non-empty file")
                        }
                        hdr = wo
                }
        }

        if err := ta.TarWriter.WriteHeader(hdr); err != nil {
                return err
        }

        if hdr.Typeflag == tar.TypeReg && hdr.Size > 0 {
                // We use system.OpenSequential to ensure we use sequential file
                // access on Windows to avoid depleting the standby list.
                // On Linux, this equates to a regular os.Open.
                file, err := OpenSequential(path)
                if err != nil {
                        return err
                }

                ta.Buffer.Reset(ta.TarWriter)
                defer ta.Buffer.Reset(nil)
                _, err = io.Copy(ta.Buffer, file)
                file.Close()
                if err != nil {
                        return err
                }
                err = ta.Buffer.Flush()
                if err != nil {
                        return err
                }
        }

        return nil
}

func createTarFile(path, extractDir string, hdr *tar.Header, reader io.Reader, Lchown bool, chownOpts *IDPair, inUserns bool) error {
        // hdr.Mode is in linux format, which we can use for sycalls,
        // but for os.Foo() calls we need the mode converted to os.FileMode,
        // so use hdrInfo.Mode() (they differ for e.g. setuid bits)
        hdrInfo := hdr.FileInfo()

        switch hdr.Typeflag {
        case tar.TypeDir:
                // Create directory unless it exists as a directory already.
                // In that case we just want to merge the two
                if fi, err := os.Lstat(path); !(err == nil && fi.IsDir()) {
                        if err := os.Mkdir(path, hdrInfo.Mode()); err != nil {
                                return err
                        }
                }

        case tar.TypeReg, tar.TypeRegA:
                // Source is regular file. We use system.OpenFileSequential to use sequential
                // file access to avoid depleting the standby list on Windows.
                // On Linux, this equates to a regular os.OpenFile
                file, err := OpenFileSequential(path, os.O_CREATE | os.O_WRONLY, hdrInfo.Mode())
                if err != nil {
                        return err
                }
                if _, err := io.Copy(file, reader); err != nil {
                        file.Close()
                        return err
                }
                file.Close()

        case tar.TypeBlock, tar.TypeChar:
                if inUserns {
                        // cannot create devices in a userns
                        return nil
                }
                // Handle this is an OS-specific way
                if err := handleTarTypeBlockCharFifo(hdr, path); err != nil {
                        return err
                }

        case tar.TypeFifo:
                // Handle this is an OS-specific way
                if err := handleTarTypeBlockCharFifo(hdr, path); err != nil {
                        return err
                }

        case tar.TypeLink:
                targetPath := filepath.Join(extractDir, hdr.Linkname)
                // check for hardlink breakout
                if !strings.HasPrefix(targetPath, extractDir) {
                        return breakoutError(fmt.Errorf("invalid hardlink %q -> %q", targetPath, hdr.Linkname))
                }
                if err := os.Link(targetPath, path); err != nil {
                        return err
                }

        case tar.TypeSymlink:
                // 	path 				-> hdr.Linkname = targetPath
                // e.g. /extractDir/path/to/symlink 	-> ../2/file	= /extractDir/path/2/file
                targetPath := filepath.Join(filepath.Dir(path), hdr.Linkname)

                // the reason we don't need to check symlinks in the path (with FollowSymlinkInScope) is because
                // that symlink would first have to be created, which would be caught earlier, at this very check:
                if !strings.HasPrefix(targetPath, extractDir) {
                        return breakoutError(fmt.Errorf("invalid symlink %q -> %q", path, hdr.Linkname))
                }
                if err := os.Symlink(hdr.Linkname, path); err != nil {
                        return err
                }

        case tar.TypeXGlobalHeader:
                logrus.Debug("PAX Global Extended Headers found and ignored")
                return nil

        default:
                return fmt.Errorf("Unhandled tar header type %d\n", hdr.Typeflag)
        }

        // Lchown is not supported on Windows.
        if Lchown && runtime.GOOS != "windows" {
                if chownOpts == nil {
                        chownOpts = &IDPair{UID: hdr.Uid, GID: hdr.Gid}
                }
                if err := os.Lchown(path, chownOpts.UID, chownOpts.GID); err != nil {
                        return err
                }
        }

        var errors []string
        for key, value := range hdr.Xattrs {
                if err := system.Lsetxattr(path, key, []byte(value), 0); err != nil {
                        if err == syscall.ENOTSUP {
                                // We ignore errors here because not all graphdrivers support
                                // xattrs *cough* old versions of AUFS *cough*. However only
                                // ENOTSUP should be emitted in that case, otherwise we still
                                // bail.
                                errors = append(errors, err.Error())
                                continue
                        }
                        return err
                }

        }

        if len(errors) > 0 {
                logrus.WithFields(logrus.Fields{
                        "errors": errors,
                }).Warn("ignored xattrs in archive: underlying filesystem doesn't support them")
        }

        // There is no LChmod, so ignore mode for symlink. Also, this
        // must happen after chown, as that can modify the file mode
        if err := handleLChmod(hdr, path, hdrInfo); err != nil {
                return err
        }

        aTime := hdr.AccessTime
        if aTime.Before(hdr.ModTime) {
                // Last access time should never be before last modified time.
                aTime = hdr.ModTime
        }

        // system.Chtimes doesn't support a NOFOLLOW flag atm
        if hdr.Typeflag == tar.TypeLink {
                if fi, err := os.Lstat(hdr.Linkname); err == nil && (fi.Mode() & os.ModeSymlink == 0) {
                        if err := system.Chtimes(path, aTime, hdr.ModTime); err != nil {
                                return err
                        }
                }
        } else if hdr.Typeflag != tar.TypeSymlink {
                if err := system.Chtimes(path, aTime, hdr.ModTime); err != nil {
                        return err
                }
        } else {
                ts := []syscall.Timespec{timeToTimespec(aTime), timeToTimespec(hdr.ModTime)}
                if err := system.LUtimesNano(path, ts); err != nil && err != system.ErrNotSupportedPlatform {
                        return err
                }
        }
        return nil
}

// TarWithOptions creates an archive from the directory at `path`, only including files whose relative
// paths are included in `options.IncludeFiles` (if non-nil) or not in `options.ExcludePatterns`.
func tarWithOptions(srcPath string, options *TarOptions) (io.ReadCloser, error) {

        // Fix the source path to work with long path names. This is a no-op
        // on platforms other than Windows.
        srcPath = fixVolumePathPrefix(srcPath)

        pm, err := newPatternMatcher(options.ExcludePatterns)
        if err != nil {
                return nil, err
        }

        pipeReader, pipeWriter := io.Pipe()

        compressWriter, err := compressStream(pipeWriter, options.Compression)
        if err != nil {
                return nil, err
        }

        go func() {
                ta := newTarAppender(
                        NewIDMappingsFromMaps(options.UIDMaps, options.GIDMaps),
                        compressWriter,
                )
                ta.WhiteoutConverter = getWhiteoutConverter(options.WhiteoutFormat)

                defer func() {
                        // Make sure to check the error on Close.
                        if err := ta.TarWriter.Close(); err != nil {
                                logrus.Errorf("Can't close tar writer: %s", err)
                        }
                        if err := compressWriter.Close(); err != nil {
                                logrus.Errorf("Can't close compress writer: %s", err)
                        }
                        if err := pipeWriter.Close(); err != nil {
                                logrus.Errorf("Can't close pipe writer: %s", err)
                        }
                }()

                // this buffer is needed for the duration of this piped stream
                defer BufioWriter32KPool.Put(ta.Buffer)

                // In general we log errors here but ignore them because
                // during e.g. a diff operation the container can continue
                // mutating the filesystem and we can see transient errors
                // from this

                stat, err := os.Lstat(srcPath)
                if err != nil {
                        return
                }

                if !stat.IsDir() {
                        // We can't later join a non-dir with any includes because the
                        // 'walk' will error if "file/." is stat-ed and "file" is not a
                        // directory. So, we must split the source path and use the
                        // basename as the include.
                        if len(options.IncludeFiles) > 0 {
                                logrus.Warn("Tar: Can't archive a file with includes")
                        }

                        dir, base := splitPathDirEntry(srcPath)
                        srcPath = dir
                        options.IncludeFiles = []string{base}
                }

                if len(options.IncludeFiles) == 0 {
                        options.IncludeFiles = []string{"."}
                }

                seen := make(map[string]bool)

                for _, include := range options.IncludeFiles {
                        rebaseName := options.RebaseNames[include]

                        walkRoot := getWalkRoot(srcPath, include)
                        filepath.Walk(walkRoot, func(filePath string, f os.FileInfo, err error) error {
                                if err != nil {
                                        logrus.Errorf("Tar: Can't stat file %s to tar: %s", srcPath, err)
                                        return nil
                                }

                                relFilePath, err := filepath.Rel(srcPath, filePath)
                                if err != nil || (!options.IncludeSourceDir && relFilePath == "." && f.IsDir()) {
                                        // Error getting relative path OR we are looking
                                        // at the source directory path. Skip in both situations.
                                        return nil
                                }

                                if options.IncludeSourceDir && include == "." && relFilePath != "." {
                                        relFilePath = strings.Join([]string{".", relFilePath}, string(filepath.Separator))
                                }

                                skip := false

                                // If "include" is an exact match for the current file
                                // then even if there's an "excludePatterns" pattern that
                                // matches it, don't skip it. IOW, assume an explicit 'include'
                                // is asking for that file no matter what - which is true
                                // for some files, like .dockerignore and Dockerfile (sometimes)
                                if include != relFilePath {
                                        skip, err = pm.Matches(relFilePath)
                                        if err != nil {
                                                logrus.Errorf("Error matching %s: %v", relFilePath, err)
                                                return err
                                        }
                                }

                                if skip {
                                        // If we want to skip this file and its a directory
                                        // then we should first check to see if there's an
                                        // excludes pattern (e.g. !dir/file) that starts with this
                                        // dir. If so then we can't skip this dir.

                                        // Its not a dir then so we can just return/skip.
                                        if !f.IsDir() {
                                                return nil
                                        }

                                        // No exceptions (!...) in patterns so just skip dir
                                        if !pm.Exclusions() {
                                                return filepath.SkipDir
                                        }

                                        dirSlash := relFilePath + string(filepath.Separator)

                                        for _, pat := range pm.Patterns() {
                                                if !pat.Exclusion() {
                                                        continue
                                                }
                                                if strings.HasPrefix(pat.String() + string(filepath.Separator), dirSlash) {
                                                        // found a match - so can't skip this dir
                                                        return nil
                                                }
                                        }

                                        // No matching exclusion dir so just skip dir
                                        return filepath.SkipDir
                                }

                                if seen[relFilePath] {
                                        return nil
                                }
                                seen[relFilePath] = true

                                // Rename the base resource.
                                if rebaseName != "" {
                                        var replacement string
                                        if rebaseName != string(filepath.Separator) {
                                                // Special case the root directory to replace with an
                                                // empty string instead so that we don't end up with
                                                // double slashes in the paths.
                                                replacement = rebaseName
                                        }

                                        relFilePath = strings.Replace(relFilePath, include, replacement, 1)
                                }

                                if err := ta.addTarFile(filePath, relFilePath); err != nil {
                                        logrus.Errorf("Can't add file %s to tar: %s", filePath, err)
                                        // if pipe is broken, stop writing tar stream to it
                                        if err == io.ErrClosedPipe {
                                                return err
                                        }
                                }
                                return nil
                        })
                }
        }()

        return pipeReader, nil
}

// cmdStream executes a command, and returns its stdout as a stream.
// If the command fails to run or doesn't complete successfully, an error
// will be returned, including anything written on stderr.
func cmdStream(cmd *exec.Cmd, input io.Reader) (io.ReadCloser, <-chan struct{}, error) {
        chdone := make(chan struct{})
        cmd.Stdin = input
        pipeR, pipeW := io.Pipe()
        cmd.Stdout = pipeW
        var errBuf bytes.Buffer
        cmd.Stderr = &errBuf

        // Run the command and return the pipe
        if err := cmd.Start(); err != nil {
                return nil, nil, err
        }

        // Copy stdout to the returned pipe
        go func() {
                if err := cmd.Wait(); err != nil {
                        pipeW.CloseWithError(fmt.Errorf("%s: %s", err, errBuf.String()))
                } else {
                        pipeW.Close()
                }
                close(chdone)
        }()

        return pipeR, chdone, nil
}

// NewTempArchive reads the content of src into a temporary file, and returns the contents
// of that file as an archive. The archive can only be read once - as soon as reading completes,
// the file will be deleted.
func NewTempArchive(src io.Reader, dir string) (*TempArchive, error) {
        f, err := ioutil.TempFile(dir, "")
        if err != nil {
                return nil, err
        }
        if _, err := io.Copy(f, src); err != nil {
                return nil, err
        }
        if _, err := f.Seek(0, 0); err != nil {
                return nil, err
        }
        st, err := f.Stat()
        if err != nil {
                return nil, err
        }
        size := st.Size()
        return &TempArchive{File: f, Size: size}, nil
}

// TempArchive is a temporary archive. The archive can only be read once - as soon as reading completes,
// the file will be deleted.
type TempArchive struct {
        *os.File
        Size   int64 // Pre-computed from Stat().Size() as a convenience
        read   int64
        closed bool
}

// Close closes the underlying file if it's still open, or does a no-op
// to allow callers to try to close the TempArchive multiple times safely.
func (archive *TempArchive) Close() error {
        if archive.closed {
                return nil
        }

        archive.closed = true

        return archive.File.Close()
}

func (archive *TempArchive) Read(data []byte) (int, error) {
        n, err := archive.File.Read(data)
        archive.read += int64(n)
        if err != nil || archive.read == archive.Size {
                archive.Close()
                os.Remove(archive.File.Name())
        }
        return n, err
}