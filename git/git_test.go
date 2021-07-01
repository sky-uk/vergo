package git_test

import (
	"errors"
	"fmt"
	"github.com/Masterminds/semver"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/stretchr/testify/assert"
	"github.com/thoas/go-funk"
	"go.uber.org/atomic"
	"math"
	"sky.uk/vergo/bump"
	. "sky.uk/vergo/git"
	. "sky.uk/vergo/internal-test"
	"testing"
	"time"
)

//nolint:scopelint,paralleltest
func TestNoTag(t *testing.T) {
	prefixes := []string{"", "app", "application"}

	for _, prefix := range prefixes {
		t.Run(prefix, func(t *testing.T) {
			r := NewTestRepo(t)
			_, err := LatestRef(r.Repo, prefix)
			assert.Regexp(t, "no tag found", err)
		})

		t.Run(prefix, func(t *testing.T) {
			r := NewTestRepo(t)
			_, err := PreviousRef(r.Repo, prefix)
			assert.Regexp(t, "no tag found", err)
		})
	}
}

//nolint:scopelint,paralleltest
func TestTagExists(t *testing.T) {
	prefixes := []string{"", "app", "application"}

	for _, prefix := range prefixes {
		t.Run(prefix, func(t *testing.T) {
			r := NewTestRepo(t)

			err := CreateTag(r.Repo, "0.0.1", prefix, false)
			assert.Nil(t, err)
			found, err := TagExists(r.Repo, prefix+"0.0.1")
			assert.Nil(t, err)
			assert.True(t, found)
		})
	}
}

//nolint:scopelint,paralleltest
func TestTagAlreadyExists(t *testing.T) {
	prefixes := []string{"", "app", "application"}

	for _, prefix := range prefixes {
		t.Run(prefix, func(t *testing.T) {
			r := NewTestRepo(t)

			err := CreateTag(r.Repo, "0.0.1", prefix, false)
			assert.Nil(t, err)

			err = CreateTag(r.Repo, "0.0.1", prefix, false)
			assert.Regexp(t, "already exists", err)
		})
	}
}

//nolint:funlen,scopelint,paralleltest
func TestListRefs(t *testing.T) {
	prefixes := []string{"", "app", "application"}
	r := NewTestRepo(t)

	t.Run("tag does not exist", func(t *testing.T) {
		a, err := ListRefs(r.Repo, "banana", ASC, math.MaxInt32)
		assert.Nil(t, err)
		assert.Equal(t, 0, len(a))
	})

	maxVersion := 3

	for _, prefix := range prefixes {
		for major := 1; major <= maxVersion; major++ {
			for minor := 1; minor <= maxVersion; minor++ {
				for patch := 1; patch <= maxVersion; patch++ {
					versionString := fmt.Sprintf("%d.%d.%d", major, minor, patch)
					aVersion := NewVersionT(t, versionString)
					err := CreateTag(r.Repo, aVersion.String(), prefix, false)
					assert.Nil(t, err)
				}
			}
		}
	}
	smallest := NewVersionT(t, "1.1.1")
	greatest := NewVersionT(t, fmt.Sprintf("%d.%d.%d", maxVersion, maxVersion, maxVersion))
	listAll := math.MaxInt32
	listSome := 10
	for _, prefix := range prefixes {
		t.Run(prefix, func(t *testing.T) {
			t.Run("asc", func(t *testing.T) {
				t.Run("listAll", func(t *testing.T) {
					refs, err := ListRefs(r.Repo, prefix, ASC, listAll)
					assert.Nil(t, err)
					assert.Equal(t, maxVersion*maxVersion*maxVersion, len(refs))
					assert.Equal(t, smallest, refs[0].Version)
					assert.Equal(t, greatest, refs[len(refs)-1].Version)
				})
				t.Run("listSome", func(t *testing.T) {
					greatest := NewVersionT(t, "2.1.1")
					refs, err := ListRefs(r.Repo, prefix, ASC, listSome)
					assert.Nil(t, err)
					assert.Equal(t, listSome, len(refs))
					assert.Equal(t, smallest, refs[0].Version)
					assert.Equal(t, greatest, refs[len(refs)-1].Version)
				})
			})
			t.Run("desc", func(t *testing.T) {
				t.Run("listAll", func(t *testing.T) {
					refs, err := ListRefs(r.Repo, prefix, DESC, listAll)
					assert.Nil(t, err)
					assert.Equal(t, maxVersion*maxVersion*maxVersion, len(refs))
					assert.Equal(t, greatest, refs[0].Version)
					assert.Equal(t, smallest, refs[len(refs)-1].Version)
				})

				t.Run("listSome", func(t *testing.T) {
					smallest := NewVersionT(t, "2.3.3")
					refs, err := ListRefs(r.Repo, prefix, DESC, listSome)
					assert.Nil(t, err)
					assert.Equal(t, listSome, len(refs))
					assert.Equal(t, greatest, refs[0].Version)
					assert.Equal(t, smallest, refs[len(refs)-1].Version)
				})
			})
		})
	}
}

