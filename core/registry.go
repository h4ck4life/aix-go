package core

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/h4ck4life/aix-go/constants"
	"github.com/h4ck4life/aix-go/utils"
	"github.com/h4ck4life/aix-go/validation"
)

// Registry manages provider configurations with caching
type Registry struct {
	mu       sync.RWMutex
	data     map[string]constants.ProviderConfig
	loadedAt time.Time
	cacheTTL time.Duration
}

// NewRegistry creates a new registry instance
func NewRegistry() *Registry {
	return &Registry{
		cacheTTL: time.Duration(constants.RegistryCacheTTL) * time.Second,
	}
}

// Load reads the registry from disk
func (r *Registry) Load() error {
	r.mu.Lock()
	defer r.mu.Unlock()

	path := constants.RegistryPath()
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			// Initialize with preconfigured providers on first run
			r.data = make(map[string]constants.ProviderConfig)
			for name, cfg := range constants.PreconfiguredProviders {
				r.data[name] = cfg
			}
			r.loadedAt = time.Now()
			return r.saveLocked()
		}
		return utils.NewFileNotFoundError(path)
	}

	var providers map[string]constants.ProviderConfig
	if err := json.Unmarshal(data, &providers); err != nil {
		return utils.NewValidationError("registry", fmt.Sprintf("failed to parse registry: %v", err))
	}

	r.data = providers
	r.loadedAt = time.Now()
	return nil
}

// saveLocked writes registry atomically (caller must hold lock)
func (r *Registry) saveLocked() error {
	path := constants.RegistryPath()
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return err
	}

	// Write to temp file, then rename for atomicity
	tmpPath := path + ".tmp"
	data, err := json.MarshalIndent(r.data, "", "  ")
	if err != nil {
		return err
	}

	if err := os.WriteFile(tmpPath, data, 0600); err != nil {
		return err
	}

	return os.Rename(tmpPath, path)
}

// Save writes the registry to disk
func (r *Registry) Save() error {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.saveLocked()
}

// isCacheValid checks if cached data is still fresh
func (r *Registry) isCacheValid() bool {
	return !r.loadedAt.IsZero() && time.Since(r.loadedAt) < r.cacheTTL
}

// ensureLoaded loads data if not cached
func (r *Registry) ensureLoaded() error {
	if r.data != nil && r.isCacheValid() {
		return nil
	}
	return r.Load()
}

// GetAll returns all providers
func (r *Registry) GetAll() (map[string]constants.ProviderConfig, error) {
	r.mu.RLock()
	if r.data != nil && r.isCacheValid() {
		defer r.mu.RUnlock()
		return copyProviders(r.data), nil
	}
	r.mu.RUnlock()

	if err := r.ensureLoaded(); err != nil {
		return nil, err
	}

	r.mu.RLock()
	defer r.mu.RUnlock()
	return copyProviders(r.data), nil
}

// GetOne returns a single provider
func (r *Registry) GetOne(name string) (constants.ProviderConfig, error) {
	if err := r.ensureLoaded(); err != nil {
		return constants.ProviderConfig{}, err
	}

	r.mu.RLock()
	defer r.mu.RUnlock()

	cfg, ok := r.data[name]
	if !ok {
		return constants.ProviderConfig{}, utils.NewValidationError("provider", fmt.Sprintf("provider '%s' not found", name))
	}
	return cfg, nil
}

