package git_test

import (
	"errors"
	"fmt"
	"github.com/Masterminds/semver"
	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"go.uber.org/atomic"
	"math"
	. "sky.uk/vergo/git"
	. "sky.uk/vergo/internal"
	"testing"
)

func init() {
	log.SetLevel(log.TraceLevel)
}

//nolint:scopelint,paralleltest
func TestNoTag(t *testing.T) {
	prefixes := []string{"", "app", "apple"}

	for _, prefix := range prefixes {
		t.Run(prefix, func(t *testing.T) {
			r := InMemoryRepositoryWithDefaultCommit(t)
			_, err := LatestRef(r, prefix)
			assert.Regexp(t, "no tag found", err)
		})
	}
}

//nolint:scopelint,paralleltest
func TestTagExists(t *testing.T) {
	prefixes := []string{"", "app", "apple"}

	for _, prefix := range prefixes {
		t.Run(prefix, func(t *testing.T) {
			r := InMemoryRepositoryWithDefaultCommit(t)

			aVersion := NewVersionT(t, "0.0.1")
			err := CreateTag(r, aVersion.String(), prefix, false)
			assert.Nil(t, err)
			found, err := TagExists(r, prefix+"0.0.1")
			assert.Nil(t, err)
			assert.True(t, found)
		})
	}
}

//nolint:scopelint,paralleltest
func TestTagAlreadyExists(t *testing.T) {
	prefixes := []string{"", "app", "apple"}

	for _, prefix := range prefixes {
		t.Run(prefix, func(t *testing.T) {
			r := InMemoryRepositoryWithDefaultCommit(t)

			aVersion := NewVersionT(t, "0.0.1")
			err := CreateTag(r, aVersion.String(), prefix, false)
			assert.Nil(t, err)

			anotherVersion := NewVersionT(t, "0.0.1")
			err = CreateTag(r, anotherVersion.String(), prefix, false)
			assert.Regexp(t, "already exists", err)
		})
	}
}

