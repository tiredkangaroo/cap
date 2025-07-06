package main

import (
	"log/slog"

	"github.com/tiredkangaroo/cap/proxy/config"
	"github.com/tiredkangaroo/cap/proxy/http"
)

// tca is trigger-conditions-actions.

type Trigger uint8

const (
	TriggerUnknown Trigger = iota
	TriggerOnRequest
	TriggerOnPerform
	TriggerOnResponse
	TriggerOnError
	TriggerOnConfigChange
	TriggerOnShutdown
)

func (t Trigger) String() string {
	switch t {
	case TriggerUnknown:
		return "Unknown"
	case TriggerOnRequest:
		return "On Request"
	case TriggerOnPerform:
		return "On Perform"
	case TriggerOnResponse:
		return "On Response"
	case TriggerOnError:
		return "On Error"
	case TriggerOnConfigChange:
		return "On Config Change"
	case TriggerOnShutdown:
		return "On Shutdown"
	default:
		return "Invalid Trigger"
	}
}

type TCAService struct {
	tcas    map[Trigger][]TCA
	db      *Database
	manager *Manager
}

// Trigger sets off all TCAs with the given trigger
func (s *TCAService) Trigger(t Trigger, ctx TCAContext) {
	s.maximizeContext(ctx)
	ctx.Config = config.DefaultConfig

	tcas, ok := s.tcas[t]
	if !ok { // no TCAs for this trigger
		return
	}
	for _, tca := range tcas {
		if !checkConditions(tca.conditions, ctx) {
			continue // skip this TCA if conditions are not met
		}
		for _, action := range tca.actions {
			if err := action(ctx); err != nil {
				slog.Error("tca action error", "err", err.Error(), "trigger", t.String())
			}
		}
	}
}

func checkConditions(conditions []ConditionFunc, ctx TCAContext) bool {
	for _, condition := range conditions {
		if !condition(ctx) {
			return false
		}
	}
	return true
}

func (s *TCAService) maximizeContext(ctx TCAContext) TCAContext {
	if ctx.Database == nil {
		ctx.Database = s.db
	}
	if ctx.Manager == nil {
		ctx.Manager = s.manager
	}
	if ctx.Config == nil {
		ctx.Config = config.DefaultConfig
	}
	return ctx
}

type TCA struct {
	trigger    Trigger
	conditions []ConditionFunc // all conditions must be true for actions to be executed
	actions    []ActionFunc
}
type ConditionFunc func(ctx TCAContext) bool
type ActionFunc func(ctx TCAContext) error

type TCAContext struct {
	Database     *Database
	Manager      *Manager
	ProxyRequest *Request
	HTTPRequest  *http.Request
	HTTPResponse *http.Response
	Error        error
	Config       *config.Config // this will just point to config.DefaultConfig
}

func NewTCAService() *TCAService {
	return &TCAService{
		tcas: make(map[Trigger][]TCA, 8),
	}
}
