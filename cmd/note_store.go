package cmd

import (
	"fmt"
	"io"
	"math/rand"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"
	"unicode/utf8"

	"agmd/pkg/registry"
)

// The user's notes, shared with the Agents app (its notes.py service):
//
//	$AGMD_HOME/note/<project>/<slug>.md   front matter: title, created, updated, tasks
//
// The file format, the slug rule and the lock are the service's; keep them in
// step with it. Every write holds an exclusive flock on $AGMD_HOME/note/.lock
// (one lock for all projects, the one the app takes). Pictures live outside
// this folder (the app's notes-media) and are never touched here; neither is
// the .trash/ folder.

var (
	noteProjectRE = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9._-]{0,80}$`)
	noteSlugRE    = regexp.MustCompile(`^[a-z0-9][a-z0-9-]{0,79}$`)
	noteTaskRE    = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9._/-]{0,120}$`)
	noteMetaKeyRE = regexp.MustCompile(`^[a-z]+$`)
	noteNonSlugRE = regexp.MustCompile(`[^a-z0-9]+`)
)

const (
	noteTitleMax  = 160
	noteBodyMax   = 512 * 1024
	noteGitignore = ".lock\n.trash/\n.*.tmp\n"
	noteTimeFmt   = "2006-01-02T15:04:05Z"
)

type note struct {
	Project string
	Slug    string
	Path    string
	Title   string
	Created string
	Updated string
	Tasks   []string
	Body    string
}

// UpdatedTime is the note's updated stamp, zero when missing or unreadable.
func (n *note) UpdatedTime() time.Time {
	t, err := time.Parse(noteTimeFmt, n.Updated)
	if err != nil {
		return time.Time{}
	}
	return t
}

func noteRoot(reg *registry.Registry) string {
	return filepath.Join(reg.BasePath, "note")
}

func checkNoteProject(project string) error {
	if !noteProjectRE.MatchString(project) {
		return newTaskError("invalid_input", "project '%s': letters, digits, . _ - (at most 81)", project)
	}
	return nil
}

func checkNoteSlug(slug string) error {
	if !noteSlugRE.MatchString(slug) {
		return newTaskError("invalid_input", "note '%s': lowercase letters, digits and - (at most 80)", slug)
	}
	return nil
}

// noteProjectDir is <root>/<project>, refused when it is a symlink.
func noteProjectDir(reg *registry.Registry, project string) (string, error) {
	if err := checkNoteProject(project); err != nil {
		return "", err
	}
	dir := filepath.Join(noteRoot(reg), project)
	if fi, err := os.Lstat(dir); err == nil && fi.Mode()&os.ModeSymlink != 0 {
		return "", newTaskError("invalid_input", "a link in the notes folder: refused")
	}
	return dir, nil
}

func notePath(reg *registry.Registry, project, slug string) (string, error) {
	dir, err := noteProjectDir(reg, project)
	if err != nil {
		return "", err
	}
	if err := checkNoteSlug(slug); err != nil {
		return "", err
	}
	return filepath.Join(dir, slug+".md"), nil
}

// lockNotes takes the one notes lock (flock on <root>/.lock) and writes the
// folder's .gitignore when it is missing, as the service does.
func lockNotes(reg *registry.Registry) (func(), error) {
	root := noteRoot(reg)
	if err := os.MkdirAll(root, 0o755); err != nil {
		return nil, fmt.Errorf("failed to create notes folder: %w", err)
	}
	f, err := os.OpenFile(filepath.Join(root, ".lock"), os.O_CREATE|os.O_RDWR|oNoFollow, 0o600)
	if err != nil {
		return nil, fmt.Errorf("failed to open notes lock: %w", err)
	}
	if err := lockFile(f); err != nil {
		f.Close()
		return nil, fmt.Errorf("failed to lock notes: %w", err)
	}
	unlock := func() {
		unlockFile(f) //nolint:errcheck
		f.Close()
	}
	ign := filepath.Join(root, ".gitignore")
	if _, err := os.Lstat(ign); os.IsNotExist(err) {
		if err := writeNoteFile(ign, []byte(noteGitignore), true); err != nil && !os.IsExist(err) {
			unlock()
			return nil, fmt.Errorf("failed to write notes .gitignore: %w", err)
		}
	}
	return unlock, nil
}

