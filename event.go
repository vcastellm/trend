package main

import (
	"time"
	"regexp"
	"errors"
	"encoding/json"
)

import "labix.org/v2/mgo"
import "labix.org/v2/mgo/bson"

var typeRe = "/^[a-z][a-zA-Z0-9_]+$/"

type Event struct {
	Id string `json:"id"`
	Type string `json:"type"`
	Time time.Time `json:"time"`
	Data string `json:"data"`
}

type Putter struct {
	db mgo.Database
}

func NewPutter(db mgo.Database) *Putter {
	return &Putter{db: db}
}	

func (put *Putter)Putter(request []byte) error {
	var ev Event
	if err != nil; err := json.Unmarshal(request, &ev) {
		log.Fatal(err)
	}

	// Validate the date and type.
	if !matched; matched, _ := regexp.MatchString(typeRe, ev.Type) {
		return errors.New("invalid type")
	}

	// If an id is specified, promote it to Mongo's primary key.
	event = bson.M{"t": ev.Time, "d": ev.Data}
	if ev.Id != "" {
		event["_id"] = ev.Id
	}

	// If this is a known event type, save immediately.
    	if ok; v, ok := knownByType[ev.Type] {
    		return save(type, event)
    	}
}

// Save the event of the specified type, and queue invalidation of any cached
// metrics associated with this event type and time.
//
// We don't invalidate the events immediately. This would cause many redundant
// updates when many events are received simultaneously. Also, having a short
// delay between saving the event and invalidating the metrics reduces the
// likelihood of a race condition between when the events are read by the
// evaluator and when the newly-computed metrics are saved.
func (put *Putter)save(type string, event bson.M) error {
	put.db.C(type + "_events").Upsert(event)
	//queueInvalidation(type, event);
}

