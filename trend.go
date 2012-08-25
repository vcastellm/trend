package main

import (
	//"encoding/json"
	"flag"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"go/build"
	"io/ioutil"
)

//import "github.com/kylelemons/go-gypsy/yaml"
//import "code.google.com/p/go.net/websocket"
import "labix.org/v2/mgo"
import "labix.org/v2/mgo/bson"

const basePkg = "github.com/seasonlabs/toukei/"

var port *int = flag.Int("p", 8080, "Port to listen.")
var putter Event

func rootDir() string {
	// find and serve static files
        p, err := build.Default.Import(basePkg, "", build.FindOnly)
        if err != nil {
                log.Fatalf("Couldn't find toukei files: %v", err)
        }
        root := p.Dir

        return root
}

func assetsDir() string {
	return rootDir() + "/assets"
}

func main() {
	flag.Parse()

	// config, err := yaml.ReadFile(rootDir() + "/config.yml")
	// if err != nil {
	// 	log.Fatalf("readfile(%q): %s", "config.yml", err)
	// }

	session, err := mgo.Dial("127.0.0.1")
        if err != nil {
                log.Fatal(err)
        }
        defer session.Close()
        db := session.DB("trend")
        putter = NewEvent(db)

	http.Handle("/assets/", http.StripPrefix("/assets/", http.FileServer(http.Dir(assetsDir()))))
	http.HandleFunc("/", root)
	http.HandleFunc("/1.0/event/put", putter)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", *port), nil))
}

func root(w http.ResponseWriter, req *http.Request) {
	t := template.Must(template.New("foo").ParseGlob(rootDir() + "/index.html"))
	if err := t.ExecuteTemplate(w, "index", req.Host+":"+req.URL.Scheme); err != nil {
		log.Fatal(err)
	}
}

func putter(w http.ResponseWriter, req *http.Request) {
	b, err := iouil.ReadAll(req.Body)

	fmt.Println(b)
	
	if err != nil {
		log.Fatal(err)
	}
	event.Putter(b)
}