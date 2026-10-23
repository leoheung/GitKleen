package git

import (
	"context"
	"slices"
	"strconv"
	"strings"
	"time"
)

// Branches are the local branches, by name.
func (r *Repo) Branches(ctx context.Context) ([]string, error) {
	out, err := r.Git(ctx, "for-each-ref", "--format=%(refname:short)", "refs/heads")
	if err != nil {
		return nil, err
	}
	return strings.Fields(out), nil
}

// BranchTrack is a local branch, its upstream and how far apart.
type BranchTrack struct {
	Name, Upstream string
	Ahead, Behind  int
}

// Tracks lists the local branches with their upstream's counts.
func (r *Repo) Tracks(ctx context.Context) ([]BranchTrack, error) {
	out, err := r.Git(ctx, "for-each-ref", "--format=%(refname:short)%00%(upstream:short)%00%(upstream:track)", "refs/heads")
	if err != nil {
		return nil, err
	}
	var tracks []BranchTrack
	for _, line := range strings.Split(strings.TrimRight(out, "\n"), "\n") {
		if line == "" {
			continue
		}
		f := strings.SplitN(line, "\x00", 3)
		if len(f) != 3 {
			continue
		}
		tr := BranchTrack{Name: f[0], Upstream: f[1]}
		tr.Ahead, tr.Behind = parseTrack(f[2])
		tracks = append(tracks, tr)
	}
	return tracks, nil
}

// parseTrack reads "ahead N" and "behind N" of a track, as
// "(ahead 1, behind 2)" or "[behind 3]".
func parseTrack(s string) (ahead, behind int) {
	f := strings.FieldsFunc(s, func(r rune) bool {
		return r == '(' || r == ')' || r == '[' || r == ']' || r == ',' || r == ' '
	})
	for i := 0; i+1 < len(f); i += 2 {
		switch f[i] {
		case "ahead":
			ahead, _ = strconv.Atoi(f[i+1])
		case "behind":
			behind, _ = strconv.Atoi(f[i+1])
		}
	}
	return ahead, behind
}

// Switch checks a local branch out.
func (r *Repo) Switch(ctx context.Context, branch string) error {
	_, err := r.Git(ctx, "switch", branch)
	return err
}

// Fetch updates every remote's branches.
func (r *Repo) Fetch(ctx context.Context) error {
	_, err := r.Git(ctx, "fetch", "--all", "--prune")
	return err
}

// PullMode is how a pull joins the upstream's commits.
type PullMode string

const (
	PullFastForward PullMode = "--ff-only"   // only when no merge is needed
	PullRebase      PullMode = "--rebase"    // replay the branch's commits on them
	PullMerge       PullMode = "--no-rebase" // merge them
	PullConfigured  PullMode = ""            // as git is set to, pull.rebase
)

// Pull brings the upstream's commits in.
func (r *Repo) Pull(ctx context.Context, mode PullMode) error {
	args := []string{"pull"}
	if mode != PullConfigured {
		args = append(args, string(mode))
	}
	_, err := r.Git(ctx, args...)
	return err
}

// Push sends the branch to its upstream, or to origin under its own name,
// which becomes its upstream, when it has none. Forced, it overwrites the
// remote's branch, if it is still where it was last fetched.
func (r *Repo) Push(ctx context.Context, st Status, force bool) error {
	args := []string{"push"}
	if force {
		args = append(args, "--force-with-lease")
	}
	if st.Upstream == "" && st.Branch != "" {
		args = append(args, "-u", "origin", st.Branch)
	}
	_, err := r.Git(ctx, args...)
	return err
}

// PushTags sends every tag to origin.
func (r *Repo) PushTags(ctx context.Context) error {
	_, err := r.Git(ctx, "push", "origin", "--tags")
	return err
}

// RemoteTags lists the names of the tags on origin, by asking it.
func (r *Repo) RemoteTags(ctx context.Context) (map[string]bool, error) {
	out, err := r.Git(ctx, "ls-remote", "--tags", "origin")
	if err != nil {
		return nil, err
	}
	tags := map[string]bool{}
	for _, line := range strings.Split(out, "\n") {
		_, name, ok := strings.Cut(line, "refs/tags/")
		if !ok || name == "" {
			continue
		}
		name = strings.TrimSuffix(name, "^{}") // an annotated tag lists its commit once more
		tags[name] = true
	}
	return tags, nil
}

