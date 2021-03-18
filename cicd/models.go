package main

import (
	"fmt"
	"path"
	"strings"
	"sync"
	"time"

	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"

	"bitbucket.org/xhumiq/go-mclib/microservice"
)

type JobManager struct {
	*microservice.App
	SvcConfig *AppConfig
	Configs   map[string]*DeployConfig
	Jobs      map[string]*DeployContext
	ErrorCh   chan error
	mutex     *sync.Mutex
}

type DeployContext struct {
	Current  *ProcessContext
	Config   DeployConfig
	Requests chan *DeployConfig
	CancelFn func()
	mutex    sync.Mutex
	reqWait  sync.WaitGroup
	Error    error
	Status   string
	History  []*ProcessContext
}

type DeployConfig struct {
	RepoName   string
	Repository string
	Branch     string
	DeployPath string
	GitPath    string
	ConfigPath string
	Commands   []*DeployCommand
	EnvPrefix  string
	Env        map[string]string
	Timeout    int
	CertPath   string
}

func NewDeployConfig(repo string, branch string, deployPath string, gitPath string, configPath string) *DeployConfig {
	cfg := DeployConfig{
		RepoName: strings.TrimSpace(repo),
		Branch:   strings.TrimSpace(branch),
		Timeout:  300,
	}
	if cfg.RepoName == "" || cfg.Branch == "" {
		checkError(errors.Errorf("Deployment configuration has an empty Repo or Branch name"))
	}
	br := cfg.Branch
	if br == "master" || br == "beta" {
		br = "release"
	}
	if br == "qa" || br == "test" {
		br = "test"
	}
	log.Info().Msgf("Branch Config %s", br)
	cfg.DeployPath = strings.TrimSpace(deployPath)
	if len(cfg.DeployPath) < 1 {
		cfg.DeployPath = path.Join(cfg.RepoName, br)
	}
	cfg.GitPath = strings.TrimSpace(gitPath)
	if len(cfg.GitPath) < 1 {
		cfg.GitPath = path.Join(cfg.RepoName, br)
	}
	cfg.ConfigPath = strings.TrimSpace(configPath)
	if len(cfg.ConfigPath) < 1 {
		cfg.ConfigPath = path.Join(cfg.RepoName, br)
	}
	return &cfg
}

type DeployCommand struct {
	Command    string                                             `json:"command,omitempty" yaml:"command,omitempty"`
	Thunk      func(job *DeployContext, pc *ProcessContext) error `json:"-" yaml:"-"`
	WorkingDir string
}

func NewShellCmd(command string, args ...interface{}) *DeployCommand {
	return &DeployCommand{fmt.Sprintf(command, args...), nil, ""}
}

func NewShellCmdDir(dir string, command string, args ...interface{}) *DeployCommand {
	return &DeployCommand{fmt.Sprintf(command, args...), nil, dir}
}

func NewShellCmds(commands []string) []*DeployCommand {
	args := []*DeployCommand{}
	for _, arg := range commands {
		args = append(args, NewShellCmd(arg))
	}
	return args
}

type ProcessContext struct {
	CmdIndex      int            `json:"cmdIndex,omitempty" yaml:"cmdIndex,omitempty"`
	Command       DeployCommand  `json:"command,omitempty" yaml:"command,omitempty"`
	Error         error          `json:"-,omitempty" yaml:"-,omitempty"`
	ErrorResponse *ResponseError `json:"error,omitempty" yaml:"error,omitempty"`
	ExitCode      int            `json:"exitCode,omitempty" yaml:"exitCode,omitempty"`
	Started       time.Time      `json:"started,omitempty" yaml:"started,omitempty"`
	Finished      *time.Time     `json:"finished,omitempty" yaml:"finished,omitempty"`
	Output        []string       `json:"output,omitempty" yaml:"output,omitempty"`
	Completed     bool           `json:"completed,omitempty" yaml:"completed,omitempty"`
}

func (pc *ProcessContext) Log(message string, args ...interface{}) {
	log.Info().Msgf(message, args)
	pc.Output = append(pc.Output, fmt.Sprintf(message, args))
}

type ProcessLineOut struct {
	Line    string `json:"line,omitempty" yaml:"line,omitempty"`
	IsError bool   `json:"isError,omitempty" yaml:"isError,omitempty"`
}

