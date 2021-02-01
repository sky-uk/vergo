package cmd_test

import (
	"github.com/stretchr/testify/assert"
	"sky.uk/vergo/cmd"
	"testing"
)

//nolint:scopelint,paralleltest
func TestShouldIgnore(t *testing.T) {
	testCases := []struct {
		tagPrefix string
		messages  []string
	}{
		{
			tagPrefix: "",
			messages: []string{
				"[vergo ignore] doc update",
				"[vergoignore] doc update",
				"[vergo:ignore] doc update",
				"[vergo-ignore] doc update",
				"[vergo_ignore] doc update",
				"[vergo/ignore] doc update",
				"[vergo\\ignore] doc update",
				"[vergoXignore] doc update",
				"@vergo@ignore doc update",
			},
		},
		{
			tagPrefix: "job-checker",
			messages: []string{
				"[vergo job-checker ignore] doc update",
				"[vergojob-checkerignore] doc update",
				"[vergo:job-checker:ignore] doc update",
				"[vergo-job-checker-ignore] doc update",
				"[vergo_job-checker_ignore] doc update",
				"[vergo/job-checker/ignore] doc update",
				"[vergo\\job-checker\\ignore] doc update",
				"[vergoXjob-checkerXignore] doc update",
				"@vergo@job-checker@ignore doc update",
			},
		},
	}
	for _, testCase := range testCases {
		for _, message := range testCase.messages {
			t.Run(testCase.tagPrefix+message, func(t *testing.T) {
				assert.True(t, cmd.IgnoreHintPresent(message, testCase.tagPrefix))
			})
		}
	}
}
