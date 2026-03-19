package lessons

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"sync"
	"time"
)

// FileStore implements LessonStore using JSONL files.
type FileStore struct {
	mu        sync.RWMutex
	localPath string // Project-local lessons file
	globalPath string // Global lessons file (~/.moai/global-lessons/)
	maxActive int
}

// NewFileStore creates a new file-based lesson store.
func NewFileStore(localPath, globalPath string, maxActive int) *FileStore {
	return &FileStore{
		localPath:  localPath,
		globalPath: globalPath,
		maxActive:  maxActive,
	}
}

// Save persists a lesson to the local JSONL file.
func (s *FileStore) Save(lesson *Lesson) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if lesson.CreatedAt.IsZero() {
		lesson.CreatedAt = time.Now()
	}
	lesson.UpdatedAt = time.Now()

	lessons, err := s.readAll(s.localPath)
	if err != nil {
		// readAll already handles os.IsNotExist internally (returns nil, nil).
		// Any error returned here is a real I/O failure; propagate it to avoid
		// silently discarding existing lesson data.
		return fmt.Errorf("read existing lessons: %w", err)
	}
	if lessons == nil {
		lessons = []*Lesson{}
	}

	// Update existing or append new
	found := false
	for i, l := range lessons {
		if l.ID == lesson.ID {
			lessons[i] = lesson
			found = true
			break
		}
	}
	if !found {
		lessons = append(lessons, lesson)
	}

	// Archive oldest if over max
	activeCount := 0
	for _, l := range lessons {
		if l.Active {
			activeCount++
		}
	}
	if activeCount > s.maxActive {
		s.archiveOldest(lessons, activeCount-s.maxActive)
	}

	return s.writeAll(s.localPath, lessons)
}

// List returns lessons matching the filter.
func (s *FileStore) List(filter LessonFilter) ([]*Lesson, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	lessons, err := s.readAll(s.localPath)
	if err != nil {
		return nil, err
	}

	// Also read global lessons if path is set
	if s.globalPath != "" {
		global, err := s.readAll(s.globalPath)
		if err == nil {
			lessons = append(lessons, global...)
		}
	}

	var result []*Lesson
	for _, l := range lessons {
		if s.matchesFilter(l, filter) {
			result = append(result, l)
		}
	}

	// Sort by hit count descending
	sort.Slice(result, func(i, j int) bool {
		return result[i].HitCount > result[j].HitCount
	})

	if filter.Limit > 0 && len(result) > filter.Limit {
		result = result[:filter.Limit]
	}

	return result, nil
}

// Get retrieves a single lesson by ID.
func (s *FileStore) Get(id string) (*Lesson, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	lessons, err := s.readAll(s.localPath)
	if err != nil {
		return nil, err
	}

	for _, l := range lessons {
		if l.ID == id {
			return l, nil
		}
	}
	return nil, fmt.Errorf("lesson not found: %s", id)
}

// Archive marks a lesson as inactive.
func (s *FileStore) Archive(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	lessons, err := s.readAll(s.localPath)
	if err != nil {
		return err
	}

	for _, l := range lessons {
		if l.ID == id {
			l.Active = false
			l.UpdatedAt = time.Now()
			return s.writeAll(s.localPath, lessons)
		}
	}
	return fmt.Errorf("lesson not found: %s", id)
}

// Count returns the number of active lessons.
func (s *FileStore) Count() (int, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	lessons, err := s.readAll(s.localPath)
	if err != nil {
		return 0, err
	}

	count := 0
	for _, l := range lessons {
		if l.Active {
			count++
		}
	}
	return count, nil
}

func (s *FileStore) readAll(path string) ([]*Lesson, error) {
	f, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("open lessons file: %w", err)
	}
	defer f.Close()

	var lessons []*Lesson
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		var l Lesson
		if err := json.Unmarshal(scanner.Bytes(), &l); err != nil {
			continue // skip malformed lines
		}
		lessons = append(lessons, &l)
	}
	return lessons, scanner.Err()
}

func (s *FileStore) writeAll(path string, lessons []*Lesson) error {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("create lessons directory: %w", err)
	}

	f, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("create lessons file: %w", err)
	}
	defer f.Close()

	enc := json.NewEncoder(f)
	for _, l := range lessons {
		if err := enc.Encode(l); err != nil {
			return fmt.Errorf("encode lesson: %w", err)
		}
	}
	return nil
}

func (s *FileStore) matchesFilter(l *Lesson, f LessonFilter) bool {
	if f.Type != "" && l.Type != f.Type {
		return false
	}
	if f.Active != nil && l.Active != *f.Active {
		return false
	}
	if f.MinHits > 0 && l.HitCount < f.MinHits {
		return false
	}
	if len(f.Tags) > 0 {
		matched := false
		for _, ft := range f.Tags {
			for _, lt := range l.Tags {
				if ft == lt {
					matched = true
					break
				}
			}
			if matched {
				break
			}
		}
		if !matched {
			return false
		}
	}
	return true
}

func (s *FileStore) archiveOldest(lessons []*Lesson, count int) {
	// Sort active lessons by updated time ascending
	var active []*Lesson
	for _, l := range lessons {
		if l.Active {
			active = append(active, l)
		}
	}
	sort.Slice(active, func(i, j int) bool {
		return active[i].UpdatedAt.Before(active[j].UpdatedAt)
	})

	archived := 0
	for _, l := range active {
		if archived >= count {
			break
		}
		l.Active = false
		archived++
	}
}
