package simulation

import (
	"fmt"
	"log"
	"math/rand"
	"reddit/actors"
	"reddit/messages"
	"sync"
	"time"

	"github.com/asynkron/protoactor-go/actor"
)

type SimulationMetrics struct {
    TotalRequests     int64
    SuccessfulRequests int64
    FailedRequests    int64
    ActionCounts      map[messages.ActionType]int64
    ResponseTimes       map[messages.ActionType][]time.Duration
    ErrorCounts         map[messages.ActionType]int64
    mu                  sync.RWMutex
}

type SimulationController struct {
    system       *actor.ActorSystem
    enginePIDs   []*actor.PID
    clientPIDs   []*actor.PID
    metrics      *SimulationMetrics
    zipf        *rand.Zipf
    pid         *actor.PID
    numEngines   int
    numClients   int
}

func NewSimulationController(system *actor.ActorSystem, numEngines int, numClients int) *SimulationController {
    // Initialize Zipf distribution for load balancing
    s := 1.1 // skewness parameter
    v := 1.0 // value parameter
    imax := uint64(numEngines) // number of engines
    zipf := rand.NewZipf(rand.New(rand.NewSource(time.Now().UnixNano())), s, v, imax)

    metrics := &SimulationMetrics{
        ActionCounts:   make(map[messages.ActionType]int64),
        ResponseTimes:  make(map[messages.ActionType][]time.Duration),
        ErrorCounts:    make(map[messages.ActionType]int64),
    }
    return &SimulationController{
        system: system,
        metrics: metrics,
        zipf: zipf,
        numEngines: numEngines,
        numClients: numClients,
    }
}

func (sc *SimulationController) updateMetrics(msg *messages.MetricsMessage) {
    sc.metrics.mu.Lock()
    defer sc.metrics.mu.Unlock()

    sc.metrics.TotalRequests++
    if msg.Success {
        sc.metrics.SuccessfulRequests++
    } else {
        sc.metrics.FailedRequests++
        sc.metrics.ErrorCounts[messages.ActionType(msg.Action)]++
    }

    actionType := messages.ActionType(msg.Action)
    sc.metrics.ActionCounts[actionType]++
    sc.metrics.ResponseTimes[actionType] = append(
        sc.metrics.ResponseTimes[actionType],
        msg.ResponseTime,
    )
}

func (sc *SimulationController) Receive(context actor.Context) {
    switch msg := context.Message().(type) {
    case *actor.Started:
        sc.pid = context.Self()
            // Start simulation
        err := sc.StartSimulation()
        if err != nil {
            log.Fatalf("Failed to start simulation: %v", err)
        }
    case *messages.MetricsMessage:
        sc.updateMetrics(msg)
    }
}

func (sc *SimulationController) getEngineActor() *actor.PID {
    // Use Zipf distribution to select engine
    index := sc.zipf.Uint64() % uint64(len(sc.enginePIDs))
    return sc.enginePIDs[index]
}

func (sc *SimulationController) StartSimulation() error {
    fmt.Println("Starting simulation...")
    // Create engine actors
    for i := 0; i < sc.numEngines; i++ {
        engineProps := actor.PropsFromProducer(func() actor.Actor {
            return actors.NewEngineActor()
        })
        enginePID := sc.system.Root.Spawn(engineProps)
        sc.enginePIDs = append(sc.enginePIDs, enginePID)
    }

    // Create client actors
    for i := 0; i < sc.numClients; i++ {
        clientProps := actor.PropsFromProducer(func() actor.Actor {
            return actors.NewClientActor(sc.getEngineActor(), sc.pid)
        })
        clientPID := sc.system.Root.Spawn(clientProps)
        sc.clientPIDs = append(sc.clientPIDs, clientPID)
    }

    return nil
}

func (sc *SimulationController) GetMetrics() map[string]interface{} {
    sc.metrics.mu.RLock()
    defer sc.metrics.mu.RUnlock()

    summary := make(map[string]interface{})
    summary["total_requests"] = sc.metrics.TotalRequests
    summary["successful_requests"] = sc.metrics.SuccessfulRequests
    summary["failed_requests"] = sc.metrics.FailedRequests

    // Calculate average response times per action
    avgResponseTimes := make(map[messages.ActionType]time.Duration)
    for action, times := range sc.metrics.ResponseTimes {
        var total time.Duration
        for _, t := range times {
            total += t
        }
        if len(times) > 0 {
            avgResponseTimes[action] = total / time.Duration(len(times))
        }
    }
    summary["avg_response_times"] = avgResponseTimes
    summary["action_counts"] = sc.metrics.ActionCounts
    summary["error_counts"] = sc.metrics.ErrorCounts

    return summary
}

