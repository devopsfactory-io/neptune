package lock

import (
	"context"

	"github.com/devopsfactory-io/neptune/internal/domain"
	"github.com/devopsfactory-io/neptune/internal/stacks"
)

// ChangedStacks returns the list of changed stack paths in run order using the Terramate SDK.
// It is a wrapper around the terramate stacks provider for backward compatibility.
// Prefer using stacks.NewProvider(cfg).ChangedStacks(ctx, cfg) so stacks_management is respected.
func ChangedStacks(ctx context.Context, cfg *domain.NeptuneConfig) (*domain.TerraformStacks, error) {
	cfg2 := *cfg
	if cfg2.Repository == nil {
		cfg2.Repository = &domain.RepositoryConfig{}
	}
	repoCopy := *cfg2.Repository
	repoCopy.StacksManagement = "terramate"
	cfg2.Repository = &repoCopy
	return stacks.NewProvider(&cfg2).ChangedStacks(ctx, cfg)
}
