package runner

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/lukaswoellhaf/jitgen/internal/diff"
)

const testFileName = "jitgen_catch_test.go"

// Result holds the outcome of a differential test run.
type Result struct {
	ParentPassed bool
	ChildPassed  bool
	IsCatch      bool
	ParentOutput string
	ChildOutput  string
}

// Run executes the generated test on both parent and child versions using git worktrees.
// It returns a Result indicating whether the test is a catch or a discard.
func Run(repoPath, parentRef, childRef, testCode, diffOutput string) (*Result, error) {
	// Determine target package directory from the diff
	pkgDir := targetPkgDir(diffOutput)

	// Create temp directory for worktrees
	tmpDir, err := os.MkdirTemp("", "jitgen-")
	if err != nil {
		return nil, fmt.Errorf("create temp dir: %w", err)
	}
	defer os.RemoveAll(tmpDir)

	parentDir := filepath.Join(tmpDir, "parent")
	childDir := filepath.Join(tmpDir, "child")

	// Create worktrees
	if err := createWorktree(repoPath, parentDir, parentRef); err != nil {
		return nil, fmt.Errorf("create parent worktree: %w", err)
	}
	defer removeWorktree(repoPath, parentDir)

	if err := createWorktree(repoPath, childDir, childRef); err != nil {
		return nil, fmt.Errorf("create child worktree: %w", err)
	}
	defer removeWorktree(repoPath, childDir)

	// Inject test file and run on parent
	parentTestDir := filepath.Join(parentDir, pkgDir)
	if err := os.WriteFile(filepath.Join(parentTestDir, testFileName), []byte(testCode), 0644); err != nil {
		return nil, fmt.Errorf("inject test into parent: %w", err)
	}
	parentOutput, parentPassed := runTest(parentTestDir)

	// Inject test file and run on child
	childTestDir := filepath.Join(childDir, pkgDir)
	if err := os.WriteFile(filepath.Join(childTestDir, testFileName), []byte(testCode), 0644); err != nil {
		return nil, fmt.Errorf("inject test into child: %w", err)
	}
	childOutput, childPassed := runTest(childTestDir)

	result := &Result{
		ParentPassed: parentPassed,
		ChildPassed:  childPassed,
		ParentOutput: parentOutput,
		ChildOutput:  childOutput,
	}
	result.IsCatch = result.ParentPassed && !result.ChildPassed

	return result, nil
}

// targetPkgDir returns the directory of the first non-test .go file touched by the diff.
// Falls back to "." (repo root) if no Go files are found.
func targetPkgDir(diffOutput string) string {
	for _, f := range diff.ParseTouchedFiles(diffOutput) {
		if strings.HasSuffix(f, ".go") && !strings.HasSuffix(f, "_test.go") {
			return filepath.Dir(f)
		}
	}
	return "."
}

func createWorktree(repoPath, worktreeDir, ref string) error {
	cmd := exec.Command("git", "worktree", "add", "--detach", worktreeDir, ref)
	cmd.Dir = repoPath
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("%s: %s", err, string(out))
	}
	return nil
}

func removeWorktree(repoPath, worktreeDir string) {
	cmd := exec.Command("git", "worktree", "remove", "--force", worktreeDir)
	cmd.Dir = repoPath
	_ = cmd.Run()
}

func runTest(dir string) (output string, passed bool) {
	cmd := exec.Command("go", "test", "-run", "(?i)^testjitgen", "-count=1", "-timeout", "30s", ".")
	cmd.Dir = dir
	out, err := cmd.CombinedOutput()
	return string(out), err == nil
}

// CreateCatchBranch creates a branch in the target repo with the catching test committed.
// The branch is based on childRef and named jitgen/catch/<childRef-basename>.
func CreateCatchBranch(repoPath, childRef, testCode, diffOutput string) (string, error) {
	// Derive a branch name from the child ref
	baseName := filepath.Base(childRef)
	branchName := "jitgen/catch/" + baseName

	pkgDir := targetPkgDir(diffOutput)

	// Create a temporary worktree for the new branch
	tmpDir, err := os.MkdirTemp("", "jitgen-catch-")
	if err != nil {
		return "", fmt.Errorf("create temp dir: %w", err)
	}
	defer os.RemoveAll(tmpDir)
	worktreeDir := filepath.Join(tmpDir, "catch")
	defer removeWorktree(repoPath, worktreeDir)

	// Delete stale branch from a previous run, if it exists
	checkCmd := exec.Command("git", "rev-parse", "--verify", branchName)
	checkCmd.Dir = repoPath
	if checkCmd.Run() == nil {
		delCmd := exec.Command("git", "branch", "-D", branchName)
		delCmd.Dir = repoPath
		if out, err := delCmd.CombinedOutput(); err != nil {
			return "", fmt.Errorf("delete stale branch: %s: %s", err, string(out))
		}
	}

	// Create a new branch from childRef
	cmd := exec.Command("git", "worktree", "add", "-b", branchName, worktreeDir, childRef)
	cmd.Dir = repoPath
	if out, err := cmd.CombinedOutput(); err != nil {
		return "", fmt.Errorf("create branch worktree: %s: %s", err, string(out))
	}

	// Write the test file
	testDir := filepath.Join(worktreeDir, pkgDir)
	testFile := filepath.Join(testDir, testFileName)
	if err := os.WriteFile(testFile, []byte(testCode), 0644); err != nil {
		return "", fmt.Errorf("write test file: %w", err)
	}

	// Stage and commit
	addCmd := exec.Command("git", "add", testFile)
	addCmd.Dir = worktreeDir
	if out, err := addCmd.CombinedOutput(); err != nil {
		return "", fmt.Errorf("git add: %s: %s", err, string(out))
	}

	commitCmd := exec.Command("git", "commit", "-m", "jitgen: add catching test for "+baseName)
	commitCmd.Dir = worktreeDir
	if out, err := commitCmd.CombinedOutput(); err != nil {
		return "", fmt.Errorf("git commit: %s: %s", err, string(out))
	}

	return branchName, nil
}

// PushCatchBranch pushes the given branch to the remote.
func PushCatchBranch(repoPath, branchName string) error {
	cmd := exec.Command("git", "push", "origin", branchName, "--force")
	cmd.Dir = repoPath
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("git push: %s: %s", err, string(out))
	}
	return nil
}
