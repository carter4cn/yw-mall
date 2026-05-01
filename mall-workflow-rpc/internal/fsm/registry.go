package fsm

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"mall-workflow-rpc/internal/model"

	"github.com/qmuntal/stateless"
)

// Definition mirrors the JSON we persist in workflow_definition.
// Each Transition entry is keyed by `from_state` → list of permitted
// (trigger, to_state) pairs. The FSM has no inline guards/actions for now —
// guards are evaluated explicitly in the Fire logic by the caller via rule-rpc.
type Definition struct {
	Code            string       `json:"code"`
	Description     string       `json:"description,omitempty"`
	InitialState    string       `json:"initial_state"`
	States          []string     `json:"states"`
	Transitions     []Transition `json:"transitions"`
	TerminalStates  []string     `json:"terminal_states,omitempty"`
}

type Transition struct {
	From    string `json:"from"`
	Trigger string `json:"trigger"`
	To      string `json:"to"`
}

// Registry caches parsed Definitions keyed by definition id.
type Registry struct {
	mu     sync.RWMutex
	defs   map[uint64]*Definition
	models model.WorkflowDefinitionModel
}

func NewRegistry(m model.WorkflowDefinitionModel) *Registry {
	return &Registry{
		defs:   make(map[uint64]*Definition),
		models: m,
	}
}

// Parse builds a Definition from the raw definition row.
// states_json holds {"initial":"REGISTERED","states":[...],"terminal":[...]}
// transitions_json holds [{from,trigger,to}, ...]
func Parse(def *model.WorkflowDefinition) (*Definition, error) {
	d := &Definition{Code: def.Code, Description: def.Description}
	var states struct {
		Initial  string   `json:"initial"`
		States   []string `json:"states"`
		Terminal []string `json:"terminal"`
	}
	if err := json.Unmarshal([]byte(def.StatesJson), &states); err != nil {
		return nil, fmt.Errorf("states_json: %w", err)
	}
	d.InitialState = states.Initial
	d.States = states.States
	d.TerminalStates = states.Terminal
	if err := json.Unmarshal([]byte(def.TransitionsJson), &d.Transitions); err != nil {
		return nil, fmt.Errorf("transitions_json: %w", err)
	}
	return d, nil
}

// Get returns a cached definition or loads it from DB.
func (r *Registry) Get(ctx context.Context, id uint64) (*Definition, error) {
	r.mu.RLock()
	if d, ok := r.defs[id]; ok {
		r.mu.RUnlock()
		return d, nil
	}
	r.mu.RUnlock()

	row, err := r.models.FindOne(ctx, id)
	if err != nil {
		return nil, err
	}
	d, err := Parse(row)
	if err != nil {
		return nil, err
	}
	r.mu.Lock()
	r.defs[id] = d
	r.mu.Unlock()
	return d, nil
}

// Bust removes a cached definition (called by RegisterDefinition on update).
func (r *Registry) Bust(id uint64) {
	r.mu.Lock()
	delete(r.defs, id)
	r.mu.Unlock()
}

// Build constructs a stateless.StateMachine bound to a single instance state.
// `currentState` is read from the workflow_instance row; transitions are
// configured for every (from→trigger→to) tuple in the definition. The caller
// fires triggers and reads the resulting state from the machine.
func Build(d *Definition, currentState string) *stateless.StateMachine {
	sm := stateless.NewStateMachine(currentState)
	// register every state mentioned, ensuring `Permit` calls are anchored.
	configured := map[string]*stateless.StateConfiguration{}
	cfg := func(s string) *stateless.StateConfiguration {
		if c, ok := configured[s]; ok {
			return c
		}
		c := sm.Configure(s)
		configured[s] = c
		return c
	}
	for _, s := range d.States {
		cfg(s)
	}
	for _, t := range d.Transitions {
		cfg(t.From).Permit(t.Trigger, t.To)
	}
	return sm
}

// FireResult captures the outcome of a single Fire invocation.
type FireResult struct {
	From     string
	To       string
	Trigger  string
	Advanced bool
	Latency  time.Duration
	Err      error
}
