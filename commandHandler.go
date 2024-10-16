package main

import (
	"encoding/json"

	"github.com/bwmarrin/discordgo"
	"github.com/google/uuid"
	"github.com/pocketbase/pocketbase/forms"
)

var (
	serverManagerPerms int64 = discordgo.PermissionManageServer

	commandsHandlers = map[string]func(s *discordgo.Session, i *discordgo.InteractionCreate){
		"gamenick": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			optionMap := parseOptions(i.ApplicationCommandData().Options)
			if nick, ok := optionMap["nick"]; ok {
				s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{Type: discordgo.InteractionResponseDeferredChannelMessageWithSource, Data: &discordgo.InteractionResponseData{Flags: discordgo.MessageFlagsEphemeral}})

				_, err := getOrCreatePlayer(i.GuildID, i.Member.User.ID, map[string]any{"nickname": nick.StringValue()})
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
				gRecord, _ := getOrCreateGuildRecord(i.GuildID)
				form := forms.NewRecordUpsert(app, gRecord)
				form.LoadData(map[string]any{
					"itemRollChannelId": rollChannel.ChannelValue(s).ID,
				})

				if err := form.Submit(); err != nil {
					app.Logger().Error("Cannot save item roll chanel", "guildId", i.GuildID, "channel", rollChannel)
					return
				}
				s.FollowupMessageCreate(i.Interaction, true, &discordgo.WebhookParams{Content: "Saved item roll channel", Flags: discordgo.MessageFlagsEphemeral})

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
	}
)
