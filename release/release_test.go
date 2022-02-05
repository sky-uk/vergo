package release_test

import (
	. "github.com/sky-uk/umc-shared/vergo/internal-test"
	"github.com/sky-uk/umc-shared/vergo/release"
	"github.com/stretchr/testify/assert"
	"testing"
)

//nolint:scopelint,paralleltest
func TestShouldVerifySkipReleaseHint(t *testing.T) {
	testCases := []struct {
		tagPrefix string
		messages  []string
	}{
		{
			tagPrefix: "",
			messages: []string{
				"[vergo:skip-release] doc update",
				"@vergo:skip-release@ doc update",
			},
		},
		{
			tagPrefix: "app",
			messages: []string{
				"[vergo:app:skip-release] doc update",
				"@vergo:app:skip-release@ doc update",
			},
		},
	}
	for _, testCase := range testCases {
		for _, message := range testCase.messages {
			t.Run(testCase.tagPrefix+message, func(t *testing.T) {
				r := NewTestRepo(t)
				assert.Nil(t, release.SkipHintPresent(r.Repo, testCase.tagPrefix))

				DoCommitWithMessage(t, r.Repo, "some content 1", message)
				assert.ErrorIs(t, release.SkipHintPresent(r.Repo, testCase.tagPrefix), release.ErrSkipRelease)

				DoCommitWithMessage(t, r.Repo, "some content 2", "another commit")
				assert.Nil(t, release.SkipHintPresent(r.Repo, testCase.tagPrefix))

				DoCommitWithMessage(t, r.Repo, "some content 3", "another commit"+message)
				assert.ErrorIs(t, release.SkipHintPresent(r.Repo, testCase.tagPrefix), release.ErrSkipRelease)
			})
		}
	}
}
