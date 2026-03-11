package datasource

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
)

func TestFileDataSource_Read(t *testing.T) {
	cfg := NewConfig()
	ds := NewFileDataSource(cfg)

	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.json")

	testData := map[string]any{"key": "value"}
	dataJSON, _ := json.Marshal(testData)
	os.WriteFile(testFile, dataJSON, 0644)

	content, taints, err := ds.Read(testFile)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if content == nil {
		t.Error("Expected non-nil content")
	}

	if len(taints) == 0 {
		t.Error("Expected taints to be set")
	}
}

func TestFileDataSource_Write(t *testing.T) {
	cfg := NewConfig()
	ds := NewFileDataSource(cfg)

	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.json")

	testData := map[string]any{"key": "value"}
	writeTaints := []string{"TEST_TAINT"}

	err := ds.Write(testFile, testData, writeTaints)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if _, err := os.Stat(testFile); err != nil {
		t.Fatalf("Expected file to exist, got %v", err)
	}
}

func TestFileDataSource_WriteReadCycle(t *testing.T) {
	cfg := NewConfig()
	ds := NewFileDataSource(cfg)

	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.json")

	testData := map[string]any{"key": "value"}
	writeTaints := []string{"CUSTOM_TAINT"}

	err := ds.Write(testFile, testData, writeTaints)
	if err != nil {
		t.Fatalf("Write failed: %v", err)
	}

	content, readTaints, err := ds.Read(testFile)
	if err != nil {
		t.Fatalf("Read failed: %v", err)
	}

	if content == nil {
		t.Error("Expected non-nil content after write-read cycle")
	}

	if len(readTaints) == 0 {
		t.Error("Expected taints to be preserved after write-read cycle")
	}
}

func TestFileDataSource_DefaultTaintOnWrite(t *testing.T) {
	cfg := NewConfig()
	ds := NewFileDataSource(cfg)

	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.json")

	testData := "simple string"

	err := ds.Write(testFile, testData, nil)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	_, taints, err := ds.Read(testFile)
	if err != nil {
		t.Fatalf("Read failed: %v", err)
	}

	hasExternal := false
	for _, t := range taints {
		if t == "EXTERNAL" {
			hasExternal = true
			break
		}
	}
	if !hasExternal {
		t.Error("Expected EXTERNAL taint when no taints provided")
	}
}

func TestFileDataSource_CreateDirectories(t *testing.T) {
	cfg := NewConfig()
	ds := NewFileDataSource(cfg)

	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "nested", "deep", "test.json")

	testData := map[string]any{"key": "value"}

	err := ds.Write(testFile, testData, []string{"TEST"})
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if _, err := os.Stat(testFile); err != nil {
		t.Fatalf("Expected file to exist with created directories, got %v", err)
	}
}

func TestFileDataSource_ReadNonExistent(t *testing.T) {
	cfg := NewConfig()
	ds := NewFileDataSource(cfg)

	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "nonexistent.json")

	_, _, err := ds.Read(testFile)
	if err == nil {
		t.Error("Expected error for non-existent file")
	}

	if !strings.Contains(err.Error(), "no such file") {
		t.Logf("Got error: %v", err)
	}
}

func TestObjectDataSource_GetObject(t *testing.T) {
	cfg := NewConfig()
	ds := NewObjectDataSource(cfg)

	bucket := "test-bucket"
	key := "test-key"

	data, taints, err := ds.GetObject(bucket, key)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if data != nil {
		t.Error("Expected nil data for non-existent object")
	}

	if len(taints) != 0 {
		t.Errorf("Expected empty taints for non-existent object, got %v", taints)
	}
}

func TestObjectDataSource_PutObject(t *testing.T) {
	cfg := NewConfig()
	ds := NewObjectDataSource(cfg)

	bucket := "test-bucket"
	key := "test-key"
	testData := map[string]any{"key": "value"}
	writeTaints := []string{"OBJECT_TAINT"}

	err := ds.PutObject(bucket, key, testData, writeTaints)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	data, taints, err := ds.GetObject(bucket, key)
	if err != nil {
		t.Fatalf("GetObject failed: %v", err)
	}

	if data == nil {
		t.Error("Expected non-nil data after PutObject")
	}

	if len(taints) == 0 {
		t.Error("Expected taints to be set after PutObject")
	}
}

