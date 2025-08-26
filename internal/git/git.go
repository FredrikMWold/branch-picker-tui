package git

import (
	"bufio"
	"errors"
	"fmt"
	"os/exec"
	"strings"
)

func runGit(args ...string) (string, error) {
	cmd := exec.Command("git", args...)
	out, err := cmd.Output()
	if err != nil {
		var ee *exec.ExitError
		if errors.As(err, &ee) {
			return "", fmt.Errorf("git %v failed: %v\n%s", args, err, string(ee.Stderr))
		}
		return "", err
	}
	return string(out), nil
}

// Branch holds info for a git branch
type Branch struct {
	Name     string
	Current  bool
	IsRemote bool
	Remote   string
	// Upstream is the configured upstream for the local branch (e.g. "origin/main").
	// Empty if no upstream is configured.
	Upstream string
}

// ListBranches lists local branches with current indicator
func ListBranches() ([]Branch, error) {
	// Sort local branches by most recent activity (committer date descending)
	// and include the configured upstream (short) for each branch.
	// Using for-each-ref gives us consistent access to %(upstream:short).
	out, err := runGit("for-each-ref", "--sort=-committerdate", "--format", "%(refname:short)\t%(upstream:short)", "refs/heads")
	if err != nil {
		return nil, err
	}
	cur, _ := CurrentBranch()
	var res []Branch
	s := bufio.NewScanner(strings.NewReader(out))
	for s.Scan() {
		line := strings.TrimSpace(s.Text())
		if line == "" {
			continue
		}
		// Expect: name\tupstream (upstream may be empty)
		parts := strings.SplitN(line, "\t", 2)
		name := strings.TrimSpace(parts[0])
		var upstream string
		if len(parts) > 1 {
			upstream = strings.TrimSpace(parts[1])
		}
		if name == "" {
			continue
		}
		res = append(res, Branch{Name: name, Current: name == cur, Upstream: upstream})
	}
	return res, nil
}

// CurrentBranch returns the current branch name (short)
func CurrentBranch() (string, error) {
	out, err := runGit("rev-parse", "--abbrev-ref", "HEAD")
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(out), nil
}

// Checkout switches to a branch
func Checkout(branch string) error {
	if branch == "" {
		return fmt.Errorf("branch required")
	}
	_, err := runGit("checkout", branch)
	return err
}

// CreateBranch creates a new branch from current HEAD
func CreateBranch(branch string) error {
	if branch == "" {
		return fmt.Errorf("branch required")
	}
	_, err := runGit("branch", branch)
	return err
}

// DeleteBranch deletes a local branch. When force is true, uses -D to force delete.
func DeleteBranch(branch string, force bool) error {
	if branch == "" {
		return fmt.Errorf("branch required")
	}
	flag := "-d"
	if force {
		flag = "-D"
	}
	_, err := runGit("branch", flag, branch)
	return err
}