//nolint:paralleltest
func TestFindLatestTagSameCommitNoPrefix(t *testing.T) {
	r := NewTestRepo(t)
	rangeEnd := 5

	for major := 1; major < rangeEnd; major++ {
		for minor := 1; minor < rangeEnd; minor++ {
			for patch := 1; patch < rangeEnd; patch++ {
				version := fmt.Sprintf("%d.%d.%d", major, minor, patch)
				t.Run(version, func(t *testing.T) {
					ref := r.CreateTag(version, r.Head().Hash())
					assert.Equal(t, ref.Hash(), r.Head().Hash())

					semverRef, err := LatestRef(r.Repo, "")
					assert.Nil(t, err)
					expectedTag := NewVersionT(t, version)
					assert.Equal(t, *expectedTag, *semverRef.Version)
					assert.Equal(t, semverRef.Ref.Hash(), r.Head().Hash())
				})
			}
		}
	}
}

//nolint:paralleltest
func TestFindPreviousTagSameCommitNoPrefix(t *testing.T) {
	r := NewTestRepo(t)

	rangeEnd := 5
	previous := atomic.NewString("")

	for major := 1; major < rangeEnd; major++ {
		for minor := 1; minor < rangeEnd; minor++ {
			for patch := 1; patch < rangeEnd; patch++ {
				version := fmt.Sprintf("%d.%d.%d", major, minor, patch)
				t.Run(version, func(t *testing.T) {
					ref := r.CreateTag(version, r.Head().Hash())
					assert.Equal(t, ref.Hash(), r.Head().Hash())

					switch version {
					case "1.1.1":
						_, err := PreviousRef(r.Repo, "")
						assert.Regexp(t, "one tag found", err)
					default:
						semverRef, err := PreviousRef(r.Repo, "")
						assert.Nil(t, err)
						println(previous)
						_, err = semver.NewVersion(previous.Load())
						if err != nil {
							str := previous.Load()
							println(str)
						}
						expectedTag := NewVersionT(t, previous.Load())
						assert.Equal(t, *expectedTag, *semverRef.Version)
						assert.Equal(t, semverRef.Ref.Hash(), r.Head().Hash())
					}
					previous.Store(version)
				})
			}
		}
	}
}

//nolint:scopelint,paralleltest
func TestFindLatestTagSameCommitWithPrefix(t *testing.T) {
	r := NewTestRepo(t)

	for major := 1; major < 5; major++ {
		for minor := 1; minor < 5; minor++ {
			for patch := 1; patch < 5; patch++ {
				versionSuffix := fmt.Sprintf("%d.%d.%d", major, minor, patch)
				t.Run(versionSuffix, func(t *testing.T) {
					tagPrefix1 := "app"
					version1 := fmt.Sprintf("%d.%d.%d", major, minor, patch)
					tag1 := fmt.Sprintf("%s%s", tagPrefix1, version1)
					ref1 := r.CreateTag(tag1, r.Head().Hash())
					assert.Equal(t, ref1.Hash(), r.Head().Hash())

					tagPrefix2 := "apple"
					version2 := fmt.Sprintf("%d.%d.%d", major*100, minor*100, patch*100)
					tag2 := fmt.Sprintf("%s%s", tagPrefix2, version2)
					ref2 := r.CreateTag(tag2, r.Head().Hash())
					PrintTags(t, r.Repo)
					assert.Equal(t, ref2.Hash(), r.Head().Hash())

					semverRef, err := LatestRef(r.Repo, tagPrefix1)
					assert.Nil(t, err)
					assert.Equal(t, *NewVersionT(t, version1), *semverRef.Version)
					assert.Equal(t, semverRef.Ref.Hash(), r.Head().Hash())

					semverRef2, err := LatestRef(r.Repo, tagPrefix2)
					assert.Nil(t, err)
					assert.Equal(t, *NewVersionT(t, version2), *semverRef2.Version)
					assert.Equal(t, ref2.Hash(), r.Head().Hash())
				})
			}
		}
	}
}

