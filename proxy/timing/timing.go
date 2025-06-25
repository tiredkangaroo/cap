package timing

import (
	"time"
)

type Order uint8

type Time string
type Subtime string

const (
	TimeNone        Time = ""
	TimeRequestInit Time = "Request Init"

	TimePrepRequest   = "Prep Request"
	TimeWaitApproval  = "Wait Approval"
	TimeDelayPeform   = "Perform Delay"
	TimeWriteRequest  = "Write Request"
	TimeReadRequest   = "Read Response"
	TimeWriteResponse = "Write Response"

	TimeProxyResponse   = "Proxy Response"
	TimeDialHost        = "Dial Host"
	TimeReadWriteTunnel = "Read/Write Tunnel"

	TimeCertGenTLSHandshake = "Cert Gen + TLS Handshake"
	TimeReadParseRequest    = "Read/Parse Request"

	TimeTotal = "Total"
)

const (
	SubtimeNone                 Subtime = ""
	SubtimeUUID                 Subtime = "UUID Generation"
	SubtimeGetClientProcessInfo Subtime = "Client Process Info"
)

type MinorTime struct {
	start    time.Time
	Duration time.Duration `json:"duration"`
}

type MajorTime struct {
	Start    time.Time     `json:"start"`
	Duration time.Duration `json:"duration"`

	MinorTimeKeys   []Subtime    `json:"minorTimeKeys"`
	MinorTimeValues []*MinorTime `json:"minorTimeValues"`
}

type Timing struct {
	MajorTimeKeys   []Time       `json:"majorTimeKeys"`
	MajorTimeValues []*MajorTime `json:"majorTimeValues"`
}

func (t *Timing) Start(key Time) {
	t.MajorTimeKeys = append(t.MajorTimeKeys, key)
	t.MajorTimeValues = append(t.MajorTimeValues, &MajorTime{
		Start: time.Now(),
	})
}

func (t *Timing) Stop() {
	lastidx := len(t.MajorTimeValues) - 1
	val := t.MajorTimeValues[lastidx]
	val.Duration = time.Since(val.Start)
}

func (t *Timing) Substart(sub Subtime) {
	lastMajorIdx := len(t.MajorTimeValues) - 1
	t.MajorTimeValues[lastMajorIdx].MinorTimeKeys = append(t.MajorTimeValues[lastMajorIdx].MinorTimeKeys, sub)
	t.MajorTimeValues[lastMajorIdx].MinorTimeValues = append(t.MajorTimeValues[lastMajorIdx].MinorTimeValues, &MinorTime{
		start: time.Now(),
	})
}

func (t *Timing) Substop() {
	lastMajorIdx := len(t.MajorTimeValues) - 1
	lastMinorIdx := len(t.MajorTimeValues[lastMajorIdx].MinorTimeValues) - 1
	minor := t.MajorTimeValues[lastMajorIdx].MinorTimeValues[lastMinorIdx]
	minor.Duration = time.Since(minor.start)
}

func (t *Timing) Export() map[string]any {
	return map[string]any{
		"majorTimeKeys":   t.MajorTimeKeys,
		"majorTimeValues": t.MajorTimeValues,
	}
}

func (t *Timing) Total() time.Duration {
	if len(t.MajorTimeValues) == 0 {
		return 0
	}
	first := t.MajorTimeValues[0]
	last := t.MajorTimeValues[len(t.MajorTimeValues)-1]
	return last.Start.Add(last.Duration).Sub(first.Start)
}

// // NOTE: fix ts (use more type-safe stuff)
// func (t *Timing) MarshalJSON() ([]byte, error) {
// 	d := []any{}
// 	for i, key := range t.MajorTimeKeys {
// 		val := t.MajorTimeValues[i]
// 		d = append(d, map[string]any{
// 			"name":     string(key),
// 			"duration": val.Duration,
// 			"subevents": func() []any {
// 				subevents := []any{}
// 				for j, subkey := range val.MinorTimeKeys {
// 					subval := val.MinorTimeValues[j]
// 					subevents = append(subevents, map[string]any{
// 						"name":     string(subkey),
// 						"duration": subval.Duration,
// 					})
// 				}
// 				return subevents
// 			}(),
// 		})
// 	}
// 	// result:
// 	// [
// 	// 	{
// 	// 	"name": "Request Init",
// 	// 	"duration": 124.56, (ns)
// 	// 	"subevents": [
// 	// 		{
// 	// 			"name": "UUID Generation",
// 	// 			"duration": 12.34 (ns)
// 	// 		},
// 	// 	}
// 	// ]

// 	return json.Marshal(d)
// }

// func (t *Timing) UnmarshalJSON(data []byte) error {
// 	type marshaledTime []struct {
// 		Name      string        `json:"name"`
// 		Duration  time.Duration `json:"duration"`
// 		Subevents []struct {
// 			Name     string        `json:"name"`
// 			Duration time.Duration `json:"duration"`
// 		} `json:"subevents"`
// 	}
// 	var mt marshaledTime
// 	if err := json.Unmarshal(data, &mt); err != nil {
// 		return err
// 	}
// 	t.MajorTimeKeys = make([]Time, len(mt))
// 	t.MajorTimeValues = make([]*MajorTime, len(mt))
// 	for i, m := range mt {
// 		t.MajorTimeKeys[i] = Time(m.Name)
// 		t.MajorTimeValues[i] = &MajorTime{
// 			Duration: m.Duration,
// 			MinorTimeKeys: func() []Subtime {
// 				keys := make([]Subtime, len(m.Subevents))
// 				for j, sub := range m.Subevents {
// 					keys[j] = Subtime(sub.Name)
// 				}
// 				return keys
// 			}(),
// 			MinorTimeValues: func() []*MinorTime {
// 				values := make([]*MinorTime, len(m.Subevents))
// 				for j, sub := range m.Subevents {
// 					values[j] = &MinorTime{Duration: sub.Duration}
// 				}
// 				return values
// 			}(),
// 		}
// 	}
// 	return nil
// }

func New() *Timing {
	t := &Timing{
		MajorTimeKeys:   make([]Time, 0, 4),
		MajorTimeValues: make([]*MajorTime, 0, 4),
	}
	return t
}
