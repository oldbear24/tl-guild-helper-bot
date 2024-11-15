package main

import (
	"encoding/json"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/pocketbase/pocketbase/forms"
	"github.com/pocketbase/pocketbase/models"
)

func handleModal(s *discordgo.Session, i *discordgo.InteractionCreate) {
	modalId := i.ModalSubmitData().CustomID
	if strings.HasPrefix(modalId, "create_roll_modal_") {
		handleCreateRollModal(modalId, s, i)
	}
	if h, ok := modalsHandlers[modalId]; ok {
		h(s, i)
	}

}

func handleCreateRollModal(modalId string, s *discordgo.Session, i *discordgo.InteractionCreate) {
	//s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{Type: discordgo.InteractionResponseDeferredChannelMessageWithSource, Data: &discordgo.InteractionResponseData{Flags: discordgo.MessageFlagsEphemeral}})
	ci := modalCache.Get(modalId)
	if ci == nil {
		app.Logger().Warn("Modal form has expired", "modalId", modalId)
		return
	}

	modalCache.Delete(ci.Key())
	var newRollItem newItemRollCacheItem
	json.Unmarshal(ci.Value(), &newRollItem)
	guildRecord, err := getOrCreateGuildRecordById(app.Dao(), i.GuildID)
	if err != nil {
		return
	}
	collection, err := app.Dao().FindCollectionByNameOrId("itemRolls")
	if err != nil {
		return
	}

	data := i.ModalSubmitData()
	newRollRecord := models.NewRecord(collection)
	form := forms.NewRecordUpsert(app, newRollRecord)
	form.LoadData(map[string]any{
		"guild":           guildRecord.Id,
		"itemName":        data.Components[0].(*discordgo.ActionsRow).Components[0].(*discordgo.TextInput).Value,
		"itemDescription": data.Components[1].(*discordgo.ActionsRow).Components[0].(*discordgo.TextInput).Value,
		"rollStart":       time.Now().UTC(),
		"rollEnd":         time.Now().UTC().AddDate(0, 0, newRollItem.ExpirationDays),
		"status":          "new",
	})
	if err := form.Submit(); err != nil {
		app.Logger().Error("Cannot create roll record", "error", err)
		return
	}
	//	err = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{Type: discordgo.InteractionResponseChannelMessageWithSource, Data: &discordgo.InteractionResponseData{Content: "Roll has been created."}})
	/*	if err != nil {
		return
	}*/
	replyEmpheralInteraction(s, i, "Saved item roll")
	deleteInteractionWithdelay(s, i, 30)
	/*go func() {
		time.Sleep(time.Second * 30)
		s.InteractionResponseDelete(i.Interaction)
	}()*/

}

var modalsHandlers = map[string]func(s *discordgo.Session, i *discordgo.InteractionCreate){
	"feedback_modal": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		guildRecord, err := getOrCreateGuildRecordById(app.Dao(), i.GuildID)
		if err != nil {
			return
		}
		collection, err := app.Dao().FindCollectionByNameOrId("feedback")
		if err != nil {
			return
		}

		data := i.ModalSubmitData()
		newRollRecord := models.NewRecord(collection)
		form := forms.NewRecordUpsert(app, newRollRecord)
		form.LoadData(map[string]any{
			"guild":           guildRecord.Id,
			"feedbackMessage": data.Components[0].(*discordgo.ActionsRow).Components[0].(*discordgo.TextInput).Value,
		})
		if err := form.Submit(); err != nil {
			app.Logger().Error("Cannot create feedback record", "error", err)
			replyEmpheralInteraction(s, i, "> There was error while processing your request :(")
			return
		}
		app.Logger().Info("Someone send feedback message", "guild", guildRecord.Id)
		replyEmpheralInteraction(s, i, "> Saved your feedback ;)")
	},
}