type ComboEvent struct {
	Actor       Actor       `json:"actor"`
	Repository  Repository  `json:"repository"`
	PullRequest PullRequest `json:"pullrequest"`
	Push        struct {
		Changes []struct {
			Forced    bool     `json:"forced"`
			Old       OldOrNew `json:"old"`
			New       OldOrNew `json:"new"`
			Closed    bool     `json:"closed"`
			Created   bool     `json:"created"`
			Truncated bool     `json:"truncated"`
			Links     `json:"links"`
			Commits   []Commit `json:"commits"`
		} `json:"changes"`
	} `json:"push"`
}

func (event ComboEvent) GetRepoBranch() (string, string) {
	repo := event.Repository.Name
	branch := "Unknown"
	if event.Push.Changes != nil {
		for _, chg := range event.Push.Changes {
			if chg.New.Type == "branch" {
				branch = chg.New.Name
				bs := strings.Split(branch, "/")
				if len(bs) > 1 {
					branch = bs[0]
				}
				return repo, branch
			}
		}
	}
	return repo, branch
}

// RepoPushEvent https://confluence.atlassian.com/bitbucket/event-payloads-740262817.html#EventPayloads-Push
type RepoPushEvent struct {
	Actor      Actor      `json:"actor"`
	Repository Repository `json:"repository"`
	Push       struct {
		Changes []struct {
			Forced    bool     `json:"forced"`
			Old       OldOrNew `json:"old"`
			New       OldOrNew `json:"new"`
			Closed    bool     `json:"closed"`
			Created   bool     `json:"created"`
			Truncated bool     `json:"truncated"`
			Links     `json:"links"`
			Commits   []Commit `json:"commits"`
		} `json:"changes"`
	} `json:"push"`
}

// Links is a common struct used in several types. Refer to the event documentation
// to find out which link types are populated in which events.
type Links struct {
	Avatar struct {
		Href string `json:"href"`
	} `json:"avatar"`
	HTML struct {
		Href string `json:"href"`
	} `json:"html"`
	Self struct {
		Href string `json:"href"`
	} `json:"self"`
	Commits struct {
		Href string `json:"href"`
	} `json:"commits"`
	Commit struct {
		Href string `json:"href"`
	} `json:"commit"`
}

// OldOrNew is used in the RepoPushEvent type
type OldOrNew struct {
	Repository struct {
		FullName string `json:"full_name"`
		UUID     string `json:"uuid"`
		Links    Links  `json:"links"`
		Name     string `json:"name"`
		Type     string `json:"type"`
	} `json:"repository"`
	Target struct {
		Date    *time.Time `json:"date"`
		Parents []struct {
			Hash  string `json:"hash"`
			Links Links  `json:"links"`
			Type  string `json:"type"`
		} `json:"parents"`
		Message string `json:"message"`
		Hash    string `json:"hash"`
		Author  Author `json:"author"`
		Links   Links  `json:"links"`
		Type    string `json:"type"`
	} `json:"target"`
	Links Links  `json:"links"`
	Name  string `json:"name"`
	Type  string `json:"type"`
}

// Author is a common struct used in several types
type Author struct {
	Raw  string `json:"raw"`
	User struct {
		Username    string `json:"username"`
		Type        string `json:"type"`
		UUID        string `json:"uuid"`
		Links       Links  `json:"links"`
		DisplayName string `json:"display_name"`
	} `json:"user"`
}

// Repository is a common struct used in several types
type Repository struct {
	Scm      string `json:"scm"`
	FullName string `json:"full_name"`
	Type     string `json:"type"`
	Website  string `json:"website"`
	Owner    struct {
		Username    string `json:"username"`
		Type        string `json:"type"`
		UUID        string `json:"uuid"`
		Links       Links  `json:"links"`
		DisplayName string `json:"display_name"`
	} `json:"owner"`
	UUID      string `json:"uuid"`
	Links     Links  `json:"links"`
	Name      string `json:"name"`
	IsPrivate bool   `json:"is_private"`
}

// Commit is a common struct used in several types
type Commit struct {
	Date    time.Time `json:"date"`
	Parents []struct {
		Hash  string `json:"hash"`
		Links Links  `json:"self"`
		Type  string `json:"type"`
	} `json:"parents"`
	Message string `json:"message"`
	Hash    string `json:"hash"`
	Author  Author `json:"author"`
	Links   Links  `json:"links"`
	Type    string `json:"type"`
}

// Actor is a common struct used in several types
type Actor struct {
	Username    string `json:"username"`
	Type        string `json:"type"`
	UUID        string `json:"uuid"`
	Links       Links  `json:"links"`
	DisplayName string `json:"display_name"`
}

