package main

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/google/uuid"
)

var (
	serverManagerPerms int64 = discordgo.PermissionManageServer

	commandsHandlers = map[string]func(s *discordgo.Session, i *discordgo.InteractionCreate){
		"gamenick": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			optionMap := parseOptions(i.ApplicationCommandData().Options)
			if nick, ok := optionMap["nick"]; ok {
				s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{Type: discordgo.InteractionResponseDeferredChannelMessageWithSource, Data: &discordgo.InteractionResponseData{Flags: discordgo.MessageFlagsEphemeral}})

				_, err := getOrCreatePlayer(app, i.GuildID, i.Member.User, map[string]any{"nickname": nick.StringValue()})
				if err != nil {
					app.Logger().Error("Could not save nickname", "error", err)
					s.FollowupMessageCreate(i.Interaction, true, &discordgo.WebhookParams{
						Flags: discordgo.MessageFlagsEphemeral,
						Embeds: []*discordgo.MessageEmbed{
							{
								Color: 16711680, /*Red*/
								Title: "Could not save your nickname",
							},
						},
					})
				}
				s.FollowupMessageCreate(i.Interaction, true, &discordgo.WebhookParams{
					Flags: discordgo.MessageFlagsEphemeral,
					Embeds: []*discordgo.MessageEmbed{
						{
							Color:  65280, /*Green*/
							Title:  "Your nickname has been saved.",
							Author: &discordgo.MessageEmbedAuthor{},
							Fields: []*discordgo.MessageEmbedField{
								{Name: "Nickname", Value: nick.StringValue()},
							},
						},
					},
				})
			}

		},
		"setrollchannel": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			optionMap := parseOptions(i.ApplicationCommandData().Options)

			if rollChannel, ok := optionMap["channel"]; ok {
				s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{Type: discordgo.InteractionResponseDeferredChannelMessageWithSource, Data: &discordgo.InteractionResponseData{Flags: discordgo.MessageFlagsEphemeral}})

				err := setGuildChannel(i, "itemRollChannelId", rollChannel.ChannelValue(s).ID)
				if err != nil {
					replyEmpheralInteraction(s, i, "Could not save item roll channel")
				}
				replyEmpheralInteraction(s, i, "Saved item roll channel")

			}

		},
		"createroll": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			modalId := "create_roll_modal_" + uuid.NewString()
			optionMap := parseOptions(i.ApplicationCommandData().Options)

			var expiration int = 3

			if expOption, ok := optionMap["expiration"]; ok {
				expiration = int(expOption.IntValue())
			}

			cacheData, _ := json.Marshal(newItemRollCacheItem{ExpirationDays: expiration})
			modalCache.Set(modalId, cacheData, 0)
			err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseModal,
				Data: &discordgo.InteractionResponseData{
					CustomID: modalId,
					Title:    "Create item roll",
					Components: []discordgo.MessageComponent{
						discordgo.ActionsRow{
							Components: []discordgo.MessageComponent{
								discordgo.TextInput{CustomID: "name", Required: true, Style: discordgo.TextInputShort, Label: "Item name"},
							},
						},
						discordgo.ActionsRow{
							Components: []discordgo.MessageComponent{
								discordgo.TextInput{CustomID: "description", Required: true, Style: discordgo.TextInputParagraph, Label: "Item description"},
							},
						},
					},
				},
			})

			if err != nil {
				app.Logger().Error("Could not create modal", "error", err)
			}
		},
		"feedback": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			modalId := "feedback_modal"
			err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseModal,
				Data: &discordgo.InteractionResponseData{
					CustomID: modalId,
					Title:    "Send feedback",
					Components: []discordgo.MessageComponent{
						discordgo.ActionsRow{
							Components: []discordgo.MessageComponent{
								discordgo.TextInput{CustomID: "feedback_text", Required: true, Style: discordgo.TextInputParagraph, Label: "Your feedback", Placeholder: "What are your thoughts..."},
							},
						},
					},
				},
			})

			if err != nil {
				app.Logger().Error("Could not create modal", "error", err)
			}
		},

		"roll": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			number := rollDice()
			responseText := fmt.Sprintf("> <@%s> dice result: %d", i.Member.User.ID, number)
			s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{Type: discordgo.InteractionResponseChannelMessageWithSource, Data: &discordgo.InteractionResponseData{
				Content: responseText,
			}})
		},
		"ss": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{Type: discordgo.InteractionResponseChannelMessageWithSource, Data: &discordgo.InteractionResponseData{
				Content: "Not implemented",
			}})
		},
		"dkp-export": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{Type: discordgo.InteractionResponseDeferredChannelMessageWithSource, Data: &discordgo.InteractionResponseData{Flags: discordgo.MessageFlagsEphemeral}})

			guild, err := s.State.Guild(i.GuildID)

			if err != nil {
				replyEmpheralInteraction(s, i, "Failed to processed your request")
				return
			}
			csvContent := ""

			for _, v := range guild.VoiceStates {
				if v.ChannelID == i.ChannelID {
					csvContent += fmt.Sprintf("%s\n", v.UserID)
				}
			}
			channel, err := s.UserChannelCreate(i.Member.User.ID)
			if err != nil {
				app.Logger().Error("Failed to create export for dkp", "guildId", i.GuildID, "member", i.Member.User.ID, "error", err)
				return
			}
			csvContent = strings.Trim(csvContent, "\n")
			reader := strings.NewReader(csvContent)
			fileNamePart := time.Now().UTC().Format("20060102150405")

			s.ChannelMessageSendComplex(channel.ID, &discordgo.MessageSend{
				Content: fmt.Sprintf("DKP Export %s", fileNamePart),
				Files: []*discordgo.File{
					{
						Name:        fmt.Sprintf("%s_dkp_export.csv", fileNamePart),
						ContentType: "text/csv",
						Reader:      reader,
					},
				}})
			replyEmpheralInteraction(s, i, "File was sent to privare message")
			deleteInteractionWithdelay(s, i, 30)
			app.Logger().Info("Created dkp-export", "guildId", i.GuildID, "member", i.Member.User.ID, "data", csvContent)
		},
	}
)
