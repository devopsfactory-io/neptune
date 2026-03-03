package github

import (
	"context"
	"strings"

	"github.com/devopsfactory-io/neptune/internal/domain"
	"github.com/devopsfactory-io/neptune/internal/git"
	"github.com/devopsfactory-io/neptune/internal/log"
)

// CheckRequirements checks if the PR meets the given requirements (approved, mergeable, undiverged, rebased).
func (c *Client) CheckRequirements(ctx context.Context, requirements []string) *domain.PRRequirementsStatus {
	if len(requirements) == 0 {
		return &domain.PRRequirementsStatus{IsCompliant: true}
	}
	log.For("github").Info("Checking requirements for PR " + c.prNum)
	log.For("github").Info("Getting PR info")
	prInfo, err := c.GetPRInfo(ctx)
	if err != nil {
		return &domain.PRRequirementsStatus{
			IsCompliant:  false,
			ErrorMessage: "Could not fetch PR information. Make sure GITHUB_TOKEN is set and has access to the repository.",
		}
	}
	log.For("github").Info("Getting PR requirements information")
	var failed []string
	for _, req := range requirements {
		log.For("github").Info("Checking requirement: " + req)
		switch req {
		case "approved":
			ok, err := c.checkApproved(ctx)
			if err != nil || !ok {
				failed = append(failed, req)
			}
		case "mergeable":
			log.For("github").Info("Checking PR mergeability")
			v, ok := prInfo.Response["mergeable"].(bool)
			if !ok || !v {
				failed = append(failed, req)
			}
		case "undiverged":
			log.For("github").Info("Checking PR branch is up to date with base")
			v, ok := prInfo.Response["mergeable_state"].(string)
			if ok && v == "behind" {
				failed = append(failed, req)
			}
		case "rebased":
			log.For("github").Info("Checking PR branch is rebased...")
			if !git.IsBranchRebased(c.cfg) {
				failed = append(failed, req)
			}
		}
	}
	log.For("github").Info("PR requirements collected")
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
