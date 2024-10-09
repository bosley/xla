// The ResourceManager is responsible for loading and managing resources such as profiles.
// Resources are referenced in the code using identifiers like "@profiles/general",
// which correspond to entries in the ResourceManager's Profiles map.
// For example, "@profiles/general" refers to the resource "general" loaded from "resources/profiles/general.yaml".

package xvm

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type Resource struct {
	Name     string
	FilePath string
	Type     string
}

type ResourceManager struct {
	Profiles  map[string]Resource
	XlaGoPath Resource
}

func NewResourceManager(path string) (ResourceManager, error) {
	profilesPath := filepath.Join(path, "profiles")

	// Add "xla_gopath" to the resources path
	xlaGoPath := filepath.Join(path, "xla_gopath")

	rm := ResourceManager{
		Profiles: make(map[string]Resource),
	}

	// Check if xlaGoPath exists and is a file
	if info, err := os.Stat(xlaGoPath); err == nil && !info.IsDir() {
		absPath, err := filepath.Abs(xlaGoPath)
		if err != nil {
			return ResourceManager{}, fmt.Errorf("failed to get absolute path of xla_gopath: %w", err)
		}
		rm.XlaGoPath.Name = "xla_gopath"
		rm.XlaGoPath.FilePath = absPath
		rm.XlaGoPath.Type = "dir"
	} else {
		// xlaGoPath does not exist or is a directory, create it
		err := os.MkdirAll(xlaGoPath, os.ModePerm)
		if err != nil {
			return ResourceManager{}, fmt.Errorf("failed to create xla_gopath directory: %w", err)
		}
	}

	// When loading, if profilesPath doesn't exist, create it and add to the resource manager
	if info, err := os.Stat(profilesPath); err != nil || !info.IsDir() {
		err = os.MkdirAll(profilesPath, os.ModePerm)
		if err != nil {
			return ResourceManager{}, fmt.Errorf("failed to create profiles directory at %s: %w", profilesPath, err)
		}
	}

	// Read all yaml files in the profiles directory
	err := filepath.Walk(profilesPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() && strings.HasSuffix(info.Name(), ".yaml") {
			name := strings.TrimSuffix(info.Name(), ".yaml")
			rm.Profiles[name] = Resource{
				Name:     name,
				FilePath: path,
				Type:     "yaml",
			}
		}
		return nil
	})

	if err != nil {
		return ResourceManager{}, fmt.Errorf("error reading profiles: %w", err)
	}

	return rm, nil
}
