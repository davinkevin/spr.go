package config

import (
	"testing"

	"github.com/ejoffe/spr/github/githubclient/gen/genclient"
	"github.com/stretchr/testify/assert"
	"gopkg.in/yaml.v3"
)

func TestEmptyConfig(t *testing.T) {
	expect := &Config{
		Repo: &RepoConfig{},
		User: &UserConfig{},
		State: &InternalState{
			MergeCheckCommit: map[string]string{},
		},
	}
	actual := EmptyConfig()
	assert.Equal(t, expect, actual)
}

func TestDefaultConfig(t *testing.T) {
	expect := &Config{
		Repo: &RepoConfig{
			GitHubRepoOwner:       "",
			GitHubRepoName:        "",
			GitHubRemote:          "origin",
			GitHubBranch:          "main",
			GitHubHost:            "github.com",
			RequireChecks:         true,
			RequireApproval:       true,
			MergeMethod:           "rebase",
			PRTemplateType:        "stack",
			PRTemplatePath:        "",
			PRTemplateInsertStart: "",
			PRTemplateInsertEnd:   "",
			ShowPrTitlesInStack:   false,
		},
		User: &UserConfig{
			ShowPRLink:       true,
			LogGitCommands:   false,
			LogGitHubCalls:   false,
			StatusBitsHeader: true,
			StatusBitsEmojis: true,
			CreateDraftPRs:   "none",
			BranchPrefix:     "spr",
		},
		State: &InternalState{
			MergeCheckCommit: map[string]string{},
		},
	}
	actual := DefaultConfig()
	assert.Equal(t, expect, actual)
}

func TestMergeMethodHelper(t *testing.T) {
	for _, tc := range []struct {
		configValue string
		expected    genclient.PullRequestMergeMethod
	}{
		{
			configValue: "rebase",
			expected:    genclient.PullRequestMergeMethod_REBASE,
		},
		{
			configValue: "",
			expected:    genclient.PullRequestMergeMethod_REBASE,
		},
		{
			configValue: "Merge",
			expected:    genclient.PullRequestMergeMethod_MERGE,
		},
		{
			configValue: "SQUASH",
			expected:    genclient.PullRequestMergeMethod_SQUASH,
		},
	} {
		tcName := tc.configValue
		if tcName == "" {
			tcName = "<EMPTY>"
		}
		t.Run(tcName, func(t *testing.T) {
			config := &Config{Repo: &RepoConfig{MergeMethod: tc.configValue}}
			actual, err := config.MergeMethod()
			assert.NoError(t, err)
			assert.Equal(t, tc.expected, actual)
		})
	}
	t.Run("invalid", func(t *testing.T) {
		config := &Config{Repo: &RepoConfig{MergeMethod: "magic"}}
		actual, err := config.MergeMethod()
		assert.Error(t, err)
		assert.Empty(t, actual)
	})
}

func TestBranchPrefix(t *testing.T) {
	t.Run("returns user default when repo is empty", func(t *testing.T) {
		cfg := &Config{
			Repo: &RepoConfig{},
			User: &UserConfig{BranchPrefix: "spr"},
		}
		assert.Equal(t, "spr", cfg.BranchPrefix())
	})

	t.Run("repo overrides user", func(t *testing.T) {
		cfg := &Config{
			Repo: &RepoConfig{BranchPrefix: "team-x"},
			User: &UserConfig{BranchPrefix: "spr"},
		}
		assert.Equal(t, "team-x", cfg.BranchPrefix())
	})

	t.Run("falls back to user when repo not set", func(t *testing.T) {
		cfg := &Config{
			Repo: &RepoConfig{},
			User: &UserConfig{BranchPrefix: "custom"},
		}
		assert.Equal(t, "custom", cfg.BranchPrefix())
	})

	t.Run("strips trailing slash", func(t *testing.T) {
		cfg := &Config{
			Repo: &RepoConfig{BranchPrefix: "johndoe/spr/"},
			User: &UserConfig{BranchPrefix: "spr"},
		}
		assert.Equal(t, "johndoe/spr", cfg.BranchPrefix())
	})
}

