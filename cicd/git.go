package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"path"

	"github.com/a8m/envsubst"
	git "github.com/libgit2/git2go/v30"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
)

func (mgr *JobManager) CreateGitCheckoutCommands(cfg DeployConfig) ([]*DeployCommand, error) {
	cmds := []*DeployCommand{}
	root := path.Join(cfg.DeployPath, "..")
	log.Info().Msgf("Deploy Paths Root: %s Repo: %s", root, cfg.DeployPath)
	if _, err := os.Stat(root); os.IsNotExist(err) {
		if err = os.MkdirAll(root, 0755); err != nil {
			return nil, errors.Wrapf(err, "Unable to create git root %s", root)
		}
	}
	if _, err := os.Stat(cfg.DeployPath); err==nil {
		if err = os.RemoveAll(cfg.DeployPath); err != nil {
			return nil, errors.Wrapf(err, "Unable to remove deploy path %s", cfg.DeployPath)
		}
	}
	cmds = append(cmds, NewShellCmdDir(root, "git clone -b %s --single-branch --depth 1 %s %s", cfg.Branch, cfg.Repository, cfg.DeployPath))
	cmds = append(cmds, &DeployCommand{"CopyConfig", copyDeployConfig, ""})
	return cmds, nil
}

func copyDeployConfig(job *DeployContext, pc *ProcessContext) error {
	cfg := job.Config
	log.Info().Msgf("Copy Config Deploy Path: %s Config: %s", cfg.DeployPath, cfg.ConfigPath)
	if _, err := os.Stat(cfg.DeployPath); os.IsNotExist(err) {
		pc.Log("Make Deployment Directory %s", cfg.DeployPath)
		if err := os.MkdirAll(cfg.DeployPath, 0755); err != nil {
			return errors.Wrapf(err, "Unable to make folder %s", cfg.DeployPath)
		}
	}
	if _, err := os.Stat(cfg.ConfigPath); !os.IsNotExist(err) {
		pc.Log("Read From Config Directory %s", cfg.ConfigPath)
		files, err := ioutil.ReadDir(cfg.ConfigPath)
		if err != nil {
			return errors.Wrapf(err, "Unable to read directory %s", cfg.ConfigPath)
		}
		for _, f := range files {
			if f.IsDir() {
				pc.Log("Skip Directory %s", f.Name())
				continue
			}
			sourcePath := path.Join(cfg.ConfigPath, f.Name())
			pc.Log("Transform %s", sourcePath)
			input, err := envsubst.ReadFile(sourcePath)
			if err != nil {
				return errors.Wrapf(err, "Unable to read file %s", sourcePath)
			}
			destinationFile := path.Join(cfg.DeployPath, path.Base(sourcePath))
			pc.Log("Write Transformed %s", destinationFile)
			err = ioutil.WriteFile(destinationFile, input, 0600)
			if err != nil {
				return errors.Wrapf(err, "Unable to write file %s", destinationFile)
			}
		}
	}
	return nil
}

func (mgr *JobManager) CheckoutGitRepo(cfg DeployConfig) error {
	if _, err := os.Stat(cfg.GitPath); os.IsNotExist(err) {
		if err = os.MkdirAll(cfg.GitPath, 0755); err != nil {
			return errors.Wrapf(err, "Unable to create git dir")
		}
	}
	var repo *git.Repository
	if _, err := os.Stat(path.Join(cfg.GitPath, "HEAD")); os.IsNotExist(err) {
		if repo, err = gitClone(cfg); err != nil {
			return err
		}
	} else if repo, err = gitFetch(cfg); err != nil {
		return err
	}
	defer repo.Free()
	if err := gitCheckout(cfg, repo); err != nil {
		return err
	}
	return nil
}

func gitFetch(cfg DeployConfig) (*git.Repository, error) {
	repo, err := git.OpenRepository(cfg.GitPath)
	if err != nil {
		return nil, errors.Wrapf(err, "Unable to fetch git repo %s", cfg.GitPath)
	}
	fetchOpts := git.FetchOptions{
		DownloadTags: git.DownloadTagsAll,
		RemoteCallbacks: git.RemoteCallbacks{
			CredentialsCallback:      credCallback(cfg),
			CertificateCheckCallback: certificateCheckCallback,
		},
	}
	org, err := repo.Remotes.Lookup("origin")
	if err != nil {
		return nil, errors.Wrapf(err, "Unable to lookup remote origin")
	}
	defer org.Free()
	if err = org.Fetch([]string{cfg.Branch}, &fetchOpts, ""); err != nil {
		return nil, errors.Wrapf(err, "Unable to fetch %s", cfg.Branch)
	}
	return repo, nil
}

