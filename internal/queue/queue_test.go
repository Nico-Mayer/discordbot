package queue_test

import (
	"sync"
	"testing"

	"github.com/disgoorg/disgolink/v3/lavalink"
	"github.com/stretchr/testify/require"

	"github.com/nico-mayer/discordbot/internal/queue"
)

func track(title string) lavalink.Track {
	return lavalink.Track{Info: lavalink.TrackInfo{Title: title}}
}

func TestQueueZeroValue(t *testing.T) {
	t.Parallel()

	var q queue.Queue

	require.Zero(t, q.Len())
	require.Empty(t, q.Tracks())

	_, ok := q.Next()
	require.False(t, ok)

	q.Add(track("a"))
	got, ok := q.Next()
	require.True(t, ok)
	require.Equal(t, "a", got.Info.Title)
}

func TestQueueAddAndNext(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		add   []lavalink.Track
		takes int
		want  []string
		left  int
	}{
		{
			name:  "empty queue yields nothing",
			takes: 1,
			left:  0,
		},
		{
			name:  "single track",
			add:   []lavalink.Track{track("a")},
			takes: 1,
			want:  []string{"a"},
			left:  0,
		},
		{
			name:  "fifo order",
			add:   []lavalink.Track{track("a"), track("b"), track("c")},
			takes: 3,
			want:  []string{"a", "b", "c"},
			left:  0,
		},
		{
			name:  "partially drained",
			add:   []lavalink.Track{track("a"), track("b"), track("c")},
			takes: 1,
			want:  []string{"a"},
			left:  2,
		},
		{
			name:  "over-drained",
			add:   []lavalink.Track{track("a")},
			takes: 3,
			want:  []string{"a"},
			left:  0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			var q queue.Queue
			q.Add(tt.add...)
			require.Equal(t, len(tt.add), q.Len())

			var got []string
			for range tt.takes {
				next, ok := q.Next()
				if !ok {
					break
				}
				got = append(got, next.Info.Title)
			}

			require.Equal(t, tt.want, got)
			require.Equal(t, tt.left, q.Len())
		})
	}
}

func TestQueueAddVariadic(t *testing.T) {
	t.Parallel()

	var q queue.Queue
	q.Add()
	require.Zero(t, q.Len())

	q.Add(track("a"), track("b"))
	q.Add(track("c"))
	require.Equal(t, 3, q.Len())
}

func TestQueueClear(t *testing.T) {
	t.Parallel()

	var q queue.Queue
	q.Add(track("a"), track("b"))
	q.Clear()

	require.Zero(t, q.Len())
	require.Empty(t, q.Tracks())

	_, ok := q.Next()
	require.False(t, ok)

	q.Clear()
	require.Zero(t, q.Len())
}

func TestQueueTracksIsACopy(t *testing.T) {
	t.Parallel()

	var q queue.Queue
	q.Add(track("a"), track("b"))

	got := q.Tracks()
	require.Equal(t, []string{"a", "b"}, []string{got[0].Info.Title, got[1].Info.Title})

	got[0] = track("hijacked")
	got = append(got, track("appended"))

	require.Equal(t, 2, q.Len())
	inQueue := q.Tracks()
	require.Equal(t, "a", inQueue[0].Info.Title)
	require.Equal(t, "b", inQueue[1].Info.Title)
	require.Len(t, inQueue, 2)
}

// Next advances with q.tracks = q.tracks[1:], so a later Add appends into the
// same backing array. The taken track is a value copy and must not change.
func TestQueueNextIsUnaffectedByLaterAdd(t *testing.T) {
	t.Parallel()

	var q queue.Queue
	q.Add(track("a"), track("b"), track("c"))

	taken, ok := q.Next()
	require.True(t, ok)
	require.Equal(t, "a", taken.Info.Title)

	for i := range 10 {
		q.Add(track("filler"))
		require.Equal(t, "a", taken.Info.Title, "taken track mutated after %d appends", i+1)
	}

	require.Equal(t, "b", q.Tracks()[0].Info.Title)
}

func TestQueueConcurrentAccess(t *testing.T) {
	t.Parallel()

	var q queue.Queue

	const goroutines = 8
	const iterations = 200

	var wg sync.WaitGroup
	for range goroutines {
		wg.Add(4)
		go func() {
			defer wg.Done()
			for range iterations {
				q.Add(track("a"))
			}
		}()
		go func() {
			defer wg.Done()
			for range iterations {
				q.Next()
			}
		}()
		go func() {
			defer wg.Done()
			for range iterations {
				q.Clear()
			}
		}()
		go func() {
			defer wg.Done()
			for range iterations {
				// Results are discarded on purpose: this goroutine exists to
				// make the race detector observe concurrent reads.
				_ = q.Len()
				_ = q.Tracks()
			}
		}()
	}
	wg.Wait()

	require.GreaterOrEqual(t, q.Len(), 0)
}
