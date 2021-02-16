package internal_test

import (
	"fmt"
	"github.com/Masterminds/semver"
	"github.com/go-git/go-billy/v5/memfs"
	"github.com/go-git/go-billy/v5/util"
	gogit "github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/go-git/go-git/v5/storage/memory"
	"github.com/stretchr/testify/assert"
	"sort"
	"testing"
	"time"
)

func NewVersionT(t *testing.T, version string) *semver.Version {
	t.Helper()
	v, err := semver.NewVersion(version)
	assert.Nil(t, err)
	return v
}

func defaultSignature() *object.Signature {
	when, _ := time.Parse(object.DateFormat, "Thu May 04 00:03:43 2017 +0200")
	return &object.Signature{
		Name:  "foo",
		Email: "foo@foo.foo",
		When:  when,
	}
}

type TestRepo struct {
	t    *testing.T
	Repo *gogit.Repository
}

func (t *TestRepo) Head() *plumbing.Reference {
	t.t.Helper()
	head, err := t.Repo.Head()
	assert.Nil(t.t, err)
	return head
}

func (t *TestRepo) Worktree() *gogit.Worktree {
	t.t.Helper()
	worktree, err := t.Repo.Worktree()
	assert.Nil(t.t, err)
	return worktree
}

func (t *TestRepo) CreateTag(name string, hash plumbing.Hash) *plumbing.Reference {
	t.t.Helper()
	ref, err := t.Repo.CreateTag(name, hash, nil)
	assert.Nil(t.t, err)
	assert.NotNil(t.t, ref)
	return ref
}

func (t *TestRepo) DoCommit(file string) {
	t.t.Helper()
	DoCommitWithMessage(t.t, t.Repo, file, file)
}

func (t *TestRepo) BranchExists(branchName string) bool {
	t.t.Helper()
	branches, err := t.Repo.Branches()
	assert.Nil(t.t, err)
	defer branches.Close()
	branchExists := false
	for {
		branch, err := branches.Next()
		if err != nil {
			break
		}
		if branch.Name().Short() == branchName {
			branchExists = true
		}
	}
	return branchExists
}

func NewTestRepo(t *testing.T) TestRepo {
	t.Helper()
	return TestRepo{
		t:    t,
		Repo: inMemoryRepositoryWithDefaultCommit(t),
	}
}

func inMemoryRepositoryWithDefaultCommit(t *testing.T) *gogit.Repository {
	t.Helper()
	r, err := gogit.Init(memory.NewStorage(), memfs.New())
	assert.Nil(t, err)
	DoCommit(t, r, "foo")
	_, err = r.Head()
	assert.Nil(t, err)
	return r
}

func PersistentRepository(t *testing.T) (*gogit.Repository, string) {
	t.Helper()
	tempDir := t.TempDir()
	r, err := gogit.PlainInit(tempDir, false)
	assert.Nil(t, err)
	return r, tempDir
}

func DoCommit(t *testing.T, r *gogit.Repository, file string) {
	t.Helper()
	DoCommitWithMessage(t, r, file, file)
}

func DoCommitWithMessage(t *testing.T, r *gogit.Repository, file, message string) {
	t.Helper()
	w, err := r.Worktree()
	assert.Nil(t, err)

	err = util.WriteFile(w.Filesystem, file, nil, 0755)
	assert.Nil(t, err)

	_, err = w.Add(file)
	assert.Nil(t, err)

	_, err = w.Commit(message, &gogit.CommitOptions{
		Author:    defaultSignature(),
		Committer: defaultSignature(),
	})
	assert.Nil(t, err)
}

func PrintTags(t *testing.T, r *gogit.Repository) {
	t.Helper()
	iter, err := r.Tags()
	assert.Nil(t, err)

	var tags []string
	err = iter.ForEach(func(r *plumbing.Reference) error {
		tags = append(tags, r.Name().String())
		return nil
	})
	assert.Nil(t, err)

	sort.Strings(tags)

	fmt.Print("tags: ") //nolint
	for _, tag := range tags {
		fmt.Print(tag, ", ") //nolint
	}
	fmt.Println() //nolint
}
