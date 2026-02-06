import type { Track } from "lavalink-client"

export function getYtThumbnailUrl(track: Track) {
	return `https://img.youtube.com/vi/${track.info.identifier}/hqdefault.jpg`
}

// export function formatDuration(d: number) {
// 	const secondsTotal = d / 1000
// 	const minutes = Math.floor(secondsTotal / 60)
// 	const restSeconds = secondsTotal - minutes * 60

// 	return `${minutes}min ${restSeconds}s`
// }

export function formatDuration(ms: number): string {
	const seconds = Math.floor(ms / 1000)
	const minutes = Math.floor(seconds / 60)
	const hours = Math.floor(minutes / 60)

	const secs = seconds % 60
	const mins = minutes % 60

	const time =
		hours > 0
			? `${hours}:${mins.toString().padStart(2, "0")}:${secs.toString().padStart(2, "0")}`
			: `${mins}:${secs.toString().padStart(2, "0")}`

	return `\`${time}\``
}
