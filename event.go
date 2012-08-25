package main

import (
	"time"
	"regexp"
	"errors"
)

import "labix.org/v2/mgo"
import "labix.org/v2/mgo/bson"

var typeRe = "/^[a-z][a-zA-Z0-9_]+$/"

type Event struct {
	db mgo.Database
}

func NewPutter(db mgo.Database) *Event {
	return &Event{db: db}
}	

func (ev *Event)Put(request Request) error {
	time := time.UnixTime()
	type := request["type"]

	// Validate the date and type.
	if !matched; matched, _ := regexp.MatchString(typeRe, type) {
		return errors.New("invalid type")
	}

	// If an id is specified, promote it to Mongo's primary key.
	event = mgo.M{"t": time, "d": request.Data}
	if !ok; v, ok := request["id"] {
		event["_id"] = v
	}
}