//nolint:funlen,scopelint,paralleltest
func TestListRefs(t *testing.T) {
	prefixes := []string{"", "app", "apple"}
	r := InMemoryRepositoryWithDefaultCommit(t)

	t.Run("tag does not exist", func(t *testing.T) {
		a, err := ListRefs(r, "banana", ASC, math.MaxInt32)
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
					err := CreateTag(r, aVersion.String(), prefix, false)
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
					refs, err := ListRefs(r, prefix, ASC, listAll)
					assert.Nil(t, err)
					assert.Equal(t, maxVersion*maxVersion*maxVersion, len(refs))
					assert.Equal(t, smallest, refs[0].Version)
					assert.Equal(t, greatest, refs[len(refs)-1].Version)
				})
				t.Run("listSome", func(t *testing.T) {
					greatest := NewVersionT(t, "2.1.1")
					refs, err := ListRefs(r, prefix, ASC, listSome)
					assert.Nil(t, err)
					assert.Equal(t, listSome, len(refs))
					assert.Equal(t, smallest, refs[0].Version)
					assert.Equal(t, greatest, refs[len(refs)-1].Version)
				})
			})
			t.Run("desc", func(t *testing.T) {
				t.Run("listAll", func(t *testing.T) {
					refs, err := ListRefs(r, prefix, DESC, listAll)
					assert.Nil(t, err)
					assert.Equal(t, maxVersion*maxVersion*maxVersion, len(refs))
					assert.Equal(t, greatest, refs[0].Version)
					assert.Equal(t, smallest, refs[len(refs)-1].Version)
				})

				t.Run("listSome", func(t *testing.T) {
					smallest := NewVersionT(t, "2.3.3")
					refs, err := ListRefs(r, prefix, DESC, listSome)
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
	r := InMemoryRepositoryWithDefaultCommit(t)
	head, err := r.Head()
	assert.Nil(t, err)

	rangeEnd := 5

	for major := 1; major < rangeEnd; major++ {
		for minor := 1; minor < rangeEnd; minor++ {
			for patch := 1; patch < rangeEnd; patch++ {
				version := fmt.Sprintf("%d.%d.%d", major, minor, patch)
				t.Run(version, func(t *testing.T) {
					ref, err := r.CreateTag(version, head.Hash(), nil)
					assert.Nil(t, err)
					assert.Equal(t, ref.Hash(), head.Hash())

					semverRef, err := LatestRef(r, "")
					assert.Nil(t, err)
					expectedTag := NewVersionT(t, version)
					assert.Equal(t, *expectedTag, *semverRef.Version)
					assert.Equal(t, semverRef.Ref.Hash(), head.Hash())
				})
			}
		}
	}
}

//nolint:paralleltest
func TestFindPreviousTagSameCommitNoPrefix(t *testing.T) {
	r := InMemoryRepositoryWithDefaultCommit(t)
	head, err := r.Head()
	assert.Nil(t, err)

	rangeEnd := 5
	previous := atomic.NewString("")

	for major := 1; major < rangeEnd; major++ {
		for minor := 1; minor < rangeEnd; minor++ {
			for patch := 1; patch < rangeEnd; patch++ {
				version := fmt.Sprintf("%d.%d.%d", major, minor, patch)
				t.Run(version, func(t *testing.T) {
					ref, err := r.CreateTag(version, head.Hash(), nil)
					assert.Nil(t, err)
					assert.Equal(t, ref.Hash(), head.Hash())

					switch version {
					case "1.1.1":
						_, err := PreviousRef(r, "")
						assert.Regexp(t, "one tag found", err)
					default:
						semverRef, err := PreviousRef(r, "")
						assert.Nil(t, err)
						println(previous)
						_, err = semver.NewVersion(previous.Load())
						if err != nil {
							str := previous.Load()
							println(str)
						}
						expectedTag := NewVersionT(t, previous.Load())
						assert.Equal(t, *expectedTag, *semverRef.Version)
						assert.Equal(t, semverRef.Ref.Hash(), head.Hash())
					}
					previous.Store(version)
				})
			}
		}
	}
}

//nolint:scopelint,paralleltest
func TestFindLatestTagSameCommitWithPrefix(t *testing.T) {
	r := InMemoryRepositoryWithDefaultCommit(t)
	head, err := r.Head()
	assert.Nil(t, err)

	for major := 1; major < 5; major++ {
		for minor := 1; minor < 5; minor++ {
			for patch := 1; patch < 5; patch++ {
				versionSuffix := fmt.Sprintf("%d.%d.%d", major, minor, patch)
				t.Run(versionSuffix, func(t *testing.T) {
					tagPrefix1 := "app"
					version1 := fmt.Sprintf("%d.%d.%d", major, minor, patch)
					tag1 := fmt.Sprintf("%s%s", tagPrefix1, version1)
					PrintTags(t, r)
					ref1, err := r.CreateTag(tag1, head.Hash(), nil)
					assert.Nil(t, err)
					PrintTags(t, r)
					assert.Equal(t, ref1.Hash(), head.Hash())

					tagPrefix2 := "apple"
					version2 := fmt.Sprintf("%d.%d.%d", major*100, minor*100, patch*100)
					tag2 := fmt.Sprintf("%s%s", tagPrefix2, version2)
					ref2, err := r.CreateTag(tag2, head.Hash(), nil)
					assert.Nil(t, err)
					PrintTags(t, r)
					assert.Equal(t, ref2.Hash(), head.Hash())

					semverRef, err := LatestRef(r, tagPrefix1)
					assert.Nil(t, err)
					expectedVersion := NewVersionT(t, version1)
					assert.Equal(t, *expectedVersion, *semverRef.Version)
					assert.Equal(t, semverRef.Ref.Hash(), head.Hash())

					semverRef2, err := LatestRef(r, tagPrefix2)
					assert.Nil(t, err)
					expectedVersion2 := NewVersionT(t, version2)
					assert.Equal(t, *expectedVersion2, *semverRef2.Version)
					assert.Equal(t, ref2.Hash(), head.Hash())
				})
			}
		}
	}
}

//nolint:scopelint,paralleltest
func TestCurrentVersionTagOnTheHead(t *testing.T) {
	prefixes := []string{"", "app", "apple"}

	for _, prefix := range prefixes {
		t.Run(prefix, func(t *testing.T) {
			r := InMemoryRepositoryWithDefaultCommit(t)

			aVersion := NewVersionT(t, "0.0.1")
			err := CreateTag(r, aVersion.String(), prefix, false)
			assert.Nil(t, err)

			cr, err := CurrentVersion(r, prefix, func(version *semver.Version) (semver.Version, error) {
				return semver.Version{}, errors.New("should not have called") //nolint
			})
			assert.Nil(t, err)
			assert.Equal(t, NewVersionT(t, "0.0.1").String(), cr.Version.String())
		})
	}
}

//nolint:scopelint,paralleltest
func TestCurrentVersionNoTagOnTheHead(t *testing.T) {
	prefixes := []string{"", "app", "apple"}

	for _, prefix := range prefixes {
		t.Run(prefix, func(t *testing.T) {
			r := InMemoryRepositoryWithDefaultCommit(t)

			aVersion := NewVersionT(t, "0.0.1")
			err := CreateTag(r, aVersion.String(), prefix, false)
			assert.Nil(t, err)

			DoCommit(t, r, "bar")
			head, err := r.Head()
			assert.Nil(t, err)
			cr, err := CurrentVersion(r, prefix, func(version *semver.Version) (semver.Version, error) {
				return version.IncMinor().SetPrerelease("SNAPSHOT")
			})
			assert.Nil(t, err)
			assert.Equal(t, SemverRef{Ref: head, Version: NewVersionT(t, "0.1.0-SNAPSHOT")}, cr)
		})
	}
}

//nolint:scopelint,paralleltest
func TestCurrentVersionNoTagOnTheHeadInvalidPrerelease(t *testing.T) {
	prefixes := []string{"", "app", "apple"}

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
				r := InMemoryRepositoryWithDefaultCommit(t)

				aVersion := NewVersionT(t, "0.0.1")
				err := CreateTag(r, aVersion.String(), prefix, false)
				assert.Nil(t, err)

				DoCommit(t, r, "bar")
				_, err = CurrentVersion(r, prefix, preRelease.fn)
				assert.Regexp(t, preRelease.error, err)
			})
		}
	}
}

//nolint:scopelint,paralleltest
func TestCheckNoRelevantChanges(t *testing.T) {
	maxLogIteration := 10
	prefixes := []string{"app", "apple"}

	for _, prefix := range prefixes {
		t.Run(prefix, func(t *testing.T) {
			r := InMemoryRepositoryWithDefaultCommit(t)

			aVersion := NewVersionT(t, "0.0.1")
			err := CreateTag(r, aVersion.String(), prefix, false)
			assert.Nil(t, err)

			DoCommit(t, r, "bar")
			err = RelevantChanges(r, prefix, prefix, maxLogIteration)
			assert.EqualError(t, err, "no relevant change")

			DoCommit(t, r, prefix)
			err = RelevantChanges(r, prefix, prefix, maxLogIteration)
			assert.Nil(t, err)
		})
	}
}
