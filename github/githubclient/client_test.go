package githubclient

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/ejoffe/spr/config"
	"github.com/ejoffe/spr/git"
	"github.com/ejoffe/spr/github"
	"github.com/ejoffe/spr/github/githubclient/fezzik_types"
	"github.com/stretchr/testify/require"
)

func TestMatchPullRequestStack(t *testing.T) {
	tests := []struct {
		name    string
		commits []git.Commit
		prs     fezzik_types.PullRequestConnection
		expect  []*github.PullRequest
	}{
		{
			name: "ThirdCommitQueue",
			commits: []git.Commit{
				{CommitID: "00000001"},
				{CommitID: "00000002"},
				{CommitID: "00000003"},
			},
			prs: fezzik_types.PullRequestConnection{
				Nodes: &fezzik_types.PullRequestsViewerPullRequestsNodes{
					{
						Id:              "2",
						HeadRefName:     "spr/master/00000002",
						BaseRefName:     "master",
						MergeQueueEntry: &fezzik_types.PullRequestsViewerPullRequestsNodesMergeQueueEntry{Id: "020"},
						Commits: fezzik_types.PullRequestsViewerPullRequestsNodesCommits{
							Nodes: &fezzik_types.PullRequestsViewerPullRequestsNodesCommitsNodes{
								{
									fezzik_types.PullRequestsViewerPullRequestsNodesCommitsNodesCommit{Oid: "1", MessageBody: "commit-id:1"},
								},
								{
									fezzik_types.PullRequestsViewerPullRequestsNodesCommitsNodesCommit{Oid: "2", MessageBody: "commit-id:2"},
								},
							},
						},
					},
				},
			},
			expect: []*github.PullRequest{
				{
					ID:         "2",
					FromBranch: "spr/master/00000002",
					ToBranch:   "master",
					Commit: git.Commit{
						CommitID:   "00000002",
						CommitHash: "2",
						Body:       "commit-id:2",
					},
					InQueue: true,
					Commits: []git.Commit{
						{CommitID: "1", CommitHash: "1", Body: "commit-id:1"},
						{CommitID: "2", CommitHash: "2", Body: "commit-id:2"},
					},
					MergeStatus: github.PullRequestMergeStatus{
						ChecksPass: github.CheckStatusPass,
					},
				},
			},
		},
		{
			name: "FourthCommitQueue",
			commits: []git.Commit{
				{CommitID: "00000001"},
				{CommitID: "00000002"},
				{CommitID: "00000003"},
				{CommitID: "00000004"},
			},
			prs: fezzik_types.PullRequestConnection{
				Nodes: &fezzik_types.PullRequestsViewerPullRequestsNodes{
					{
						Id:              "2",
						HeadRefName:     "spr/master/00000002",
						BaseRefName:     "master",
						MergeQueueEntry: &fezzik_types.PullRequestsViewerPullRequestsNodesMergeQueueEntry{Id: "020"},
						Commits: fezzik_types.PullRequestsViewerPullRequestsNodesCommits{
							Nodes: &fezzik_types.PullRequestsViewerPullRequestsNodesCommitsNodes{
								{
									fezzik_types.PullRequestsViewerPullRequestsNodesCommitsNodesCommit{Oid: "1", MessageBody: "commit-id:1"},
								},
								{
									fezzik_types.PullRequestsViewerPullRequestsNodesCommitsNodesCommit{Oid: "2", MessageBody: "commit-id:2"},
								},
							},
						},
					},
					{
						Id:          "3",
						HeadRefName: "spr/master/00000003",
						BaseRefName: "spr/master/00000002",
						Commits: fezzik_types.PullRequestsViewerPullRequestsNodesCommits{
							Nodes: &fezzik_types.PullRequestsViewerPullRequestsNodesCommitsNodes{
								{
									fezzik_types.PullRequestsViewerPullRequestsNodesCommitsNodesCommit{Oid: "3", MessageBody: "commit-id:3"},
								},
							},
						},
					},
				},
			},
			expect: []*github.PullRequest{
				{
					ID:         "2",
					FromBranch: "spr/master/00000002",
					ToBranch:   "master",
					Commit: git.Commit{
						CommitID:   "00000002",
						CommitHash: "2",
						Body:       "commit-id:2",
					},
					InQueue: true,
					Commits: []git.Commit{
						{CommitID: "1", CommitHash: "1", Body: "commit-id:1"},
						{CommitID: "2", CommitHash: "2", Body: "commit-id:2"},
					},
					MergeStatus: github.PullRequestMergeStatus{
						ChecksPass: github.CheckStatusPass,
					},
				},
				{
					ID:         "3",
					FromBranch: "spr/master/00000003",
					ToBranch:   "spr/master/00000002",
					Commit: git.Commit{
						CommitID:   "00000003",
						CommitHash: "3",
						Body:       "commit-id:3",
					},
					Commits: []git.Commit{
						{CommitID: "3", CommitHash: "3", Body: "commit-id:3"},
					},
					MergeStatus: github.PullRequestMergeStatus{
						ChecksPass: github.CheckStatusPass,
					},
				},
			},
		},
		{
			name:    "Empty",
			commits: []git.Commit{},
			prs:     fezzik_types.PullRequestConnection{},
			expect:  []*github.PullRequest{},
		},
		{
			name:    "FirstCommit",
			commits: []git.Commit{{CommitID: "00000001"}},
			prs:     fezzik_types.PullRequestConnection{},
			expect:  []*github.PullRequest{},
		},
		{
			name: "SecondCommit",
			commits: []git.Commit{
				{CommitID: "00000001"},
				{CommitID: "00000002"},
			},
			prs: fezzik_types.PullRequestConnection{
				Nodes: &fezzik_types.PullRequestsViewerPullRequestsNodes{
					{
						Id:          "1",
						HeadRefName: "spr/master/00000001",
						BaseRefName: "master",
						Commits: fezzik_types.PullRequestsViewerPullRequestsNodesCommits{
							Nodes: &fezzik_types.PullRequestsViewerPullRequestsNodesCommitsNodes{
								{
									fezzik_types.PullRequestsViewerPullRequestsNodesCommitsNodesCommit{Oid: "1"},
								},
							},
						},
					},
				},
			},
			expect: []*github.PullRequest{
				{
					ID:         "1",
					FromBranch: "spr/master/00000001",
					ToBranch:   "master",
					Commit: git.Commit{
						CommitID:   "00000001",
						CommitHash: "1",
					},
					MergeStatus: github.PullRequestMergeStatus{
						ChecksPass: github.CheckStatusPass,
					},
				},
			},
		},
		{
			name: "ThirdCommit",
			commits: []git.Commit{
				{CommitID: "00000001"},
				{CommitID: "00000002"},
				{CommitID: "00000003"},
			},
			prs: fezzik_types.PullRequestConnection{
				Nodes: &fezzik_types.PullRequestsViewerPullRequestsNodes{
					{
						Id:          "1",
						HeadRefName: "spr/master/00000001",
						BaseRefName: "master",
						Commits: fezzik_types.PullRequestsViewerPullRequestsNodesCommits{
							Nodes: &fezzik_types.PullRequestsViewerPullRequestsNodesCommitsNodes{
								{
									fezzik_types.PullRequestsViewerPullRequestsNodesCommitsNodesCommit{Oid: "1"},
								},
							},
						},
					},
					{
						Id:          "2",
						HeadRefName: "spr/master/00000002",
						BaseRefName: "spr/master/00000001",
						Commits: fezzik_types.PullRequestsViewerPullRequestsNodesCommits{
							Nodes: &fezzik_types.PullRequestsViewerPullRequestsNodesCommitsNodes{
								{
									fezzik_types.PullRequestsViewerPullRequestsNodesCommitsNodesCommit{Oid: "2"},
								},
							},
						},
					},
				},
			},
			expect: []*github.PullRequest{
				{
					ID:         "1",
					FromBranch: "spr/master/00000001",
					ToBranch:   "master",
					Commit: git.Commit{
						CommitID:   "00000001",
						CommitHash: "1",
					},
					MergeStatus: github.PullRequestMergeStatus{
						ChecksPass: github.CheckStatusPass,
					},
				},
				{
					ID:         "2",
					FromBranch: "spr/master/00000002",
					ToBranch:   "spr/master/00000001",
					Commit: git.Commit{
						CommitID:   "00000002",
						CommitHash: "2",
					},
					MergeStatus: github.PullRequestMergeStatus{
						ChecksPass: github.CheckStatusPass,
					},
				},
			},
		},
		{
			name:    "RemoveOnlyCommit",
			commits: []git.Commit{},
			prs: fezzik_types.PullRequestConnection{
				Nodes: &fezzik_types.PullRequestsViewerPullRequestsNodes{
					{
						Id:          "1",
						HeadRefName: "spr/master/00000001",
						BaseRefName: "master",
						Commits: fezzik_types.PullRequestsViewerPullRequestsNodesCommits{
							Nodes: &fezzik_types.PullRequestsViewerPullRequestsNodesCommitsNodes{
								{
									fezzik_types.PullRequestsViewerPullRequestsNodesCommitsNodesCommit{Oid: "1"},
								},
							},
						},
					},
				},
			},
			expect: []*github.PullRequest{},
		},
		{
			name: "RemoveTopCommit",
			commits: []git.Commit{
				{CommitID: "00000001"},
				{CommitID: "00000002"},
			},
			prs: fezzik_types.PullRequestConnection{
				Nodes: &fezzik_types.PullRequestsViewerPullRequestsNodes{
					{
						Id:          "1",
						HeadRefName: "spr/master/00000001",
						BaseRefName: "master",
						Commits: fezzik_types.PullRequestsViewerPullRequestsNodesCommits{
							Nodes: &fezzik_types.PullRequestsViewerPullRequestsNodesCommitsNodes{
								{
									fezzik_types.PullRequestsViewerPullRequestsNodesCommitsNodesCommit{Oid: "1"},
								},
							},
						},
					},
					{
						Id:          "3",
						HeadRefName: "spr/master/00000003",
						BaseRefName: "spr/master/00000002",
						Commits: fezzik_types.PullRequestsViewerPullRequestsNodesCommits{
							Nodes: &fezzik_types.PullRequestsViewerPullRequestsNodesCommitsNodes{
								{
									fezzik_types.PullRequestsViewerPullRequestsNodesCommitsNodesCommit{Oid: "2"},
								},
							},
						},
					},
					{
						Id:          "2",
						HeadRefName: "spr/master/00000002",
						BaseRefName: "spr/master/00000001",
						Commits: fezzik_types.PullRequestsViewerPullRequestsNodesCommits{
							Nodes: &fezzik_types.PullRequestsViewerPullRequestsNodesCommitsNodes{
								{
									fezzik_types.PullRequestsViewerPullRequestsNodesCommitsNodesCommit{Oid: "2"},
								},
							},
						},
					},
				},
			},
			expect: []*github.PullRequest{
				{
					ID:         "1",
					FromBranch: "spr/master/00000001",
					ToBranch:   "master",
					Commit: git.Commit{
						CommitID:   "00000001",
						CommitHash: "1",
					},
					MergeStatus: github.PullRequestMergeStatus{
						ChecksPass: github.CheckStatusPass,
					},
				},
				{
					ID:         "2",
					FromBranch: "spr/master/00000002",
					ToBranch:   "spr/master/00000001",
					Commit: git.Commit{
						CommitID:   "00000002",
						CommitHash: "2",
					},
					MergeStatus: github.PullRequestMergeStatus{
						ChecksPass: github.CheckStatusPass,
					},
				},
			},
		},
		{
			name: "RemoveMiddleCommit",
			commits: []git.Commit{
				{CommitID: "00000001"},
				{CommitID: "00000003"},
			},
			prs: fezzik_types.PullRequestConnection{
				Nodes: &fezzik_types.PullRequestsViewerPullRequestsNodes{
					{
						Id:          "1",
						HeadRefName: "spr/master/00000001",
						BaseRefName: "master",
						Commits: fezzik_types.PullRequestsViewerPullRequestsNodesCommits{
							Nodes: &fezzik_types.PullRequestsViewerPullRequestsNodesCommitsNodes{
								{
									fezzik_types.PullRequestsViewerPullRequestsNodesCommitsNodesCommit{Oid: "1"},
								},
							},
						},
					},
					{
						Id:          "2",
						HeadRefName: "spr/master/00000002",
						BaseRefName: "spr/master/00000001",
						Commits: fezzik_types.PullRequestsViewerPullRequestsNodesCommits{
							Nodes: &fezzik_types.PullRequestsViewerPullRequestsNodesCommitsNodes{
								{
									fezzik_types.PullRequestsViewerPullRequestsNodesCommitsNodesCommit{Oid: "2"},
								},
							},
						},
					},
					{
						Id:          "3",
						HeadRefName: "spr/master/00000003",
						BaseRefName: "spr/master/00000002",
						Commits: fezzik_types.PullRequestsViewerPullRequestsNodesCommits{
							Nodes: &fezzik_types.PullRequestsViewerPullRequestsNodesCommitsNodes{
								{
									fezzik_types.PullRequestsViewerPullRequestsNodesCommitsNodesCommit{Oid: "3"},
								},
							},
						},
					},
				},
			},
			expect: []*github.PullRequest{
				{
					ID:         "1",
					FromBranch: "spr/master/00000001",
					ToBranch:   "master",
					Commit: git.Commit{
						CommitID:   "00000001",
						CommitHash: "1",
					},
					MergeStatus: github.PullRequestMergeStatus{
						ChecksPass: github.CheckStatusPass,
					},
				},
				{
					ID:         "2",
					FromBranch: "spr/master/00000002",
					ToBranch:   "spr/master/00000001",
					Commit: git.Commit{
						CommitID:   "00000002",
						CommitHash: "2",
					},
					MergeStatus: github.PullRequestMergeStatus{
						ChecksPass: github.CheckStatusPass,
					},
				},
				{
					ID:         "3",
					FromBranch: "spr/master/00000003",
					ToBranch:   "spr/master/00000002",
					Commit: git.Commit{
						CommitID:   "00000003",
						CommitHash: "3",
					},
					MergeStatus: github.PullRequestMergeStatus{
						ChecksPass: github.CheckStatusPass,
					},
				},
			},
		},
		{
			name: "RemoveBottomCommit",
			commits: []git.Commit{
				{CommitID: "00000002"},
				{CommitID: "00000003"},
			},
			prs: fezzik_types.PullRequestConnection{
				Nodes: &fezzik_types.PullRequestsViewerPullRequestsNodes{
					{
						Id:          "1",
						HeadRefName: "spr/master/00000001",
						BaseRefName: "master",
						Commits: fezzik_types.PullRequestsViewerPullRequestsNodesCommits{
							Nodes: &fezzik_types.PullRequestsViewerPullRequestsNodesCommitsNodes{
								{
									fezzik_types.PullRequestsViewerPullRequestsNodesCommitsNodesCommit{Oid: "1"},
								},
							},
						},
					},
					{
						Id:          "2",
						HeadRefName: "spr/master/00000002",
						BaseRefName: "spr/master/00000001",
						Commits: fezzik_types.PullRequestsViewerPullRequestsNodesCommits{
							Nodes: &fezzik_types.PullRequestsViewerPullRequestsNodesCommitsNodes{
								{
									fezzik_types.PullRequestsViewerPullRequestsNodesCommitsNodesCommit{Oid: "2"},
								},
							},
						},
					},
					{
						Id:          "3",
						HeadRefName: "spr/master/00000003",
						BaseRefName: "spr/master/00000002",
						Commits: fezzik_types.PullRequestsViewerPullRequestsNodesCommits{
							Nodes: &fezzik_types.PullRequestsViewerPullRequestsNodesCommitsNodes{
								{
									fezzik_types.PullRequestsViewerPullRequestsNodesCommitsNodesCommit{Oid: "3"},
								},
							},
						},
					},
				},
			},
			expect: []*github.PullRequest{
				{
					ID:         "1",
					FromBranch: "spr/master/00000001",
					ToBranch:   "master",
					Commit: git.Commit{
						CommitID:   "00000001",
						CommitHash: "1",
					},
					MergeStatus: github.PullRequestMergeStatus{
						ChecksPass: github.CheckStatusPass,
					},
				},

				{
					ID:         "2",
					FromBranch: "spr/master/00000002",
					ToBranch:   "spr/master/00000001",
					Commit: git.Commit{
						CommitID:   "00000002",
						CommitHash: "2",
					},
					MergeStatus: github.PullRequestMergeStatus{
						ChecksPass: github.CheckStatusPass,
					},
				},
				{
					ID:         "3",
					FromBranch: "spr/master/00000003",
					ToBranch:   "spr/master/00000002",
					Commit: git.Commit{
						CommitID:   "00000003",
						CommitHash: "3",
					},
					MergeStatus: github.PullRequestMergeStatus{
						ChecksPass: github.CheckStatusPass,
					},
				},
			},
		},
	}

	for _, tc := range tests {
		repoConfig := &config.RepoConfig{}
		t.Run(tc.name, func(t *testing.T) {
			actual := matchPullRequestStack(repoConfig, "spr", "master", tc.commits, tc.prs)
			require.Equal(t, tc.expect, actual)
		})
	}
}