// PushTag sends one tag to origin.
func (r *Repo) PushTag(ctx context.Context, name string) error {
	_, err := r.Git(ctx, "push", "origin", name)
	return err
}

// DeleteRemoteTag deletes a tag on origin.
func (r *Repo) DeleteRemoteTag(ctx context.Context, name string) error {
	_, err := r.Git(ctx, "push", "origin", "--delete", name)
	return err
}

// Tag is a tag: its name, the commit it points at, and when it was made.
type Tag struct {
	Name, Hash string
	Time       time.Time
	Subject    string
}

// Tags lists the tags, newest first.
func (r *Repo) Tags(ctx context.Context) ([]Tag, error) {
	out, err := r.Git(ctx, "for-each-ref", "refs/tags", "--sort=-creatordate",
		"--format=%(refname:short)\x1f%(*objectname)\x1f%(objectname)\x1f%(creatordate:unix)\x1f%(subject)")
	if err != nil {
		return nil, err
	}
	var list []Tag
	for _, rec := range strings.Split(out, "\n") {
		f := strings.Split(rec, "\x1f")
		if len(f) != 5 {
			continue
		}
		sec, _ := strconv.ParseInt(f[3], 10, 64)
		hash := f[1] // an annotated tag names the commit it points at here, its own object beside
		if hash == "" {
			hash = f[2] // a lightweight tag's object is the commit
		}
		list = append(list, Tag{Name: f[0], Hash: hash, Time: time.Unix(sec, 0), Subject: f[4]})
	}
	slices.SortStableFunc(list, func(a, b Tag) int { return b.Time.Compare(a.Time) })
	return list, nil
}

// Refs sums up where HEAD and every ref point, to tell when the history
// changed.
func (r *Repo) Refs(ctx context.Context) string {
	out, _ := r.Git(ctx, "show-ref", "--head")
	return out
}

// RemoteBranches are the remotes' branches, "origin/main", without the
// remotes' HEADs.
func (r *Repo) RemoteBranches(ctx context.Context) ([]string, error) {
	out, err := r.Git(ctx, "for-each-ref", "--format=%(refname:short)", "refs/remotes")
	if err != nil {
		return nil, err
	}
	var branches []string
	for _, b := range strings.Fields(out) {
		if !strings.HasSuffix(b, "/HEAD") && strings.Contains(b, "/") {
			branches = append(branches, b)
		}
	}
	return branches, nil
}

// CreateBranch makes a branch at start, HEAD when empty, and checks it out
// when asked.
func (r *Repo) CreateBranch(ctx context.Context, name, start string, checkout bool) error {
	args := []string{"branch", name}
	if checkout {
		args = []string{"switch", "-c", name}
	}
	if start != "" {
		args = append(args, start)
	}
	_, err := r.Git(ctx, args...)
	return err
}

// CheckoutRemote checks a remote's branch out as a local branch of its
// name, which tracks it.
func (r *Repo) CheckoutRemote(ctx context.Context, remoteBranch string) error {
	_, err := r.Git(ctx, "switch", "--track", remoteBranch)
	return err
}

// DeleteBranch deletes a local branch; without force, only one merged.
func (r *Repo) DeleteBranch(ctx context.Context, name string, force bool) error {
	flag := "-d"
	if force {
		flag = "-D"
	}
	_, err := r.Git(ctx, "branch", flag, name)
	return err
}

// NotMerged tells whether DeleteBranch failed for a branch not merged.
func NotMerged(err error) bool {
	return err != nil && strings.Contains(err.Error(), "not fully merged")
}

// DeleteRemoteBranch deletes "origin/name" on its remote.
func (r *Repo) DeleteRemoteBranch(ctx context.Context, remoteBranch string) error {
	remote, name, _ := strings.Cut(remoteBranch, "/")
	_, err := r.Git(ctx, "push", remote, "--delete", name)
	return err
}

// RenameBranch renames a local branch.
func (r *Repo) RenameBranch(ctx context.Context, from, to string) error {
	_, err := r.Git(ctx, "branch", "-m", from, to)
	return err
}

// ValidBranchName tells whether git takes name for a branch.
func (r *Repo) ValidBranchName(ctx context.Context, name string) bool {
	_, err := r.Git(ctx, "check-ref-format", "--branch", name)
	return err == nil && name != ""
}
