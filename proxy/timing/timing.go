package timing

import (
	"time"
)

type Order uint8

type Time string
type Subtime string
type SetStateFuncType func(string)

const (
	// invalid
	TimeNone Time = ""

	// HTTP: read proxy request -> init -> perform request -> save request body -> save response body -> write response
	// HTTPS (no MITM): read proxy request -> init -> send 200 -> wait approval -> delay perform -> tunnel
	// HTTPS (MITM): read proxy request -> init -> send 200 -> cert gen + handshake -> read request -> perform request -> save request body -> save response body -> write response

	// TimeReadProxyRequest is the time taken to read the request to the proxy. For HTTP connections, this is the only
	// read request, since the request is sent in full to the proxy. For HTTPS connections, this is the time taken to
	// read the CONNECT request.
	TimeReadProxyRequest Time = "Read Proxy Request"
	// TimeRequestInit is the time taken to initialize the proxy request.
	TimeRequestInit Time = "Request Init"

	// TimeSaveRequestBody is the time taken to save the request body to the database.
	TimeSaveRequestBody Time = "Save Request Body"
	// TimePerformRequest is the time taken to perform the request.
	TimePerformRequest Time = "Perform Request"
	// TimeSaveResponseBody is the time taken to save the response body to the database.
	TimeSaveResponseBody Time = "Save Response Body"
	// TimeWriteResponse is the time taken to write the response to the connection.
	TimeWriteResponse Time = "Write Response"

	// TimeSendProxyResponse is the time taken to send a response to the client. This is used for HTTPS connections
	// where the proxy sends a 200 OK response to the client before establishing a secure tunnel.
	TimeSendProxyResponse Time = "Send Proxy Response"

	// TimeWaitApproval is the time taken to wait for approval from the client. This should only be used for tunnel requests
	// since they don't call Perform.
	TimeWaitApproval Time = "Wait Approval"
	// TimeDelayPerform is the time taken to delay the request before performing it. This should only
	// be used for tunnel requests since they don't call Perform.
	TimeDelayPerform Time = "Perform Delay"
	// TimeTunnel is the time taken in the tunnel.
	TimeTunnel Time = "Tunnel"

	// TimeCertGenTLSHandshake is the time taken to generate a certificate and perform a TLS handshake. This is used for
	// HTTPS connections where the proxy is acting as the intended host (MITM).
	TimeCertGenTLSHandshake Time = "Cert Gen + TLS Handshake"
	// TimeReadRequest is the time taken to read the request from the client after the TLS handshake. This is used for HTTPS
	// connections where the proxy is acting as the intended host (MITM).
	TimeReadRequest Time = "Read Request"

	TimeTotal = "Total"
)

const (
	SubtimeNone Subtime = ""

	// Init
	SubtimeUUID                 Subtime = "UUID Generation"
	SubtimeGetClientProcessInfo Subtime = "Client Process Info"

	// Perform
	SubtimeWaitApproval Subtime = "Wait Approval"
	SubtimeDelayPerform Subtime = "Perform Delay"
	// SubtimeDialHost includes the time taken to dial the host for both HTTP and HTTPS (TLS) connections.
	SubtimeDialHost     Subtime = "Dial Host"
	SubtimeWriteRequest Subtime = "Write Request"
	SubtimeReadResponse Subtime = "Read Response"
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
	setStateFunc SetStateFuncType // this func will be called at the start of each major time key and minor time key with the key name

	MajorTimeKeys   []Time       `json:"majorTimeKeys"`
	MajorTimeValues []*MajorTime `json:"majorTimeValues"`
}

func (t *Timing) Start(key Time) {
	t.setStateFunc(string(key)) // call the state function with the key name
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
	t.setStateFunc(string(sub)) // call the state function with the sub key name
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

func New(setStateFunc SetStateFuncType) *Timing {
	t := &Timing{
		setStateFunc:    setStateFunc,
		MajorTimeKeys:   make([]Time, 0, 4),
		MajorTimeValues: make([]*MajorTime, 0, 4),
	}
	return t
}
