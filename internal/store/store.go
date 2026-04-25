// Package store provides bbolt-backed persistent storage for escalation history,
// turn tracking, and predictive routing data.
package store

import (
	"encoding/binary"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	bolt "go.etcd.io/bbolt"
)

var (
	bucketEscalations = []byte("escalations")
	bucketTurns       = []byte("turns")
	bucketSessions    = []byte("sessions")
)

// Store manages the bbolt database for escalation data.
type Store struct {
	db *bolt.DB
}

// EscalationEvent represents a model escalation or de-escalation.
type EscalationEvent struct {
	ID        int64     `json:"id"`
	Timestamp time.Time `json:"timestamp"`
	FromModel string    `json:"from_model"`
	ToModel   string    `json:"to_model"`
	TaskType  string    `json:"task_type"`
	Reason    string    `json:"reason"`
}

// Turn represents a conversation turn for circular reasoning detection.
type Turn struct {
	Timestamp time.Time `json:"timestamp"`
	Model     string    `json:"model"`
	Concepts  string    `json:"concepts"`
}

// TaskTypeStats holds aggregated statistics for a task type.
type TaskTypeStats struct {
	TaskType    string  `json:"TaskType"`
	Escalations int     `json:"Escalations"`
	Successes   int     `json:"Successes"`
	SuccessRate float64 `json:"SuccessRate"`
}

// Open creates or opens the bbolt database at the given directory.
func Open(dataDir string) (*Store, error) {
	if err := os.MkdirAll(dataDir, 0755); err != nil {
		return nil, fmt.Errorf("creating data dir: %w", err)
	}

	dbPath := filepath.Join(dataDir, "escalation.db")
	db, err := bolt.Open(dbPath, 0600, &bolt.Options{Timeout: 1 * time.Second})
	if err != nil {
		return nil, fmt.Errorf("opening database: %w", err)
	}

	// Create buckets
	err = db.Update(func(tx *bolt.Tx) error {
		for _, b := range [][]byte{bucketEscalations, bucketTurns, bucketSessions} {
			if _, err := tx.CreateBucketIfNotExists(b); err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		db.Close()
		return nil, fmt.Errorf("creating buckets: %w", err)
	}

	return &Store{db: db}, nil
}

// Close closes the database.
func (s *Store) Close() error {
	return s.db.Close()
}

// itob encodes a uint64 as big-endian bytes (sortable key).
func itob(v uint64) []byte {
	b := make([]byte, 8)
	binary.BigEndian.PutUint64(b, v)
	return b
}

// btoi decodes big-endian bytes to uint64.
func btoi(b []byte) uint64 {
	return binary.BigEndian.Uint64(b)
}

// LogEscalation records an escalation or de-escalation event.
func (s *Store) LogEscalation(fromModel, toModel, taskType, reason string) error {
	event := EscalationEvent{
		Timestamp: time.Now(),
		FromModel: fromModel,
		ToModel:   toModel,
		TaskType:  taskType,
		Reason:    reason,
	}
	return s.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket(bucketEscalations)
		id, _ := b.NextSequence()
		event.ID = int64(id)
		data, err := json.Marshal(event)
		if err != nil {
			return err
		}
		return b.Put(itob(id), data)
	})
}

// LogTurn records a conversation turn with extracted concepts.
func (s *Store) LogTurn(model, concepts string) error {
	turn := Turn{
		Timestamp: time.Now(),
		Model:     model,
		Concepts:  concepts,
	}
	return s.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket(bucketTurns)
		id, _ := b.NextSequence()
		data, err := json.Marshal(turn)
		if err != nil {
			return err
		}
		if err := b.Put(itob(id), data); err != nil {
			return err
		}

		// Prune: keep last 100 turns
		count := 0
		c := b.Cursor()
		for k, _ := c.Last(); k != nil; k, _ = c.Prev() {
			count++
		}
		if count > 100 {
			toDelete := count - 100
			c2 := b.Cursor()
			for k, _ := c2.First(); k != nil && toDelete > 0; k, _ = c2.Next() {
				b.Delete(k)
				toDelete--
			}
		}
		return nil
	})
}

