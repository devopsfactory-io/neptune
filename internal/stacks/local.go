package stacks

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/devopsfactory-io/neptune/internal/domain"
	"github.com/devopsfactory-io/neptune/internal/git"
)

// stackHclFilename is the filename used for discovery.
const stackHclFilename = "stack.hcl"

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
		order, err := topologicalOrder(local.Stacks)
		if err != nil {
			return nil, err
		}
		return order, nil
	}
	// discovery: find directories containing stack.hcl, parse depends_on, resolve and expand, then order
	return discoverStackHclOrdered(rootDir)
}

// resolveDepPath resolves a depends_on entry to a repo-root-relative path.
// stackPath is the stack's directory (repo-root-relative). If dep is relative (contains ".." or
// starts with "./" or "."), it is resolved relative to the stack's directory; otherwise it is
// treated as repo-root-relative. Returned path uses forward slashes.
func resolveDepPath(rootDir, stackPath, dep string) string {
	isRelative := strings.Contains(dep, "..") || strings.HasPrefix(dep, "./") || (len(dep) > 0 && dep[0] == '.')
	if isRelative {
		joined := filepath.Join(rootDir, filepath.FromSlash(stackPath), dep)
		rel, err := filepath.Rel(rootDir, joined)
		if err != nil {
			return filepath.ToSlash(filepath.Clean(dep))
		}
		return filepath.ToSlash(filepath.Clean(rel))
	}
	return filepath.ToSlash(filepath.Clean(dep))
}

// expandDirDeps expands dependency paths to concrete stack paths. For each dep in depPaths (repo-
// root-relative), adds every stack in allPaths that equals dep or is under dep (dep is a directory
// of stacks). Returned list has no duplicates and uses forward slashes.
func expandDirDeps(allPaths []string, depPaths []string) []string {
	seen := make(map[string]bool)
	var result []string
	for _, dep := range depPaths {
		dep = filepath.ToSlash(filepath.Clean(dep))
		for _, p := range allPaths {
			p := filepath.ToSlash(p)
			if p == dep || strings.HasPrefix(p+"/", dep+"/") {
				if !seen[p] {
					seen[p] = true
					result = append(result, p)
				}
			}
		}
	}
	return result
}

// topologicalOrder returns stack paths in run order (dependencies first).
// It detects cycles and returns an error (wrapped in a sentinel or we need to change signature).
// For now we keep the same signature and add cycle detection that returns a partial order and
// no error (or we add error return). Plan says "optional but recommended" - we add cycle detection
// and on cycle return error. So we need topologicalOrder to return ([]string, error).
func topologicalOrder(entries []domain.StackEntry) ([]string, error) {
	pathToDeps := make(map[string][]string)
	for _, e := range entries {
		pathToDeps[e.Path] = e.DependsOn
	}
	var order []string
	visited := make(map[string]bool)
	visiting := make(map[string]bool)
	var visit func(path string) error
	visit = func(path string) error {
		if visiting[path] {
			return fmt.Errorf("cycle in stack dependencies involving %q", path)
		}
		if visited[path] {
			return nil
		}
		visiting[path] = true
		defer func() { visiting[path] = false }()
		for _, dep := range pathToDeps[path] {
			if err := visit(dep); err != nil {
				return err
			}
		}
		visited[path] = true
		order = append(order, path)
		return nil
	}
	for _, e := range entries {
		if err := visit(e.Path); err != nil {
			return nil, err
		}
	}
	return order, nil
}

// walkAndCollectStacks walks rootDir and collects directories containing stack.hcl files.
func walkAndCollectStacks(rootDir string) ([]string, error) {
	var paths []string
	err := filepath.Walk(rootDir, func(path string, info os.FileInfo, walkErr error) error {
		if walkErr != nil {
			return nil
		}
		if info.IsDir() {
			if info.Name() == ".git" {
				return filepath.SkipDir
			}
			return nil
		}
		if info.Name() == stackHclFilename {
			dir := filepath.Dir(path)
			rel, err := filepath.Rel(rootDir, dir)
			if err != nil {
				return nil
			}
			rel = filepath.ToSlash(rel)
			if rel != "." && rel != "" {
				paths = append(paths, rel)
			}
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	sort.Strings(paths)
	return paths, nil
}

// resolveAndBuildEntries parses stack.hcl for each path, resolves dependencies, and builds entries for topological sort.
func resolveAndBuildEntries(rootDir string, paths []string) ([]domain.StackEntry, error) {
	var entries []domain.StackEntry
	for _, p := range paths {
		hclPath := filepath.Join(rootDir, filepath.FromSlash(p), stackHclFilename)
		_, rawDeps, err := ParseStackHcl(hclPath)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", hclPath, err)
		}
		var resolved []string
		for _, d := range rawDeps {
			resolved = append(resolved, resolveDepPath(rootDir, p, d))
		}
		concreteDeps := expandDirDeps(paths, resolved)
		entries = append(entries, domain.StackEntry{Path: p, DependsOn: concreteDeps})
	}
	return entries, nil
}

// discoverStackHclOrdered discovers stacks with stack.hcl, parses depends_on, resolves relative
// paths, expands directory deps, and returns stack paths in topological order. Returns error on
// parse failure or dependency cycle.
func discoverStackHclOrdered(rootDir string) ([]string, error) {
	paths, err := walkAndCollectStacks(rootDir)
	if err != nil {
		return nil, err
	}
	if len(paths) == 0 {
		return nil, nil
	}
	entries, err := resolveAndBuildEntries(rootDir, paths)
	if err != nil {
		return nil, err
	}
	return topologicalOrder(entries)
}
