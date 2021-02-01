package main

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/rs/zerolog/log"
	"ntc.org/mclib/common"
)

type validationContext struct {
	BaseDir    string
	BaseFile   string
	FileName   string
	RelPath    string
	LangId     string
	Content    []byte
	Corrected  []byte
	MatchIndex [][]int
	Level      int
	Links      []*linkContext
}

func (ctx *validationContext) Skip(link *linkContext, msg string, args ...interface{}) {
	ctx.Links = append(ctx.Links, link)
	link.Result = "skipped"
	link.Message = msg
	if len(args) > 0 {
		link.Message = fmt.Sprintf(msg, args)
	}
	log.Trace().Msg("Skipped: " + link.Message)
}

func (ctx *validationContext) SkippedCount() int {
	cnt := 0
	for _, l := range ctx.Links {
		if l.Result == "skipped" {
			cnt++
		}
	}
	return cnt
}

func (ctx *validationContext) CorrectedCount() int {
	cnt := 0
	for _, l := range ctx.Links {
		if l.Result == "corrected" {
			cnt++
		}
	}
	return cnt
}

func (ctx *validationContext) GoodCount() int {
	cnt := 0
	for _, l := range ctx.Links {
		if l.Result == "good" {
			cnt++
		}
	}
	return cnt
}

func (ctx *validationContext) NotFoundCount() int {
	cnt := 0
	for _, l := range ctx.Links {
		if l.Result == "not found" {
			cnt++
		}
	}
	return cnt
}

type linkContext struct {
	Parent         *validationContext
	Range          []int
	FullLink       string
	Link           string
	LinkType       string
	NormalizedPath string
	BaseFile       string
	FullPath       string
	Ext            string
	MatchLink      []string
	Corrected      string
	CorrectedPath  string
	AttemptedPaths []string
	Result         string
	Message        string
}

func (link *linkContext) Check(path, nlink string, msg string, args ...interface{}) bool {
	link.Message = msg
	if len(args) > 0 {
		link.Message = fmt.Sprintf(msg, args...)
	}
	found := false
	if strings.HasPrefix(path, "/ntc/web/zion-web/images") ||
		strings.HasPrefix(path, "/ntc/web/zion-web/en/images") ||
		strings.HasPrefix(path, "/ntc/web/zion-web/en/pdf") ||
		strings.HasPrefix(path, "/ntc/web/zion-web/en/video") ||
		strings.HasPrefix(path, "/ntc/web/zion-web/zh/images") ||
		strings.HasPrefix(path, "/ntc/web/zion-web/zh/pdf") ||
		strings.HasPrefix(path, "/ntc/web/zion-web/zh/video"){
		path = strings.Replace(path, "/zion-web/", "/zion-media/zion/", -1)
		nlink = "https://media.ziongjcc.org" + path[len("/ntc/web/zion-media"):]
		found = common.FileExists(path)
	}else{
		found = common.FileExists(path)
	}
	if !found && strings.ToLower(nlink) != nlink {
		lpath := filepath.Join(strings.ToLower(filepath.Dir(path)),filepath.Base(path))
		if common.FileExists(lpath){
			nlink = filepath.Join(strings.ToLower(filepath.Dir(nlink)),filepath.Base(nlink))
			path = lpath
			found = true
		}
	}
	if !found {
		link.Result = "not found"
		link.AttemptedPaths = append(link.AttemptedPaths, path)
		return false
	}
	if link.Link != nlink {
		link.Correct(path, nlink, msg, args...)
	} else {
		link.Parent.Links = append(link.Parent.Links, link)
		log.Trace().Str("FilePath", path).Msg("Good: " + link.Message)
		link.Result = "good"
	}
	return true
}

func (link *linkContext) Correct(path, nlink string, msg string, args ...interface{}) {
	link.Message = msg
	if len(args) > 0 {
		link.Message = fmt.Sprintf(msg, args...)
	}
	link.Parent.Links = append(link.Parent.Links, link)
	log.Trace().Str("Corrected Path", path).
		Str("Corrected Link", nlink).
		Str("Page", link.Parent.FileName).
		Msg("Corrected: " + link.Message)
	link.Result = "corrected"
	link.Corrected = nlink
	link.CorrectedPath = path
	m2m := REImgHref.FindStringSubmatchIndex(link.FullLink)
	link.Range[0] += m2m[4]
	link.Range[1] = link.Range[0] + m2m[5] - m2m[4]
}