// writeNoteFile writes path, through a temp file named .<name>.<pid>.<n>.tmp
// (ignored by the notes .gitignore). exclusive fails when path exists.
func writeNoteFile(path string, data []byte, exclusive bool) error {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	tmp := filepath.Join(dir, fmt.Sprintf(".%s.%d.%d.tmp",
		strings.TrimSuffix(filepath.Base(path), ".md"), os.Getpid(), rand.Int63()))
	f, err := os.OpenFile(tmp, os.O_WRONLY|os.O_CREATE|os.O_EXCL|oNoFollow, 0o600)
	if err != nil {
		return err
	}
	if _, err := f.Write(data); err != nil {
		f.Close()
		os.Remove(tmp)
		return err
	}
	if err := f.Close(); err != nil {
		os.Remove(tmp)
		return err
	}
	if exclusive {
		defer os.Remove(tmp)
		return os.Link(tmp, path)
	}
	if err := os.Rename(tmp, path); err != nil {
		os.Remove(tmp)
		return err
	}
	return nil
}

func noteNow() string {
	return time.Now().UTC().Format(noteTimeFmt)
}

func noteSlugify(title string) string {
	s := strings.Trim(noteNonSlugRE.ReplaceAllString(strings.ToLower(title), "-"), "-")
	if len(s) > 60 {
		s = s[:60]
	}
	s = strings.TrimRight(s, "-")
	if s == "" {
		return "note"
	}
	return s
}

func cleanNoteTitle(t string) string {
	t = strings.Join(strings.Fields(t), " ")
	if utf8.RuneCountInString(t) > noteTitleMax {
		t = string([]rune(t)[:noteTitleMax])
	}
	return t
}

func cleanNoteTasks(s string) []string {
	var out []string
	for _, t := range strings.Split(s, ",") {
		t = strings.TrimSpace(t)
		if !noteTaskRE.MatchString(t) {
			continue
		}
		dup := false
		for _, o := range out {
			dup = dup || o == t
		}
		if !dup {
			out = append(out, t)
		}
	}
	if len(out) > 20 {
		out = out[:20]
	}
	return out
}

// parseNote splits a note file into its front matter and body; a file
// without front matter is all body.
func parseNote(text string) (map[string]string, string) {
	meta := map[string]string{}
	if !strings.HasPrefix(text, "---\n") {
		return meta, text
	}
	end := strings.Index(text[4:], "\n---\n")
	if end >= 0 {
		end += 4
	} else if strings.HasSuffix(text, "\n---") {
		end = len(text) - 4
	}
	if end < 0 {
		return meta, text
	}
	for _, line := range strings.Split(text[4:end], "\n") {
		k, v, ok := strings.Cut(line, ":")
		if ok && noteMetaKeyRE.MatchString(strings.TrimSpace(k)) {
			meta[strings.TrimSpace(k)] = strings.TrimSpace(v)
		}
	}
	if end+5 >= len(text) {
		return meta, ""
	}
	return meta, text[end+5:]
}

func renderNote(meta map[string]string, body string) string {
	created, updated := meta["created"], meta["updated"]
	if created == "" {
		created = noteNow()
	}
	if updated == "" {
		updated = noteNow()
	}
	head := []string{"title: " + cleanNoteTitle(meta["title"]), "created: " + created, "updated: " + updated}
	if tasks := cleanNoteTasks(meta["tasks"]); len(tasks) > 0 {
		head = append(head, "tasks: "+strings.Join(tasks, ", "))
	}
	body = strings.ReplaceAll(body, "\r\n", "\n")
	if body != "" && !strings.HasSuffix(body, "\n") {
		body += "\n"
	}
	return "---\n" + strings.Join(head, "\n") + "\n---\n" + body
}

func readNoteRaw(path string) (string, error) {
	f, err := os.OpenFile(path, os.O_RDONLY|oNoFollow, 0)
	if err != nil {
		if os.IsNotExist(err) {
			return "", newTaskError("not_found", "no such note")
		}
		return "", newTaskError("invalid_input", "the note cannot be read")
	}
	defer f.Close()
	if fi, err := f.Stat(); err != nil || !fi.Mode().IsRegular() {
		return "", newTaskError("invalid_input", "the note cannot be read")
	}
	raw, err := io.ReadAll(io.LimitReader(f, noteBodyMax+4096))
	if err != nil {
		return "", newTaskError("invalid_input", "the note cannot be read")
	}
	return strings.ToValidUTF8(string(raw), "�"), nil
}

func noteFromRaw(project, slug, path, raw string) *note {
	meta, body := parseNote(raw)
	title := cleanNoteTitle(meta["title"])
	if title == "" {
		title = slug
	}
	return &note{Project: project, Slug: slug, Path: path, Title: title, Created: meta["created"],
		Updated: meta["updated"], Tasks: cleanNoteTasks(meta["tasks"]), Body: body}
}

func readNote(reg *registry.Registry, project, slug string) (*note, error) {
	path, err := notePath(reg, project, slug)
	if err != nil {
		return nil, err
	}
	raw, err := readNoteRaw(path)
	if err != nil {
		return nil, err
	}
	return noteFromRaw(project, slug, path, raw), nil
}

