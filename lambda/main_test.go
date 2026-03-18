package main

import "testing"

// TestShouldDispatchPullRequest verifies the deduplication logic that prevents
// a double-dispatch when a PR is created with a label already attached.
//
// GitHub fires TWO events when `gh pr create --label <label>` is used:
//   - pull_request.opened  (pull_request.labels already contains the label)
//   - pull_request.labeled (label.name == the new label)
//
// The fix: when a prLabel gate is active, suppress opened/reopened events and
// rely on the labeled event as the sole initial trigger. synchronize and
// ready_for_review are still gated by label presence because no labeled event
// accompanies a push.
func TestShouldDispatchPullRequest(t *testing.T) {
	tests := []struct {
		name       string
		prLabel    string
		action     string
		addedLabel string
		labels     []string
		want       bool
	}{
		// No label gate configured — all non-labeled actions dispatch.
		{
			name:   "no label gate, opened",
			action: "opened",
			want:   true,
		},
		{
			name:   "no label gate, synchronize",
			action: "synchronize",
			want:   true,
		},
		{
			name:   "no label gate, ready_for_review",
			action: "ready_for_review",
			want:   true,
		},
		// No label gate, labeled event with addedLabel — no gate means no match.
		{
			name:       "no label gate, labeled event",
			action:     "labeled",
			addedLabel: "neptune",
			want:       false,
		},

		// Label gate active; opened/reopened are suppressed to avoid double dispatch.
		{
			name:    "label gate, opened with label present — suppressed",
			prLabel: "neptune",
			action:  "opened",
			labels:  []string{"neptune"},
			want:    false,
		},
		{
			name:    "label gate, opened without label — suppressed",
			prLabel: "neptune",
			action:  "opened",
			labels:  []string{},
			want:    false,
		},
		{
			name:    "label gate, reopened with label present — dispatches",
			prLabel: "neptune",
			action:  "reopened",
			labels:  []string{"neptune"},
			want:    true,
		},
		{
			name:    "label gate, reopened without label — suppressed",
			prLabel: "neptune",
			action:  "reopened",
			labels:  []string{},
			want:    false,
		},

		// Label gate active; labeled event is the sole trigger for PR creation.
		{
			name:       "label gate, labeled event matching gate — dispatches",
			prLabel:    "neptune",
			action:     "labeled",
			addedLabel: "neptune",
			labels:     []string{"neptune"},
			want:       true,
		},
		{
			name:       "label gate, labeled event non-matching — suppressed",
			prLabel:    "neptune",
			action:     "labeled",
			addedLabel: "other",
			labels:     []string{"other"},
			want:       false,
		},

		// Label gate active; synchronize/ready_for_review check label presence.
		{
			name:    "label gate, synchronize with label — dispatches",
			prLabel: "neptune",
			action:  "synchronize",
			labels:  []string{"neptune", "extra"},
			want:    true,
		},
		{
			name:    "label gate, synchronize without label — suppressed",
			prLabel: "neptune",
			action:  "synchronize",
			labels:  []string{"other"},
			want:    false,
		},
		{
			name:    "label gate, ready_for_review with label — dispatches",
			prLabel: "neptune",
			action:  "ready_for_review",
			labels:  []string{"neptune"},
			want:    true,
		},
		{
			name:    "label gate, ready_for_review without label — suppressed",
			prLabel: "neptune",
			action:  "ready_for_review",
			labels:  []string{},
			want:    false,
		},

		// Edge: labeled event when no gate — should not dispatch.
		{
			name:       "no gate, labeled event for neptune — suppressed",
			prLabel:    "",
			action:     "labeled",
			addedLabel: "neptune",
			want:       false,
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := shouldDispatchPullRequest(tc.prLabel, tc.action, tc.addedLabel, tc.labels)
			if got != tc.want {
				t.Errorf("shouldDispatchPullRequest(%q, %q, %q, %v) = %v, want %v",
					tc.prLabel, tc.action, tc.addedLabel, tc.labels, got, tc.want)
			}
		})
	}
}