func TestCreateDraftPRsYAMLBackwardCompat(t *testing.T) {
	for _, tc := range []struct {
		name     string
		yaml     string
		expected string
	}{
		{"legacy bool true", "createDraftPRs: true", "true"},
		{"legacy bool false", "createDraftPRs: false", "false"},
		{"string none", "createDraftPRs: none", "none"},
		{"string all", "createDraftPRs: all", "all"},
		{"string allExceptNext", "createDraftPRs: allExceptNext", "allExceptNext"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var cfg UserConfig
			err := yaml.Unmarshal([]byte(tc.yaml), &cfg)
			assert.NoError(t, err)
			assert.Equal(t, tc.expected, cfg.CreateDraftPRs)
		})
	}
}

func TestShouldDraftPR(t *testing.T) {
	for _, tc := range []struct {
		name            string
		configValue     string
		isClosestToBase bool
		expected        bool
	}{
		// "none" mode
		{"none - closest to base", "none", true, false},
		{"none - not closest to base", "none", false, false},

		// "all" mode
		{"all - closest to base", "all", true, true},
		{"all - not closest to base", "all", false, true},

		// "allExceptNext" mode
		{"allExceptNext - closest to base", "allExceptNext", true, false},
		{"allExceptNext - not closest to base", "allExceptNext", false, true},

		// backward compat: bool "true" maps to "all"
		{"true (legacy) - closest to base", "true", true, true},
		{"true (legacy) - not closest to base", "true", false, true},

		// backward compat: bool "false" maps to "none"
		{"false (legacy) - closest to base", "false", true, false},
		{"false (legacy) - not closest to base", "false", false, false},

		// empty string defaults to "none"
		{"empty string - closest to base", "", true, false},
		{"empty string - not closest to base", "", false, false},

		// case insensitivity
		{"ALLEXCEPTNEXT (uppercase)", "ALLEXCEPTNEXT", false, true},
		{"All (mixed case)", "All", true, true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			cfg := &Config{User: &UserConfig{CreateDraftPRs: tc.configValue}}
			actual := cfg.ShouldDraftPR(tc.isClosestToBase)
			assert.Equal(t, tc.expected, actual)
		})
	}
}

func TestNormalizeConfig(t *testing.T) {
	t.Run("PRTemplatePath provided sets PRTemplateType to custom", func(t *testing.T) {
		cfg := &Config{
			Repo: &RepoConfig{
				PRTemplateType: "stack",
				PRTemplatePath: "/path/to/template.md",
			},
		}
		cfg.Normalize()
		assert.Equal(t, "custom", cfg.Repo.PRTemplateType)
		assert.Equal(t, "/path/to/template.md", cfg.Repo.PRTemplatePath)
	})

	t.Run("PRTemplatePath empty does not change PRTemplateType", func(t *testing.T) {
		cfg := &Config{
			Repo: &RepoConfig{
				PRTemplateType: "stack",
				PRTemplatePath: "",
			},
		}
		cfg.Normalize()
		assert.Equal(t, "stack", cfg.Repo.PRTemplateType)
		assert.Equal(t, "", cfg.Repo.PRTemplatePath)
	})

	t.Run("PRTemplatePath provided overrides existing PRTemplateType", func(t *testing.T) {
		cfg := &Config{
			Repo: &RepoConfig{
				PRTemplateType: "why_what",
				PRTemplatePath: "/custom/template.md",
			},
		}
		cfg.Normalize()
		assert.Equal(t, "custom", cfg.Repo.PRTemplateType)
		assert.Equal(t, "/custom/template.md", cfg.Repo.PRTemplatePath)
	})

	t.Run("DefaultConfig with PRTemplatePath sets PRTemplateType to custom", func(t *testing.T) {
		cfg := DefaultConfig()
		cfg.Repo.PRTemplatePath = "/path/to/template.md"
		cfg.Normalize()
		assert.Equal(t, "custom", cfg.Repo.PRTemplateType)
	})
}