// RepoForkEvent https://confluence.atlassian.com/bitbucket/event-payloads-740262817.html#EventPayloads-Fork
type RepoForkEvent struct {
	Actor      Actor      `json:"actor"`
	Repository Repository `json:"repository"`
	Fork       Repository `json:"fork"`
}

// Comment https://confluence.atlassian.com/bitbucket/event-payloads-740262817.html#EventPayloads-entity_comment
type Comment struct {
	ID     int `json:"id"`
	Parent struct {
		ID int `json:"id"`
	} `json:"parent"`
	Content struct {
		Raw    string `json:"raw"`
		HTML   string `json:"html"`
		Markup string `json:"markup"`
	} `json:"content"`
	Inline struct {
		Path string      `json:"path"`
		From interface{} `json:"from"`
		To   int         `json:"to"`
	} `json:"inline"`
	CreatedOn *time.Time `json:"created_on"`
	UpdatedOn *time.Time `json:"updated_on"`
	Links     Links      `json:"links"`
}

// RepoCommitCommentCreatedEvent https://confluence.atlassian.com/bitbucket/event-payloads-740262817.html#EventPayloads-CommitCommentCreated
type RepoCommitCommentCreatedEvent struct {
	Actor      Actor      `json:"actor"`
	Comment    Comment    `json:"comment"`
	Repository Repository `json:"repository"`
	Commit     struct {
		Hash string `json:"hash"`
	} `json:"commit"`
}

// A RepoCommitStatusEvent is not a BB event. This is the base for several CommitStatus* events.
type RepoCommitStatusEvent struct {
	Actor        Actor      `json:"actor"`
	Repository   Repository `json:"repository"`
	CommitStatus struct {
		Name        string     `json:"name"`
		Description string     `json:"description"`
		State       string     `json:"state"`
		Key         string     `json:"key"`
		URL         string     `json:"url"`
		Type        string     `json:"type"`
		CreatedOn   *time.Time `json:"created_on"`
		UpdatedOn   *time.Time `json:"updated_on"`
		Links       Links      `json:"links"`
	} `json:"commit_status"`
}

// RepoCommitStatusCreatedEvent https://confluence.atlassian.com/bitbucket/event-payloads-740262817.html#EventPayloads-CommitStatusCreated
type RepoCommitStatusCreatedEvent struct {
	RepoCommitStatusEvent
}

// RepoCommitStatusUpdatedEvent https://confluence.atlassian.com/bitbucket/event-payloads-740262817.html#EventPayloads-CommitStatusUpdated
type RepoCommitStatusUpdatedEvent struct {
	RepoCommitStatusEvent
}

// An IssueEvent is not a BB event. This is the base for several Issue* events.
type IssueEvent struct {
	Actor      Actor      `json:"actor"`
	Issue      Issue      `json:"issue"`
	Repository Repository `json:"repository"`
}

// IssueCreatedEvent https://confluence.atlassian.com/bitbucket/event-payloads-740262817.html#EventPayloads-Created
type IssueCreatedEvent struct {
	IssueEvent
}

// IssueUpdatedEvent https://confluence.atlassian.com/bitbucket/event-payloads-740262817.html#EventPayloads-Updated
type IssueUpdatedEvent struct {
	IssueEvent
	Comment Comment `json:"comment"`
	Changes struct {
		Status struct {
			Old string `json:"old"`
			New string `json:"new"`
		} `json:"status"`
	} `json:"changes"`
}

// IssueCommentCreatedEvent https://confluence.atlassian.com/bitbucket/event-payloads-740262817.html#EventPayloads-CommentCreated
type IssueCommentCreatedEvent struct {
	IssueEvent
	Comment Comment `json:"comment"`
}

// Issue https://confluence.atlassian.com/bitbucket/event-payloads-740262817.html#EventPayloads-entity_issue
type Issue struct {
	ID        int    `json:"id"`
	Component string `json:"component"`
	Title     string `json:"title"`
	Content   struct {
		Raw    string `json:"raw"`
		HTML   string `json:"html"`
		Markup string `json:"markup"`
	} `json:"content"`
	Priority  string `json:"priority"`
	State     string `json:"state"`
	Type      string `json:"type"`
	Milestone struct {
		Name string `json:"name"`
	} `json:"milestone"`
	Version struct {
		Name string `json:"name"`
	} `json:"version"`
	CreatedOn *time.Time `json:"created_on"`
	UpdatedOn *time.Time `json:"updated_on"`
	Links     Links      `json:"links"`
}

