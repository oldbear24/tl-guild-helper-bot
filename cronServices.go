package main

import (
	"fmt"
	"sort"
	"strconv"
	"sync"

	"github.com/bwmarrin/discordgo"
	"github.com/pocketbase/dbx"
	"github.com/pocketbase/pocketbase/forms"
	"github.com/pocketbase/pocketbase/tools/types"
)

var sendItemRollsMutex sync.Mutex
var closeItemRollsMutex sync.Mutex

func sendItemRolls() {
	if !sendItemRollsMutex.TryLock() {
		return
	}
	defer sendItemRollsMutex.Unlock()
	records, err := app.Dao().FindRecordsByFilter("itemRolls", "status = 'new' && rollStart <= @now", "", 0, 0)
	if err != nil {
		return
	}
	for _, v := range records {

		guildRecord, _ := app.Dao().FindRecordById("guilds", v.GetString("guild"))

		mess, err := discord.ChannelMessageSendComplex(guildRecord.GetString("itemRollChannelId"), &discordgo.MessageSend{
			Embeds: []*discordgo.MessageEmbed{{
				Type:        discordgo.EmbedTypeArticle,
				Title:       v.GetString("itemName"),
				Description: fmt.Sprintf("%s\n\nEnding  <t:%d:R>", v.GetString("itemDescription"), v.GetDateTime("rollEnd").Time().Unix()),
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
			"status":    "in_progress",
		})
		if err := form.Submit(); err != nil {
			discord.ChannelMessageDelete(mess.ChannelID, mess.ID)
		}

	}

}
func closeItemRolls() {
	if !closeItemRollsMutex.TryLock() {
		return
	}
	defer closeItemRollsMutex.Unlock()

	records, err := app.Dao().FindRecordsByFilter("itemRolls", "status = 'in_progress' && rollEnd <= @now", "", 0, 0)
	if err != nil {
		return
	}
	for _, rollRecord := range records {
		guildRecord, _ := app.Dao().FindRecordById("guilds", rollRecord.GetString("guild"))

		var playerRolls []playerRollRecord
		err = app.Dao().DB().Select("players.userId as userId", "players.nickname as nickname", "itemPlayerRolls.rolledNumber as rolledNumber", "itemPlayerRolls.created as created").From("itemPlayerRolls").InnerJoin("players", dbx.NewExp("players.id = itemPlayerRolls.player")).Where(dbx.NewExp("roll={:roll}", dbx.Params{"roll": rollRecord.Id})).All(&playerRolls)
		if err != nil {
			app.Logger().Error("Sql query error", "error", err)
		}
		if len(playerRolls) == 0 {
			continue
		}
		sort.Slice(playerRolls, func(i, j int) bool {

			if playerRolls[i].Roll == playerRolls[j].Roll && playerRolls[i].Created.Time().Before(playerRolls[j].Created.Time()) {
				return true
			}
			return playerRolls[i].Roll > playerRolls[j].Roll

		})
		var fields []*discordgo.MessageEmbedField
		for i, playerRoll := range playerRolls {
			fields = append(fields, &discordgo.MessageEmbedField{
				Name:  strconv.Itoa(i + 1),
				Value: fmt.Sprintf("<@%s> (%s): %d", playerRoll.DiscordUserId, playerRoll.Nickname, playerRoll.Roll),
			})

		}
		_, err := discord.ChannelMessageSendEmbed(guildRecord.GetString("itemRollChannelId"), &discordgo.MessageEmbed{
			Type:        discordgo.EmbedTypeArticle,
			Title:       "Roll result for " + rollRecord.GetString("itemName"),
			Description: rollRecord.GetString("itemDescription"),
			Fields:      fields,
		})
		if err != nil {
			app.Logger().Error("Cannot send embeded", "error", err)
			continue
		}
		form := forms.NewRecordUpsert(app, rollRecord)
		form.LoadData(map[string]any{
			"status": "ended",
		})
		if err = form.Submit(); err != nil {
			app.Logger().Error("Error saving roll state", "error", err)
		}
	}
}

type playerRollRecord struct {
	DiscordUserId string         `db:"userId"`
	Nickname      string         `db:"nickname"`
	Roll          int            `db:"rolledNumber"`
	Created       types.DateTime `db:"created"`
}
