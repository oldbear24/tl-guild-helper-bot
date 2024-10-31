package main

import (
	"github.com/bwmarrin/discordgo"
	"github.com/pocketbase/dbx"
	"github.com/pocketbase/pocketbase/daos"
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
		app.Logger().Error("Could not create/update guild record", "error", err)
		return nil, err
	}
	return guildRecord, err
}
func getOrCreateGuildRecordById(txDao *daos.Dao, discordGuildId string) (*models.Record, error) {
	guildRecord, err := txDao.FindFirstRecordByFilter("guilds", "guild_id={:gId}", dbx.Params{"gId": discordGuildId})
	if err != nil {
		collection, _ := txDao.FindCollectionByNameOrId("guilds")
		guildRecord = models.NewRecord(collection)

	}
	form := forms.NewRecordUpsert(app, guildRecord)
	form.SetDao(txDao)
	form.LoadData(map[string]any{
		"guild_id": discordGuildId,
	})
	if err := form.Submit(); err != nil {
		app.Logger().Error("Could not create/update guild record", "error", err)
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
func getOrCreatePlayer(txDao *daos.Dao, guildId string, discordUser *discordgo.User, data map[string]any) (*models.Record, error) {
	guildRecord, err := getOrCreateGuildRecordById(txDao, guildId)
	if err != nil {
		app.Logger().Error("Could not create/update player record missing guild record", "error", err)
		return nil, err
	}
	nickRecord, err := txDao.FindFirstRecordByFilter("players", "guild = {:guildId} && userId = {:userId}", dbx.Params{"guildId": guildRecord.Id, "userId": discordUser.ID})
	if err != nil {
		collection, err := txDao.FindCollectionByNameOrId("players")
		if err != nil {
			return nil, err
		}
		nickRecord = models.NewRecord(collection)
		data["guild"] = guildRecord.Id
		data["userId"] = discordUser.ID

	}
	data["name"] = discordUser.Username
	form := forms.NewRecordUpsert(app, nickRecord)
	form.SetDao(txDao)
	form.LoadData(data)
	if err := form.Submit(); err != nil {
		app.Logger().Error("Could not create/update player record", "error", err)
		return nil, err
	}

	return nickRecord, nil
}

func registerUserOnEvent(eventId, guildId, playerId, regType string) {
	el, err := app.Dao().FindFirstRecordByData("eventLogs", "eventId", eventId)
	if err != nil {
		return
	}
	member, err := discord.GuildMember(guildId, playerId)
	if err != nil {
		return
	}
	pl, err := getOrCreatePlayer(app.Dao(), guildId, member.User, map[string]any{})
	if err != nil {
		return
	}
	playerLogRecord, err := app.Dao().FindFirstRecordByFilter("eventPlayerLogs", "eventLog={:el} && player={:pl}", dbx.Params{"el": el.Id, "pl": pl.Id})
	if err != nil {
		collection, _ := app.Dao().FindCollectionByNameOrId("eventPlayerLogs")
		playerLogRecord = models.NewRecord(collection)
	}

	form := forms.NewRecordUpsert(app, playerLogRecord)

	form.LoadData(map[string]any{
		"eventLog": el.Id,
		"player":   pl.Id,
		"status":   regType,
	})
	form.Submit()
}

func updateGuildPlayer(guildRecord *models.Record) {
	members, err := discord.GuildMembers(guildRecord.GetString("guild_id"), "", 1000)
	if err != nil {
		app.Logger().Error("Cannot get info about guild members", "error", err)
	}
	app.Dao().RunInTransaction(func(txDao *daos.Dao) error {

		_, err := txDao.DB().Update("players", dbx.Params{"active": false}, dbx.NewExp("guild={:guild}", dbx.Params{"guild": guildRecord.Id})).Execute()
		if err != nil {
			app.Logger().Error("Could not se active false on players", "error", err)
			return err
		}
		for _, v := range members {
			if v.User.Bot {
				continue
			}
			nick := ""
			if v.Nick == "" {
				nick = v.User.GlobalName
			} else {
				nick = v.Nick
			}

			_, err := getOrCreatePlayer(txDao, guildRecord.GetString("guild_id"), v.User, map[string]any{"serverNick": nick, "active": "true"})
			if err != nil {
				return err
			}
		}
		return nil
	})

}
