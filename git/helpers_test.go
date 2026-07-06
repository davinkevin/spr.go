package git

import (
	"context"
	"strings"
	"testing"

	"github.com/ejoffe/spr/config"
	"github.com/stretchr/testify/assert"
)

func TestBranchNameRegex(t *testing.T) {
	tests := []struct {
		prefix string
		input  string
		branch string
		commit string
	}{
		{prefix: "spr", input: "spr/b1/deadbeef", branch: "b1", commit: "deadbeef"},
		{prefix: "spr", input: "spr/main/abcd1234", branch: "main", commit: "abcd1234"},
		{prefix: "custom", input: "custom/main/deadbeef", branch: "main", commit: "deadbeef"},
		{prefix: "my-team", input: "my-team/develop/abcd1234", branch: "develop", commit: "abcd1234"},
	}

	for _, tc := range tests {
		t.Run(tc.input, func(t *testing.T) {
			matches := BranchNameRegex(tc.prefix).FindStringSubmatch(tc.input)
			assert.NotNil(t, matches)
			assert.Equal(t, tc.branch, matches[1])
			assert.Equal(t, tc.commit, matches[2])
		})
	}
}

func TestBranchNameRegexNoMatch(t *testing.T) {
	tests := []struct {
		prefix string
		input  string
	}{
		{prefix: "spr", input: "other/main/deadbeef"},
		{prefix: "custom", input: "spr/main/deadbeef"},
	}

	for _, tc := range tests {
		t.Run(tc.input, func(t *testing.T) {
			matches := BranchNameRegex(tc.prefix).FindStringSubmatch(tc.input)
			assert.Nil(t, matches)
		})
	}
}

func TestBranchNameFromCommit(t *testing.T) {
	tests := []struct {
		name     string
		prefix   string
		branch   string
		commitID string
		expected string
	}{
		{
			name:     "default prefix",
			prefix:   "spr",
			branch:   "main",
			commitID: "deadbeef",
			expected: "spr/main/deadbeef",
		},
		{
			name:     "custom prefix",
			prefix:   "my-team",
			branch:   "develop",
			commitID: "abcd1234",
			expected: "my-team/develop/abcd1234",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			cfg := config.EmptyConfig()
			cfg.User.BranchPrefix = tc.prefix
			cfg.Repo.GitHubBranch = tc.branch

			commit := Commit{CommitID: tc.commitID}
			result := BranchNameFromCommit(cfg, commit)
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestBranchNameFromCommitRepoOverride(t *testing.T) {
	cfg := config.EmptyConfig()
	cfg.User.BranchPrefix = "spr"
	cfg.Repo.BranchPrefix = "team-x"
	cfg.Repo.GitHubBranch = "main"

	commit := Commit{CommitID: "deadbeef"}
	result := BranchNameFromCommit(cfg, commit)
	assert.Equal(t, "team-x/main/deadbeef", result)
}

// mergeBaseGit is a minimal in-package fake of GitInterface used to exercise
// mergeBaseWithTarget without importing mockgit (which would create an import
// cycle with package git).
type mergeBaseGit struct {
	base    string
	gotArgs []string
}

func (f *mergeBaseGit) MustGit(args string, output *string) {
	f.gotArgs = append(f.gotArgs, args)
	if strings.HasPrefix(args, "merge-base") && output != nil {
		// real git prints the sha followed by a trailing newline
		*output = f.base + "\n"
	}
}
func (f *mergeBaseGit) Git(args string, output *string) error { f.MustGit(args, output); return nil }
func (f *mergeBaseGit) GitWithEditor(args string, output *string, editorCmd string) error {
	f.MustGit(args, output)
	return nil
}
func (f *mergeBaseGit) RootDir() string                                             { return "" }
func (f *mergeBaseGit) DeleteRemoteBranch(ctx context.Context, branch string) error { return nil }

func TestMergeBaseWithTarget(t *testing.T) {
	cfg := config.EmptyConfig()
	cfg.Repo.GitHubRemote = "origin"
	cfg.Repo.GitHubBranch = "master"

	f := &mergeBaseGit{base: "abc1234def5678"}
	got := mergeBaseWithTarget(cfg, f)

	assert.Equal(t, "abc1234def5678", got, "trailing newline should be trimmed")
	assert.Equal(t, []string{"merge-base origin/master HEAD"}, f.gotArgs,
		"reword must rebase onto the merge-base, not the target branch tip")
}
