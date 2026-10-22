package main

import (
	"context"
	"fmt"
	"slices"
	"strings"

	"github.com/egoist/mygo"
	"github.com/egoist/mygo/ui"

	"mygit/internal/git"
	loc "mygit/locales"
)

// prompt asks for a name, of a branch or a tag, in a dialog.
type prompt struct {
	open                 bool
	title, button, value string
	message, msgVal      string // a second, optional field: a tag's message
	check                string // an option, as a checkbox, when given
	checked              bool
	ok                   func(value, message string)
}

func (a *app) ask1(title, button, value string, ok func(string)) {
	a.prompt = prompt{open: true, title: title, button: button, value: value,
		ok: func(v, _ string) { ok(v) }}
}

func (a *app) promptView(c *ui.Context) {
	p := &a.prompt
	ui.Modal(c, &p.open, func() {
		ui.Column(c).Gap(12).Width(360).Children(func() {
			ui.Text(c, p.title).Bold().FontSize(14)
			submitted := ui.TextInput(c, &p.value).Label(p.title).AutoFocus().Submitted()
			if p.message != "" {
				ui.TextInput(c, &p.msgVal).Placeholder(p.message).Label(p.message)
			}
			if p.check != "" {
				ui.Checkbox(c, &p.checked, p.check)
			}
			ui.Row(c).Gap(8).Justify(ui.End).Children(func() {
				if ui.Button(c, loc.Cancel()).Clicked() {
					p.open = false
				}
				ok := strings.TrimSpace(p.value) != ""
				if (ui.PrimaryButton(c, p.button).Disabled(!ok).Clicked() || submitted) && ok {
					p.open = false
					p.ok(strings.TrimSpace(p.value), strings.TrimSpace(p.msgVal))
				}
			})
		})
	})
}

// branchItems are what a local branch's menu does with it.
func (a *app) branchItems(m *ui.Menu, b string) {
	cur := a.status.Branch
	if b != cur && a.chose(m.Item(loc.BranchSwitch())) {
		a.switchTo(b)
	}
	if b != cur {
		a.mergeItems(m, b)
	}
	if a.chose(m.Item(loc.BranchNewFrom())) {
		a.newBranch(b)
	}
	if a.chose(m.Item(loc.BranchHistory())) {
		a.setScope(git.Scope{Ref: b})
	}
	m.Separator()
	if a.chose(m.Item(loc.RenameMenu())) {
		a.ask1(loc.BranchRenameTitle(b), loc.Rename(), b, func(to string) {
			a.run(loc.OpRenameBranch(), func(ctx context.Context) error { return a.repo.RenameBranch(ctx, b, to) }, nil)
		})
	}
	if b != cur && a.chose(m.Item(loc.Delete())) {
		a.deleteBranch(b)
	}
}

// remoteItems are what a remote branch's menu does with it.
func (a *app) remoteItems(m *ui.Menu, rb string) {
	if a.chose(m.Item(loc.BranchCheckoutLocal())) {
		a.run(loc.OpCheckout(rb), func(ctx context.Context) error { return a.repo.CheckoutRemote(ctx, rb) }, nil)
	}
	a.mergeItems(m, rb)
	if a.chose(m.Item(loc.BranchNewFrom())) {
		a.newBranch(rb)
	}
	m.Separator()
	if a.chose(m.Item(loc.BranchDeleteRemote())) {
		a.ask(loc.BranchDeleteRemoteTitle(rb), loc.BranchDeleteRemoteMessage(), loc.Delete(), func() {
			a.run(loc.OpDeleteRemoteBranch(), func(ctx context.Context) error { return a.repo.DeleteRemoteBranch(ctx, rb) }, nil)
		})
	}
}

// chose reports the item chosen, as whatever the branches' menus do is
// done with it.
func (a *app) chose(it *ui.MenuItem) bool {
	return it.Chosen()
}

func (a *app) switchTo(b string) {
	a.run(loc.OpSwitchTo(b), func(ctx context.Context) error { return a.repo.Switch(ctx, b) }, nil)
}