var dontNeedPreRelease = func(version *semver.Version) (semver.Version, error) {
	return semver.Version{}, errors.New("should not have called") //nolint
}

//nolint:scopelint,paralleltest
func TestCurrentVersionTagOnTheHead(t *testing.T) {
	prefixes := []string{"", "app", "application"}

	for _, prefix := range prefixes {
		t.Run(prefix, func(t *testing.T) {
			r := NewTestRepo(t)

			err := CreateTag(r.Repo, "0.0.1", prefix, false)
			assert.Nil(t, err)

			cr, err := CurrentVersion(r.Repo, prefix, dontNeedPreRelease)
			assert.Nil(t, err)
			assert.Equal(t, NewVersionT(t, "0.0.1").String(), cr.Version.String())
		})
	}
}

//nolint:scopelint,paralleltest
func TestCurrentVersionNoTagOnTheHead(t *testing.T) {
	prefixes := []string{"", "app", "application"}

	for _, prefix := range prefixes {
		t.Run(prefix, func(t *testing.T) {
			r := NewTestRepo(t)

			err := CreateTag(r.Repo, "0.0.1", prefix, false)
			assert.Nil(t, err)

			r.DoCommit("bar")
			cr, err := CurrentVersion(r.Repo, prefix, func(version *semver.Version) (semver.Version, error) {
				return version.IncMinor().SetPrerelease("SNAPSHOT")
			})
			assert.Nil(t, err)
			assert.Equal(t, SemverRef{Ref: r.Head(), Version: NewVersionT(t, "0.1.0-SNAPSHOT")}, cr)
		})
	}
}

//nolint:scopelint,paralleltest
func TestCurrentVersionWithCheckoutOlderRelease(t *testing.T) {
	prefixes := []string{"", "app", "application"}

	for _, prefix := range prefixes {
		t.Run(prefix, func(t *testing.T) {
			r := NewTestRepo(t)
			err := CreateTag(r.Repo, "0.0.1", prefix, false)
			assert.Nil(t, err)
			checkoutHash := r.Head().Hash()

			r.DoCommit("bar")
			err = CreateTag(r.Repo, "0.0.2", prefix, false)
			assert.Nil(t, err)

			{
				cr, err := CurrentVersion(r.Repo, prefix, dontNeedPreRelease)
				assert.Nil(t, err)
				assert.Equal(t, NewVersionT(t, "0.0.2"), cr.Version)
			}

			wt, err := r.Repo.Worktree()
			assert.Nil(t, err)
			err = wt.Checkout(&git.CheckoutOptions{Hash: checkoutHash})
			assert.Nil(t, err)

			{
				cr, err := CurrentVersion(r.Repo, prefix, dontNeedPreRelease)
				assert.Nil(t, err)
				assert.Equal(t, checkoutHash, cr.Ref.Hash())
				assert.Equal(t, NewVersionT(t, "0.0.1"), cr.Version)
			}
		})
	}
}

