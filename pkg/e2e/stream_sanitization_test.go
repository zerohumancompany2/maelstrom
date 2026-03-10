package e2e

import (
	"testing"
	"time"

	"github.com/maelstrom/v3/pkg/mail"
	security "github.com/maelstrom/v3/pkg/security"
)

func TestE2E_StreamSanitization_MultiChunkLatency(t *testing.T) {
	runtime := NewE2ERuntime()
	if err := runtime.Start(); err != nil {
		t.Fatalf("Failed to start runtime: %v", err)
	}
	defer runtime.Stop()

	dmzAgent := runtime.CreateAgent("dmz-agent", mail.DMZBoundary, security.TaintPolicy{
		AllowedForBoundary: []security.BoundaryType{security.DMZBoundary, security.OuterBoundary},
	})
	if dmzAgent == nil {
		t.Fatal("Failed to create DMZ agent")
	}

	session, err := runtime.StartStreamingSession("dmz-agent", mail.OuterBoundary)
	if err != nil {
		t.Fatalf("Failed to start streaming session: %v", err)
	}

	numChunks := 10
	measuredLatencies := make([]time.Duration, numChunks)
	chunkSizes := []int{1024, 2048, 3072, 4096, 5120, 6144, 7168, 8192, 9216, 10240}

	for i := 0; i < numChunks; i++ {
		chunkData := makeChunkData(chunkSizes[i])
		taints := []string{"TOOL_OUTPUT"}
		if i%2 == 0 {
			taints = append(taints, "INNER_ONLY")
		}

		start := time.Now()
		latency, err := runtime.SendStreamChunk(session, chunkData, taints)
		measuredLatencies[i] = time.Since(start)

		if err != nil {
			t.Fatalf("Failed to send chunk %d: %v", i, err)
		}

		totalProcessingTime := measuredLatencies[i] + latency
		if totalProcessingTime >= 50*time.Millisecond {
			t.Errorf("Chunk %d sanitization latency %v exceeds 50ms limit", i, totalProcessingTime)
		}
	}

	lastChunk := &session.chunks[len(session.chunks)-1]
	lastChunk.IsFinal = true

	result, err := runtime.EndStreamSession(session)
	if err != nil {
		t.Fatalf("Failed to end streaming session: %v", err)
	}

	if len(result.Chunks) != numChunks {
		t.Errorf("Expected %d chunks, got %d", numChunks, len(result.Chunks))
	}

	for i, chunk := range result.Chunks {
		if chunk.Sequence != i {
			t.Errorf("Chunk %d has sequence %d, expected %d", i, chunk.Sequence, i)
		}
		if len(chunk.Data) != chunkSizes[i] {
			t.Errorf("Chunk %d has size %d, expected %d", i, len(chunk.Data), chunkSizes[i])
		}
	}

	if !result.Chunks[len(result.Chunks)-1].IsFinal {
		t.Error("Last chunk IsFinal flag not set")
	}

	sumOfPerChunkLatencies := time.Duration(0)
	for _, lat := range measuredLatencies {
		sumOfPerChunkLatencies += lat
	}

	if result.TotalLatency < sumOfPerChunkLatencies {
		t.Errorf("Total latency %v less than sum of per-chunk latencies %v (indicates buffering)", result.TotalLatency, sumOfPerChunkLatencies)
	}
}

func makeChunkData(size int) string {
	data := make([]byte, size)
	for i := range data {
		data[i] = 'A' + byte((i % 26))
	}
	return string(data)
}
