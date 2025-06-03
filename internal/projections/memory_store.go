package projections

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
)

// MemoryReadModelStore is an in-memory implementation of ReadModelStore
type MemoryReadModelStore struct {
	mu   sync.RWMutex
	data map[string]map[string][]byte // collection -> id -> json data
}

// NewMemoryReadModelStore creates a new in-memory read model store
func NewMemoryReadModelStore() *MemoryReadModelStore {
	return &MemoryReadModelStore{
		data: make(map[string]map[string][]byte),
	}
}

// Save saves a read model
func (s *MemoryReadModelStore) Save(ctx context.Context, collection string, id string, data interface{}) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Serialize data to JSON
	jsonData, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("failed to marshal data: %w", err)
	}

	// Ensure collection exists
	if s.data[collection] == nil {
		s.data[collection] = make(map[string][]byte)
	}

	// Save data
	s.data[collection][id] = jsonData
	return nil
}

// Get retrieves a read model
func (s *MemoryReadModelStore) Get(ctx context.Context, collection string, id string, result interface{}) error {
	s.mu.RLock()
	defer s.mu.RUnlock()

	// Check if collection exists
	collectionData, exists := s.data[collection]
	if !exists {
		return fmt.Errorf("collection %s not found", collection)
	}

	// Check if document exists
	jsonData, exists := collectionData[id]
	if !exists {
		return fmt.Errorf("document %s not found in collection %s", id, collection)
	}

	// Deserialize data
	if err := json.Unmarshal(jsonData, result); err != nil {
		return fmt.Errorf("failed to unmarshal data: %w", err)
	}

	return nil
}

// Query queries read models (simplified - returns all documents in collection)
func (s *MemoryReadModelStore) Query(ctx context.Context, collection string, filter interface{}) ([]interface{}, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	// Check if collection exists
	collectionData, exists := s.data[collection]
	if !exists {
		return []interface{}{}, nil
	}

	// Return all documents (no filtering in this simple implementation)
	results := make([]interface{}, 0, len(collectionData))
	for _, jsonData := range collectionData {
		var doc interface{}
		if err := json.Unmarshal(jsonData, &doc); err != nil {
			return nil, fmt.Errorf("failed to unmarshal document: %w", err)
		}
		results = append(results, doc)
	}

	return results, nil
}

// Delete deletes a read model
func (s *MemoryReadModelStore) Delete(ctx context.Context, collection string, id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Check if collection exists
	if collectionData, exists := s.data[collection]; exists {
		delete(collectionData, id)
	}

	return nil
}

// Clear clears all data in a collection
func (s *MemoryReadModelStore) Clear(ctx context.Context, collection string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Clear the collection
	delete(s.data, collection)
	return nil
}
