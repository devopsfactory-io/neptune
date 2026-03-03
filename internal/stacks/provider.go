package stacks

import (
	"context"
	"strings"

	"github.com/devopsfactory-io/neptune/internal/domain"
)

// Provider returns the list of changed stack paths in run order.
// The implementation is selected by repository.stacks_management (terramate or local).
type Provider interface {
	ChangedStacks(ctx context.Context, cfg *domain.NeptuneConfig) (*domain.TerraformStacks, error)
}

// NewProvider returns a StacksProvider for the given config.
// Default is terramate when stacks_management is empty.
func NewProvider(cfg *domain.NeptuneConfig) Provider {
	sm := strings.TrimSpace(strings.ToLower(cfg.Repository.StacksManagement))
	if sm == "" {
		sm = "terramate"
	}
	if sm == "local" {
		return &LocalProvider{}
	}
	return &TerramateProvider{}
}
