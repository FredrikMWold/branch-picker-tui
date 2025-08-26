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
}

// ListBranches lists local branches with current indicator
func ListBranches() ([]Branch, error) {
	// Sort local branches by most recent activity (committer date descending)
	out, err := runGit("branch", "--sort=-committerdate", "--format", "%(refname:short)")
	if err != nil {
		return nil, err
	}
	cur, _ := CurrentBranch()
	var res []Branch
	s := bufio.NewScanner(strings.NewReader(out))
	for s.Scan() {
		name := strings.TrimSpace(s.Text())
		if name == "" {
			continue
		}
		res = append(res, Branch{Name: name, Current: name == cur})
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
