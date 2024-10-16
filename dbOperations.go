package main

import (
	"github.com/pocketbase/dbx"
	"github.com/pocketbase/pocketbase/forms"
	"github.com/pocketbase/pocketbase/models"
)

// Gets guild record and if it does not exists creates it
func getGuildRecord(discordGuildId string) (*models.Record, error) {
	guildRecord, err := app.Dao().FindFirstRecordByFilter("guilds", "guild_id={:gId}", dbx.Params{"gId": discordGuildId})
	return guildRecord, err
}
func getOrCreateGuildRecord(discordGuildId string) (*models.Record, error) {
	guildRecord, err := getGuildRecord(discordGuildId)
	if err != nil {
		guildRecord, err = createGuildRecord(discordGuildId)
		if err != nil {
			app.Logger().Error("Could get guild info", "eventType", "send guild event", "error", err)
			return nil, err
		}
	}
	return guildRecord, err
}
func getTargetEventChannel(sourceChannel, guildId string) string {
	targetChannels := []struct {
		Id string `db:"annoucementChannelId"`
	}{}
	app.DB()
	err := app.DB().Select("annoucementChannelId").From("eventAnnouncementsConfig").Where(dbx.NewExp("guild = {:gId}", dbx.Params{"gId": guildId})).AndWhere(dbx.NewExp("channelId = {:channel}", dbx.Params{"channel": sourceChannel})).Limit(1).All(&targetChannels) //app.FindFirstRecordByFilter("eventAnnouncementsConfig", "guild = {:gId} && channelId = {:channel}", dbx.Params{"gId": guildId, "channel": sourceChannel})
	if err != nil {
		app.Logger().Debug("Could not find event target channel", "error", err)
		return ""
	}
	if len(targetChannels) > 0 {
		return targetChannels[0].Id
	} else {
		return ""
	}

}
func createGuildRecord(discordGuildId string) (*models.Record, error) {
	collection, err := app.Dao().FindCollectionByNameOrId("guilds")
	if err != nil {
		return nil, err
	}
	guildRecord := models.NewRecord(collection)
	form := forms.NewRecordUpsert(app, guildRecord)

	form.LoadData(map[string]any{
		"guild_id": discordGuildId,
	})
	if err := form.Submit(); err != nil {
		return nil, err
	}
	return guildRecord, nil
}
func getOrCreatePlayer(guildId, discordUserId string, data map[string]any) (*models.Record, error) {
	guildRecord, err := getOrCreateGuildRecord(guildId)
	if err != nil {
		return nil, err
	}
	nickRecord, err := app.Dao().FindFirstRecordByFilter("players", "guild = {:guildId} && userId = {:userId}", dbx.Params{"guildId": guildRecord.Id, "userId": discordUserId})
	if err != nil {
		collection, err := app.Dao().FindCollectionByNameOrId("players")
		if err != nil {
			return nil, err
		}
		nickRecord = models.NewRecord(collection)
		data["guild"] = guildRecord.Id
		data["userId"] = discordUserId

	}
	form := forms.NewRecordUpsert(app, nickRecord)

	form.LoadData(data)
	if err := form.Submit(); err != nil {
		return nil, err
	}

	return nickRecord, nil
}
