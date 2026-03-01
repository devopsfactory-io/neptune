package stacks

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"neptune/internal/domain"
	"neptune/internal/git"
)

// LocalProvider returns changed stacks from config or stack.hcl discovery, filtered by git changes.
type LocalProvider struct{}

// ListAllLocal returns all stack paths in run order for stacks_management: local (no git filter).
// Use for "neptune stacks list" without --changed.
func ListAllLocal(cfg *domain.NeptuneConfig) ([]string, error) {
	rootDir, err := GetRepoRoot()
	if err != nil {
		return nil, err
	}
	p := &LocalProvider{}
	return p.listAllStacks(rootDir, cfg)
}

// ChangedStacks returns the list of changed stack paths in run order.
func (p *LocalProvider) ChangedStacks(ctx context.Context, cfg *domain.NeptuneConfig) (*domain.TerraformStacks, error) {
	_ = ctx
	rootDir, err := GetRepoRoot()
	if err != nil {
		return nil, err
	}
	allStacks, err := p.listAllStacks(rootDir, cfg)
	if err != nil {
		return nil, err
	}
	if len(allStacks) == 0 {
		return &domain.TerraformStacks{Stacks: nil}, nil
	}
	baseRef := "origin/" + cfg.Repository.Branch
	if baseRef == "origin/" {
		baseRef = "origin/master"
	}
	changedPaths, err := git.ChangedPaths(rootDir, baseRef)
	if err != nil {
		return nil, err
	}
	changedSet := make(map[string]bool)
	for _, path := range changedPaths {
		changedSet[path] = true
	}
	// Filter to stacks that have at least one changed file under their path (or the stack path itself).
	var changedStacks []string
	for _, stackPath := range allStacks {
		if stackPath == "" {
			continue
		}
		prefix := stackPath + "/"
		for changedPath := range changedSet {
			if changedPath == stackPath || strings.HasPrefix(changedPath, prefix) {
				changedStacks = append(changedStacks, stackPath)
				break
			}
		}
	}
	return &domain.TerraformStacks{Stacks: changedStacks}, nil
}

// listAllStacks returns all stack paths in run order (config-based or discovery).
func (p *LocalProvider) listAllStacks(rootDir string, cfg *domain.NeptuneConfig) ([]string, error) {
	local := cfg.Repository.LocalStacks
	source := "discovery"
	if local != nil && local.Source != "" {
		source = local.Source
	}
	if source == "config" {
		if local == nil || len(local.Stacks) == 0 {
			return nil, errors.New("local_stacks.source is config but local_stacks.stacks is empty")
		}
		return topologicalOrder(local.Stacks), nil
	}
	// discovery: find directories containing stack.hcl
	return discoverStackHcl(rootDir), nil
}

// topologicalOrder returns stack paths in run order (dependencies first).
func topologicalOrder(entries []domain.StackEntry) []string {
	pathToDeps := make(map[string][]string)
	for _, e := range entries {
		pathToDeps[e.Path] = e.DependsOn
	}
	var order []string
	visited := make(map[string]bool)
	var visit func(path string)
	visit = func(path string) {
		if visited[path] {
			return
		}
		visited[path] = true
		for _, dep := range pathToDeps[path] {
			visit(dep)
		}
		order = append(order, path)
	}
	for _, e := range entries {
		visit(e.Path)
	}
	return order
}

// discoverStackHcl walks rootDir for directories that contain stack.hcl (relative paths from root).
func discoverStackHcl(rootDir string) []string {
	var stacks []string
	_ = filepath.Walk(rootDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		if info.IsDir() {
			if info.Name() == ".git" {
				return filepath.SkipDir
			}
			return nil
		}
		if info.Name() == "stack.hcl" {
			dir := filepath.Dir(path)
			rel, err := filepath.Rel(rootDir, dir)
			if err != nil {
				return nil
			}
			rel = filepath.ToSlash(rel)
			if rel != "." && rel != "" {
				stacks = append(stacks, rel)
			}
		}
		return nil
	})
	sort.Strings(stacks)
	return stacks
}
