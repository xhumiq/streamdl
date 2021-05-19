// +build windows

package main

func init(){
	println("Windows Setup")
}

const(
	DEF_FFMPEG_BIN = "file://c:/ffmpeg/bin"
	DEF_REC_BIN = "file://c:/youtube-dl/bin"
)
