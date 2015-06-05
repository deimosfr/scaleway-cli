package main

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
)

// ScalewayCache is used not to query the API to resolve full identifiers
type ScalewayCache struct {
	// Images contains names of Scaleway images indexed by identifier
	Images map[string]string `json:"images"`

	// Snapshots contains names of Scaleway snapshots indexed by identifier
	Snapshots map[string]string `json:"snapshots"`

	// Bootscripts contains names of Scaleway bootscripts indexed by identifier
	Bootscripts map[string]string `json:"bootscripts"`

	// Servers contains names of Scaleway C1 servers indexed by identifier
	Servers map[string]string `json:"servers"`

	// Path is the path to the cache file
	Path string `json:"-"`

	// Modified tells if the cache needs to be overwritten or not
	Modified bool `json:"-"`

	// Lock allows ScalewayCache to be used concurrently
	Lock sync.Mutex `json:"-"`
}

const (
	// IdentifierServer is the type key of cached server objects
	IdentifierServer = iota
	// IdentifierImage is the type key of cached image objects
	IdentifierImage
	// IdentifierSnapshot is the type key of cached snapshot objects
	IdentifierSnapshot
	// IdentifierBootscript is the type key of cached bootscript objects
	IdentifierBootscript
)

// ScalewayIdentifier is a unique identifier on Scaleway
type ScalewayIdentifier struct {
	// Identifier is a unique identifier on
	Identifier string

	// Type of the identifier
	Type int
}

// NewScalewayCache loads a per-user cache
func NewScalewayCache() (*ScalewayCache, error) {
	homeDir := os.Getenv("HOME") // *nix
	if homeDir == "" {           // Windows
		homeDir = os.Getenv("USERPROFILE")
	}
	if homeDir == "" {
		homeDir = "/tmp"
	}
	cachePath := filepath.Join(homeDir, ".scw-cache.db")
	_, err := os.Stat(cachePath)
	if os.IsNotExist(err) {
		return &ScalewayCache{
			Images:      make(map[string]string),
			Snapshots:   make(map[string]string),
			Bootscripts: make(map[string]string),
			Servers:     make(map[string]string),
			Path:        cachePath,
		}, nil
	} else if err != nil {
		return nil, err
	}
	file, err := ioutil.ReadFile(cachePath)
	if err != nil {
		return nil, err
	}
	var cache ScalewayCache
	cache.Path = cachePath
	err = json.Unmarshal(file, &cache)
	if err != nil {
		return nil, err
	}
	if cache.Images == nil {
		cache.Images = make(map[string]string)
	}
	if cache.Snapshots == nil {
		cache.Snapshots = make(map[string]string)
	}
	if cache.Servers == nil {
		cache.Servers = make(map[string]string)
	}
	if cache.Bootscripts == nil {
		cache.Bootscripts = make(map[string]string)
	}
	return &cache, nil
}

// Save atomically overwrites the current cache database
func (c *ScalewayCache) Save() error {
	c.Lock.Lock()
	defer c.Lock.Unlock()

	if c.Modified {
		file, err := ioutil.TempFile("", "")
		if err != nil {
			return err
		}
		encoder := json.NewEncoder(file)
		err = encoder.Encode(*c)
		if err != nil {
			return err
		}
		return os.Rename(file.Name(), c.Path)
	}
	return nil
}

// LookUpImages attempts to return identifiers matching a pattern
func (c *ScalewayCache) LookUpImages(needle string) []string {
	c.Lock.Lock()
	defer c.Lock.Unlock()

	var res []string
	needle = regexp.MustCompile(`^user/`).ReplaceAllString(needle, "")
	// FIXME: if 'user/' is in needle, only watch for a user image
	nameRegex := regexp.MustCompile(`(?i)` + regexp.MustCompile(`[_-]`).ReplaceAllString(needle, ".*"))
	for identifier, name := range c.Images {
		if strings.HasPrefix(identifier, needle) || nameRegex.MatchString(name) {
			res = append(res, identifier)
		}
	}
	return res
}

// LookUpSnapshots attempts to return identifiers matching a pattern
func (c *ScalewayCache) LookUpSnapshots(needle string) []string {
	c.Lock.Lock()
	defer c.Lock.Unlock()

	var res []string
	needle = regexp.MustCompile(`^user/`).ReplaceAllString(needle, "")
	nameRegex := regexp.MustCompile(`(?i)` + regexp.MustCompile(`[_-]`).ReplaceAllString(needle, ".*"))
	for identifier, name := range c.Snapshots {
		if strings.HasPrefix(identifier, needle) || nameRegex.MatchString(name) {
			res = append(res, identifier)
		}
	}
	return res
}

// LookUpBootscripts attempts to return identifiers matching a pattern
func (c *ScalewayCache) LookUpBootscripts(needle string) []string {
	c.Lock.Lock()
	defer c.Lock.Unlock()

	var res []string
	nameRegex := regexp.MustCompile(`(?i)` + regexp.MustCompile(`[_-]`).ReplaceAllString(needle, ".*"))
	for identifier, name := range c.Bootscripts {
		if strings.HasPrefix(identifier, needle) || nameRegex.MatchString(name) {
			res = append(res, identifier)
		}
	}
	return res
}

