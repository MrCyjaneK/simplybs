package utils

import (
	"log"
	"os"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
)

func optimizeGitRepo(repoPath string) {
	log.Printf("Optimizing repository at %s", repoPath)

	repo, err := git.PlainOpen(repoPath)
	if err != nil {
		log.Printf("Warning: Failed to open repository at %s: %v", repoPath, err)
		return
	}

	remotes, err := repo.Remotes()
	if err == nil {
		for _, remote := range remotes {
			err = repo.DeleteRemote(remote.Config().Name)
			if err != nil {
				log.Printf("Warning: Failed to remove remote %s in %s: %v", remote.Config().Name, repoPath, err)
			}
		}
	}

	refs, err := repo.References()
	if err == nil {
		head, headErr := repo.Head()
		if headErr == nil {
			err = refs.ForEach(func(ref *plumbing.Reference) error {
				if ref.Name().IsBranch() && ref.Hash() != head.Hash() {
					return repo.Storer.RemoveReference(ref.Name())
				}
				return nil
			})
			if err != nil {
				log.Printf("Warning: Failed to remove some branches in %s: %v", repoPath, err)
			}
		}
	}

	log.Printf("Repository optimization completed for %s", repoPath)
}

func resolveRef(repo *git.Repository, refStr string) (plumbing.Hash, error) {
	if len(refStr) == 40 {
		hash := plumbing.NewHash(refStr)
		_, err := repo.CommitObject(hash)
		if err == nil {
			return hash, nil
		}
	}

	tagRef, err := repo.Tag(refStr)
	if err == nil {
		return tagRef.Hash(), nil
	}

	branchRef, err := repo.Reference(plumbing.ReferenceName("refs/heads/"+refStr), true)
	if err == nil {
		return branchRef.Hash(), nil
	}

	remoteRef, err := repo.Reference(plumbing.ReferenceName("refs/remotes/origin/"+refStr), true)
	if err == nil {
		return remoteRef.Hash(), nil
	}

	ref, err := repo.Reference(plumbing.ReferenceName(refStr), true)
	if err == nil {
		return ref.Hash(), nil
	}

	return plumbing.ZeroHash, err
}

func DownloadGit(packageName, path, url, expectedSha256 string) error {
	log.Printf("Downloading %s to %s", url, path)

	repo, err := git.PlainClone(path, false, &git.CloneOptions{
		URL:      url,
		Progress: os.Stdout,
	})
	if err != nil {
		return err
	}

	worktree, err := repo.Worktree()
	if err != nil {
		return err
	}

	log.Printf("Checking out reference %s", expectedSha256)
	targetHash, err := resolveRef(repo, expectedSha256)
	if err != nil {
		return err
	}

	err = worktree.Checkout(&git.CheckoutOptions{
		Hash: targetHash,
	})
	if err != nil {
		return err
	}

	log.Printf("Cleaning up and optimizing repository...")

	optimizeGitRepo(path)

	log.Printf("Successfully downloaded and optimized %s", path)
	return nil
}
