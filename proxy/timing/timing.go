package timing

import (
	"time"
)

type Time = uint8
type Order = uint8

const (
	TimeRequestInit Time = iota

	TimePrepRequest
	TimeWaitApproval
	TimeDelayPeform
	TimeRequestPerform
	TimeDumpResponse
	TimeWriteResponse

	TimeProxyResponse
	TimeDialHost
	TimeReadWriteTunnel

	TimeCertGenTLSHandshake
	TimeReadParseRequest

	TimeTotal
)

const (
	OrderHTTP Order = iota
	OrderHTTPS
	OrderHTTPSMITM
)

type Timing struct {
	m     map[Time]time.Duration
	order Order
}

func (t *Timing) Start(key Time) func() {
	s := time.Now()
	return func() {
		if _, ok := t.m[key]; ok {
			panic("duplicate done call for timing key")
		}
		t.m[key] = time.Since(s)
	}
}

func (t *Timing) Export() map[string]any {
	return map[string]any{
		"order": t.order,
		"times": t.m,
	}
}

func New(secure, mitm bool) *Timing {
	t := &Timing{
		m: make(map[Time]time.Duration),
	}
	if secure && mitm {
		t.order = OrderHTTPSMITM // https mitm
	} else if secure {
		t.order = OrderHTTPS // https
	} else {
		t.order = OrderHTTP // http
	}
	return t
}
