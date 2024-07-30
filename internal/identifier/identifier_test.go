package identifier_test

import (
	"chariottakehome/internal/identifier"
	"sync"
	"testing"
	"unicode/utf8"
)

// Unit Tests
func TestIdentifiersUniqueness(t *testing.T) {
	idUniqueMap := make(map[identifier.Identifier]bool)

	for i := 0; i < 10000; i++ {
		id, err := identifier.New()
		if err != nil {
			t.Fatal("Failed to generate identifier")
		}

		if found := idUniqueMap[id]; found {
			t.Fatalf("Identifier not unique: %s", id)
		}

		idUniqueMap[id] = true
	}
}

func TestIdentifiersMonotonicity(t *testing.T) {
	startId, err := identifier.New()
	if err != nil {
		t.Fatal("Failed to generate identifier")
	}

	prevId := startId

	for i := 0; i < 10000; i++ {
		currId, err := identifier.New()

		if err != nil {
			t.Fatal("Failed to generate identifier")
		}

		if currId.String() <= prevId.String() {
			t.Fatalf("The current identifier '%s' is less than the previous identifier '%s'", currId, prevId)
		}
		prevId = currId
	}
}

func TestIdentifiersLength(t *testing.T) {
	id, err := identifier.New()
	if err != nil {
		t.Fatal("Failed to generate identifier")
	}
	idStr := id.String()

	if utf8.RuneCountInString(idStr) != 20 {
		t.Fatalf("Identifier '%s' character count is not equal to 20 characters, %d", id, len(idStr))
	}

	if len(idStr) != 20 {
		t.Fatalf("Identifier '%s' byte count is not equal to 20, %d", id, len(idStr))
	}
}

func TestHighConcurrentIdentifierUniqueness(t *testing.T) {
	const (
		numRoutines   = 1000
		idsPerRoutine = 10000
	)

	var (
		wg    sync.WaitGroup
		mutex sync.Mutex
	)

	idChannel := make(chan identifier.Identifier, numRoutines*idsPerRoutine)
	idUniqueMap := make(map[identifier.Identifier]bool)

	generateIdentifiers := func(wg *sync.WaitGroup, idChannel chan identifier.Identifier, n int) {
		defer wg.Done()
		for i := 0; i < n; i++ {
			id, err := identifier.New()
			if err != nil {
				panic(err)
			}
			idChannel <- id
		}
	}

	for i := 0; i < numRoutines; i++ {
		wg.Add(1)
		go generateIdentifiers(&wg, idChannel, idsPerRoutine)
	}

	go func() {
		wg.Wait()
		close(idChannel)
	}()

	for id := range idChannel {
		mutex.Lock()
		if exists := idUniqueMap[id]; exists {
			t.Fatalf("Duplicate identifier found: %s", id)
		}
		idUniqueMap[id] = true
		mutex.Unlock()
	}
}

func TestValidFromString(t *testing.T) {
	const validString string = "c-0UJqBRQX36SNln7ilE"

	id, err := identifier.FromString(validString)
	if err != nil {
		t.Fatalf("Failed to generate identifier from string %s: %s", validString, err)
	}

	if id.String() != validString {
		t.Fatalf("Failed to generate identifier with exact string %s", validString)
	}
}

func TestInvalidFromString(t *testing.T) {
	badStrings := []string{"c-0UJqB:QX36SNln7ilE", "c-", "x-LZ4QX0D7002KHGBPM3"}

	for _, badStr := range badStrings {
		_, err := identifier.FromString(badStr)
		if err == nil {
			t.Fatalf("Failed to reject bad string %s", badStr)
		}
	}
}

func TestValidFromBytes(t *testing.T) {
	validBytes := []byte("c-0UJqBRQX36SNln7ilE")

	id, err := identifier.FromBytes(validBytes)
	if err != nil {
		t.Fatalf("Failed to generate identifier from bytes %b: %s", validBytes, err)
	}

	if id.Bytes() != [20]byte(validBytes) {
		t.Fatalf("Failed to generate identifier with exact bytes %b", validBytes)
	}
}

func TestInvalidFromBytes(t *testing.T) {
	badBytes := [][]byte{
		[]byte("c-LZ4:X0D7002KHGBPM3"), []byte("c-"), []byte("x-LZ4QX0D7002KHGBPM3")}

	for _, badB := range badBytes {
		_, err := identifier.FromBytes(badB)
		if err == nil {
			t.Fatalf("Failed to reject bad string %b", badB)
		}
	}
}

func TestValidate(t *testing.T) {
	validId, err := identifier.FromString("c-0UJqBRQX36SNln7ilE")
	if err != nil {
		panic(err)
	}

	validRes, err := identifier.Validate(validId)
	if err != nil {
		panic(err)
	}

	if !validRes {
		t.Fatalf("Failed to recognize '%s' as valid", validId)
	}

	validRes, err = identifier.Validate(identifier.Identifier{})
	if err != nil {
		panic(err)
	}

	if validRes {
		t.Fatal("Failed to reject invalid identifier")
	}
}

// Benchmarking
func BenchmarkNew(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_, _ = identifier.New()
	}
}

func BenchmarkFromString(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_, _ = identifier.FromString("c-0UJqBRQX36SNln7ilE")
	}
}

func BenchmarkFromBytes(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_, _ = identifier.FromBytes([]byte("c-0UJqBRQX36SNln7ilE"))
	}
}

func BenchmarkValidate(b *testing.B) {
	id, err := identifier.New()
	if err != nil {
		panic(err)
	}
	for i := 0; i < b.N; i++ {
		_, _ = identifier.Validate(id)
	}
}
