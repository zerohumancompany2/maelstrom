package main

import (
	"bufio"
	"fmt"
	"log"
	"os"

	"github.com/maelstrom/v3/pkg/statechart"
)

type ChatMessage struct {
	Sender  string
	Content string
}

type ChatSession struct {
	Content []ChatMessage
}

type DemoApplicationContext struct {
	Data map[string]any // We'll be using string -> string
}

func (m DemoApplicationContext) Get(key string, callerBoundary string) (any, []string, error) {
	val, exists := m.Data[key]
	if !exists {
		return nil, nil, fmt.Errorf("key %s not found", key)
	}
	return val, []string{}, nil
}

func (m DemoApplicationContext) Set(
	key string,
	value any,
	taints []string,
	callerBoundary string,
) error {
	m.Data[key] = value
	// TODO maybe throw error if key exists?
	return nil
}

func (DemoApplicationContext) Namespace() string { return "demo" }

func NewDemoApplicationContext() *DemoApplicationContext {
	return &DemoApplicationContext{
		Data: make(map[string]any),
	}
}

func main() {
	// create the engine
	engine := statechart.NewEngine()

	// create the application context
	appCtx := NewDemoApplicationContext()
	appCtx.Set("session", ChatSession{}, []string{}, "")

	// register actions
	engine.RegisterAction("logEntry", func(ctx statechart.RuntimeContext, appCtx statechart.ApplicationContext, ev statechart.Event) error {
		fmt.Printf("Received event %v", ev)
		return nil
	})

	// define chart
	def := statechart.ChartDefinition{
		ID:      "user-chat-session",
		Version: "1.0.0",
		Root: &statechart.Node{
			ID: "root",
			Children: map[string]*statechart.Node{
				"idle": {
					ID: "idle",
					Transitions: []statechart.Transition{
						{Event: "user-sends", Target: "llm-inference", Actions: []string{"logEntry"}},
					},
					IsInitial: true,
				},
				"llm-inference": {
					ID:          "llm-inference",
					Transitions: []statechart.Transition{{Event: "", Target: "idle"}},
					IsInitial:   false,
				},
			},
		},
		InitialState: "root/idle",
	}

	rtID, err := engine.Spawn(def, appCtx)
	engine.Control(rtID, statechart.CmdStart)

	if err != nil {
		panic(fmt.Sprintf("cannot spawn runtime engine with err %s", err))
	}

	fmt.Println("Chat demo initialized")

	// stand up stdio reader
	reader := bufio.NewReader(os.Stdin)

	// Enter chat loop.
	for {
		fmt.Print("> ")
		line, err := reader.ReadString('\n')
		if err != nil {
			log.Fatal(err)
		}
		fmt.Print(line)
		evt := statechart.Event{
			Type: "user-sends", Payload: line, CorrelationID: "", Source: "", TargetPath: "",
		}
		engine.Dispatch(rtID, evt)
	}
}
