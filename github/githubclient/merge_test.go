package githubclient

import (
	"context"
	"errors"
	"testing"

	"github.com/ejoffe/spr/config"
	"github.com/ejoffe/spr/git"
	"github.com/ejoffe/spr/github"
	"github.com/ejoffe/spr/github/githubclient/gen/genclient"
	"github.com/stretchr/testify/require"
)

func TestMergePullRequest_PassesAuthorEmailFromGitConfig(t *testing.T) {
	tests := []struct {
		name        string
		mergeQueue  bool
		gitEmail    string
		gitErr      error
		wantEmailFn func(t *testing.T, gotMerge *string, gotAuto *string)
	}{
		{
			name:     "merge: forwards user.email as AuthorEmail",
			gitEmail: "kdavin@gradle.com\n",
			wantEmailFn: func(t *testing.T, gotMerge *string, gotAuto *string) {
				require.NotNil(t, gotMerge, "MergePullRequestInput.AuthorEmail should be set")
				require.Equal(t, "kdavin@gradle.com", *gotMerge)
				require.Nil(t, gotAuto, "AutoMerge path should not have been used")
			},
		},
		{
			name:     "merge: trims whitespace and surrounding noise",
			gitEmail: "  kdavin@gradle.com  ",
			wantEmailFn: func(t *testing.T, gotMerge *string, gotAuto *string) {
				require.NotNil(t, gotMerge)
				require.Equal(t, "kdavin@gradle.com", *gotMerge)
			},
		},
		{
			name:     "merge: leaves AuthorEmail nil when git config is empty",
			gitEmail: "",
			wantEmailFn: func(t *testing.T, gotMerge *string, gotAuto *string) {
				require.Nil(t, gotMerge, "AuthorEmail should be omitted when no email is configured")
			},
		},
		{
			name:     "merge: leaves AuthorEmail nil when git config returns an error",
			gitEmail: "",
			gitErr:   errors.New("git config user.email exited with status 1"),
			wantEmailFn: func(t *testing.T, gotMerge *string, gotAuto *string) {
				require.Nil(t, gotMerge)
			},
		},
		{
			name:       "auto-merge: forwards user.email as AuthorEmail",
			mergeQueue: true,
			gitEmail:   "kdavin@gradle.com",
			wantEmailFn: func(t *testing.T, gotMerge *string, gotAuto *string) {
				require.Nil(t, gotMerge, "Standard merge path should not have been used")
				require.NotNil(t, gotAuto, "EnablePullRequestAutoMergeInput.AuthorEmail should be set")
				require.Equal(t, "kdavin@gradle.com", *gotAuto)
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			fakeAPI := &fakeGenClient{}
			fakeGit := &fakeGitConfigEmail{response: tc.gitEmail, err: tc.gitErr}

			c := &client{
				config: &config.Config{
					Repo: &config.RepoConfig{MergeQueue: tc.mergeQueue},
					User: &config.UserConfig{},
				},
				api: fakeAPI,
			}

			pr := &github.PullRequest{ID: "PR_ID", Number: 42, Title: "test"}

			c.MergePullRequest(context.Background(), fakeGit, pr, genclient.PullRequestMergeMethod_MERGE)

			require.Equal(t, "config user.email", fakeGit.gotArgs,
				"client should query the git user.email config key")

			var gotMergeEmail, gotAutoEmail *string
			if fakeAPI.mergeCalled {
				gotMergeEmail = fakeAPI.mergeInput.AuthorEmail
				require.Equal(t, "PR_ID", fakeAPI.mergeInput.PullRequestId)
			}
			if fakeAPI.autoMergeCalled {
				gotAutoEmail = fakeAPI.autoMergeInput.AuthorEmail
				require.Equal(t, "PR_ID", fakeAPI.autoMergeInput.PullRequestId)
			}
			tc.wantEmailFn(t, gotMergeEmail, gotAutoEmail)
		})
	}
}

// fakeGitConfigEmail records the args of the single git call MergePullRequest is
// expected to make and serves a canned response for `git config user.email`.
type fakeGitConfigEmail struct {
	response string
	err      error
	gotArgs  string
}

func (f *fakeGitConfigEmail) Git(args string, output *string) error {
	f.gotArgs = args
	if output != nil {
		*output = f.response
	}
	return f.err
}

func (f *fakeGitConfigEmail) GitWithEditor(args string, output *string, editorCmd string) error {
	return f.Git(args, output)
}

func (f *fakeGitConfigEmail) MustGit(args string, output *string) {
	if err := f.Git(args, output); err != nil {
		panic(err)
	}
}

func (f *fakeGitConfigEmail) RootDir() string { return "" }

func (f *fakeGitConfigEmail) DeleteRemoteBranch(_ context.Context, _ string) error { return nil }

var _ git.GitInterface = (*fakeGitConfigEmail)(nil)

// fakeGenClient captures the input of the two merge mutations under test and is
// a no-op for every other operation on the genclient.Client interface.
// Embedding a nil genclient.Client makes unrelated methods panic if accidentally
// reached, which keeps the tests honest about what the merge path actually calls.
type fakeGenClient struct {
	genclient.Client

	mergeCalled     bool
	mergeInput      genclient.MergePullRequestInput
	autoMergeCalled bool
	autoMergeInput  genclient.EnablePullRequestAutoMergeInput
}

func (f *fakeGenClient) MergePullRequest(_ context.Context, input genclient.MergePullRequestInput) (*genclient.MergePullRequestResponse, error) {
	f.mergeCalled = true
	f.mergeInput = input
	return &genclient.MergePullRequestResponse{}, nil
}

func (f *fakeGenClient) AutoMergePullRequest(_ context.Context, input genclient.EnablePullRequestAutoMergeInput) (*genclient.AutoMergePullRequestResponse, error) {
	f.autoMergeCalled = true
	f.autoMergeInput = input
	return &genclient.AutoMergePullRequestResponse{}, nil
}
