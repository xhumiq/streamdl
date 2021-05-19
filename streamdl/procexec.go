package main

import (
	"bitbucket.org/xhumiq/go-mclib/common"
	"bufio"
	"fmt"
	"github.com/rs/zerolog/log"
	"io"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"strings"
	"time"
)

const(
	DEF_FFMPEG_OPTION = "-err_detect ignore_err -vcodec copy -acodec copy -movflags faststart"
	DEF_REC_OPTION = "-f 96 <url> -o <out>"
	DEF_TEMP_PATH = "/tmp/streams"
)

var (
	chInterrupt = make(chan os.Signal, 1)
)

func init(){
	signal.Notify(chInterrupt, os.Interrupt)
}

func execCommand(cmd *exec.Cmd, duration time.Duration) error{
	stdout, err := cmd.StdoutPipe()
	if err!=nil{
		return err
	}
	stderr, err := cmd.StderrPipe()
	if err!=nil{
		return err
	}
	startIOStreams(stdout, stderr)
	if err = cmd.Start(); err!=nil{
		return err
	}
	defer func(){
		cmd.Process.Kill()
		time.Sleep(2 * time.Second)
	}()
	done := make(chan error)
	go func() { done <- cmd.Wait() }()
	select {
	case err := <-done:
		if err != nil {
			return err
		}
		return nil
	case <-chInterrupt:
		cmd.Process.Signal(os.Interrupt)
		log.Info().Msgf("Signal Interrupt")
	case <-time.After(duration):
		cmd.Process.Signal(os.Interrupt)
	}
	select {
	case err := <-done:
		return err
	case <-chInterrupt:
		cmd.Process.Kill()
	case <-time.After(15 * time.Second):
		cmd.Process.Kill()
	}
	return nil
}

func startIOStreams(stdout io.ReadCloser, stderr io.ReadCloser) {
	go func() {
		scanner := bufio.NewScanner(stdout)
		for scanner.Scan() {
			b := scanner.Bytes()
			if len(b) < 1{
				continue
			}
			line := strings.TrimSpace(string(b))
			if len(line) < 1{
				continue
			}
			log.Debug().Msg(line)
		}
	}()

	go func() {
		scanner := bufio.NewScanner(stderr)
		for scanner.Scan() {
			b := scanner.Bytes()
			if len(b) < 1{
				continue
			}
			line := strings.TrimSpace(string(b))
			if len(line) < 1{
				continue
			}
			log.Debug().Msg(line)
		}
	}()
}

func createFileName(prefix, suffix string)string{
	if suffix != ""{
		suffix = "_" + suffix
	}
	if prefix!=""{
		prefix = prefix + "_"
	}
	return prefix + time.Now().Format("06-01-02.15.04") + suffix + ".mp4"
}

func createFFMpegCmd(ffmpeg, inpFile, outFile string, delIfExists bool) (*exec.Cmd, error){
	if !common.FileExists(inpFile){
		return nil, fmt.Errorf("Input file %s doesn't exist", inpFile)
	}
	if common.FileExists(outFile){
		if delIfExists{
			return nil, fmt.Errorf("Output file %s already exists", outFile)
		}
		os.Remove(outFile)
	}else{
		dirPath := filepath.Dir(outFile)
		if !common.DirExists(dirPath){
			err := os.MkdirAll(dirPath, 0755)
			if err!=nil{
				return nil, err
			}
		}
	}
	opts := append([]string{"-i", inpFile}, strings.Split(DEF_FFMPEG_OPTION, " ")...)
	opts = append(opts, outFile)
	log.Info().Str("FFMpeg: %s", ffmpeg).
		Str("Input: %s", inpFile).
		Str("Output: %s", outFile).
		Str("Cmd: %s", ffmpeg + " " + strings.Join(opts, " ")).
		Msg("Executing FFMpeg Fix Video")
	return exec.Command(ffmpeg, opts...), nil
}

func createRecCmd(rec, url, outFile string, delIfExists bool) (*exec.Cmd, error){
	if common.FileExists(outFile){
		if delIfExists{
			return nil, fmt.Errorf("Output file %s already exists", outFile)
		}
		os.Remove(outFile)
	}else{
		dirPath := filepath.Dir(outFile)
		if !common.DirExists(dirPath){
			err := os.MkdirAll(dirPath, 0755)
			if err!=nil{
				return nil, err
			}
		}
	}
	opts := strings.Split(DEF_REC_OPTION, " ")
	for i, _ := range opts{
		opts[i] = strings.Replace(opts[i], "<url>", url, 1)
		opts[i] = strings.Replace(opts[i], "<out>", outFile, 1)
	}
	log.Info().Str("Recorder", rec).
		Str("Url", url).
		Str("Output", outFile).
		Str("Cmd", rec + " " + strings.Join(opts, " ")).
		Msg("Creating Record Stream Command")
	return exec.Command(rec, opts...), nil
}

