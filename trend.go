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
	"encoding/json"
)

//import "github.com/kylelemons/go-gypsy/yaml"
import "code.google.com/p/go.net/websocket"
import "labix.org/v2/mgo"

const basePkg = "github.com/seasonlabs/trend/"

var port *int = flag.Int("p", 8080, "Port to listen.")
var putter *Putter

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

	http.Handle("/assets/", http.StripPrefix("/assets/", http.FileServer(http.Dir(assetsDir()))))
	http.HandleFunc("/", root)
	http.HandleFunc("/1.0/event/put", putHttp)
	http.Handle("/1.0/event/ws/put", websocket.Handler(putWs))
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", *port), nil))
}

func root(w http.ResponseWriter, req *http.Request) {
	t := template.Must(template.New("foo").ParseGlob(rootDir() + "/index.html"))
	if err := t.ExecuteTemplate(w, "index", req.Host+":"+req.URL.Scheme); err != nil {
		log.Fatal(err)
	}
}

func putHttp(w http.ResponseWriter, req *http.Request) {
	session, err := mgo.Dial("127.0.0.1")
        if err != nil {
                panic(err)
        }
        defer session.Close()
        db := session.DB("trend")

        putter = NewPutter(db)
	
	b, err := ioutil.ReadAll(req.Body)
	if err != nil {
		panic(err)
	}

	fmt.Println(string(b))

	var events []Event
	if err := json.Unmarshal(b, &events); err != nil {
		fmt.Fprintf(w, "{\"error\": \"%s\"}", err)
	}

	for _, event := range events {
		if err := putter.Put(event); err != nil {
			fmt.Fprintf(w, "{\"error\": \"%s\"}", err)
		}
	}
}

func putWs(ws *websocket.Conn) {
	session, err := mgo.Dial("127.0.0.1")
        if err != nil {
                panic(err)
        }
        defer session.Close()
        db := session.DB("trend")

        putter = NewPutter(db)

	var event Event
	for {
		if err := websocket.JSON.Receive(ws, &event); err != nil {
			websocket.Message.Send(ws, fmt.Sprintf("{\"error\": \"%s\"}", err))
			break
		}
		fmt.Printf("received: %#v\n", event)
		
		if err := putter.Put(event); err != nil {
			websocket.Message.Send(ws, fmt.Sprintf("{\"error\": \"%s\"}", err))
			break
		}
	}
}