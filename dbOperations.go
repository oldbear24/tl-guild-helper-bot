package main

import (
	"github.com/bwmarrin/discordgo"
	"github.com/pocketbase/dbx"
	"github.com/pocketbase/pocketbase/core"
)

func getOrCreateGuildRecord(discordGuild *discordgo.Guild) (*core.Record, error) {
	guildRecord, err := app.FindFirstRecordByFilter("guilds", "guild_id={:gId}", dbx.Params{"gId": discordGuild.ID})
	if err != nil {
		collection, _ := app.FindCollectionByNameOrId("guilds")
		guildRecord = core.NewRecord(collection)

	}

	guildRecord.Load(map[string]any{
		"guild_id": discordGuild.ID,
		"name":     discordGuild.Name,
	})
	if err := app.Save(guildRecord); err != nil {
		app.Logger().Error("Could not create/update guild record", "error", err)
		return nil, err
	}
	return guildRecord, err
}
func getOrCreateGuildRecordById(txApp core.App, discordGuildId string) (*core.Record, error) {
	guildRecord, err := txApp.FindFirstRecordByFilter("guilds", "guild_id={:gId}", dbx.Params{"gId": discordGuildId})
	if err != nil {
		collection, _ := txApp.FindCollectionByNameOrId("guilds")
		guildRecord = core.NewRecord(collection)

	}

	guildRecord.Load(map[string]any{
		"guild_id": discordGuildId,
	})
	if err := txApp.Save(guildRecord); err != nil {
		app.Logger().Error("Could not create/update guild record", "error", err)
		return nil, err
	}
	return guildRecord, err
}
func getTargetEventChannel(sourceChannel, guildId string) string {
	if sourceChannel == "" {
		return ""
	}
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
func getOrCreatePlayer(txApp core.App, guildId string, discordUser *discordgo.User, data map[string]any) (*core.Record, error) {
	guildRecord, err := getOrCreateGuildRecordById(txApp, guildId)
	if err != nil {
		app.Logger().Error("Could not create/update player record missing guild record", "error", err)
		return nil, err
	}
	nickRecord, err := txApp.FindFirstRecordByFilter("players", "guild = {:guildId} && userId = {:userId}", dbx.Params{"guildId": guildRecord.Id, "userId": discordUser.ID})
	if err != nil {
		collection, err := txApp.FindCollectionByNameOrId("players")
		if err != nil {
			return nil, err
		}
		nickRecord = core.NewRecord(collection)
		data["guild"] = guildRecord.Id
		data["userId"] = discordUser.ID

	}
	data["name"] = discordUser.Username
	nickRecord.Load(data)
	if err := txApp.Save(nickRecord); err != nil {
		app.Logger().Error("Could not create/update player record", "error", err)
		return nil, err
	}

	return nickRecord, nil
}

func registerUserOnEvent(eventId, guildId, playerId, regType string) {
	el, err := app.FindFirstRecordByData("eventLogs", "eventId", eventId)
	if err != nil {
		return
	}
	member, err := discord.GuildMember(guildId, playerId)
	if err != nil {
		return
	}
	pl, err := getOrCreatePlayer(app, guildId, member.User, map[string]any{})
	if err != nil {
		return
	}
	playerLogRecord, err := app.FindFirstRecordByFilter("eventPlayerLogs", "eventLog={:el} && player={:pl}", dbx.Params{"el": el.Id, "pl": pl.Id})
	if err != nil {
		collection, _ := app.FindCollectionByNameOrId("eventPlayerLogs")
		playerLogRecord = core.NewRecord(collection)
	}

	playerLogRecord.Load(map[string]any{
		"eventLog": el.Id,
		"player":   pl.Id,
		"status":   regType,
	})
	app.Save(playerLogRecord)
}

func updateGuildPlayer(guildRecord *core.Record) {
	members, err := discord.GuildMembers(guildRecord.GetString("guild_id"), "", 1000)
	if err != nil {
		app.Logger().Error("Cannot get info about guild members", "error", err)
	}
	app.RunInTransaction(func(txApp core.App) error {

		_, err := txApp.DB().Update("players", dbx.Params{"active": false}, dbx.NewExp("guild={:guild}", dbx.Params{"guild": guildRecord.Id})).Execute()
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

			_, err := getOrCreatePlayer(txApp, guildRecord.GetString("guild_id"), v.User, map[string]any{"serverNick": nick, "active": "true"})
			if err != nil {
				return err
			}
		}
		return nil
	})

}