// listNotes gives a project's notes, last updated first.
func listNotes(reg *registry.Registry, project string) ([]*note, error) {
	dir, err := noteProjectDir(reg, project)
	if err != nil {
		return nil, err
	}
	entries, err := os.ReadDir(dir)
	if err != nil && !os.IsNotExist(err) {
		return nil, err
	}
	var out []*note
	for _, e := range entries {
		slug, ok := strings.CutSuffix(e.Name(), ".md")
		if !ok || !noteSlugRE.MatchString(slug) || !e.Type().IsRegular() {
			continue
		}
		path := filepath.Join(dir, e.Name())
		raw, err := readNoteRaw(path)
		if err != nil {
			continue
		}
		out = append(out, noteFromRaw(project, slug, path, raw))
	}
	sort.SliceStable(out, func(i, j int) bool {
		ti, tj := out[i].UpdatedTime(), out[j].UpdatedTime()
		if !ti.Equal(tj) {
			return ti.After(tj)
		}
		return out[i].Slug < out[j].Slug
	})
	return out, nil
}

type noteProjectCount struct {
	Name  string
	Notes int
}

// listNoteProjects gives the projects that have notes, most notes first.
func listNoteProjects(reg *registry.Registry) ([]noteProjectCount, error) {
	entries, err := os.ReadDir(noteRoot(reg))
	if err != nil && !os.IsNotExist(err) {
		return nil, err
	}
	var out []noteProjectCount
	for _, e := range entries {
		if !e.IsDir() || !noteProjectRE.MatchString(e.Name()) {
			continue
		}
		files, err := os.ReadDir(filepath.Join(noteRoot(reg), e.Name()))
		if err != nil {
			continue
		}
		n := 0
		for _, f := range files {
			slug, ok := strings.CutSuffix(f.Name(), ".md")
			if ok && noteSlugRE.MatchString(slug) && f.Type().IsRegular() {
				n++
			}
		}
		if n > 0 {
			out = append(out, noteProjectCount{e.Name(), n})
		}
	}
	sort.SliceStable(out, func(i, j int) bool {
		if out[i].Notes != out[j].Notes {
			return out[i].Notes > out[j].Notes
		}
		return strings.ToLower(out[i].Name) < strings.ToLower(out[j].Name)
	})
	return out, nil
}

// createNote makes a new note; its slug comes from the title (-2, -3… when taken).
func createNote(reg *registry.Registry, project, title, body, tasks string) (*note, error) {
	title = cleanNoteTitle(title)
	if title == "" {
		return nil, newTaskError("invalid_input", "a note needs a title")
	}
	if len(body) > noteBodyMax {
		return nil, newTaskError("invalid_input", "the note is too long (512 KB at most)")
	}
	if _, err := noteProjectDir(reg, project); err != nil {
		return nil, err
	}
	base := noteSlugify(title)
	unlock, err := lockNotes(reg)
	if err != nil {
		return nil, err
	}
	defer unlock()
	for n := 1; n < 200; n++ {
		slug := base
		if n > 1 {
			slug = fmt.Sprintf("%s-%d", base[:min(len(base), 56)], n)
		}
		path, err := notePath(reg, project, slug)
		if err != nil {
			return nil, err
		}
		if _, err := os.Lstat(path); !os.IsNotExist(err) {
			continue
		}
		t := noteNow()
		text := renderNote(map[string]string{"title": title, "created": t, "updated": t, "tasks": tasks}, body)
		if err := writeNoteFile(path, []byte(text), true); err != nil {
			return nil, fmt.Errorf("failed to write note: %w", err)
		}
		return noteFromRaw(project, slug, path, text), nil
	}
	return nil, newTaskError("conflict", "too many notes with this title")
}

// appendNote adds text as a new paragraph at the end of the note.
func appendNote(reg *registry.Registry, project, slug, text string) (*note, error) {
	text = strings.Trim(text, "\n")
	if strings.TrimSpace(text) == "" {
		return nil, newTaskError("invalid_input", "nothing to add")
	}
	path, err := notePath(reg, project, slug)
	if err != nil {
		return nil, err
	}
	unlock, err := lockNotes(reg)
	if err != nil {
		return nil, err
	}
	defer unlock()
	raw, err := readNoteRaw(path)
	if err != nil {
		return nil, err
	}
	meta, body := parseNote(raw)
	sep := ""
	if strings.TrimSpace(body) != "" {
		sep = "\n\n"
	}
	body = strings.TrimRight(body, "\n") + sep + text + "\n"
	if len(body) > noteBodyMax {
		return nil, newTaskError("invalid_input", "the note is too long (512 KB at most)")
	}
	meta["updated"] = noteNow()
	out := renderNote(meta, body)
	if err := writeNoteFile(path, []byte(out), false); err != nil {
		return nil, fmt.Errorf("failed to write note: %w", err)
	}
	return noteFromRaw(project, slug, path, out), nil
}