// RecentTurns returns the last N turns (newest first).
func (s *Store) RecentTurns(n int) ([]Turn, error) {
	var turns []Turn
	s.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(bucketTurns)
		c := b.Cursor()
		count := 0
		for k, v := c.Last(); k != nil && count < n; k, v = c.Prev() {
			var t Turn
			if err := json.Unmarshal(v, &t); err == nil {
				turns = append(turns, t)
			}
			count++
		}
		return nil
	})
	return turns, nil
}

// CountRecentAttempts counts recent turns on a specific model (in the last N turns).
func (s *Store) CountRecentAttempts(model string, last int) (int, error) {
	count := 0
	s.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(bucketTurns)
		c := b.Cursor()
		seen := 0
		for k, v := c.Last(); k != nil && seen < last; k, v = c.Prev() {
			var t Turn
			if err := json.Unmarshal(v, &t); err == nil {
				if t.Model == model {
					count++
				}
			}
			seen++
		}
		return nil
	})
	return count, nil
}

// TaskTypeStatsAll returns escalation statistics grouped by task type.
func (s *Store) TaskTypeStatsAll() ([]TaskTypeStats, error) {
	escByType := make(map[string]int)
	succByType := make(map[string]int)

	s.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(bucketEscalations)
		return b.ForEach(func(k, v []byte) error {
			var e EscalationEvent
			if err := json.Unmarshal(v, &e); err != nil {
				return nil
			}
			if e.Reason == "success" {
				succByType[e.TaskType]++
			} else {
				escByType[e.TaskType]++
			}
			return nil
		})
	})

	var stats []TaskTypeStats
	for taskType, esc := range escByType {
		st := TaskTypeStats{
			TaskType:    taskType,
			Escalations: esc,
			Successes:   succByType[taskType],
		}
		if esc > 0 {
			st.SuccessRate = float64(st.Successes) / float64(esc) * 100
		}
		stats = append(stats, st)
	}
	return stats, nil
}

// EscalationCountForType returns how many times a task type has been escalated (excluding successes).
func (s *Store) EscalationCountForType(taskType string) (int, error) {
	count := 0
	s.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(bucketEscalations)
		return b.ForEach(func(k, v []byte) error {
			var e EscalationEvent
			if err := json.Unmarshal(v, &e); err != nil {
				return nil
			}
			if e.TaskType == taskType && e.Reason != "success" {
				count++
			}
			return nil
		})
	})
	return count, nil
}

// TotalStats returns aggregate statistics.
func (s *Store) TotalStats() (escalations int, deescalations int, turns int, err error) {
	s.db.View(func(tx *bolt.Tx) error {
		tx.Bucket(bucketEscalations).ForEach(func(k, v []byte) error {
			var e EscalationEvent
			if err := json.Unmarshal(v, &e); err != nil {
				return nil
			}
			if e.Reason == "success" {
				deescalations++
			} else {
				escalations++
			}
			return nil
		})
		turns = tx.Bucket(bucketTurns).Stats().KeyN
		return nil
	})
	return
}

// RecentEscalations returns the last N escalation events (newest first).
func (s *Store) RecentEscalations(n int) ([]EscalationEvent, error) {
	var events []EscalationEvent
	s.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(bucketEscalations)
		c := b.Cursor()
		count := 0
		for k, v := c.Last(); k != nil && count < n; k, v = c.Prev() {
			var e EscalationEvent
			if err := json.Unmarshal(v, &e); err == nil {
				events = append(events, e)
			}
			count++
		}
		return nil
	})
	return events, nil
}

// SetSession stores a key-value pair for session state.
func (s *Store) SetSession(key, value string) error {
	return s.db.Update(func(tx *bolt.Tx) error {
		return tx.Bucket(bucketSessions).Put([]byte(key), []byte(value))
	})
}

// GetSession retrieves a session value. Returns empty string if not found.
func (s *Store) GetSession(key string) (string, error) {
	var value string
	s.db.View(func(tx *bolt.Tx) error {
		v := tx.Bucket(bucketSessions).Get([]byte(key))
		if v != nil {
			value = string(v)
		}
		return nil
	})
	return value, nil
}

// DeleteSession removes a session key.
func (s *Store) DeleteSession(key string) error {
	return s.db.Update(func(tx *bolt.Tx) error {
		return tx.Bucket(bucketSessions).Delete([]byte(key))
	})
}
