package bot

import (
	"github.com/disgoorg/disgolink/v3/lavalink"
	"github.com/disgoorg/snowflake/v2"
)

type Queue struct {
	Tracks []lavalink.Track
}

func (q *Queue) Add(tracks ...lavalink.Track) {
	q.Tracks = append(q.Tracks, tracks...)
}

func (q *Queue) Next() (lavalink.Track, bool) {
	if len(q.Tracks) == 0 {
		return lavalink.Track{}, false
	}
	track := q.Tracks[0]
	q.Tracks = q.Tracks[1:]
	return track, true
}

func (q *Queue) Clear() {
	q.Tracks = nil
}

type QueueManager struct {
	queues map[snowflake.ID]*Queue
}

func (qm *QueueManager) Get(guildID snowflake.ID) *Queue {
	queue, ok := qm.queues[guildID]
	if !ok {
		queue = &Queue{}
		qm.queues[guildID] = queue
	}
	return queue
}

func (qm *QueueManager) Delete(guildID snowflake.ID) {
	delete(qm.queues, guildID)
}
