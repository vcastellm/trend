package main

import (
	"time"
)

var (
	second10 = 10 * time.Second
	minute5 = 5 * time.Minute
	day = 24 * time.Hour
)

type Tier struct {
	Key time.Duration
	Floor func(d int64) time.Time
	Ceil func(t Tier, date int64) time.Time
	Step func(d int64) time.Time
	Next *Tier
	Size int
}

func (t Tier) tierCeil(date int64) time.Time {
  return t.Step(t.Floor(time.Unix(date - 1, 0).Unix()).Unix())
}

var Tiers = map[time.Duration]Tier{
	second10: Tier{ 
		Key: second10,
    		Floor: func(d int64) time.Time { return time.Unix(d / int64(second10) * int64(second10), 0) },
    		Ceil: Tier.tierCeil,
    		Step: func(d int64) time.Time { return time.Unix(+d + int64(second10), 0) },
	},
	time.Minute: Tier{
    		Key: time.Minute,
    		Floor: func(d int64) time.Time { return time.Unix(d / int64(time.Minute) * int64(time.Minute), 0) },
    		Ceil: Tier.tierCeil,
    		Step: func(d int64) time.Time { return time.Unix(+d + int64(time.Minute), 0) },
	},
	minute5: Tier{
    		Key: minute5,
    		Floor: func(d int64) time.Time { return time.Unix(d / int64(minute5) * int64(minute5), 0) },
    		Ceil: Tier.tierCeil,
    		Step: func(d int64) time.Time { return time.Unix(+d + int64(minute5), 0) },
	},
	time.Hour: Tier{
		Key: time.Hour,
    		Floor: func(d int64) time.Time { return time.Unix(d / int64(time.Hour) * int64(time.Hour), 0) },
    		Ceil: Tier.tierCeil,
    		Step: func(d int64) time.Time { return time.Unix(+d + int64(time.Hour), 0) },
    		//Next: &Tiers[minute5],
    		Size: 12,
	},
	day: Tier{
    		Key: day,
    		Floor: func(d int64) time.Time { return time.Unix(d / int64(day) * int64(day), 0) },
    		Ceil: Tier.tierCeil,
    		Step: func(d int64) time.Time { return time.Unix(+d + int64(day), 0) },
    		//Next: &Tiers[time.Hour],
    		Size: 24,
	},
}
