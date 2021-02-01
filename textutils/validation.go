package main

import (
	"bytes"
	"github.com/rs/zerolog/log"
	"golang.org/x/text/encoding/traditionalchinese"
	"golang.org/x/text/transform"
	"io"
	"io/ioutil"
	"net/url"
	"ntc.org/mclib/common"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

func walkPath(base string, fix bool) error{
	nf, cf, tot := 0, 0, 0
	nbase := "/ntc/web/zion-web"
	filepath.Walk(base, func(path string, info os.FileInfo, err error) error {
		if info == nil{
			return nil
		}
		if info.IsDir() || strings.Contains(path, ".git"){
			return nil
		}
		fpath := path
		lpath := strings.ToLower(path)
		if !strings.HasSuffix(lpath, "htm") && !strings.HasSuffix(lpath, "html") && !strings.HasSuffix(lpath, "css"){
			return nil
		}
		rel, err := validateFile(base, fpath[len(base):])
		if rel == nil || err!=nil{
			return err
		}
		nfc := rel.NotFoundCount()
		cc := rel.CorrectedCount()
		for _, l := range rel.Links{
			if l.Corrected != ""{
				//println(l.Link, l.Range[0], l.Range[1])
			}
		}
		if cc > 0 && fix{
			bf := filepath.Join(nbase, rel.RelPath)
			nd := filepath.Dir(bf)
			if !common.DirExists(nd){
				if err = os.MkdirAll(nd, 0755); err!=nil{
					return err
				}
			}
			nc := fixPage(rel.Content, rel.Links)
			if err = ioutil.WriteFile(bf, nc, 0644); err!=nil{
				return err
			}
		}
		nf += nfc
		cf += cc
		tot += len(rel.Links)
		return nil
	})
	println("Walk %d / %d - %d", nf, tot, cf)
	return nil
}

func fixPage(content []byte, links []*linkContext) []byte{
	nc, pi := bytes.Buffer{}, 0
	for _, l := range links{
		if l.Result == "corrected" && l.Corrected != l.Link{
			nc.Write(content[pi:l.Range[0]])
			io.WriteString(&nc, strings.Replace(l.Corrected, " ", "%20", -1))
			pi = l.Range[1]
		}
	}
	if pi < len(content) - 1{
		nc.Write(content[pi:])
	}
	return nc.Bytes()
}

func validateFile(base, relPath string) (rel *validationContext, err error){
	ctx := validationContext{
		BaseDir:    base,
		FileName:   filepath.Join(base, relPath),
		RelPath:    relPath,
		LangId: relPath[:2],
	}
	ctx.BaseFile = filepath.Base(ctx.FileName)
	if ctx.LangId != "zh" && ctx.LangId != "en" {
		log.Warn().Msgf("Skipped: File not found in either en or zh root folders: %s", ctx.FileName)
		return
	}
	if strings.Contains(ctx.FileName, "/html player/") || strings.Contains(ctx.FileName, "/youtube/") || strings.Contains(ctx.FileName, "/en/media/"){
		//return
	}
	if ctx.FileName != "/ntc/web/zion-web/zh/musee/third/musee_third_tour.html"{
		//return nil, nil
	}
	path := relPath[3:]
	dir := filepath.Dir(path)
	ctx.Level = len(strings.Split(dir, "/"))
	if dir == "." {
		ctx.Level = 0
	}
	if ctx.Level < 1 {
		log.Warn().Msgf("Skipped: File found in root folder: %s", ctx.FileName)
		return
	}
	if ctx.Content, err = ioutil.ReadFile(ctx.FileName); err != nil {
		return
	}
	rel, err = validate(ctx)
	return
}

func validate(context validationContext) (*validationContext, error){
	context.MatchIndex = REImgHref.FindAllIndex(context.Content, -1)
	ma := context.MatchIndex
	for _, m := range ma {
		sm := string(context.Content[m[0]:m[1]])
		smm := REImgHref.FindStringSubmatch(sm)
		lm := linkContext{
			Parent: &context,
			FullLink: sm,
			LinkType: strings.ToLower(smm[1]),
			MatchLink: smm,
			Link: strings.TrimSpace(smm[2]),
			Range: m,
		}
		if (lm.LinkType == "title") || (lm.LinkType == "value") || (lm.LinkType == "alt") {
			context.Skip(&lm, "Link Type %s: %s", lm.LinkType, lm.FullLink)
			continue
		}
		if lm.LinkType == "" && strings.HasPrefix(sm, "url") {
			lm.LinkType = "url"
		}
		if lm.Link == "" {
			lm.Link = strings.TrimSpace(smm[3])
		}
		if lm.Link == "" || strings.Contains(lm.Link, "\n") || strings.HasPrefix(lm.Link, "http") || strings.HasPrefix(lm.Link, "mailto") {
			context.Skip(&lm, "External link found %s -- %s", lm.Link, lm.FullLink)
			continue
		}
		var err error
		if lm.Link, err = url.QueryUnescape(lm.Link); err != nil {
			return nil, err
		}
		if (lm.LinkType != "href") && (lm.LinkType != "src") && (lm.LinkType != "background") && (lm.LinkType != "url") && (lm.LinkType != "file") && (lm.LinkType != "image") {
			log.Warn().Str("Page", context.FileName).Msgf("Unexpected Link Type %s -- %s", lm.LinkType, lm.FullLink)
		}

		lm.BaseFile = filepath.Base(lm.Link)
		fe := strings.Split(lm.BaseFile, ".")
		lm.Ext = fe[len(fe)-1]
		dp := filepath.Dir(context.FileName)
		lm.FullPath = filepath.Join(dp, lm.Link)
		if strings.HasPrefix(lm.Link, "/") {
			lm.FullPath = filepath.Join(context.BaseDir, lm.Link)
		}
		if lm.Check(lm.FullPath, lm.Link, "Raw Link: %s", lm.FullLink) {
			continue
		}
		if (lm.LinkType == "image"){
			if strings.Contains(lm.FullLink, "/video/") && lm.Ext == "jpg"{
				continue
			}
		}
		link := lm.Link
		if strings.HasPrefix(link, "#") {
			link = strings.TrimLeft(link, "#")
			if link == context.BaseFile{
				link = "#"
				lm.Correct(link, link, "Pound sign on link: %s", link)
				continue
			}else{
				fp := filepath.Join(dp, link)
				if lm.Check(fp, link, "Pound sign on link: %s", link){
					continue
				}
			}
		}
		if strings.Contains(link, "\\"){
			link = strings.Replace(link, "\\", "/", -1)
			fp := filepath.Join(dp, link)
			if lm.Check(fp, link, "Pound sign on link: %s", link){
				continue
			}
		}
		if strings.ToLower(lm.Ext) != lm.Ext {
			lm.Ext = strings.ToLower(lm.Ext)
			link = link[:len(link)-len(lm.Ext)] + lm.Ext
			obf := lm.BaseFile
			lm.BaseFile = filepath.Base(link)
			fp := lm.FullPath[:len(lm.FullPath)-len(lm.Ext)] + lm.Ext
			if lm.Check(fp, link, "Lower case extention correction: %s ==> %s",  obf, lm.BaseFile) {
				continue
			}
		}
		if !strings.Contains(lm.Link, "/"){
			plink := link
			if (lm.Ext == "gif" || lm.Ext == "jpg" || lm.Ext == "png"){
				link = "../../images/" + link
			}
			if (lm.Ext == "mp4" || lm.Ext == "flv" || lm.Ext == "ogv"){
				link = "../../video/" + link
			}
			if (lm.Ext == "css"){
				link = "../../styles/" + link
			}
			if plink != link{
				fp := filepath.Join(dp, link)
				if lm.Check(fp, link, "Media leaf file put into media sub foler: %s", link) {
					continue
				}
			}
		}
		if lm.Ext == "mp4" && strings.Contains(lm.FullPath, "ogv"){
			lm.Ext = "ogv"
			link = link[:len(link)-len(lm.Ext)] + lm.Ext
			obf := lm.BaseFile
			lm.BaseFile = filepath.Base(link)
			fp := lm.FullPath[:len(lm.FullPath)-len(lm.Ext)] + lm.Ext
			if lm.Check(fp, link, "Mp4 extention correction: %s ==> %s",  obf, lm.BaseFile) {
				continue
			}
		}
		if lm.BaseFile == "logo.jpg"{
			lm.Ext = "gif"
			link = link[:len(link)-len(lm.Ext)] + lm.Ext
			obf := lm.BaseFile
			lm.BaseFile = filepath.Base(link)
			fp := lm.FullPath[:len(lm.FullPath)-len(lm.Ext)] + lm.Ext
			if lm.Check(fp, link, "logo.jpg extention correction: %s ==> %s",  obf, lm.BaseFile) {
				continue
			}
		}
		pma := REParentDir.FindAllString(link, -1)
		pl := len(pma)
		if pl > 0 || (lm.Ext!=".html" && lm.Ext!=".htm"){
			if context.Level != pl && len(link) > (3*pl)+1 {
				ssmi := NormalizeCase(link[(3 * pl):])
				for i := 0; i < context.Level; i++ {
					ssmi = "../" + ssmi
				}
				fp := filepath.Join(dp, ssmi)
				if lm.Check(fp, ssmi,"Normalize to language root. Link: %s ==> %s", link, ssmi) {
					continue
				}
			}
			if (context.Level + 1) != pl && len(link) > (3*pl)+1 {
				ssmi := NormalizeCase(link[(3 * pl):])
				for i := 0; i < (context.Level + 1); i++ {
					ssmi = "../" + ssmi
				}
				fp := filepath.Join(dp, ssmi)
				if lm.Check(fp, ssmi, "Normalize to site root. Link: %s ==> %s", link, ssmi) {
					continue
				}
			}
		}
		if (lm.Ext == "mp4" || lm.Ext == "flv" || lm.Ext == "ogv") {
			dp2 := filepath.Dir(lm.FullPath)
			if !strings.HasSuffix(dp2, "ogv") && !strings.HasSuffix(dp2, "mp4") && !strings.HasSuffix(dp2, "flv") {
				ldp := filepath.Dir(link)
				link = filepath.Join(ldp, lm.Ext, lm.BaseFile)
				fp := filepath.Join(dp2, lm.Ext, lm.BaseFile)
				if lm.Check(fp,link,"Link video files to extension subfolder. Link: %s => %s", lm.Link, link) {
					continue
				}
			}
		}
		context.Links = append(context.Links, &lm)
		log.Warn().Str("NotFound", lm.FullPath).Str("Page", context.FileName).Msgf("Not Found: %s", lm.FullLink)
	}
	return &context, nil
}

func NormalizeCase(url string)string{
	urls := strings.Split(url, "/")
	if len(urls) < 2{
		return url
	}
	lroot := strings.ToLower(urls[0])
	if lroot==urls[0]{
		return url
	}
	if lroot == "scripts"{
		urls[0] = "scripts"
		return strings.Join(urls, "/")
	}
	if lroot == "images"{
		urls[0] = "images"
		return strings.Join(urls, "/")
	}
	if lroot == "pdf"{
		urls[0] = "pdf"
		return strings.Join(urls, "/")
	}
	return url
}

var (
	REImgHref = regexp.MustCompile(`(?i)(?:(\w+)\s*=["(']?|url(?:=|["(']))(?:([^"&)']+\.\w{2,4})["&)'])`)
	REParentDir = regexp.MustCompile(`\.\./`)
)

func Decodebig5(s []byte) ([]byte, error) {
	I := bytes.NewReader(s)
	O := transform.NewReader(I, traditionalchinese.Big5.NewDecoder())
	d, e := ioutil.ReadAll(O)
	if e != nil {
		return nil, e
	}
	return d, nil
}
