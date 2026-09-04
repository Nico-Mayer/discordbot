package music

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/handler"
	"github.com/disgoorg/disgo/handler/middleware"
	"github.com/disgoorg/disgo/rest"
	"github.com/disgoorg/snowflake/v2"
)

// varDeferred marks an interaction whose acknowledgement was already sent, so
// the error handler edits it instead of trying to create a second message.
const varDeferred = "deferred"

// NewRouter wires the slash commands to the service. Handlers stay thin: read
// options, call the service, send an embed.
func NewRouter(s *Service, logger *slog.Logger) *handler.Mux {
	h := &commands{service: s, logger: logger}

	mux := handler.New()
	mux.Use(logCommand(logger), guardGuild(s.GuildID(), logger))
	mux.Error(handleError(logger))
	mux.NotFound(handleNotFound(logger))

	// Loading a track can outlast Discord's initial response window, so /play
	// acknowledges first and edits that acknowledgement with the result.
	mux.Group(func(r handler.Router) {
		r.Use(middleware.Defer(discord.InteractionTypeApplicationCommand, false, false), markDeferred)
		r.SlashCommand("/play", h.play)
	})

	mux.SlashCommand("/pause", h.pause)
	mux.SlashCommand("/stop", h.stop)
	mux.SlashCommand("/skip", h.skip)
	mux.SlashCommand("/now-playing", h.nowPlaying)
	mux.SlashCommand("/queue", h.queue)

	return mux
}

type commands struct {
	service *Service
	logger  *slog.Logger
}

func (c *commands) play(data discord.SlashCommandInteractionData, e *handler.CommandEvent) error {
	req := PlayRequest{Identifier: data.String(optionPlayName)}
	if state, ok := e.Client().Caches.VoiceState(c.service.GuildID(), e.User().ID); ok {
		req.VoiceChannelID = state.ChannelID
	}

	result, err := c.service.Enqueue(e.Ctx, req)
	if err != nil {
		return err
	}

	embed := trackEmbed(result.Track)
	if result.Queued {
		embed = queuedEmbed(result.Track, result.Position)
	}
	if _, err := e.UpdateInteractionResponse(messageUpdate(embed)); err != nil {
		return fmt.Errorf("edit /play acknowledgement: %w", err)
	}
	return nil
}

func (c *commands) pause(_ discord.SlashCommandInteractionData, e *handler.CommandEvent) error {
	paused, err := c.service.Pause(e.Ctx)
	if err != nil {
		return err
	}
	if paused {
		return reply(e, pausedEmbed())
	}
	return reply(e, resumedEmbed())
}

func (c *commands) stop(_ discord.SlashCommandInteractionData, e *handler.CommandEvent) error {
	if err := c.service.Stop(e.Ctx); err != nil {
		return err
	}
	return reply(e, stoppedEmbed())
}

func (c *commands) skip(_ discord.SlashCommandInteractionData, e *handler.CommandEvent) error {
	if _, err := c.service.Skip(e.Ctx); err != nil {
		return err
	}
	return reply(e, skippedEmbed())
}

func (c *commands) nowPlaying(_ discord.SlashCommandInteractionData, e *handler.CommandEvent) error {
	track, position, err := c.service.NowPlaying()
	if err != nil {
		return err
	}
	return reply(e, nowPlayingEmbed(track, position))
}

func (c *commands) queue(_ discord.SlashCommandInteractionData, e *handler.CommandEvent) error {
	return reply(e, queueEmbed(c.service.Queue()))
}

func reply(e *handler.CommandEvent, embed discord.Embed) error {
	if err := e.CreateMessage(discord.MessageCreate{Embeds: []discord.Embed{embed}}); err != nil {
		return fmt.Errorf("send reply: %w", err)
	}
	return nil
}

func messageUpdate(embed discord.Embed) discord.MessageUpdate {
	embeds := []discord.Embed{embed}
	return discord.MessageUpdate{Embeds: &embeds}
}

