package main

import (
	"fmt"
	"io"
	"log"
	"os"

	"github.com/hovertank3d/toloka"
)

var (
	login = os.Getenv("TOLOKA_USERNAME")
	pass  = os.Getenv("TOLOKA_PASSWORD")
)

func usage() {
	fmt.Printf("usage: %s [-o <out .torrent>] <search query> [-]\n", os.Args[0])
	os.Exit(0)
}

func main() {
	if len(os.Args) == 1 {
		usage()
	}

	var (
		searchQeury string
		outFile     io.Writer
		outName     string
	)

	outFile = os.Stdout

	skip := false
	for i, arg := range os.Args[1:] {
		if skip {
			skip = false
			continue
		}

		if arg == "-o" {
			if len(os.Args) < i+1 {
				usage()
			}
			outName = os.Args[i+2]
			skip = true
			continue
		}

		searchQeury += arg + " "
	}

	login := toloka.LoginData{
		Username: login,
		Password: pass,
	}

	tk := toloka.New()
	if err := tk.Login(login); err != nil {
		log.Fatal(err)
	}

	links, err := tk.Search(searchQeury)
	if err != nil {
		log.Fatal(err)
	}
	if len(links) == 0 {
		fmt.Println("no results")
		return
	}

	torrent, err := tk.Parse(links[0])
	if err != nil {
		log.Fatal(err)
	}

	reader, err := tk.TorrentReader(torrent)
	if err != nil {
		log.Fatal(err)
	}

	data, err := io.ReadAll(reader)
	if err != nil {
		log.Fatal(err)
	}

	if outName != "" {
		f, err := os.OpenFile(outName, os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			log.Fatal(err)
		}

		outFile = f
	}

	if _, err := outFile.Write(data); err != nil {
		log.Fatal(err)
	}
}
