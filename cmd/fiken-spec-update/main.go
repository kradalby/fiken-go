// Command fiken-spec-update fetches the canonical Fiken OpenAPI spec
// from api.fiken.no, diffs it against the vendored copy, and (on
// confirm) replaces the vendored file plus the provenance record in
// api/SOURCE.txt. Optional --apply skips the prompt.
package main

import (
	"crypto/sha256"
	"encoding/hex"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"time"
)

const (
	specURL  = "https://api.fiken.no/api/v2/docs/swagger.yaml"
	specPath = "api/fiken-openapi.yaml"
	sourceP  = "api/SOURCE.txt"
)

func main() {
	apply := flag.Bool("apply", false, "skip the diff prompt and overwrite the vendored spec")
	diffOnly := flag.Bool("diff-only", false, "print the diff (or '(no changes)') and exit; do not write")
	flag.Parse()

	if err := run(*apply, *diffOnly); err != nil {
		fmt.Fprintf(os.Stderr, "fiken-spec-update: %v\n", err)
		os.Exit(1)
	}
}

func run(apply, diffOnly bool) error {
	tmp, err := os.CreateTemp("", "fiken-spec-*.yaml")
	if err != nil {
		return fmt.Errorf("temp file: %w", err)
	}
	tmpPath := tmp.Name()
	defer os.Remove(tmpPath)
	tmp.Close()

	if err := fetch(specURL, tmpPath); err != nil {
		return err
	}

	same, err := equalFiles(tmpPath, specPath)
	if err != nil {
		return err
	}
	if same {
		fmt.Println("(no changes)")
		return nil
	}

	if err := showDiff(specPath, tmpPath); err != nil {
		return err
	}

	if diffOnly {
		return nil
	}

	if !apply {
		fmt.Print("\nApply changes? [y/N]: ")
		var ans string
		fmt.Scanln(&ans)
		if ans != "y" && ans != "Y" {
			fmt.Println("aborted.")
			return nil
		}
	}

	if err := os.Rename(tmpPath, specPath); err != nil {
		if err := copyFile(tmpPath, specPath); err != nil {
			return err
		}
	}

	digest, err := computeSHA256(specPath)
	if err != nil {
		return err
	}
	f, err := os.Create(sourceP)
	if err != nil {
		return err
	}
	defer f.Close()
	if err := renderSource(f, specURL, time.Now().UTC().Format(time.RFC3339), digest); err != nil {
		return err
	}
	fmt.Printf("updated %s and %s (sha256 %s)\n", specPath, sourceP, digest)
	return nil
}

func fetch(url, dst string) error {
	resp, err := http.Get(url)
	if err != nil {
		return fmt.Errorf("get %s: %w", url, err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("get %s: %s", url, resp.Status)
	}
	f, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer f.Close()
	_, err = io.Copy(f, resp.Body)
	return err
}

func equalFiles(a, b string) (bool, error) {
	if _, err := os.Stat(b); os.IsNotExist(err) {
		return false, nil
	}
	da, err := computeSHA256(a)
	if err != nil {
		return false, err
	}
	db, err := computeSHA256(b)
	if err != nil {
		return false, err
	}
	return da == db, nil
}

func computeSHA256(path string) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer f.Close()
	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return "", err
	}
	return hex.EncodeToString(h.Sum(nil)), nil
}

func renderSource(w io.Writer, url, fetchedAt, sha string) error {
	_, err := fmt.Fprintf(
		w,
		"source-url: %s\nfetched-at: %s\nsha256:     %s\n",
		url, fetchedAt, sha,
	)
	return err
}

func showDiff(oldPath, newPath string) error {
	bin, err := exec.LookPath("difft")
	args := []string{oldPath, newPath}
	if err == nil {
		cmd := exec.Command(bin, args...)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		_ = cmd.Run()
		return nil
	}
	cmd := exec.Command("git", "--no-pager", "diff", "--no-index", "--color=always", oldPath, newPath)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	_ = cmd.Run()
	return nil
}

func copyFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()
	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()
	_, err = io.Copy(out, in)
	return err
}
