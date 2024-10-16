package main

import (
	"sync"

	"github.com/bwmarrin/discordgo"
	"github.com/pocketbase/pocketbase/forms"
)

var sendItemRollsMutex sync.Mutex

func sendItemRolls() {
	if !sendItemRollsMutex.TryLock() {
		return
	}
	defer sendItemRollsMutex.Unlock()
	records, err := app.Dao().FindRecordsByFilter("itemRolls", "sent = false && rollStart <= @now", "", 0, 0)
	if err != nil {
		return
	}
	for _, v := range records {

		guildRecord, _ := app.Dao().FindRecordById("guilds", v.GetString("guild"))

		mess, err := discord.ChannelMessageSendComplex(guildRecord.GetString("itemRollChannelId"), &discordgo.MessageSend{
			Embeds: []*discordgo.MessageEmbed{{
				Type:        discordgo.EmbedTypeArticle,
				Title:       v.GetString("itemName"),
				Description: v.GetString("itemDescription"),
			}},
			Components: []discordgo.MessageComponent{
				discordgo.ActionsRow{Components: []discordgo.MessageComponent{
					discordgo.Button{Label: "Roll 🎲", CustomID: "roll_button", Style: discordgo.PrimaryButton},
				}},
			},
		})

		if err != nil {
			println(err.Error())
			return
		}
		form := forms.NewRecordUpsert(app, v)
		form.LoadData(map[string]any{
			"messageId": mess.ID,
			"sent":      true,
		})
		if err := form.Submit(); err != nil {
			discord.ChannelMessageDelete(mess.ChannelID, mess.ID)
		}

	}

}
