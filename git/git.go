package git

import (
	"fmt"
	"os/exec"
	"time"
)

// CommitAndPush stages all changes, commits with a timestamp, and pushes
func CommitAndPush(backupDir string, dryRun bool) error {
	if dryRun {
		fmt.Printf("[dry-run] git -C %s add -A\n", backupDir)
		fmt.Printf("[dry-run] git -C %s diff --quiet || git -C %s commit -m \"gitbak backup: %s\"\n", backupDir, backupDir, time.Now().Format("2006-01-02 15:04:05"))
		fmt.Printf("[dry-run] git -C %s push\n", backupDir)
		return nil
	}

	// First add all changes
	cmdAdd := exec.Command("git", "-C", backupDir, "add", "-A")
	if out, err := cmdAdd.CombinedOutput(); err != nil {
		return fmt.Errorf("git add failed: %v - %s", err, string(out))
	}

	// Check if there are any changes to commit
	cmdDiff := exec.Command("git", "-C", backupDir, "diff", "--quiet", "--cached")
	if err := cmdDiff.Run(); err == nil {
		// No changes to commit
		fmt.Println("No changes to commit")
		return nil
	}

	// If we get here, there are changes to commit
	msg := fmt.Sprintf("gitbak backup: %s", time.Now().Format("2006-01-02 15:04:05"))
	cmdCommit := exec.Command("git", "-C", backupDir, "commit", "-m", msg)
	if out, err := cmdCommit.CombinedOutput(); err != nil {
		return fmt.Errorf("git commit failed: %v - %s", err, string(out))
	}

	// Only push if there was a commit
	cmdPush := exec.Command("git", "-C", backupDir, "push")
	if out, err := cmdPush.CombinedOutput(); err != nil {
		return fmt.Errorf("git push failed: %v - %s", err, string(out))
	}
	return nil
}