// A PullRequestEvent is not a BB event. This is the base for several PullRequest* events.
type PullRequestEvent struct {
	Actor       Actor       `json:"actor"`
	PullRequest PullRequest `json:"pullrequest"`
	Repository  Repository  `json:"repository"`
}

// PullRequestCreatedEvent https://confluence.atlassian.com/bitbucket/event-payloads-740262817.html#EventPayloads-Created
type PullRequestCreatedEvent struct {
	PullRequestEvent
}

// PullRequestUpdatedEvent https://confluence.atlassian.com/bitbucket/event-payloads-740262817.html#EventPayloads-Updated.1
type PullRequestUpdatedEvent struct {
	PullRequestEvent
}

// PullRequestApprovedEvent https://confluence.atlassian.com/bitbucket/event-payloads-740262817.html#EventPayloads-Approved
type PullRequestApprovedEvent struct {
	PullRequestEvent
	Approval Approval `json:"approval"`
}

// PullRequestApprovalRemovedEvent https://confluence.atlassian.com/bitbucket/event-payloads-740262817.html#EventPayloads-ApprovalRemoved
type PullRequestApprovalRemovedEvent struct {
	PullRequestEvent
	Approval Approval `json:"approval"`
}

// PullRequestMergedEvent https://confluence.atlassian.com/bitbucket/event-payloads-740262817.html#EventPayloads-Merged
type PullRequestMergedEvent struct {
	PullRequestEvent
}

// PullRequestDeclinedEvent https://confluence.atlassian.com/bitbucket/event-payloads-740262817.html#EventPayloads-Declined
type PullRequestDeclinedEvent struct {
	PullRequestEvent
}

// A PullRequestCommentEvent doesn't exist. It is used as the base for several real events.
type PullRequestCommentEvent struct {
	PullRequestEvent
	Comment Comment `json:"comment"`
}

// PullRequestCommentCreatedEvent https://confluence.atlassian.com/bitbucket/event-payloads-740262817.html#EventPayloads-CommentCreated.1
type PullRequestCommentCreatedEvent struct {
	PullRequestCommentEvent
}

// PullRequestCommentUpdatedEvent https://confluence.atlassian.com/bitbucket/event-payloads-740262817.html#EventPayloads-CommentUpdated
type PullRequestCommentUpdatedEvent struct {
	PullRequestCommentEvent
}

// PullRequestCommentDeletedEvent https://confluence.atlassian.com/bitbucket/event-payloads-740262817.html#EventPayloads-CommentDeleted
type PullRequestCommentDeletedEvent struct {
	PullRequestCommentEvent
}

// An Approval is used in pull requests
type Approval struct {
	Date *time.Time `json:"date"`
	User User       `json:"user"`
}

// PullRequest https://confluence.atlassian.com/bitbucket/event-payloads-740262817.html#EventPayloads-entity_pullrequest
type PullRequest struct {
	ID          int    `json:"id"`
	Title       string `json:"title"`
	Description string `json:"description"`
	State       string `json:"state"`
	Author      User   `json:"author"`
	Source      struct {
		Branch struct {
			Name string `json:"name"`
		} `json:"branch"`
		Commit struct {
			Hash string `json:"hash"`
		} `json:"commit"`
		Repository Repository `json:"repository"`
	} `json:"source"`
	Destination struct {
		Branch struct {
			Name string `json:"name"`
		} `json:"branch"`
		Commit struct {
			Hash string `json:"hash"`
		} `json:"commit"`
		Repository Repository `json:"repository"`
	} `json:"destination"`
	MergeCommit struct {
		Hash string `json:"hash"`
	} `json:"merge_commit"`
	Participants      []Participant `json:"participants"`
	Reviewers         []User        `json:"reviewers"`
	CloseSourceBranch bool          `json:"close_source_branch"`
	ClosedBy          User          `json:"closed_by"`
	Reason            string        `json:"reason"`
	CreatedOn         *time.Time    `json:"created_on"`
	UpdatedOn         *time.Time    `json:"updated_on"`
	Links             Links         `json:"links"`
}

// Participant is the actual structure returned in PullRequest events
// Note: this doesn't match the docs!?
type Participant struct {
	Role     string `json:"role"`
	Type     string `json:"type"`
	Approved bool   `json:"approved"`
	User     User   `json:"user"`
}

// User https://confluence.atlassian.com/bitbucket/event-payloads-740262817.html#EventPayloads-entity_userUser
type User struct {
	Type        string `json:"type"`
	Username    string `json:"username"`
	DisplayName string `json:"display_name"`
	UUID        string `json:"uuid"`
	Links       Links  `json:"links"`
}