func gitCheckout(cfg DeployConfig, repo *git.Repository) error {
	if _, err := os.Stat(cfg.DeployPath); !os.IsNotExist(err) {
		if err = os.RemoveAll(cfg.DeployPath); err != nil {
			return errors.Wrapf(err, "Unable to delete deploy folder %s", cfg.DeployPath)
		}
	}
	if err := os.MkdirAll(cfg.DeployPath, 0755); err != nil {
		return errors.Wrapf(err, "Unable to make folder %s", cfg.DeployPath)
	}
	checkoutOpts := git.CheckoutOpts{
		Strategy:        git.CheckoutForce | git.CheckoutRecreateMissing | git.CheckoutAllowConflicts | git.CheckoutUseTheirs,
		TargetDirectory: cfg.DeployPath,
	}
	br, err := repo.LookupBranch("origin/"+cfg.Branch, git.BranchRemote)
	defer br.Free()
	if err != nil {
		return errors.Wrapf(err, "Unable to get branch %s", "origin/"+cfg.Branch)
	}
	commit, err := repo.LookupCommit(br.Target())
	defer commit.Free()
	if err != nil {
		return errors.Wrapf(err, "Unable to lookup commit for branch %s", "origin/"+cfg.Branch)
	}
	if err := repo.CheckoutHead(&checkoutOpts); err != nil {
		return errors.Wrapf(err, "Unable to checkout head target %s", cfg.DeployPath)
	}
	if err := repo.ResetToCommit(commit, git.ResetHard, &checkoutOpts); err != nil {
		return errors.Wrapf(err, "Unable to reset head target %s", cfg.DeployPath)
	}
	if _, err := os.Stat(cfg.ConfigPath); !os.IsNotExist(err) {
		if err = copyConfig(cfg); err != nil {
			return err
		}
	}
	return nil
}

func copyConfig(cfg DeployConfig) error {
	files, err := ioutil.ReadDir(cfg.ConfigPath)
	if err != nil {
		return errors.Wrapf(err, "Unable to read directory %s", cfg.ConfigPath)
	}
	for _, f := range files {
		if f.IsDir() {
			continue
		}
		sourcePath := path.Join(cfg.ConfigPath, f.Name())
		input, err := envsubst.ReadFile(sourcePath)
		if err != nil {
			return errors.Wrapf(err, "Unable to read file %s", sourcePath)
		}
		destinationFile := path.Join(cfg.DeployPath, path.Base(sourcePath))
		err = ioutil.WriteFile(destinationFile, input, 0600)
		if err != nil {
			return errors.Wrapf(err, "Unable to write file %s", destinationFile)
		}
	}
	return nil
}

func gitClone(cfg DeployConfig) (*git.Repository, error) {
	cloneOptions := &git.CloneOptions{
		Bare:           true,
		CheckoutBranch: cfg.Branch,
		FetchOptions: &git.FetchOptions{
			DownloadTags: git.DownloadTagsAll,
			RemoteCallbacks: git.RemoteCallbacks{
				CredentialsCallback:      credCallback(cfg),
				CertificateCheckCallback: certificateCheckCallback,
			},
		},
	}
	repo, err := git.Clone(cfg.Repository, cfg.GitPath, cloneOptions)
	if err != nil {
		return nil, errors.Wrapf(err, "Unable to clone %s %s", cfg.Repository, cfg.GitPath)
	}
	return repo, nil
}

func credCallback(cfg DeployConfig) func(url string, username_from_url string, allowed_types git.CredType) (*git.Cred, error) {
	fmt.Printf("Cert %s\n", cfg.CertPath+".pub")
	return func(url string, username_from_url string, allowed_types git.CredType) (*git.Cred, error) {
		return git.NewCredSshKey("git", cfg.CertPath+".pub", cfg.CertPath, "")
	}
}

// Made this one just return 0 during troubleshooting...
func certificateCheckCallback(cert *git.Certificate, valid bool, hostname string) git.ErrorCode {
	return 0
}
