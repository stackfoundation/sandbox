package image

// Modified from https://github.com/moby/moby/blob/1009e6a40b295187e038b67e184e9c0384d95538/pkg/archive/whiteouts.go
// Licensed under the Apache License Version 2.0

// Whiteouts are files with a special meaning for the layered filesystem.
// Docker uses AUFS whiteout files inside exported archives. In other
// filesystems these files are generated/handled on tar creation/extraction.

// WhiteoutPrefix prefix means file is a whiteout. If this is followed by a
// filename this means that file has been removed from the base layer.
const WhiteoutPrefix = ".wh."
