package utils

import "github.com/disgoorg/disgo/discord"

func GetAvatarUrl(user discord.User) string {
	if user.AvatarURL() == nil {
		return user.DefaultAvatarURL()
	}

	return *user.AvatarURL()
}
