package bot

import (
	"fmt"
	"log/slog"
	"time"

	"github.com/bwmarrin/discordgo"
)

var commands = []*discordgo.ApplicationCommand{
	{
		Name:        "next",
		Description: "Avisa quando o mapa do servidor trocar.",
	},
}

// registerCommands registers the slash commands in Discord.
func (b *Bot) registerCommands() error {
	slog.Info("Registering commands...")
	for _, cmd := range commands {
		_, err := b.Session.ApplicationCommandCreate(b.Session.State.User.ID, b.config.GuildID, cmd)
		if err != nil {
			return err
		}
	}
	slog.Info("Commands registered")
	return nil
}

// registerHandlers sets up the callback for slash commands and buttons.
func (b *Bot) registerHandlers() {
	b.Session.AddHandler(func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		switch i.Type {
		case discordgo.InteractionApplicationCommand:
			if i.ApplicationCommandData().Name == "next" {
				b.handleNextCommand(s, i)
			}
		case discordgo.InteractionMessageComponent:
			if i.MessageComponentData().CustomID == "notify_me" {
				b.handleNotifyMeButton(s, i)
			}
		}
	})
}

func (b *Bot) handleNextCommand(s *discordgo.Session, i *discordgo.InteractionCreate) {
	userID := i.Member.User.ID
	if i.Member == nil && i.User != nil {
		userID = i.User.ID // Fallback for DMs
	}

	channelID := i.ChannelID

	// Respond immediately with a deferred message (thinking state)
	err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseDeferredChannelMessageWithSource,
	})
	if err != nil {
		slog.Error("Failed to defer interaction", "err", err)
		return
	}

	// The callback that will be called when the map changes
	callback := func(message string) {
		_, err := s.ChannelMessageSend(channelID, message)
		if err != nil {
			slog.Error("Failed to send map change notification", "err", err, "channel", channelID)
		}
	}

	// Start monitor and get initial details
	alreadyRunning, details, err := b.monitor.StartMonitor(b.ctx, userID, callback)
	if err != nil {
		slog.Error("Falha ao iniciar o monitoramento", "user", userID, "err", err)
		errorMsg := "Erro ao consultar o servidor: " + err.Error()
		_, _ = s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
			Content: &errorMsg,
		})
		return
	}

	if alreadyRunning {
		// Respond ephemerally indicating they were added to the list
		msg := "O monitoramento já está rolando! Adicionei você à lista."
		_, _ = s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
			Content: &msg,
		})
		return
	}

	// Calculate ETA and format message
	eta := details.StartTime.Add(1 * time.Hour).Unix()
	msg := fmt.Sprintf("Você será avisado quando trocar de mapa!\n**Mapa atual:** %s\n**Estimativa para acabar em:** <t:%d:R>", details.LayerClassname, eta)

	_, err = s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
		Content: &msg,
		Components: &[]discordgo.MessageComponent{
			discordgo.ActionsRow{
				Components: []discordgo.MessageComponent{
					discordgo.Button{
						Emoji:    &discordgo.ComponentEmoji{Name: "🔔"},
						Label:    "Me avise também!",
						Style:    discordgo.PrimaryButton,
						CustomID: "notify_me",
					},
				},
			},
		},
	})
	if err != nil {
		slog.Error("Failed to edit interaction message", "err", err)
	}
}

func (b *Bot) handleNotifyMeButton(s *discordgo.Session, i *discordgo.InteractionCreate) {
	userID := i.Member.User.ID
	if i.Member == nil && i.User != nil {
		userID = i.User.ID // Fallback for DMs
	}

	err := b.monitor.AddSubscriber(userID)
	msg := "Você foi adicionado à lista! Também te avisarei quando o mapa mudar."
	if err != nil {
		msg = "Erro: " + err.Error()
	}

	err = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: msg,
			Flags:   discordgo.MessageFlagsEphemeral,
		},
	})
	if err != nil {
		slog.Error("Failed to respond to notify_me button", "err", err)
	}
}