// LookUpServers attempts to return identifiers matching a pattern
func (c *ScalewayCache) LookUpServers(needle string) []string {
	c.Lock.Lock()
	defer c.Lock.Unlock()

	var res []string
	nameRegex := regexp.MustCompile(`(?i)` + regexp.MustCompile(`[_-]`).ReplaceAllString(needle, ".*"))
	for identifier, name := range c.Servers {
		if strings.HasPrefix(identifier, needle) || nameRegex.MatchString(name) {
			res = append(res, identifier)
		}
	}
	return res
}

// LookUpIdentifiers attempts to return identifiers matching a pattern
func (c *ScalewayCache) LookUpIdentifiers(needle string) []ScalewayIdentifier {
	result := []ScalewayIdentifier{}

	for _, identifier := range c.LookUpServers(needle) {
		result = append(result, ScalewayIdentifier{
			Identifier: identifier,
			Type:       IdentifierServer,
		})
	}

	for _, identifier := range c.LookUpImages(needle) {
		result = append(result, ScalewayIdentifier{
			Identifier: identifier,
			Type:       IdentifierImage,
		})
	}

	for _, identifier := range c.LookUpSnapshots(needle) {
		result = append(result, ScalewayIdentifier{
			Identifier: identifier,
			Type:       IdentifierSnapshot,
		})
	}

	for _, identifier := range c.LookUpBootscripts(needle) {
		result = append(result, ScalewayIdentifier{
			Identifier: identifier,
			Type:       IdentifierBootscript,
		})
	}

	return result
}

// InsertServer registers a server in the cache
func (c *ScalewayCache) InsertServer(identifier, name string) {
	c.Lock.Lock()
	defer c.Lock.Unlock()

	currentName, exists := c.Servers[identifier]
	if !exists || currentName != name {
		c.Servers[identifier] = name
		c.Modified = true
	}
}

// RemoveServer removes a server from the cache
func (c *ScalewayCache) RemoveServer(identifier string) {
	c.Lock.Lock()
	defer c.Lock.Unlock()

	delete(c.Servers, identifier)
	c.Modified = true
}

// ClearServers removes all servers from the cache
func (c *ScalewayCache) ClearServers() {
	c.Lock.Lock()
	defer c.Lock.Unlock()

	c.Servers = make(map[string]string)
	c.Modified = true
}

// InsertImage registers an image in the cache
func (c *ScalewayCache) InsertImage(identifier, name string) {
	c.Lock.Lock()
	defer c.Lock.Unlock()

	currentName, exists := c.Images[identifier]
	if !exists || currentName != name {
		c.Images[identifier] = name
		c.Modified = true
	}
}

// RemoveImage removes a server from the cache
func (c *ScalewayCache) RemoveImage(identifier string) {
	c.Lock.Lock()
	defer c.Lock.Unlock()

	delete(c.Images, identifier)
	c.Modified = true
}

// ClearImages removes all images from the cache
func (c *ScalewayCache) ClearImages() {
	c.Lock.Lock()
	defer c.Lock.Unlock()

	c.Images = make(map[string]string)
	c.Modified = true
}

// InsertSnapshot registers an snapshot in the cache
func (c *ScalewayCache) InsertSnapshot(identifier, name string) {
	c.Lock.Lock()
	defer c.Lock.Unlock()

	currentName, exists := c.Snapshots[identifier]
	if !exists || currentName != name {
		c.Snapshots[identifier] = name
		c.Modified = true
	}
}

// RemoveSnapshot removes a server from the cache
func (c *ScalewayCache) RemoveSnapshot(identifier string) {
	c.Lock.Lock()
	defer c.Lock.Unlock()

	delete(c.Snapshots, identifier)
	c.Modified = true
}

// ClearSnapshots removes all snapshots from the cache
func (c *ScalewayCache) ClearSnapshots() {
	c.Lock.Lock()
	defer c.Lock.Unlock()

	c.Snapshots = make(map[string]string)
	c.Modified = true
}

// InsertBootscript registers an bootscript in the cache
func (c *ScalewayCache) InsertBootscript(identifier, name string) {
	c.Lock.Lock()
	defer c.Lock.Unlock()

	currentName, exists := c.Bootscripts[identifier]
	if !exists || currentName != name {
		c.Bootscripts[identifier] = name
		c.Modified = true
	}
}

// RemoveBootscript removes a bootscript from the cache
func (c *ScalewayCache) RemoveBootscript(identifier string) {
	c.Lock.Lock()
	defer c.Lock.Unlock()

	delete(c.Bootscripts, identifier)
	c.Modified = true
}

// ClearBootscripts removes all bootscripts from the cache
func (c *ScalewayCache) ClearBootscripts() {
	c.Lock.Lock()
	defer c.Lock.Unlock()

	c.Bootscripts = make(map[string]string)
	c.Modified = true
}

// GetNbServers returns the number of servers in the cache
func (c *ScalewayCache) GetNbServers() int {
	c.Lock.Lock()
	defer c.Lock.Unlock()

	return len(c.Servers)
}

// GetNbImages returns the number of images in the cache
func (c *ScalewayCache) GetNbImages() int {
	c.Lock.Lock()
	defer c.Lock.Unlock()

	return len(c.Images)
}

// GetNbSnapshots returns the number of snapshots in the cache
func (c *ScalewayCache) GetNbSnapshots() int {
	c.Lock.Lock()
	defer c.Lock.Unlock()

	return len(c.Snapshots)
}

// GetNbBootscripts returns the number of bootscripts in the cache
func (c *ScalewayCache) GetNbBootscripts() int {
	c.Lock.Lock()
	defer c.Lock.Unlock()

	return len(c.Bootscripts)
}