// branchesPanel is the column beside the workspace: tabs for the
// branches, the stash and the tags. The branches' tab has the filter
// narrowing the list and the way to add a branch, the local branches in
// folders by their names' first segment, and the remote ones under them.
// A row's click switches to its branch, its menu acts on it.
func (a *app) branchesPanel(c *ui.Context) {
	t := c.Theme()
	ui.Column(c).Fill().Children(func() {
		ui.Box(c).Padding(8, 10, 0, 10).Children(func() {
			ui.Segmented(c, &a.branchesTab, loc.BranchTab(), loc.StashTab(), loc.TagsTab()).AlignSelf(ui.Center)
		})
		switch a.branchesTab {
		case 1:
			a.stashPanel(c)
		case 2:
			a.tagsPanel(c)
		default:
			filterBar(c, &a.branchFilter, loc.FilterBranches(), loc.BranchNew(), func() { a.newBranch("") })
			local, remote := matchFilter(a.branches, a.branchFilter), matchFilter(a.remotes, a.branchFilter)
			ui.Scroll(c).Gap(1).Children(func() {
				section := func(title string, n int, top float32) {
					ui.Row(c).Gap(6).Padding(top, 8, 2, 8).AlignItems(ui.Center).Children(func() {
						ui.Text(c, title).FontSize(11).Bold().TextColor(t.TextMuted)
						ui.Badge(c, fmt.Sprint(n))
					})
				}
				section(loc.BranchesLocal(), len(local), 8)
				// The local branches, in folders by their names' first
				// segment, the folders mixed with the plain branches by name.
				a.groupRows(c, local, 0, func(c *ui.Context, name, show string, indent float32) {
					a.branchLine(c, name, show, indent)
				})
				section(loc.BranchesRemote(), len(remote), 44)
				// The remote branches, in folders the same way, without
				// their remote's name in front.
				a.groupRows(c, remote, 1, func(c *ui.Context, name, show string, indent float32) {
					branchMenuRow(c.Key(name), name, show, false, indent, func(m *ui.Menu) { a.remoteItems(m, name) })
				})
				if len(local)+len(remote) == 0 {
					ui.Text(c, loc.BranchesNone()).TextColor(t.TextMuted).Padding(8)
				}
			})
		}
	})
}

// tagsPanel lists the tags; the Tags tab of the branches' column shows
// it. A row's click shows the commit it points at.
func (a *app) tagsPanel(c *ui.Context) {
	t := c.Theme()
	ui.Column(c).Grow(1).Basis(0).Children(func() {
		filterBar(c, &a.tagFilter, loc.FilterTags(), "", nil)
		tags := matchTags(a.tags, a.tagFilter)
		l := ui.List(c.Key("tags"), &a.tagList, len(tags), func(i int) {
			tg := tags[i]
			row := ui.ButtonBase(c.Key(tg.Name)).Padding(6, 10).Label(tg.Name).Justify(ui.SpaceBetween)
			if i == a.tagSel {
				// the accent the list draws under it, hidden by an opaque soft blue
				row.Background(t.Accent.Alpha(0.14).Over(t.Surface)).TextColor(t.Text)
			}
			row.Children(func() {
				ui.Column(c).Gap(2).Children(func() {
					ui.Text(c, tg.Name).SingleLine()
					ui.Text(c, tg.Hash[:8]).FontSize(11).TextColor(t.TextMuted).SingleLine()
				})
				ui.Text(c, tg.Time.Format("2006-01-02")).FontSize(11).TextColor(t.TextMuted).SingleLine()
			})
			if row.Clicked() {
				a.tagSel = i
				a.showTag(tg)
			}
		}).Grow(1).Basis(0).Children(func() {
			if len(tags) == 0 {
				ui.Text(c, loc.TagsNone()).TextColor(t.TextMuted).Padding(12)
			}
		})
		if l.Changed() && a.tagSel >= 0 && a.tagSel < len(tags) {
			a.showTag(tags[a.tagSel])
		}
	})
}

// matchFilter narrows branch names to those the filter matches, in lower
// case; all of them while it is empty.
func matchFilter(names []string, q string) []string {
	if q == "" {
		return names
	}
	q = strings.ToLower(q)
	var out []string
	for _, n := range names {
		if strings.Contains(strings.ToLower(n), q) {
			out = append(out, n)
		}
	}
	return out
}

// matchTags narrows the tags to those whose names the filter matches, in
// lower case; all of them while it is empty.
func matchTags(list []git.Tag, q string) []git.Tag {
	if q == "" {
		return list
	}
	q = strings.ToLower(q)
	var out []git.Tag
	for _, tg := range list {
		if strings.Contains(strings.ToLower(tg.Name), q) {
			out = append(out, tg)
		}
	}
	return out
}

