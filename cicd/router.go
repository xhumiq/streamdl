package main

import (
	"fmt"
	"github.com/go-chi/chi"
	"github.com/go-chi/render"
	"github.com/pkg/errors"
	"io/ioutil"
	"net/http"
	"strings"

	"bitbucket.org/xhumiq/go-mclib/nechi"
)

func NewApi(mgr *JobManager) *nechi.WebChi {
	app := nechi.NewWebApp(&mgr.AppStatus, &mgr.SvcConfig.Http, nil)
	app.ApiHealth("/healthcheck", HealthCheck)
	app.Get("/version", func(r *http.Request, w http.ResponseWriter) error {
		w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
		render.Status(r, 200)
		status := GetStatus(mgr)
		render.JSON(w, r, status)
		return nil
	})
	app.Get("/cert", func(r *http.Request, w http.ResponseWriter) error {
		w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
		render.Status(r, 200)
		b, err := ioutil.ReadFile("/root/.ssh/id_rsa.pub") // just pass the file name
		if err != nil {
			return err
		}
		render.PlainText(w, r, string(b))
		return nil
	})
	app.Get("/jobs/{repo}/{branch}", func(r *http.Request, w http.ResponseWriter) error {
		resp := &ProcessStatus{
			RepoName: strings.TrimSpace(chi.URLParam(r, "repo")),
			Branch:   strings.TrimSpace(chi.URLParam(r, "branch")),
		}
		repo, _ := mgr.GetStatus(resp.RepoName, resp.Branch)
		if repo == nil {
			name := strings.ToUpper(resp.RepoName + "/" + resp.Branch)
			println("config", name)
			cfg, _ := mgr.Configs[name]
			if cfg == nil {
				resp.SetError(fmt.Errorf("Repo, Branch not started or configured %s %s", resp.RepoName, resp.Branch))
				println("Repo and Branch not found Deocde Json")
			} else {
				resp = NewProcessStatus(*cfg, nil)
				resp.Status = "Listening"
			}
		} else {
			resp = NewProcessStatusFromContext(repo)
			resp.CommandStatus = nil
			for _, cmd := range repo.History {
				cd := NewCommandStatus(cmd)
				if cd == nil {
					continue
				}
				resp.Logs = append(resp.Logs, cd)
			}
		}
		render.JSON(w, r, resp)
		return nil
	})

	app.Post("/jobs/{repo}/{branch}", func(r *http.Request, w http.ResponseWriter) error {
		resp := &ProcessStatus{
			RepoName: strings.TrimSpace(chi.URLParam(r, "repo")),
			Branch:   strings.TrimSpace(chi.URLParam(r, "branch")),
		}
		ctx, err := mgr.QueueDeployment(resp.RepoName, resp.Branch)
		if ctx != nil {
			resp = NewProcessStatusFromContext(ctx)
			resp.Status = "Requested"
			resp.CommandStatus = nil
			resp.Logs = nil
		}
		if err != nil {
			resp.SetError(errors.Wrapf(err, "Unable to queue deployment %s %s", resp.RepoName, resp.Branch))
			fmt.Printf("Error Queue %+v\n", err)
		}
		render.JSON(w, r, resp)
		return nil
	})

	app.Post("/events/bb/{repo}", func(r *http.Request, w http.ResponseWriter) error {
		println("Post Bitbucket Repo")
		resp := &ProcessStatus{
			RepoName: strings.TrimSpace(chi.URLParam(r, "repo")),
		}
		event := &ComboEvent{}
		err := render.DecodeJSON(r.Body, event)
		if err != nil {
			resp.SetError(errors.Wrapf(err, "Unable to decode body"))
			render.JSON(w, r, resp)
			println("Deocde Json", err.Error())
			return nil
		}
		resp.RepoName, resp.Branch = event.GetRepoBranch()
		ctx, err := mgr.QueueDeployment(resp.RepoName, resp.Branch)
		if ctx != nil {
			resp = NewProcessStatusFromContext(ctx)
			resp.Status = "Requested"
			resp.CommandStatus = nil
			resp.Logs = nil
		}
		if err != nil {
			resp.SetError(errors.Wrapf(err, "Unable to queue deployment %s %s", resp.RepoName, resp.Branch))
			fmt.Printf("Error Queue %+v\n", err)
		}
		render.JSON(w, r, resp)
		return nil
	})
	return app
}

func HealthCheck() ([]string, []error) {
	logs := []string{}
	errs := []error{}
	return logs, errs
}
