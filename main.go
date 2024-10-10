package main

import (
	"fmt"
	"log"
	"os"
	"runtime"
	"strings"

	"github.com/bwmarrin/discordgo"
	_ "github.com/oldbear24/tl-guild-helper-bot/migrations"
	"github.com/pocketbase/dbx"
	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/core"
	"github.com/pocketbase/pocketbase/forms"
	"github.com/pocketbase/pocketbase/models"
	"github.com/pocketbase/pocketbase/plugins/migratecmd"
)

var (
	app *pocketbase.PocketBase
)

func main() {
	app = pocketbase.New()
	var botToken string
	app.RootCmd.PersistentFlags().StringVar(&botToken, "token", "", "Bot token")
	app.RootCmd.ParseFlags(os.Args[1:])
	if envToken := os.Getenv("TLGH_BOT_TOKEN"); envToken != "" {
		botToken = envToken
	}
	isGoRun := strings.HasPrefix(os.Args[0], os.TempDir())
	migratecmd.MustRegister(app, app.RootCmd, migratecmd.Config{
		Automigrate: isGoRun,
	})

	discord, _ := discordgo.New("Bot " + botToken)
	app.OnAfterBootstrap().Add(func(e *core.BootstrapEvent) error {

		discord.LogLevel = func() int {
			switch e.App.Settings().Logs.MinLevel {
			case -4:
				return discordgo.LogDebug
			case 0:
				return discordgo.LogInformational
			case 4:
				return discordgo.LogWarning
			case 8:
				return discordgo.LogError
			default:
				return 0

			}

		}()
		if os.Args[1] == "serve" {
			if botToken == "" {
				log.Fatal("You mast pass bot token throught --token flag or TLGH_BOT_TOKEN environment variable")
			}
			err := discord.Open()
			if err != nil {
				log.Fatalf("could not open session: %s", err)
			}
		}
		return nil

	})
	discordgo.Logger = func(msgL, caller int, format string, a ...interface{}) {

		pc, file, line, _ := runtime.Caller(caller)

		files := strings.Split(file, "/")
		file = files[len(files)-1]

		name := runtime.FuncForPC(pc).Name()
		fns := strings.Split(name, ".")
		name = fns[len(fns)-1]
		msg := fmt.Sprintf(format, a...)
		if app.IsBootstrapped() {
			switch msgL {
			case discordgo.LogDebug:
				app.Logger().Debug("Discord bot: "+msg, "file", file, "line", line, "name", name, "type", "discordBot")
			case discordgo.LogInformational:
				app.Logger().Info("Discord bot: "+msg, "file", file, "line", line, "name", name, "type", "discordBot")
			case discordgo.LogWarning:
				app.Logger().Warn("Discord bot: "+msg, "file", file, "line", line, "name", name, "type", "discordBot")
			case discordgo.LogError:
				app.Logger().Error("Discord bot: "+msg, "file", file, "line", line, "name", name, "type", "discordBot")
			}
		} else {
			log.Printf("[DG%d] %s:%d:%s() %s\n", msgL, file, line, name, msg)
		}

	}
	discord.AddHandler(func(s *discordgo.Session, i *discordgo.GuildCreate) {

		_, err := getGuildRecord(i.ID)
		if err != nil {
			_, err := createGuildRecord(i.ID)
			if err != nil {
				app.Logger().Error("Could not create guild record on join", "guild", i.ID)
				return
			} else {
				app.Logger().Info("Created guild record", "guild", i.ID)
			}
		}
	})
	discord.AddHandler(func(s *discordgo.Session, i *discordgo.GuildScheduledEventCreate) {
		if i.EntityType == discordgo.GuildScheduledEventEntityTypeExternal {
			return
		}
		guildRecord, err := getGuildRecord(i.GuildID)
		if err != nil {
			guildRecord, err = createGuildRecord(i.GuildID)
			if err != nil {
				app.Logger().Error("Could get guild info", "eventType", "send guild event", "error", err)
				return
			}
		}
		targetChannel := getTargetEventChannel(i.ChannelID, guildRecord.Id)
		if targetChannel == "" {
			return
		}
		_, err = s.ChannelMessageSend(targetChannel, "https://discord.com/events/"+i.GuildID+"/"+i.ID)
		if err != nil {
			app.Logger().Error("Could not sent discord message!", "channel", targetChannel, "guild", i.GuildID)
			return
		}
		app.Logger().Info("Sent event info", "Guild", i.GuildID, "event", i, "targetChannelId", targetChannel)

	})

	if err := app.Start(); err != nil {
		log.Fatal(err)
	}
	err := discord.Close()
	if err != nil {
		log.Printf("could not close session gracefully: %s", err)
	}
}

// Gets guild record and if it does not exists creates it
func getGuildRecord(discordGuildId string) (*models.Record, error) {
	guildRecord, err := app.Dao().FindFirstRecordByFilter("guilds", "guild_id={:gId}", dbx.Params{"gId": discordGuildId})
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
	guildRecord.Set("guild_id", discordGuildId)
	if err := form.Submit(); err != nil {
		return nil, err
	}
	return guildRecord, nil
}
