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
			s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{Type: discordgo.InteractionResponseChannelMessageWithSource, Data: &discordgo.InteractionResponseData{Flags: discordgo.MessageFlagsEphemeral, Content: "You cannot roll again!"}})
		} else {
			s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{Type: discordgo.InteractionResponseChannelMessageWithSource, Data: &discordgo.InteractionResponseData{Flags: discordgo.MessageFlagsEphemeral, Content: "Your roll has been saved"}})
		}
		go func() {
			time.Sleep(time.Second * 30)
			s.InteractionResponseDelete(i.Interaction)
		}()
	},
}