// logCommand records which command ran and where.
func logCommand(logger *slog.Logger) handler.Middleware {
	return func(next handler.Handler) handler.Handler {
		return func(e *handler.InteractionEvent) error {
			logger.InfoContext(e.Ctx, "handling command",
				slog.String("command", commandName(e)),
				slog.Any("guild_id", e.GuildID()),
				slog.Any("user_id", e.User().ID),
			)
			return next(e)
		}
	}
}

// markDeferred runs after middleware.Defer, so the flag is only set once the
// acknowledgement actually went out.
var markDeferred handler.Middleware = func(next handler.Handler) handler.Handler {
	return func(e *handler.InteractionEvent) error {
		e.Vars[varDeferred] = "true"
		return next(e)
	}
}

// guardGuild refuses commands from any guild other than the configured one.
// Slash commands are registered per guild, so this only fires if the bot is
// added elsewhere and Discord still routes an interaction here.
func guardGuild(configured snowflake.ID, logger *slog.Logger) handler.Middleware {
	return func(next handler.Handler) handler.Handler {
		return func(e *handler.InteractionEvent) error {
			if err := checkGuild(configured, e.GuildID()); err != nil {
				logger.WarnContext(e.Ctx, "refusing command from an unconfigured guild",
					slog.Any("guild_id", e.GuildID()),
					slog.String("command", commandName(e)),
				)
				return err
			}
			return next(e)
		}
	}
}

// checkGuild reports whether an interaction may be handled at all.
func checkGuild(configured snowflake.ID, actual *snowflake.ID) error {
	if actual == nil || *actual != configured {
		return ErrForeignGuild
	}
	return nil
}

// handleError is the single place a failed command becomes a visible reply, so
// no interaction is ever left unanswered.
func handleError(logger *slog.Logger) handler.ErrorHandler {
	return func(e *handler.InteractionEvent, err error) {
		msg, known := UserMessage(err)
		name := commandName(e)

		if known {
			logger.InfoContext(e.Ctx, "command failed", slog.String("command", name), slog.Any("error", err))
		} else {
			logger.ErrorContext(e.Ctx, "command failed", slog.String("command", name), slog.Any("error", err))
		}

		if replyErr := replyError(e, e.Vars[varDeferred] == "true", msg); replyErr != nil {
			logger.ErrorContext(e.Ctx, "could not report the failure to the caller",
				slog.String("command", name),
				slog.Any("error", replyErr),
			)
		}
	}
}

// errorReplier is the part of an interaction event needed to report a failure.
// *handler.InteractionEvent satisfies it.
type errorReplier interface {
	CreateMessage(messageCreate discord.MessageCreate, opts ...rest.RequestOpt) error
	UpdateInteractionResponse(messageUpdate discord.MessageUpdate, opts ...rest.RequestOpt) (*discord.Message, error)
}

func replyError(r errorReplier, deferred bool, msg string) error {
	embed := errorEmbed(msg)
	if deferred {
		if _, err := r.UpdateInteractionResponse(messageUpdate(embed)); err != nil {
			return fmt.Errorf("edit acknowledgement with the failure: %w", err)
		}
		return nil
	}
	err := r.CreateMessage(discord.MessageCreate{
		Embeds: []discord.Embed{embed},
		Flags:  discord.MessageFlagEphemeral,
	})
	if err != nil {
		return fmt.Errorf("send failure reply: %w", err)
	}
	return nil
}

// handleNotFound logs an unrecognised command without crashing.
func handleNotFound(logger *slog.Logger) handler.NotFoundHandler {
	return func(e *handler.InteractionEvent) error {
		logUnknownCommand(e.Ctx, logger, commandName(e))
		return nil
	}
}

func logUnknownCommand(ctx context.Context, logger *slog.Logger, name string) {
	logger.WarnContext(ctx, "unknown command", slog.String("command", name))
}

func commandName(e *handler.InteractionEvent) string {
	i, ok := e.Interaction.(discord.ApplicationCommandInteraction)
	if !ok {
		return ""
	}
	return i.Data.CommandName()
}
