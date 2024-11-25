package main

import (
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/pocketbase/dbx"
	"github.com/pocketbase/pocketbase/core"
)

var messageComponentHandlers = map[string]func(s *discordgo.Session, i *discordgo.InteractionCreate){
	"roll_button": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		rollRecord, err := app.FindFirstRecordByFilter("itemRolls", "messageId={:message}", dbx.Params{"message": i.Message.ID})
		if err != nil {
			return
		}
		rollEnd := rollRecord.GetDateTime("rollEnd").Time()
		if rollEnd.Compare(time.Now().UTC()) > 0 {
			replyEmpheralInteraction(s, i, "This roll has already ended")
			deleteInteractionWithdelay(s, i, 30)
			return
		}
		rollResult := rollDice()
		player, err := getOrCreatePlayer(app, i.GuildID, i.Member.User, map[string]any{})
		if err != nil {
			return
		}
		collection, err := app.FindCollectionByNameOrId("itemPlayerRolls")
		if err != nil {
			return
		}
		rRecord := core.NewRecord(collection)

		rRecord.Load(map[string]any{
			"roll":         rollRecord.Id,
			"player":       player.Id,
			"rolledNumber": rollResult,
		})
		if err = app.Save(rRecord); err != nil {
			replyEmpheralInteraction(s, i, "You cannot roll again!")
		} else {
			replyEmpheralInteraction(s, i, "Your roll has been saved")
		}
		deleteInteractionWithdelay(s, i, 30)
	},
}