func TestObjectDataSource_PutGetCycle(t *testing.T) {
	cfg := NewConfig()
	ds := NewObjectDataSource(cfg)

	bucket := "test-bucket"
	key := "test-key"
	testData := "simple string"
	writeTaints := []string{"CUSTOM_TAINT"}

	err := ds.PutObject(bucket, key, testData, writeTaints)
	if err != nil {
		t.Fatalf("PutObject failed: %v", err)
	}

	data, readTaints, err := ds.GetObject(bucket, key)
	if err != nil {
		t.Fatalf("GetObject failed: %v", err)
	}

	if data == nil {
		t.Error("Expected non-nil data after put-get cycle")
	}

	if len(readTaints) == 0 {
		t.Error("Expected taints to be preserved after put-get cycle")
	}
}

func TestObjectDataSource_MultipleBuckets(t *testing.T) {
	cfg := NewConfig()
	ds := NewObjectDataSource(cfg)

	bucket1 := "bucket-1"
	bucket2 := "bucket-2"
	key := "shared-key"

	testData1 := map[string]any{"bucket": 1}
	testData2 := map[string]any{"bucket": 2}

	err := ds.PutObject(bucket1, key, testData1, []string{"TAINT1"})
	if err != nil {
		t.Fatalf("PutObject to bucket1 failed: %v", err)
	}

	err = ds.PutObject(bucket2, key, testData2, []string{"TAINT2"})
	if err != nil {
		t.Fatalf("PutObject to bucket2 failed: %v", err)
	}

	data1, _, _ := ds.GetObject(bucket1, key)
	data2, _, _ := ds.GetObject(bucket2, key)

	if data1 == nil || data2 == nil {
		t.Error("Expected data in both buckets")
	}
}

func TestObjectDataSource_DefaultTaintOnPut(t *testing.T) {
	cfg := NewConfig()
	ds := NewObjectDataSource(cfg)

	bucket := "test-bucket"
	key := "test-key"

	err := ds.PutObject(bucket, key, "data", nil)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	_, taints, _ := ds.GetObject(bucket, key)

	hasExternal := false
	for _, t := range taints {
		if t == "EXTERNAL" {
			hasExternal = true
			break
		}
	}
	if !hasExternal {
		t.Error("Expected EXTERNAL taint when no taints provided")
	}
}

func TestMemoryDataSource_Get(t *testing.T) {
	cfg := NewConfig()
	ds := NewMemoryDataSource(cfg)

	data, taints, err := ds.Get("nonexistent-key")
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if data != nil {
		t.Error("Expected nil data for non-existent key")
	}

	if len(taints) != 0 {
		t.Errorf("Expected empty taints for non-existent key, got %v", taints)
	}
}

func TestMemoryDataSource_Put(t *testing.T) {
	cfg := NewConfig()
	ds := NewMemoryDataSource(cfg)

	key := "test-key"
	testData := map[string]any{"key": "value"}
	writeTaints := []string{"MEMORY_TAINT"}

	err := ds.Put(key, testData, writeTaints)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	data, taints, err := ds.Get(key)
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}

	if data == nil {
		t.Error("Expected non-nil data after Put")
	}

	if len(taints) == 0 {
		t.Error("Expected taints to be set after Put")
	}
}

func TestMemoryDataSource_PutGetCycle(t *testing.T) {
	cfg := NewConfig()
	ds := NewMemoryDataSource(cfg)

	key := "test-key"
	testData := "simple string"
	writeTaints := []string{"CUSTOM_TAINT"}

	err := ds.Put(key, testData, writeTaints)
	if err != nil {
		t.Fatalf("Put failed: %v", err)
	}

	data, readTaints, err := ds.Get(key)
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}

	if data == nil {
		t.Error("Expected non-nil data after put-get cycle")
	}

	if len(readTaints) == 0 {
		t.Error("Expected taints to be preserved after put-get cycle")
	}
}

func TestMemoryDataSource_ThreadSafe(t *testing.T) {
	cfg := NewConfig()
	ds := NewMemoryDataSource(cfg)

	var wg sync.WaitGroup
	numGoroutines := 100

	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			key := "key-" + string(rune('a'+id%26))
			testData := map[string]any{"id": id}
			err := ds.Put(key, testData, []string{"TAINT"})
			if err != nil {
				t.Errorf("Put failed for goroutine %d: %v", id, err)
			}
		}(i)
	}

	wg.Wait()

	for i := 0; i < numGoroutines; i++ {
		key := "key-" + string(rune('a'+i%26))
		_, _, err := ds.Get(key)
		if err != nil {
			t.Errorf("Get failed for key %s: %v", key, err)
		}
	}
}