// groupRows lists the branches in folders by their names' first segment,
// dropping leading segments first (a remote's name), the folders mixed
// with the plain branches by name. row draws a branch's row, indented
// under a folder.
func (a *app) groupRows(c *ui.Context, names []string, drop int, row func(c *ui.Context, name, show string, indent float32)) {
	byFolder := map[string][]string{}
	plain := map[string][]string{}
	rests := map[string]string{} // the names without their leading segments
	var keys []string
	for _, n := range names {
		rest := n
		for i := 0; i < drop; i++ {
			if _, r, ok := strings.Cut(rest, "/"); ok {
				rest = r
			}
		}
		rests[n] = rest
		if folder, _, ok := strings.Cut(rest, "/"); ok {
			if byFolder[folder] == nil {
				keys = append(keys, folder)
			}
			byFolder[folder] = append(byFolder[folder], n)
		} else {
			keys = append(keys, rest)
			plain[rest] = append(plain[rest], n)
		}
	}
	slices.Sort(keys)
	for _, key := range keys {
		if ns := plain[key]; ns != nil {
			for _, n := range ns {
				row(c.Key(n), n, key, 0)
			}
			continue
		}
		fold := fmt.Sprint(drop, "/", key) // the local and the remote folders fold apart
		a.folderRow(c.Key("folder "+fold), fold, key, len(byFolder[key]))
		if !a.foldFolders[fold] {
			for _, n := range byFolder[key] {
				row(c.Key(n), n, strings.TrimPrefix(rests[n], key+"/"), 12)
			}
		}
	}
}

// folderRow is a folder of branches in the column: the first segment of
// their names, with the arrow that shows or hides them. It has no
// background, as a heading.
func (a *app) folderRow(c *ui.Context, fold, name string, n int) {
	t := c.Theme()
	folded := a.foldFolders[fold]
	row := ui.ButtonBase(c).Height(26).PaddingX(8).Gap(6).AlignItems(ui.Center).Label(name)
	row.Children(func() {
		chevron := "chevron-down"
		if folded {
			chevron = "chevron-right"
		}
		ui.Icon(c, icons[chevron]).Size(14, 14)
		ui.Text(c, name).SingleLine().Grow(1).Basis(0)
		ui.Text(c, fmt.Sprint(n)).FontSize(11).TextColor(t.TextMuted)
	})
	if row.Clicked() {
		a.foldFolders[fold] = !folded
	}
}

// branchLine is a local branch's row in the column, indented in a
// folder, with how far it is from its upstream at the end. show is its
// name without the folder's, as the folder names it above. A click
// switches to it, its menu acts on it.
func (a *app) branchLine(c *ui.Context, name, show string, indent float32) {
	e := branchRowElement(c.Key(name), name, show, name == a.status.Branch, trackText(a.tracks[name]), indent)
	if e.Clicked() {
		a.switchTo(name)
	}
	e.ContextMenu(func(m *ui.Menu) { a.branchItems(m, name) })
}

// trackText is how far a branch is from its upstream, for its row's end.
func trackText(tr git.BranchTrack) string {
	var parts []string
	if tr.Ahead > 0 {
		parts = append(parts, fmt.Sprintf("↑%d", tr.Ahead))
	}
	if tr.Behind > 0 {
		parts = append(parts, fmt.Sprintf("↓%d", tr.Behind))
	}
	return strings.Join(parts, " ")
}

func branchRowElement(c *ui.Context, name, show string, current bool, tail string, indent float32) ui.Element {
	t := c.Theme()
	row := ui.ButtonBase(c).Height(26).PaddingX(8).PaddingStart(8 + indent).Gap(6).Radius(5).AlignItems(ui.Center).Label(name)
	if row.Hovered() {
		row.Background(t.Text.Alpha(0.08))
	}
	if current {
		row.Background(t.Accent.Alpha(0.14))
	}
	row.Children(func() {
		mark := ""
		if current {
			mark = "✓"
		}
		ui.Text(c, mark).Width(12).TextColor(t.Accent)
		ui.Text(c, show).SingleLine().Grow(1).Basis(0)
		if tail != "" {
			ui.Text(c, tail).FontSize(11).TextColor(t.TextMuted).SingleLine()
		}
	})
	return row
}

// branchMenuRow is a branch's row, whose click opens its menu.
func branchMenuRow(c *ui.Context, name, show string, current bool, indent float32, items func(*ui.Menu)) {
	branchRowElement(c, name, show, current, "", indent).Menu(items)
}

// mergeItems merge a branch into the current one, or rebase the current
// one onto it.
func (a *app) mergeItems(m *ui.Menu, rev string) {
	cur := a.status.Branch
	if cur == "" {
		cur = "HEAD"
	}
	if a.chose(m.Item(loc.CompareWith(cur))) {
		a.showCompare(loc.CompareTitle(cur, shortRev(rev)), "HEAD", rev)
	}
	if a.chose(m.Item(loc.MergeInto(cur))) {
		a.run(loc.OpMerge(shortRev(rev)), func(ctx context.Context) error { return a.repo.Merge(ctx, rev) }, nil)
	}
	if a.chose(m.Item(loc.RebaseOnto(cur))) {
		a.run(loc.OpRebaseOnto(shortRev(rev)), func(ctx context.Context) error { return a.repo.Rebase(ctx, rev) }, nil)
	}
}

