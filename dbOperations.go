package main

import (
	"github.com/bwmarrin/discordgo"
	"github.com/pocketbase/dbx"
	"github.com/pocketbase/pocketbase/forms"
	"github.com/pocketbase/pocketbase/models"
)

func getOrCreateGuildRecord(discordGuild *discordgo.Guild) (*models.Record, error) {
	guildRecord, err := app.Dao().FindFirstRecordByFilter("guilds", "guild_id={:gId}", dbx.Params{"gId": discordGuild.ID})
	if err != nil {
		collection, _ := app.Dao().FindCollectionByNameOrId("guilds")
		guildRecord = models.NewRecord(collection)

	}
	form := forms.NewRecordUpsert(app, guildRecord)

	form.LoadData(map[string]any{
		"guild_id": discordGuild.ID,
		"name":     discordGuild.Name,
	})
	if err := form.Submit(); err != nil {
		return nil, err
	}
	return guildRecord, err
}
func getOrCreateGuildRecordById(discordGuildId string) (*models.Record, error) {
	guildRecord, err := app.Dao().FindFirstRecordByFilter("guilds", "guild_id={:gId}", dbx.Params{"gId": discordGuildId})
	if err != nil {
		collection, _ := app.Dao().FindCollectionByNameOrId("guilds")
		guildRecord = models.NewRecord(collection)

	}
	form := forms.NewRecordUpsert(app, guildRecord)

	form.LoadData(map[string]any{
		"guild_id": discordGuildId,
	})
	if err := form.Submit(); err != nil {
		return nil, err
	}
	return guildRecord, err
}
func getTargetEventChannel(sourceChannel, guildId string) string {
	targetChannels := []struct {
		Id string `db:"annoucementChannelId"`
	}{}
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
func getOrCreatePlayer(guildId string, discordUser *discordgo.User, data map[string]any) (*models.Record, error) {
	guildRecord, err := getOrCreateGuildRecordById(guildId)
	if err != nil {
		return nil, err
	}
	nickRecord, err := app.Dao().FindFirstRecordByFilter("players", "guild = {:guildId} && userId = {:userId}", dbx.Params{"guildId": guildRecord.Id, "userId": discordUser.ID})
	if err != nil {
		collection, err := app.Dao().FindCollectionByNameOrId("players")
		if err != nil {
			return nil, err
		}
		nickRecord = models.NewRecord(collection)
		data["guild"] = guildRecord.Id
		data["userId"] = discordUser.ID

	}
	data["name"] = discordUser.Username
	form := forms.NewRecordUpsert(app, nickRecord)
	form.LoadData(data)
	if err := form.Submit(); err != nil {
		return nil, err
	}

	return nickRecord, nil
}