func TestComputeRequiredCheckStatus(t *testing.T) {
	strPtr := func(s string) *string { return &s }

	tests := []struct {
		name           string
		contexts       []checkContextNode
		requiredChecks map[string]bool
		expect         github.CheckStatus
	}{
		// === Basic cases ===
		{
			name:           "NoContexts_NoRequired",
			contexts:       []checkContextNode{},
			requiredChecks: map[string]bool{"ci": true},
			expect:         github.CheckStatusPending, // required check hasn't reported yet
		},
		{
			name:           "NoContexts_EmptyRequired",
			contexts:       []checkContextNode{},
			requiredChecks: map[string]bool{},
			expect:         github.CheckStatusPass,
		},

		// === CheckRun states ===
		{
			name: "CheckRun_RequiredPasses",
			contexts: []checkContextNode{
				{TypeName: "CheckRun", Name: "ci", Status: "COMPLETED", Conclusion: strPtr("SUCCESS")},
			},
			requiredChecks: map[string]bool{"ci": true},
			expect:         github.CheckStatusPass,
		},
		{
			name: "CheckRun_RequiredFails",
			contexts: []checkContextNode{
				{TypeName: "CheckRun", Name: "ci", Status: "COMPLETED", Conclusion: strPtr("FAILURE")},
			},
			requiredChecks: map[string]bool{"ci": true},
			expect:         github.CheckStatusFail,
		},
		{
			name: "CheckRun_RequiredPending",
			contexts: []checkContextNode{
				{TypeName: "CheckRun", Name: "ci", Status: "IN_PROGRESS"},
			},
			requiredChecks: map[string]bool{"ci": true},
			expect:         github.CheckStatusPending,
		},
		{
			name: "CheckRun_RequiredQueued",
			contexts: []checkContextNode{
				{TypeName: "CheckRun", Name: "ci", Status: "QUEUED"},
			},
			requiredChecks: map[string]bool{"ci": true},
			expect:         github.CheckStatusPending,
		},
		{
			name: "CheckRun_NeutralPasses",
			contexts: []checkContextNode{
				{TypeName: "CheckRun", Name: "ci", Status: "COMPLETED", Conclusion: strPtr("NEUTRAL")},
			},
			requiredChecks: map[string]bool{"ci": true},
			expect:         github.CheckStatusPass,
		},
		{
			name: "CheckRun_SkippedPasses",
			contexts: []checkContextNode{
				{TypeName: "CheckRun", Name: "ci", Status: "COMPLETED", Conclusion: strPtr("SKIPPED")},
			},
			requiredChecks: map[string]bool{"ci": true},
			expect:         github.CheckStatusPass,
		},
		{
			name: "CheckRun_TimedOut",
			contexts: []checkContextNode{
				{TypeName: "CheckRun", Name: "ci", Status: "COMPLETED", Conclusion: strPtr("TIMED_OUT")},
			},
			requiredChecks: map[string]bool{"ci": true},
			expect:         github.CheckStatusFail,
		},
		{
			name: "CheckRun_Cancelled",
			contexts: []checkContextNode{
				{TypeName: "CheckRun", Name: "ci", Status: "COMPLETED", Conclusion: strPtr("CANCELLED")},
			},
			requiredChecks: map[string]bool{"ci": true},
			expect:         github.CheckStatusFail,
		},
		{
			name: "CheckRun_NilConclusion",
			contexts: []checkContextNode{
				{TypeName: "CheckRun", Name: "ci", Status: "COMPLETED", Conclusion: nil},
			},
			requiredChecks: map[string]bool{"ci": true},
			expect:         github.CheckStatusFail,
		},

		// === StatusContext states ===
		{
			name: "StatusContext_RequiredSuccess",
			contexts: []checkContextNode{
				{TypeName: "StatusContext", Context: "ci/build", State: "SUCCESS"},
			},
			requiredChecks: map[string]bool{"ci/build": true},
			expect:         github.CheckStatusPass,
		},
		{
			name: "StatusContext_RequiredFailure",
			contexts: []checkContextNode{
				{TypeName: "StatusContext", Context: "ci/build", State: "FAILURE"},
			},
			requiredChecks: map[string]bool{"ci/build": true},
			expect:         github.CheckStatusFail,
		},
		{
			name: "StatusContext_RequiredPending",
			contexts: []checkContextNode{
				{TypeName: "StatusContext", Context: "ci/build", State: "PENDING"},
			},
			requiredChecks: map[string]bool{"ci/build": true},
			expect:         github.CheckStatusPending,
		},
		{
			name: "StatusContext_RequiredExpected",
			contexts: []checkContextNode{
				{TypeName: "StatusContext", Context: "ci/build", State: "EXPECTED"},
			},
			requiredChecks: map[string]bool{"ci/build": true},
			expect:         github.CheckStatusPending,
		},

		// === Filtering: non-required checks never block, failures are a warning ===
		{
			name: "NonRequiredFails_RequiredPasses",
			contexts: []checkContextNode{
				{TypeName: "CheckRun", Name: "required-ci", Status: "COMPLETED", Conclusion: strPtr("SUCCESS")},
				{TypeName: "CheckRun", Name: "optional-lint", Status: "COMPLETED", Conclusion: strPtr("FAILURE")},
			},
			requiredChecks: map[string]bool{"required-ci": true},
			expect:         github.CheckStatusWarning,
		},
		{
			name: "NonRequiredPending_RequiredPasses",
			contexts: []checkContextNode{
				{TypeName: "CheckRun", Name: "required-ci", Status: "COMPLETED", Conclusion: strPtr("SUCCESS")},
				{TypeName: "CheckRun", Name: "optional-lint", Status: "IN_PROGRESS"},
			},
			requiredChecks: map[string]bool{"required-ci": true},
			expect:         github.CheckStatusPass,
		},
		{
			name: "NonRequiredFails_RequiredPending",
			contexts: []checkContextNode{
				{TypeName: "CheckRun", Name: "required-ci", Status: "IN_PROGRESS"},
				{TypeName: "CheckRun", Name: "optional-lint", Status: "COMPLETED", Conclusion: strPtr("FAILURE")},
			},
			requiredChecks: map[string]bool{"required-ci": true},
			expect:         github.CheckStatusPending,
		},
		{
			name: "MixedTypes_NonRequiredFails_RequiredPasses",
			contexts: []checkContextNode{
				{TypeName: "CheckRun", Name: "required-ci", Status: "COMPLETED", Conclusion: strPtr("SUCCESS")},
				{TypeName: "StatusContext", Context: "ci/build", State: "SUCCESS"},
				{TypeName: "CheckRun", Name: "optional-lint", Status: "COMPLETED", Conclusion: strPtr("FAILURE")},
				{TypeName: "StatusContext", Context: "optional/coverage", State: "FAILURE"},
			},
			requiredChecks: map[string]bool{"required-ci": true, "ci/build": true},
			expect:         github.CheckStatusWarning,
		},
		{
			name: "MixedTypes_RequiredFails",
			contexts: []checkContextNode{
				{TypeName: "CheckRun", Name: "required-ci", Status: "COMPLETED", Conclusion: strPtr("SUCCESS")},
				{TypeName: "StatusContext", Context: "ci/build", State: "FAILURE"},
			},
			requiredChecks: map[string]bool{"required-ci": true, "ci/build": true},
			expect:         github.CheckStatusFail,
		},

		// === Precedence ===
		{
			name: "FailTakesPrecedenceOverPending",
			contexts: []checkContextNode{
				{TypeName: "CheckRun", Name: "ci-1", Status: "COMPLETED", Conclusion: strPtr("FAILURE")},
				{TypeName: "CheckRun", Name: "ci-2", Status: "IN_PROGRESS"},
			},
			requiredChecks: map[string]bool{"ci-1": true, "ci-2": true},
			expect:         github.CheckStatusFail,
		},
		{
			name: "AllRequiredPass_MultipleChecks",
			contexts: []checkContextNode{
				{TypeName: "CheckRun", Name: "ci-1", Status: "COMPLETED", Conclusion: strPtr("SUCCESS")},
				{TypeName: "CheckRun", Name: "ci-2", Status: "COMPLETED", Conclusion: strPtr("SUCCESS")},
				{TypeName: "CheckRun", Name: "ci-3", Status: "COMPLETED", Conclusion: strPtr("SUCCESS")},
			},
			requiredChecks: map[string]bool{"ci-1": true, "ci-2": true, "ci-3": true},
			expect:         github.CheckStatusPass,
		},

		// === Missing required check (not yet reported) ===
		{
			name: "RequiredCheckNotReportedYet",
			contexts: []checkContextNode{
				{TypeName: "CheckRun", Name: "ci-1", Status: "COMPLETED", Conclusion: strPtr("SUCCESS")},
			},
			requiredChecks: map[string]bool{"ci-1": true, "ci-2": true},
			expect:         github.CheckStatusPending,
		},

		// === Real-world scenario: Semaphore + GitHub Actions ===
		{
			name: "RealWorld_SemaphoreFails_OnlyItRequired",
			contexts: []checkContextNode{
				{TypeName: "StatusContext", Context: "ci/semaphoreci/pr: test", State: "FAILURE"},
				{TypeName: "CheckRun", Name: "PR Policy Review", Status: "COMPLETED", Conclusion: strPtr("NEUTRAL")},
				{TypeName: "CheckRun", Name: "review", Status: "COMPLETED", Conclusion: strPtr("CANCELLED")},
				{TypeName: "CheckRun", Name: "repl-controller", Status: "COMPLETED", Conclusion: strPtr("SUCCESS")},
			},
			requiredChecks: map[string]bool{"ci/semaphoreci/pr: test": true},
			expect:         github.CheckStatusFail,
		},
		{
			name: "RealWorld_SemaphorePasses_OthersFail",
			contexts: []checkContextNode{
				{TypeName: "StatusContext", Context: "ci/semaphoreci/pr: test", State: "SUCCESS"},
				{TypeName: "CheckRun", Name: "review", Status: "COMPLETED", Conclusion: strPtr("CANCELLED")},
				{TypeName: "CheckRun", Name: "optional", Status: "COMPLETED", Conclusion: strPtr("FAILURE")},
			},
			requiredChecks: map[string]bool{"ci/semaphoreci/pr: test": true},
			expect:         github.CheckStatusWarning,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			actual := computeRequiredCheckStatus(tc.contexts, tc.requiredChecks)
			require.Equal(t, tc.expect, actual)
		})
	}
}

