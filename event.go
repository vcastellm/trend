package main

import (
	"errors"
	"fmt"
	"regexp"
	"time"
)

import "labix.org/v2/mgo"
import "labix.org/v2/mgo/bson"

var (
	typeRe                        = "^[a-z][a-zA-Z0-9_]+$"
	knownByType                   = make(map[string]bool)
	eventsToSaveByType            = make(map[string][]Event)
	timesToInvalidateByTierByType = make(map[string]map[time.Duration][]int64)
)

type Event struct {
	Id   string                 `bson:"_id,omitempty" json:"id"`
	Type string                 `bson:"" json:"type"`
	Time time.Time              `bson:"t" json:"time"`
	Data map[string]interface{} `bson:"d" json:"data"`
}

type Putter struct {
	db *mgo.Database
}

func NewPutter(db *mgo.Database) *Putter {
	return &Putter{db: db}
}

func (putr *Putter) Put(event Event) error {
	eventType := event.Type
	event.Type = ""

	// Validate the date and type.
	if matched, _ := regexp.MatchString(typeRe, eventType); !matched {
		return errors.New("invalid type")
	}

	fmt.Println(event)

	if event.Time.IsZero() {
		event.Time = bson.Now()
	}

	// If this is a known event type, save immediately.
	if _, ok := knownByType[eventType]; ok {
		return putr.save(eventType, event)
	}

	// If someone is already creating the event collection for this new type,
	// then append this event to the queue for later save.
	if _, ok := eventsToSaveByType[eventType]; ok {
		eventsToSaveByType[eventType] = append(eventsToSaveByType[eventType], event)
	}

	// Otherwise, it's up to us to see if the collection exists, verify the
	// associated indexes, create the corresponding metrics collection, and save
	// any events that have queued up in the interim!

	// First add the new event to the queue.
	eventsToSaveByType[eventType] = append(eventsToSaveByType[eventType], event)

	// If the events collection exists, then we assume the metrics & indexes do
	// too. Otherwise, we must create the required collections and indexes. Note
	// that if you want to customize the size of the capped metrics collection,
	// or add custom indexes, you can still do all that by hand.
	// Save any pending events to the new collection.
	saveEvents := func() {
		knownByType[eventType] = true

		for _, eventToSave := range eventsToSaveByType[eventType] {
			putr.save(eventType, eventToSave)
		}

		delete(eventsToSaveByType, event.Type)
	}

	names, _ := putr.db.CollectionNames()
	for _, name := range names {
		if name == eventType+"_events" {
			saveEvents()
			return nil
		}
	}

	events := putr.db.C(eventType + "_events")
	// Events are indexed by time.
	events.EnsureIndex(mgo.Index{Key: []string{"t"}})

	// Create a capped collection for metrics. Three indexes are required: one
	// for finding metrics, one (_id) for updating, and one for invalidation.
	metrics := putr.db.C(eventType + "_metrics")
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
func (put *Putter) save(eventType string, event Event) error {
	return put.db.C(eventType + "_events").Insert(event)
	queueInvalidation(eventType, event)

	return nil
}

func queueInvalidation(eventType string, event Event) error {
	eventTime := event.Time

	if timesToInvalidateByTier, ok := timesToInvalidateByTierByType[eventType]; ok {
		for tier := range Tiers {
			var tierTimes []int64 = timesToInvalidateByTier[tier]
			tierTime := Tiers[tier].Floor(eventTime).Unix()
			i := bisect(tierTimes, tierTime)

			if i >= len(tierTimes) {
				tierTimes = append(tierTimes, tierTime)
			} else if tierTimes[i] > tierTime {
				fmt.Println("**********")
				fmt.Println(i)
			}
		}
	} else {
		timesToInvalidateByTierByType[eventType] = make(map[time.Duration][]int64)

		for tier := range Tiers {
			timesToInvalidateByTierByType[eventType][tier] = []int64{Tiers[tier].Floor(eventTime).Unix()}
		}
	}

	return nil
}
