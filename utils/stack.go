package utils

import (
	"fmt"
	"sync"
	"time"
)

// Stack represents a stack data structure

// Example usage:
// stack := utils.NewStack(100, 10, 1)
// stack.Write("1", "2")
// stack.Write("3", "4", "5")
// stack.Write("6", "7", "8", "9", "10")
// for {
// 	cnt := stack.GetCapacity()
// 	if cnt > 0 {
// 		stack.Print()
// 		utils.SleepMilli(500)
// 	} else {
// 		break
// 	}
// }
// cleanup something before exit
// utils.AddGracefulExit(func() error {
// 	stack.StopCleanupRoutine() // stop cleanup routine
// 	utils.Log().Info("exit")
// 	utils.SleepSec(1)
// 	return nil
// })

// StackEntry represents an entry in the stack
type StackEntry struct {
	Timestamp int64
	Value     string
}

// Stack struct with a mutex for concurrent safety
type Stack struct {
	data            []StackEntry
	maxSize         int
	expireAfter     int64
	mu              sync.Mutex
	stopChan        chan struct{}
	cleanupInterval int // interval in seconds for cleanup
}

// NewStack creates a new Stack with a specified maximum size, expiration duration, and cleanup interval
func NewStack(maxSize int, expireAfter int64, cleanupInterval int) *Stack {
	stack := &Stack{
		data:            make([]StackEntry, 0, maxSize),
		maxSize:         maxSize,
		expireAfter:     expireAfter,
		stopChan:        make(chan struct{}),
		cleanupInterval: cleanupInterval,
	}

	go stack.startCleanupRoutine()

	return stack
}

// Write adds strings to the stack with a Unix timestamp and removes excess elements if necessary
func (s *Stack) Write(strings ...string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Calculate the space needed and insert in reverse order
	newEntries := make([]StackEntry, len(strings))
	currentTime := time.Now().Unix()
	for i := len(strings) - 1; i >= 0; i-- {
		newEntries[len(strings)-1-i] = StackEntry{
			Timestamp: currentTime,
			Value:     strings[i],
		}
	}

	// Insert the new entries at the beginning
	s.data = append(newEntries, s.data...)

	// Trim the slice if it exceeds maxSize
	if len(s.data) > s.maxSize {
		s.data = s.data[:s.maxSize]
	}
}

// removeExpiredEntries removes elements that have expired based on the current time
func (s *Stack) removeExpiredEntries() {
	s.mu.Lock()
	defer s.mu.Unlock()

	currentTime := time.Now().Unix()

	// fast cleanup, from back to front, compare the current timestamp and the expiration timestamp of the element
	// find the first expired element, and then discard all elements after the element after that element
	// finally, discard all elements after the first expired element
	lastIndex := 0
	for i := len(s.data) - 1; i >= 0; i-- {
		expiredTime := s.data[i].Timestamp + s.expireAfter
		if currentTime <= expiredTime {
			lastIndex = i + 1
			break
		}
	}
	s.data = s.data[:lastIndex]
}

// startCleanupRoutine starts a goroutine that periodically removes expired elements
func (s *Stack) startCleanupRoutine() {
	ticker := time.NewTicker(time.Duration(s.cleanupInterval) * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			s.removeExpiredEntries()
		case <-s.stopChan:
			return
		}
	}
}

// StopCleanupRoutine stops the cleanup goroutine
func (s *Stack) StopCleanupRoutine() {
	close(s.stopChan)
}

// Print prints the elements of the stack
func (s *Stack) Print() {
	s.mu.Lock()
	defer s.mu.Unlock()

	fmt.Println("Stack contents:")
	for _, entry := range s.data {
		fmt.Printf("%d: %s\n", entry.Timestamp, entry.Value)
	}
}

// GetData returns a pointer to the underlying data slice
func (s *Stack) GetData() *[]StackEntry {
	s.mu.Lock()
	defer s.mu.Unlock()
	return &s.data
}

// GetCapacity returns the current capacity of the stack
func (s *Stack) GetCapacity() int {
	s.mu.Lock()
	defer s.mu.Unlock()
	return len(s.data)
}
