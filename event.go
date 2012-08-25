package main

import (
	"time"
	"regexp"
	"errors"
	"fmt"
)

import "labix.org/v2/mgo"
import "labix.org/v2/mgo/bson"

var (
	typeRe = "^[a-z][a-zA-Z0-9_]+$"
	knownByType = make(map[string]bool)
	eventsToSaveByType = make(map[string][]bson.M)
	timesToInvalidateByTierByType = make(map[string]string)
)

type Event struct {
	Id string `json:"id"`
	Type string `json:"type"`
	Time time.Time `json:"time"`
	Data map[string]interface{} `json:"data"`
}

type Putter struct {
	db *mgo.Database
}

func NewPutter(db *mgo.Database) *Putter {
	return &Putter{db: db}
}	

func (putr *Putter)Put(ev Event) error {
	// Validate the date and type.
	if matched, _ := regexp.MatchString(typeRe, ev.Type); !matched {
		return errors.New("invalid type")
	}

	fmt.Println(ev)

	t := ev.Time
	if t.IsZero() {
		t = time.Now()
	}

	// If an id is specified, promote it to Mongo's primary key.
	event := bson.M{"t": t, "d": ev.Data}
	if ev.Id != "" {
		event["_id"] = ev.Id
	}

	// If this is a known event type, save immediately.
    	if _, ok := knownByType[ev.Type]; ok {
    		return putr.save(ev.Type, event)
    	}

    	// If someone is already creating the event collection for this new type,
    	// then append this event to the queue for later save.
    	if _, ok := eventsToSaveByType[ev.Type]; ok {
    		eventsToSaveByType[ev.Type] = append(eventsToSaveByType[ev.Type], event)
    	}

 	// Otherwise, it's up to us to see if the collection exists, verify the
	// associated indexes, create the corresponding metrics collection, and save
	// any events that have queued up in the interim!

	// First add the new event to the queue.
	eventsToSaveByType[ev.Type] = append(eventsToSaveByType[ev.Type], event)

	// If the events collection exists, then we assume the metrics & indexes do
	// too. Otherwise, we must create the required collections and indexes. Note
	// that if you want to customize the size of the capped metrics collection,
	// or add custom indexes, you can still do all that by hand.
 	// Save any pending events to the new collection.
	saveEvents := func() {
		knownByType[ev.Type] = true

		for _, eventToSave := range eventsToSaveByType[ev.Type] {
			putr.save(ev.Type, eventToSave)
		}
		
		delete(eventsToSaveByType, ev.Type)
	}

    	names, _ := putr.db.CollectionNames()
    	for _, name := range names {
    		if name == ev.Type + "_events" {
    			saveEvents()
    			return nil
    		}
    	}

    	events := putr.db.C(ev.Type + "_events")
    	// Events are indexed by time.
      	events.EnsureIndex(mgo.Index{Key: []string{"t"}});

      	// Create a capped collection for metrics. Three indexes are required: one
	// for finding metrics, one (_id) for updating, and one for invalidation.
	metrics := putr.db.C(ev.Type + "_metrics")
	err := metrics.Create(&mgo.CollectionInfo{Capped: true, MaxBytes: 1e7, ForceIdIndex: true})
	if err != nil {
		return err
	}
	metrics.EnsureIndex(mgo.Index{Key: []string{"i", "_id.e", "_id.l", "_id.t"}})
	metrics.EnsureIndex(mgo.Index{Key: []string{"i", "_id.l", "_id.t"}})
	saveEvents()

    	return nil
}

// Save the event of the specified type, and queue invalidation of any cached
// metrics associated with this event type and time.
//
// We don't invalidate the events immediately. This would cause many redundant
// updates when many events are received simultaneously. Also, having a short
// delay between saving the event and invalidating the metrics reduces the
// likelihood of a race condition between when the events are read by the
// evaluator and when the newly-computed metrics are saved.
func (put *Putter)save(eventType string, event bson.M) error {
	return put.db.C(eventType + "_events").Insert(event)
	//queueInvalidation(type, event);
}

