package bindings

import (
	"context"
	"fmt"

	"aDex-UI/internal/services/colorscheme"
	"aDex-UI/internal/models"
)

// ColorSchemeService provides Wails bindings for the color scheme service
type ColorSchemeService struct {
	service *colorscheme.Service
	ctx     context.Context
}

// NewColorSchemeService creates a new color scheme service binding
func NewColorSchemeService(service *colorscheme.Service, ctx context.Context) *ColorSchemeService {
	return &ColorSchemeService{
		service: service,
		ctx:     ctx,
	}
}

// GetSchemes returns all available color schemes
func (c *ColorSchemeService) GetSchemes() (map[string]*models.ColorScheme, error) {
	return c.service.GetSchemes(), nil
}

// GetScheme returns a specific color scheme by ID
func (c *ColorSchemeService) GetScheme(id string) (*models.ColorScheme, error) {
	scheme := c.service.GetScheme(id)
	if scheme == nil {
		return nil, fmt.Errorf("color scheme with ID '%s' not found", id)
	}
	return scheme, nil
}

// CreateScheme creates a new color scheme
func (c *ColorSchemeService) CreateScheme(scheme *models.ColorScheme) error {
	return c.service.CreateScheme(scheme)
}

// UpdateScheme updates an existing color scheme
func (c *ColorSchemeService) UpdateScheme(scheme *models.ColorScheme) error {
	return c.service.UpdateScheme(scheme)
}

// DeleteScheme deletes a color scheme
func (c *ColorSchemeService) DeleteScheme(id string) error {
	return c.service.DeleteScheme(id)
}

// SetDefaultScheme sets the default color scheme
func (c *ColorSchemeService) SetDefaultScheme(id string) error {
	return c.service.SetDefaultScheme(id)
}

// GetDefaultScheme returns the default color scheme
func (c *ColorSchemeService) GetDefaultScheme() (*models.ColorScheme, error) {
	scheme := c.service.GetDefaultScheme()
	if scheme == nil {
		return nil, fmt.Errorf("no default color scheme found")
	}
	return scheme, nil
}

// GetConfig returns the color scheme configuration
func (c *ColorSchemeService) GetConfig() (*models.ColorSchemeConfig, error) {
	return c.service.GetConfig(), nil
}

// UpdateConfig updates the color scheme configuration
func (c *ColorSchemeService) UpdateConfig(config *models.ColorSchemeConfig) error {
	return c.service.UpdateConfig(config)
}

// GetSchemePreview generates a preview for a color scheme
func (c *ColorSchemeService) GetSchemePreview(id string) (*models.ColorSchemePreview, error) {
	return c.service.GetSchemePreview(id)
}

// ValidateScheme validates a color scheme
func (c *ColorSchemeService) ValidateScheme(scheme *models.ColorScheme) (*models.ColorSchemeValidationResult, error) {
	return c.service.ValidateScheme(scheme), nil
}