// SetOne adds or updates a provider
func (r *Registry) SetOne(name string, cfg constants.ProviderConfig) error {
	if err := validation.ValidateProviderName(name); err != nil {
		return err
	}
	if err := validation.ValidateURL(cfg.BaseURL); err != nil {
		return err
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	if r.data == nil {
		r.data = make(map[string]constants.ProviderConfig)
	}
	r.data[name] = cfg
	r.loadedAt = time.Now()
	return r.saveLocked()
}

// MergeOne merges incoming config into existing provider, updating only non-zero fields
func (r *Registry) MergeOne(name string, incoming constants.ProviderConfig) error {
	if err := validation.ValidateProviderName(name); err != nil {
		return err
	}

	// Load existing data first so a missing Load() doesn't silently wipe the registry.
	if err := r.ensureLoaded(); err != nil {
		return err
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	if r.data == nil {
		r.data = make(map[string]constants.ProviderConfig)
	}

	existing := r.data[name]

	if incoming.BaseURL != "" {
		if err := validation.ValidateURL(incoming.BaseURL); err != nil {
			return err
		}
		existing.BaseURL = incoming.BaseURL
	}
	if incoming.TokenVar != "" {
		existing.TokenVar = incoming.TokenVar
	}
	if incoming.ModelName != "" {
		existing.ModelName = incoming.ModelName
	}
	if incoming.DefaultModels != nil {
		if existing.DefaultModels == nil {
			existing.DefaultModels = make(map[string]string)
		}
		for k, v := range incoming.DefaultModels {
			existing.DefaultModels[k] = v
		}
	}

	r.data[name] = existing
	r.loadedAt = time.Now()
	return r.saveLocked()
}

// RemoveOne removes a provider
func (r *Registry) RemoveOne(name string) error {
	if err := validation.ValidateProviderName(name); err != nil {
		return err
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	if r.data == nil {
		return utils.NewValidationError("provider", fmt.Sprintf("provider '%s' not found", name))
	}
	if _, ok := r.data[name]; !ok {
		return utils.NewValidationError("provider", fmt.Sprintf("provider '%s' not found", name))
	}
	delete(r.data, name)
	r.loadedAt = time.Now()
	return r.saveLocked()
}

// RenameOne renames a provider
func (r *Registry) RenameOne(oldName, newName string) error {
	if err := validation.ValidateProviderName(oldName); err != nil {
		return err
	}
	if err := validation.ValidateProviderName(newName); err != nil {
		return err
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	if r.data == nil {
		return utils.NewValidationError("provider", fmt.Sprintf("provider '%s' not found", oldName))
	}

	cfg, ok := r.data[oldName]
	if !ok {
		return utils.NewValidationError("provider", fmt.Sprintf("provider '%s' not found", oldName))
	}

	delete(r.data, oldName)
	r.data[newName] = cfg
	r.loadedAt = time.Now()
	return r.saveLocked()
}

// SetModelName sets a custom model name for a provider
func (r *Registry) SetModelName(name, model string) error {
	if err := r.ensureLoaded(); err != nil {
		return err
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	cfg, ok := r.data[name]
	if !ok {
		return utils.NewValidationError("provider", fmt.Sprintf("provider '%s' not found", name))
	}

	cfg.ModelName = model
	r.data[name] = cfg
	r.loadedAt = time.Now()
	return r.saveLocked()
}

// SetDefaultModel sets a default model alias for a provider
func (r *Registry) SetDefaultModel(name, alias, model string) error {
	if err := validation.ValidateModelAlias(alias); err != nil {
		return err
	}
	if err := r.ensureLoaded(); err != nil {
		return err
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	cfg, ok := r.data[name]
	if !ok {
		return utils.NewValidationError("provider", fmt.Sprintf("provider '%s' not found", name))
	}

	if cfg.DefaultModels == nil {
		cfg.DefaultModels = make(map[string]string)
	}
	cfg.DefaultModels[alias] = model
	r.data[name] = cfg
	r.loadedAt = time.Now()
	return r.saveLocked()
}

// ClearCache clears the in-memory cache
func (r *Registry) ClearCache() {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.loadedAt = time.Time{}
}

// copyProviders creates a shallow copy of the providers map
func copyProviders(src map[string]constants.ProviderConfig) map[string]constants.ProviderConfig {
	dst := make(map[string]constants.ProviderConfig, len(src))
	for k, v := range src {
		dst[k] = v
	}
	return dst
}