//nolint:scopelint,paralleltest
func TestCurrentVersionWithAnnotatedTags(t *testing.T) {
	prefixes := []string{"", "app", "application"}

	for _, prefix := range prefixes {
		t.Run(prefix, func(t *testing.T) {
			r := NewTestRepo(t)
			tagger := &object.Signature{
				Name:  "test",
				Email: "test@test.com",
				When:  time.Now(),
			}
			err := CreateTagWithMessage(r.Repo, "0.0.1", prefix, "test message", tagger, false)
			assert.Nil(t, err)
			checkoutHash := r.Head().Hash()

			r.DoCommit("bar")
			err = CreateTagWithMessage(r.Repo, "0.0.2", prefix, "test message", tagger, false)
			assert.Nil(t, err)

			{
				cr, err := CurrentVersion(r.Repo, prefix, dontNeedPreRelease)
				assert.Nil(t, err)
				assert.Equal(t, NewVersionT(t, "0.0.2"), cr.Version)
			}

			wt, err := r.Repo.Worktree()
			assert.Nil(t, err)
			err = wt.Checkout(&git.CheckoutOptions{Hash: checkoutHash})
			assert.Nil(t, err)

			{
				cr, err := CurrentVersion(r.Repo, prefix, dontNeedPreRelease)
				assert.Nil(t, err)
				assert.Equal(t, checkoutHash, cr.Ref.Hash())
				assert.Equal(t, NewVersionT(t, "0.0.1"), cr.Version)
			}
		})
	}
}

//nolint:scopelint,paralleltest
func TestCurrentVersionNoTagOnTheHeadInvalidPrerelease(t *testing.T) {
	prefixes := []string{"", "app", "application"}

	preReleases := []struct {
		fn    PreRelease
		error string
	}{
		{
			fn: func(version *semver.Version) (semver.Version, error) {
				return version.SetPrerelease("EARLY")
			},
			error: "must create a greater version",
		},
		{
			fn: func(version *semver.Version) (semver.Version, error) {
				return *version, nil
			},
			error: "must have prerelease part",
		},
	}

	for _, prefix := range prefixes {
		for _, preRelease := range preReleases {
			t.Run(prefix, func(t *testing.T) {
				r := NewTestRepo(t)

				err := CreateTag(r.Repo, "0.0.1", prefix, false)
				assert.Nil(t, err)

				r.DoCommit("bar")
				_, err = CurrentVersion(r.Repo, prefix, preRelease.fn)
				assert.Regexp(t, preRelease.error, err)
			})
		}
	}
}

//nolint:scopelint,paralleltest
func TestListNoTag(t *testing.T) {
	maxListSize := 10
	prefixes := []string{"", "app", "application"}
	for _, prefix := range prefixes {
		t.Run(prefix, func(t *testing.T) {
			r := NewTestRepo(t)
			refs, err := ListRefs(r.Repo, prefix, DESC, maxListSize)
			assert.Nil(t, err)
			assert.True(t, len(refs) == 0)
		})
	}
}

//nolint:scopelint,paralleltest
func TestList(t *testing.T) {
	var mainBranch = []string{"master"}
	prefixes := []string{"", "app", "application"}
	for _, prefix := range prefixes {
		t.Run(prefix, func(t *testing.T) {
			r := NewTestRepo(t)
			_, err := bump.Bump(r.Repo, prefix, "minor", mainBranch, false)
			assert.Nil(t, err)
			r.DoCommit("jo")
			_, err = bump.Bump(r.Repo, prefix, "minor", mainBranch, false)
			assert.Nil(t, err)
			getSemver := func(ref SemverRef) string { return ref.Version.String() }

			t.Run("list-all", func(t *testing.T) {
				{
					refs, err := ListRefs(r.Repo, prefix, ASC, 2)
					assert.Nil(t, err)
					assert.Equal(t, funk.Map(refs, getSemver), []string{"0.1.0", "0.2.0"})
				}
				{
					refs, err := ListRefs(r.Repo, prefix, DESC, 2)
					assert.Nil(t, err)
					assert.Equal(t, funk.Map(refs, getSemver), []string{"0.2.0", "0.1.0"})
				}
			})
			t.Run("list-1", func(t *testing.T) {
				{
					refs, err := ListRefs(r.Repo, prefix, ASC, 1)
					assert.Nil(t, err)
					assert.Equal(t, funk.Map(refs, getSemver), []string{"0.1.0"})
				}
				{
					refs, err := ListRefs(r.Repo, prefix, DESC, 1)
					assert.Nil(t, err)
					assert.Equal(t, funk.Map(refs, getSemver), []string{"0.2.0"})
				}
			})
		})
	}
}
