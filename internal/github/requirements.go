package github

import (
	"context"
	"strings"

	"neptune/internal/domain"
	"neptune/internal/git"
)

// CheckRequirements checks if the PR meets the given requirements (approved, mergeable, undiverged, rebased).
func (c *Client) CheckRequirements(ctx context.Context, requirements []string) *domain.PRRequirementsStatus {
	if len(requirements) == 0 {
		return &domain.PRRequirementsStatus{IsCompliant: true}
	}
	prInfo, err := c.GetPRInfo(ctx)
	if err != nil {
		return &domain.PRRequirementsStatus{
			IsCompliant:  false,
			ErrorMessage: "Could not fetch PR information. Make sure GITHUB_TOKEN is set and has access to the repository.",
		}
	}
	var failed []string
	for _, req := range requirements {
		switch req {
		case "approved":
			ok, err := c.checkApproved(ctx)
			if err != nil || !ok {
				failed = append(failed, req)
			}
		case "mergeable":
			if v, _ := prInfo.Response["mergeable"].(bool); !v {
				failed = append(failed, req)
			}
		case "undiverged":
			if v, _ := prInfo.Response["mergeable_state"].(string); v == "behind" {
				failed = append(failed, req)
			}
		case "rebased":
			if !git.IsBranchRebased(c.cfg) {
				failed = append(failed, req)
			}
		}
	}
	compliant := len(failed) == 0
	msg := ""
	if !compliant {
		msg = "PR does not meet the following requirements: " + strings.Join(failed, ", ")
	}
	return &domain.PRRequirementsStatus{
		IsCompliant:        compliant,
		FailedRequirements: failed,
		ErrorMessage:       msg,
	}
}

func (c *Client) checkApproved(ctx context.Context) (bool, error) {
	reviews, err := c.ListReviews(ctx)
	if err != nil {
		return false, err
	}
	for _, r := range reviews {
		if r != nil && r.State != nil && strings.ToLower(*r.State) == "approved" {
			return true, nil
		}
	}
	return false, nil
}