func TestFetchRequiredChecksStatus(t *testing.T) {
	strPtr := func(s string) *string { return &s }
	contexts := func(nodes ...checkContextNode) map[string]any {
		return map[string]any{"commits": map[string]any{"nodes": []any{
			map[string]any{"commit": map[string]any{"statusCheckRollup": map[string]any{
				"contexts": map[string]any{"nodes": nodes},
			}}},
		}}}
	}
	passed := func(name string) checkContextNode {
		return checkContextNode{TypeName: "CheckRun", Name: name, Status: "COMPLETED", Conclusion: strPtr("SUCCESS")}
	}
	failed := func(name string) checkContextNode {
		return checkContextNode{TypeName: "CheckRun", Name: name, Status: "COMPLETED", Conclusion: strPtr("FAILURE")}
	}
	target := func(required ...string) map[string]any {
		return map[string]any{"ref": map[string]any{"refUpdateRule": map[string]any{
			"requiredStatusCheckContexts": required,
		}}}
	}
	prs := []*github.PullRequest{{ID: "PR_1", Number: 1}, {ID: "PR_2", Number: 2}}
	data := func(target any) map[string]any {
		return map[string]any{
			"target": target,
			"pr_1":   contexts(passed("build"), failed("lint")),
			"pr_2":   contexts(failed("build"), passed("lint")),
		}
	}

	tests := []struct {
		name            string
		configured      []string
		data            map[string]any
		expectTargetRef bool
		expect          map[int]github.CheckStatus
	}{
		{
			name:            "RequiredFromTargetBranch",
			data:            data(target("build")),
			expectTargetRef: true,
			expect:          map[int]github.CheckStatus{1: github.CheckStatusWarning, 2: github.CheckStatusFail},
		},
		{
			name:            "ConfiguredOverridesTargetBranch",
			configured:      []string{"lint"},
			data:            data(nil),
			expectTargetRef: false,
			expect:          map[int]github.CheckStatus{1: github.CheckStatusFail, 2: github.CheckStatusWarning},
		},
		{
			name:            "NoProtection",
			data:            data(map[string]any{"ref": map[string]any{"refUpdateRule": nil}}),
			expectTargetRef: true,
			expect:          nil,
		},
		{
			name:            "UnknownTargetBranch",
			data:            data(map[string]any{"ref": nil}),
			expectTargetRef: true,
			expect:          nil,
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var query string
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				var req graphqlRequest
				require.NoError(t, json.NewDecoder(r.Body).Decode(&req))
				query = req.Query
				require.NoError(t, json.NewEncoder(w).Encode(map[string]any{"data": tc.data}))
			}))
			defer server.Close()

			c := &client{
				config: &config.Config{Repo: &config.RepoConfig{
					GitHubRepoOwner: "owner",
					GitHubRepoName:  "repo",
					GitHubBranch:    "main",
					RequiredChecks:  tc.configured,
				}, User: &config.UserConfig{}},
				graphqlEndpoint: server.URL,
				httpClient:      server.Client(),
			}

			actual := c.fetchRequiredChecksStatus(context.Background(), prs)
			require.Equal(t, tc.expect, actual)
			require.Equal(t, tc.expectTargetRef, strings.Contains(query, `ref(qualifiedName: "refs/heads/main")`))
		})
	}
}
