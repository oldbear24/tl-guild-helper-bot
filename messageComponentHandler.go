package main

import (
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/pocketbase/dbx"
	"github.com/pocketbase/pocketbase/forms"
	"github.com/pocketbase/pocketbase/models"
)

var messageComponentHandlers = map[string]func(s *discordgo.Session, i *discordgo.InteractionCreate){
	"roll_button": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		rollRecord, err := app.Dao().FindFirstRecordByFilter("itemRolls", "messageId={:message}", dbx.Params{"message": i.Message.ID})
		if err != nil {
			return
		}
		rollEnd := rollRecord.GetTime("rollEnd")
		if rollEnd.Compare(time.Now().UTC()) > 0 {
			replyEmpheralInteraction(s, i, "This roll has already ended")
			deleteInteractionWithdelay(s, i, 30)
			return
		}
		rollResult := rollDice()
		player, err := getOrCreatePlayer(i.GuildID, i.Member.User.ID, map[string]any{})
		if err != nil {
			return
		}
		collection, err := app.Dao().FindCollectionByNameOrId("itemPlayerRolls")
		if err != nil {
			return
		}
		rRecord := models.NewRecord(collection)

		form := forms.NewRecordUpsert(app, rRecord)
		form.LoadData(map[string]any{
			"roll":       rollRecord.Id,
			"player":     player.Id,
			"rollNumber": rollResult,
		})
		if err = form.Submit(); err != nil {
			replyEmpheralInteraction(s, i, "You cannot roll again!")
		} else {
			replyEmpheralInteraction(s, i, "Your roll has been saved")
		}
		deleteInteractionWithdelay(s, i, 30)
	},
}
