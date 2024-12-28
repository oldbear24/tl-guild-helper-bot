package main

import (
	"encoding/base64"
	"fmt"
	"io"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/adhocore/gronx"
	"github.com/bwmarrin/discordgo"
	"github.com/pocketbase/dbx"
	"github.com/pocketbase/pocketbase/tools/types"
)

var sendItemRollsMutex sync.Mutex
var closeItemRollsMutex sync.Mutex
var createEventsMutex sync.Mutex

func sendItemRolls() {
	if !sendItemRollsMutex.TryLock() {
		return
	}
	defer sendItemRollsMutex.Unlock()
	records, err := app.FindRecordsByFilter("itemRolls", "status = 'new' && rollStart <= @now", "", 0, 0)
	if err != nil {
		return
	}
	for _, v := range records {

		guildRecord, _ := app.FindRecordById("guilds", v.GetString("guild"))

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
		v.Load(map[string]any{
			"messageId": mess.ID,
			"status":    "in_progress",
		})
		if err := app.Save(v); err != nil {
			discord.ChannelMessageDelete(mess.ChannelID, mess.ID)
		}

	}

}
func closeItemRolls() {
	if !closeItemRollsMutex.TryLock() {
		return
	}
	defer closeItemRollsMutex.Unlock()

	records, err := app.FindRecordsByFilter("itemRolls", "status = 'in_progress' && rollEnd <= @now", "", 0, 0)
	if err != nil {
		return
	}
	for _, rollRecord := range records {
		guildRecord, _ := app.FindRecordById("guilds", rollRecord.GetString("guild"))

		var playerRolls []playerRollRecord
		err = app.DB().Select("players.userId as userId", "players.nickname as nickname", "itemPlayerRolls.rolledNumber as rolledNumber", "itemPlayerRolls.created as created").From("itemPlayerRolls").InnerJoin("players", dbx.NewExp("players.id = itemPlayerRolls.player")).Where(dbx.NewExp("roll={:roll}", dbx.Params{"roll": rollRecord.Id})).All(&playerRolls)
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
		rollRecord.Load(map[string]any{
			"status": "ended",
		})
		if err = app.Save(rollRecord); err != nil {
			app.Logger().Error("Error saving roll state", "error", err)
		}
	}
}
func createEvents() {
	if !createEventsMutex.TryLock() {
		return
	}
	defer createEventsMutex.Unlock()
	records, err := app.FindRecordsByFilter("plannedEvents", "enabled = true && lastPlanned<=@now", "", 0, 0)
	if err != nil {
		app.Logger().Error("Could not retrieved planned events", "err", err)
	}
	for _, rec := range records {
		guildRecord, _ := app.FindRecordById("guilds", rec.GetString("guild"))
		eventDate, err := gronx.NextTickAfter(rec.GetString("startExp"), time.Now().UTC().Add(time.Minute*5), true)
		if err != nil {
			app.Logger().Error("Could not parse cron expresion from plannedEvent record", "exp", rec.GetString("startExp"), "record", rec, "error", err)
		}
		imageName := rec.GetString("image")
		imageString := ""
		if imageName != "" {
			path := filepath.Join(rec.BaseFilesPath(), imageName)

			ext := strings.ReplaceAll(filepath.Ext(path), ".", "")
			// initialize the filesystem
			fsys, err := app.NewFilesystem()
			if err != nil {
				app.Logger().Error("Error while opening fs", "path", path, "error", err)
				return
			}
			defer fsys.Close()
			r, err := fsys.GetFile(path)
			if err != nil {
				app.Logger().Error("Error while opening file", "path", path, "error", err)
				return
			}
			defer r.Close()

			data, _ := io.ReadAll(r)
			imageData := base64.StdEncoding.EncodeToString(data)

			imageString = fmt.Sprintf("data:image/%s;base64,%s", ext, imageData)
		}
		event, err := discord.GuildScheduledEventCreate(guildRecord.GetString("guild_id"), &discordgo.GuildScheduledEventParams{
			ChannelID:          rec.GetString("channel"),
			Name:               rec.GetString("name"),
			Description:        rec.GetString("description"),
			ScheduledStartTime: &eventDate,
			EntityType:         discordgo.GuildScheduledEventEntityTypeVoice,
			PrivacyLevel:       discordgo.GuildScheduledEventPrivacyLevelGuildOnly,
			Image:              imageString,
		})
		if err != nil {
			app.Logger().Error("Cannot create guild event", "plannedEvent", rec, "error", err)
			return
		}
		rec.Set("lastPlanned", eventDate)
		if err := app.Save(rec); err != nil {
			app.Logger().Error("Could not save schedule record", "record", rec, "error", err)
		}

		app.Logger().Info("Scheduled guild event", "event", event)
	}

}

/*
	func proccesEventVote() {
		guildRecord, _ := app.Dao().FindRecordById("guilds", conf.GetString("guild"))
		nextVotingTime := cronexpr.MustParse(conf.GetString("eventVotingEndTime")).Next(time.Now()).UTC()
		event, err := discord.GuildScheduledEventCreate(guildRecord.GetString("guild_id"), &discordgo.GuildScheduledEventParams{
			ChannelID:          conf.GetString("eventChannelId"),
			Name:               conf.GetString("name"),
			Description:        conf.GetString("description"),
			ScheduledStartTime: &nextTime,
			EntityType:         discordgo.GuildScheduledEventEntityTypeVoice,
			PrivacyLevel:       discordgo.GuildScheduledEventPrivacyLevelGuildOnly,
		})
		if err != nil {
			app.Logger().Error("Cannot create guild event", "event_config", conf.Id, "error", err)
			return
		}
	}
*/

func refreshGuildsMembers() {
	guilds, _ := app.FindRecordsByFilter("guilds", "", "", 0, 0)
	for _, guild := range guilds {
		updateGuildPlayer(guild)
	}
}

func autoDeleteOldEventMessages() {
	timeToDelete := time.Now().UTC().Add(time.Hour * 2)
	recordsToDelete, err := app.FindRecordsByFilter("eventLogs", "start< {:date} && deleted = false", "", 0, 0, dbx.Params{"date": timeToDelete})
	if err != nil {
		app.Logger().Error("Searching for eventLogs recod that needs to be deleted failed!", "error", err)
		return
	}
	for _, record := range recordsToDelete {
		messageId := record.GetString("announcementMessageId")
		messageChannelId := record.GetString("announcementMessageChannelId")
		if messageId != "" && messageChannelId != "" {
			err := discord.ChannelMessageDelete(messageChannelId, messageId, discordgo.WithAuditLogReason("Auto event delete"))
			if err != nil {
				app.Logger().Error("Could not delete event message", "event", record, "error", err, "deleteDate", timeToDelete)
			} else {
				app.Logger().Info("Sucesfully delete message for event announcement", "event", record, "deleteDate", timeToDelete)
			}
		} else {
			app.Logger().Warn("Could not delete event message because some parameters are missing", "event", record, "deleteDate", timeToDelete)
		}
		record.Set("deleted", true)
		app.Save(record)

	}
}

type playerRollRecord struct {
	DiscordUserId string         `db:"userId"`
	Nickname      string         `db:"nickname"`
	Roll          int            `db:"rolledNumber"`
	Created       types.DateTime `db:"created"`
}