func TestMemoryDataSource_DefaultTaintOnPut(t *testing.T) {
	cfg := NewConfig()
	ds := NewMemoryDataSource(cfg)

	key := "test-key"

	err := ds.Put(key, "data", nil)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	_, taints, _ := ds.Get(key)

	hasExternal := false
	for _, t := range taints {
		if t == "EXTERNAL" {
			hasExternal = true
			break
		}
	}
	if !hasExternal {
		t.Error("Expected EXTERNAL taint when no taints provided")
	}
}

func TestNetworkDataSource_Fetch(t *testing.T) {
	cfg := NewConfig()
	ds := NewNetworkDataSource(cfg)

	_, taints, err := ds.Fetch("http://example.com/api")
	if err == nil {
		t.Log("Note: Network fetch succeeded (network available)")
	} else {
		t.Logf("Network fetch failed (expected in isolated env): %v", err)
	}

	hasExternal := false
	for _, t := range taints {
		if t == "EXTERNAL" {
			hasExternal = true
			break
		}
	}
	if !hasExternal {
		t.Error("Expected EXTERNAL taint on network fetch")
	}
}

func TestNetworkDataSource_FetchAlwaysExternalTaint(t *testing.T) {
	cfg := NewConfig()
	ds := NewNetworkDataSource(cfg)

	_, taints, _ := ds.Fetch("http://example.com/test")

	hasExternal := false
	for _, t := range taints {
		if t == "EXTERNAL" {
			hasExternal = true
			break
		}
	}
	if !hasExternal {
		t.Error("NetworkDataSource must always attach EXTERNAL taint")
	}
}

func TestNetworkDataSource_FetchInvalidURL(t *testing.T) {
	cfg := NewConfig()
	ds := NewNetworkDataSource(cfg)

	_, taints, err := ds.Fetch("http://invalid.invalid.invalid.invalid")
	if err == nil {
		t.Log("Note: Invalid URL fetch succeeded")
	} else {
		t.Logf("Invalid URL fetch failed as expected: %v", err)
	}

	hasExternal := false
	for _, t := range taints {
		if t == "EXTERNAL" {
			hasExternal = true
			break
		}
	}
	if !hasExternal {
		t.Error("Expected EXTERNAL taint even on fetch error")
	}
}

func TestFileDataSource_StringData(t *testing.T) {
	cfg := NewConfig()
	ds := NewFileDataSource(cfg)

	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.txt")

	testData := "plain text content"

	err := ds.Write(testFile, testData, []string{"TEXT_TAINT"})
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	content, _, err := ds.Read(testFile)
	if err != nil {
		t.Fatalf("Read failed: %v", err)
	}

	strContent, ok := content.(string)
	if !ok {
		t.Error("Expected string content for plain text file")
	}

	if strContent != testData {
		t.Errorf("Expected %q, got %q", testData, strContent)
	}
}

func TestObjectDataSource_StringData(t *testing.T) {
	cfg := NewConfig()
	ds := NewObjectDataSource(cfg)

	bucket := "test-bucket"
	key := "test-key"
	testData := "plain string data"

	err := ds.PutObject(bucket, key, testData, []string{"STRING_TAINT"})
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	content, _, err := ds.GetObject(bucket, key)
	if err != nil {
		t.Fatalf("GetObject failed: %v", err)
	}

	strContent, ok := content.(string)
	if !ok {
		t.Error("Expected string content")
	}

	if strContent != testData {
		t.Errorf("Expected %q, got %q", testData, strContent)
	}
}

func TestMemoryDataSource_Overwrite(t *testing.T) {
	cfg := NewConfig()
	ds := NewMemoryDataSource(cfg)

	key := "test-key"

	err := ds.Put(key, "first", []string{"FIRST"})
	if err != nil {
		t.Fatalf("First Put failed: %v", err)
	}

	data1, _, _ := ds.Get(key)
	if data1 != "first" {
		t.Errorf("Expected 'first', got %v", data1)
	}

	err = ds.Put(key, "second", []string{"SECOND"})
	if err != nil {
		t.Fatalf("Second Put failed: %v", err)
	}

	data2, _, _ := ds.Get(key)
	if data2 != "second" {
		t.Errorf("Expected 'second' after overwrite, got %v", data2)
	}
}
