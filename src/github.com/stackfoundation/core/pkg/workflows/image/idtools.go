package image

// Modified from https://github.com/moby/moby/blob/1009e6a40b295187e038b67e184e9c0384d95538/pkg/idtools/idtools.go
// Licensed under the Apache License Version 2.0

import "fmt"

// IDMap contains a single entry for user namespace range remapping. An array
// of IDMap entries represents the structure that will be provided to the Linux
// kernel for creating a user namespace.
type IDMap struct {
        ContainerID int `json:"container_id"`
        HostID      int `json:"host_id"`
        Size        int `json:"size"`
}

// IDPair is a UID and GID pair
type IDPair struct {
        UID int
        GID int
}

// IDMappings contains a mappings of UIDs and GIDs
type IDMappings struct {
        uids []IDMap
        gids []IDMap
}

// NewIDMappingsFromMaps creates a new mapping from two slices
// Deprecated: this is a temporary shim while transitioning to IDMapping
func NewIDMappingsFromMaps(uids []IDMap, gids []IDMap) *IDMappings {
        return &IDMappings{uids: uids, gids: gids}
}

// toContainer takes an id mapping, and uses it to translate a
// host ID to the remapped ID. If no map is provided, then the translation
// assumes a 1-to-1 mapping and returns the passed in id
func toContainer(hostID int, idMap []IDMap) (int, error) {
        if idMap == nil {
                return hostID, nil
        }
        for _, m := range idMap {
                if (hostID >= m.HostID) && (hostID <= (m.HostID + m.Size - 1)) {
                        contID := m.ContainerID + (hostID - m.HostID)
                        return contID, nil
                }
        }
        return -1, fmt.Errorf("Host ID %d cannot be mapped to a container ID", hostID)
}

// ToContainer returns the container UID and GID for the host uid and gid
func (i *IDMappings) ToContainer(pair IDPair) (int, int, error) {
        uid, err := toContainer(pair.UID, i.uids)
        if err != nil {
                return -1, -1, err
        }
        gid, err := toContainer(pair.GID, i.gids)
        return uid, gid, err
}

// Empty returns true if there are no id mappings
func (i *IDMappings) Empty() bool {
        return len(i.uids) == 0 && len(i.gids) == 0
}
