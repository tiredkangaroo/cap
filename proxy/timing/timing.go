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

	TimePrepRequest    = "Prep Request"
	TimeWaitApproval   = "Wait Approval"
	TimeDelayPeform    = "Perform Delay"
	TimeRequestPerform = "Perform Request"
	TimeDumpResponse   = "Dump Response"
	TimeWriteResponse  = "Write Response"

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
	start    time.Time
	Duration time.Duration `json:"duration"`

	MinorTimeKeys   []Subtime    `json:"minorTimeKeys"`
	MinorTimeValues []*MinorTime `json:"minorTimeValues"`
}

type Timing struct {
	majorTimeKeys   []Time
	majorTimeValues []*MajorTime
}

func (t *Timing) Start(key Time) {
	t.majorTimeKeys = append(t.majorTimeKeys, key)
	t.majorTimeValues = append(t.majorTimeValues, &MajorTime{
		start: time.Now(),
	})
}

func (t *Timing) Stop() {
	lastidx := len(t.majorTimeValues) - 1
	val := t.majorTimeValues[lastidx]
	val.Duration = time.Since(val.start)
}

func (t *Timing) Substart(sub Subtime) {
	lastMajorIdx := len(t.majorTimeValues) - 1
	t.majorTimeValues[lastMajorIdx].MinorTimeKeys = append(t.majorTimeValues[lastMajorIdx].MinorTimeKeys, sub)
	t.majorTimeValues[lastMajorIdx].MinorTimeValues = append(t.majorTimeValues[lastMajorIdx].MinorTimeValues, &MinorTime{
		start: time.Now(),
	})
}

func (t *Timing) Substop() {
	lastMajorIdx := len(t.majorTimeValues) - 1
	lastMinorIdx := len(t.majorTimeValues[lastMajorIdx].MinorTimeValues) - 1
	minor := t.majorTimeValues[lastMajorIdx].MinorTimeValues[lastMinorIdx]
	minor.Duration = time.Since(minor.start)
}

func (t *Timing) Export() map[string]any {
	return map[string]any{
		"majorTimeKeys":   t.majorTimeKeys,
		"majorTimeValues": t.majorTimeValues,
	}
}

func New(secure, mitm bool) *Timing {
	t := &Timing{
		majorTimeKeys:   make([]Time, 0, 4),
		majorTimeValues: make([]*MajorTime, 0, 4),
	}
	return t
}
