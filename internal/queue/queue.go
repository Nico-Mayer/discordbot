// Package queue holds the tracks waiting to play for the single guild this bot serves.
package queue

import (
	"slices"
	"sync"

	"github.com/disgoorg/disgolink/v3/lavalink"
)

// Queue is a FIFO track queue. Reads and writes arrive from concurrent gateway
// and player event handlers, so every method takes the mutex. The zero value is
// ready to use.
type Queue struct {
	mu     sync.Mutex
	tracks []lavalink.Track
}

// Add appends tracks to the back of the queue.
func (q *Queue) Add(tracks ...lavalink.Track) {
	q.mu.Lock()
	defer q.mu.Unlock()
	q.tracks = append(q.tracks, tracks...)
}

// Next removes and returns the track at the front of the queue. It reports false
// when the queue is empty.
func (q *Queue) Next() (lavalink.Track, bool) {
	q.mu.Lock()
	defer q.mu.Unlock()
	if len(q.tracks) == 0 {
		return lavalink.Track{}, false
	}
	track := q.tracks[0]
	q.tracks = q.tracks[1:]
	return track, true
}

// Clear discards every queued track.
func (q *Queue) Clear() {
	q.mu.Lock()
	defer q.mu.Unlock()
	q.tracks = nil
}

// Len reports how many tracks are queued.
func (q *Queue) Len() int {
	q.mu.Lock()
	defer q.mu.Unlock()
	return len(q.tracks)
}

// Tracks returns the queued tracks in play order as a copy, so a caller cannot
// mutate the queue through the returned slice.
func (q *Queue) Tracks() []lavalink.Track {
	q.mu.Lock()
	defer q.mu.Unlock()
	return slices.Clone(q.tracks)
}
