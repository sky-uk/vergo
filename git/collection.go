package git

import (
	"github.com/Masterminds/semver/v3"
	"github.com/go-git/go-git/v5/plumbing"
)

type SemverRef struct {
	Version *semver.Version
	Ref     *plumbing.Reference
}

// EmptyRef empty constructor
// nolint:gochecknoglobals
var EmptyRef = SemverRef{}

// EmptyRefList empty constructor
// nolint:gochecknoglobals
var EmptyRefList []SemverRef

// SemverRefColl is a collection of Version instances and implements the sort
// interface. See the sort package for more details.
// https://golang.org/pkg/sort/
type SemverRefColl []SemverRef

// Len returns the length of a collection. The number of Version instances
// on the slice.
func (c SemverRefColl) Len() int {
	return len(c)
}

// Less is needed for the sort interface to compare two Version objects on the
// slice. If checks if one is less than the other.
func (c SemverRefColl) Less(i, j int) bool {
	return c[i].Version.LessThan(c[j].Version)
}

// Swap is needed for the sort interface to replace the Version objects
// at two different positions in the slice.
func (c SemverRefColl) Swap(i, j int) {
	c[i], c[j] = c[j], c[i]
}
