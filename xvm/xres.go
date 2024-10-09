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
	Profiles map[string]Resource
}

func NewResourceManager(path string) (ResourceManager, error) {
	profilesPath := filepath.Join(path, "profiles")

	// Check if the profiles directory exists
	if info, err := os.Stat(profilesPath); err != nil || !info.IsDir() {
		return ResourceManager{}, fmt.Errorf("profiles directory does not exist at %s", profilesPath)
	}

	rm := ResourceManager{
		Profiles: make(map[string]Resource),
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