// newBranch asks for a branch's name, made at start (HEAD when empty) and
// checked out.
func (a *app) newBranch(start string) {
	title := loc.NewBranch()
	if start != "" {
		title = loc.NewBranchFrom(shortRev(start))
	}
	a.ask1(title, loc.CreateAndSwitch(), "", func(name string) {
		a.run(loc.OpNewBranch(name), func(ctx context.Context) error {
			return a.repo.CreateBranch(ctx, name, start, true)
		}, nil)
	})
}

// deleteBranch deletes a merged branch at once, and asks before deleting
// one whose commits would be lost.
func (a *app) deleteBranch(b string) {
	a.run(loc.OpDeleteBranch(b), func(ctx context.Context) error {
		err := a.repo.DeleteBranch(ctx, b, false)
		if git.NotMerged(err) {
			a.update(func() {
				a.ask(loc.BranchNotMergedTitle(b), loc.BranchNotMergedMessage(), loc.DeleteAnyway(), func() {
					a.run(loc.OpDeleteBranch(b), func(ctx context.Context) error { return a.repo.DeleteBranch(ctx, b, true) }, nil)
				})
			})
			return nil
		}
		return err
	}, nil)
}

// commitMenu is a commit's context menu in the history.
func (a *app) commitMenu(m *ui.Menu, cm git.Commit) {
	if m.Item(loc.CopyHash()).Chosen() {
		mygo.Clipboard.WriteText(cm.Hash)
	}
	if m.Item(loc.CopyMessage()).Chosen() {
		mygo.Clipboard.WriteText(cm.Subject)
	}
	m.Separator()
	if m.Item(loc.CommitNewBranch()).Chosen() {
		a.newBranch(cm.Hash)
	}
	if m.Item(loc.CommitNewTag()).Chosen() {
		a.prompt = prompt{open: true, title: loc.NewTagTitle(cm.Short()), button: loc.Create(), message: loc.TagMessage(),
			ok: func(name, msg string) {
				a.run(loc.OpNewTag(name), func(ctx context.Context) error { return a.repo.Tag(ctx, name, cm.Hash, msg) }, nil)
			}}
	}
	if m.Item(loc.CompareWorktree()).Chosen() {
		a.showCompare(loc.CompareTitle(cm.Short(), loc.Worktree()), cm.Hash, "")
	}
	if m.Item(loc.CheckoutCommit()).Chosen() {
		a.run(loc.OpCheckout(cm.Short()), func(ctx context.Context) error { return a.repo.CheckoutDetached(ctx, cm.Hash) }, nil)
	}
	m.Separator()
	if m.Item(loc.CherryPick()).Chosen() {
		a.run(loc.CherryPick()+" "+cm.Short(), func(ctx context.Context) error { return a.repo.CherryPick(ctx, cm) }, nil)
	}
	if m.Item(loc.RevertCommit()).Chosen() {
		a.run(loc.OpRevert(cm.Short()), func(ctx context.Context) error { return a.repo.Revert(ctx, cm) }, nil)
	}
	a.mergeItems(m, cm.Hash)
	if m.Item(loc.RebaseFromHere()).Chosen() {
		a.planRebase(cm)
	}
	if a.status.Branch != "" {
		m.Submenu(loc.ResetTo(a.status.Branch), func(m *ui.Menu) {
			reset := func(mode git.ResetMode) {
				a.run(loc.OpReset(), func(ctx context.Context) error { return a.repo.Reset(ctx, cm.Hash, mode) }, nil)
			}
			if m.Item(loc.ResetSoft()).Chosen() {
				reset(git.ResetSoft)
			}
			if m.Item(loc.ResetMixed()).Chosen() {
				reset(git.ResetMixed)
			}
			if m.Item(loc.ResetHard()).Chosen() {
				a.ask(loc.ResetHardTitle(cm.Short()), loc.ResetHardMessage(), loc.ResetHardButton(), func() { reset(git.ResetHard) })
			}
		})
	}
	for _, ref := range cm.Refs {
		if tag, ok := strings.CutPrefix(ref, "tag: "); ok && m.Item(loc.DeleteTag(tag)).Chosen() {
			a.run(loc.OpDeleteTag(tag), func(ctx context.Context) error { return a.repo.DeleteTag(ctx, tag) }, nil)
		}
	}
}

// shortRev shortens a hash, and leaves a branch's name as it is.
func shortRev(rev string) string {
	if len(rev) == 40 && strings.Trim(rev, "0123456789abcdef") == "" {
		return rev[:8]
	}
	return rev
}
