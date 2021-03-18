package main

func GetStatus(mgr *JobManager) *SystemStatus{
  status := SystemStatus{
    Version:    version,
    GitHash:    gitHash,
    BuildStamp: buildStamp,
    Config:     *mgr.SvcConfig,
  }
  status.Statuses = make(map[string]*ProcessStatus)
  for _, repo := range mgr.Jobs {
    s := NewProcessStatusFromContext(repo)
    if s.CommandStatus != nil {
      s.CommandStatus.Output = nil
    }
    status.Statuses[repo.Config.RepoName+"/"+repo.Config.Branch] = s
  }
  for name, cfg := range mgr.Configs {
    println("name", name)
    _, ok := status.Statuses[cfg.RepoName+"/"+cfg.Branch]
    if !ok {
      s := NewProcessStatus(*cfg, nil)
      s.Status = "Listening"
      status.Statuses[cfg.RepoName+"/"+cfg.Branch] = s
    }
  }
  return &status
}
