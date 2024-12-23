package main

import (
	"fmt"
	"log"
	"os"

	"github.com/hovertank3d/toloka"
)

func main() {
	if len(os.Args) != 3 {
		fmt.Printf("usage: %s <username> <password>\n", os.Args[0])
		os.Exit(1)
	}

	login := toloka.LoginData{
		Username: os.Args[1],
		Password: os.Args[2],
	}

	tk := toloka.New()
	if err := tk.Login(login); err != nil {
		log.Fatal(err)
	}

	links, err := tk.Search("jojo's bizarre adventure")
	if err != nil {
		log.Fatal(err)
	}

	for _, link := range links {
		torrent, err := tk.Parse(link)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Println(torrent)
	}
}
