package main

import (
	"fmt"
	"log"
	"math/rand/v2"
	"os"
	"runtime"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/jellydator/ttlcache/v3"

	_ "github.com/oldbear24/tl-guild-helper-bot/migrations"
	"github.com/pocketbase/dbx"
	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/apis"
	"github.com/pocketbase/pocketbase/core"
	"github.com/pocketbase/pocketbase/plugins/migratecmd"
)

var (
	app        *pocketbase.PocketBase
	discord    *discordgo.Session
	modalCache = ttlcache.New[string, []byte](
		ttlcache.WithTTL[string, []byte](30 * time.Minute),
	)
)

func main() {
	app = pocketbase.New()
	var botToken string
	var disableBot bool

	app.Cron().MustAdd("send_item_rolls", "* * * * *", sendItemRolls)
	app.Cron().MustAdd("close_item_rolls", "* * * * *", closeItemRolls)
	app.Cron().MustAdd("create_events", "* * * * *", createEvents)
	app.Cron().MustAdd("get_guilds_members", "0 * * * *", refreshGuildsMembers)
	app.Cron().MustAdd("auto_delete_old_event_messages", "*/15 * * * *", autoDeleteOldEventMessages)
	app.Cron().MustAdd("notify_event_start", "* * * * *", notifyEventStart)
	app.RootCmd.PersistentFlags().StringVar(&botToken, "token", "", "Bot token")
	go modalCache.Start()
	app.RootCmd.PersistentFlags().BoolVar(&disableBot, "db", false, "Disables bot startup")
	app.RootCmd.ParseFlags(os.Args[1:])
	if envToken := os.Getenv("TLGH_BOT_TOKEN"); envToken != "" {
		botToken = envToken
	}
	isGoRun := strings.HasPrefix(os.Args[0], os.TempDir())
	migratecmd.MustRegister(app, app.RootCmd, migratecmd.Config{
		Automigrate: isGoRun,
	})

	discord, _ = discordgo.New("Bot " + botToken)
	discord.Identify.Intents = discord.Identify.Intents | discordgo.IntentGuildMembers
	app.OnBootstrap().BindFunc(func(e *core.BootstrapEvent) error {
		if err := e.Next(); err != nil {
			return err
		}
		discord.LogLevel = func() int {
			switch e.App.Settings().Logs.MinLevel {
			case -4:
				return discordgo.LogDebug
			case 0:
				return discordgo.LogWarning /*Lots of unimportant logs*/
			case 4:
				return discordgo.LogWarning
			case 8:
				return discordgo.LogError
			default:
				return 0

			}

		}()
		if os.Args[1] == "serve" && !disableBot {
			if botToken == "" {
				log.Fatal("You mast pass bot token throught --token flag or TLGH_BOT_TOKEN environment variable")
			}
			err := discord.Open()
			if err != nil {
				log.Fatalf("could not open session: %s", err)
			}
			discord.ApplicationCommandDelete(discord.State.User.ID, "", "") //TODO: delete only command that are not registered
			for _, comand := range commands {
				_, err := discord.ApplicationCommandCreate(discord.State.User.ID, "", comand)
				if err != nil {
					log.Fatalf("Cannot create slash command: %v", err)
				}
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
		guildRecord, err := getOrCreateGuildRecord(i.Guild)
		if err != nil {
			app.Logger().Error("Could not create guild record on join", "guild", i.ID)
			return
		} else {
			app.Logger().Info("Created guild record", "guild", i.ID)
		}

		updateGuildPlayer(guildRecord)

	})
	// Components are part of interactions, so we register InteractionCreate handler
	discord.AddHandler(func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		switch i.Type {
		case discordgo.InteractionApplicationCommand:
			if h, ok := commandsHandlers[i.ApplicationCommandData().Name]; ok {
				h(s, i)
			}
		case discordgo.InteractionMessageComponent:
			if h, ok := messageComponentHandlers[i.MessageComponentData().CustomID]; ok {
				h(s, i)
			}
		case discordgo.InteractionModalSubmit:
			handleModal(s, i)

		}

	})
	discord.AddHandler(func(s *discordgo.Session, i *discordgo.GuildScheduledEventUpdate) {
		guildRecord, err := getOrCreateGuildRecordById(app, i.GuildID)
		if err != nil {
			return
		}
		createOrUpdateEventLogRecord(guildRecord, i.ID, i.Name, i.Description, i.ScheduledStartTime, i.ChannelID, i.Image)
	})

	discord.AddHandler(func(s *discordgo.Session, i *discordgo.GuildScheduledEventCreate) {
		/*if i.EntityType == discordgo.GuildScheduledEventEntityTypeExternal {
			///	return
		}*/
		guildRecord, err := getOrCreateGuildRecordById(app, i.GuildID)
		if err != nil {
			return
		}
		createOrUpdateEventLogRecord(guildRecord, i.ID, i.Name, i.Description, i.ScheduledStartTime, i.ChannelID, i.Image)

	})

	discord.AddHandler(func(s *discordgo.Session, i *discordgo.GuildScheduledEventUserAdd) {

		registerUserOnEvent(i.GuildScheduledEventID, i.GuildID, i.UserID, "registered")
	})
	discord.AddHandler(func(s *discordgo.Session, i *discordgo.GuildScheduledEventUserRemove) {
		registerUserOnEvent(i.GuildScheduledEventID, i.GuildID, i.UserID, "unregistered")
	})
	discord.AddHandler(func(s *discordgo.Session, i *discordgo.GuildScheduledEventDelete) {
		record, err := app.FindFirstRecordByData("eventLogs", "eventId", i.ID)
		if err != nil {
			return
		}
		err = app.Delete(record)
		if err != nil {
			app.Logger().Error("Could not delete event from log", "error", err)
			return
		}
	})
	app.OnServe().BindFunc(func(se *core.ServeEvent) error {
		// register "GET /hello/{name}" route (allowed for everyone)
		se.Router.GET("/api/tlgh/get-nicknames/{guildId}", func(e *core.RequestEvent) error {
			guildId := e.Request.PathValue("guildId")
			if guildId == "" {
				return e.BadRequestError("Missing guildId", nil)
			}
			type User struct {
				Id       string `db:"userId" json:"id"`
				Nickname string `db:"serverNick" json:"nickname"`
			}
			usersResponse := []User{}
			guildRecord, err := app.FindFirstRecordByData("guilds", "guild_id", guildId)
			if err != nil {
				return e.NotFoundError("Guild could not be found", err)
			}
			e.App.DB().Select("userId", "serverNick").From("players").Where(dbx.NewExp("guild={:guildId} and active=true", dbx.Params{"guildId": guildRecord.Id})).All(&usersResponse)
			return e.JSON(200, usersResponse)
		}).Bind(apis.RequireAuth())
		return se.Next()
	})
	defer func() {
		err := discord.Close() //TODO: Check if bot is running
		if err != nil {
			log.Printf("could not close session gracefully: %s", err)
		}
	}()
	if err := app.Start(); err != nil {
		log.Fatal(err)
	}

}

func parseOptions(options []*discordgo.ApplicationCommandInteractionDataOption) map[string]*discordgo.ApplicationCommandInteractionDataOption {
	optionMap := make(map[string]*discordgo.ApplicationCommandInteractionDataOption, len(options))
	for _, opt := range options {
		optionMap[opt.Name] = opt
	}
	return optionMap
}

func rollDice() int {

	return rand.IntN(99) + 1
}

type newItemRollCacheItem struct {
	ExpirationDays int `json:"expirationDays"`
}

func replyEmpheralInteraction(s *discordgo.Session, i *discordgo.InteractionCreate, text string) {
	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{Type: discordgo.InteractionResponseChannelMessageWithSource, Data: &discordgo.InteractionResponseData{Flags: discordgo.MessageFlagsEphemeral, Content: text}})

}
func deleteInteractionWithdelay(s *discordgo.Session, i *discordgo.InteractionCreate, delaySeconds int64) {
	go func() {
		time.Sleep(time.Second * time.Duration(delaySeconds))
		s.InteractionResponseDelete(i.Interaction)
	}()
}

func setGuildChannel(i *discordgo.InteractionCreate, channelDbName, channelId string) error {
	gRecord, _ := getOrCreateGuildRecordById(app, i.GuildID)
	gRecord.Load(map[string]any{
		channelDbName: channelId,
	})

	if err := app.Save(gRecord); err != nil {
		app.Logger().Error("Cannot save guild chanel", "guildId", i.GuildID, "channel_db_name", channelDbName, "channel", channelId, "error", err)
		return err
	}
	app.Logger().Info("Save guild chanel", "guildId", i.GuildID, "channel_db_name", channelDbName, "channel", channelId)
	return nil
}
func createOrUpdateEventLogRecord(guildRecord *core.Record, id, name, description string, start time.Time, channelId string, imageID string) {
	guildId := guildRecord.GetString("guild_id")
	var cSMess *discordgo.Message
	logRecord, err := app.FindFirstRecordByData("eventLogs", "eventId", id)
	if err != nil {
		collection, _ := app.FindCollectionByNameOrId("eventLogs")
		logRecord = core.NewRecord(collection)
		var targetChannel string
		if channelId != "" {
			targetChannel = getTargetEventChannel(channelId, guildRecord.Id)
		} else {
			targetChannel = guildRecord.GetString("defaultAnnouncemenetChannel")
		}
		if targetChannel != "" {

			mention := ""
			guildMentionRole := guildRecord.GetString("announcemenetRoleId")
			if guildMentionRole != "" {
				mention = fmt.Sprintf("<@&%s>\n", guildMentionRole)
			}

			cSMess, err = discord.ChannelMessageSend(targetChannel, fmt.Sprintf("> # [%s](https://discord.com/events/%s/%s)\n||%s||", name, guildId, id, mention))
			if err != nil {
				app.Logger().Error("Could not sent discord message!", "channel", targetChannel, "guild", guildId, "error", err)
				return
			}

			app.Logger().Info("Sent event info", "Guild", guildId, "record", logRecord, "targetChannelId", targetChannel)
		}
	}
	messageId := ""
	messageChannel := ""
	if cSMess != nil {
		messageId = cSMess.ID
		messageChannel = cSMess.ChannelID
	}
	logRecord.Load(map[string]any{
		"eventName":                    name,
		"guild":                        guildRecord.Id,
		"eventId":                      id,
		"start":                        start,
		"description":                  description,
		"announcementMessageId":        messageId,
		"announcementMessageChannelId": messageChannel,
		"imageId":                      imageID,
	})
	if err := app.Save(logRecord); err != nil {
		app.Logger().Error("Could not save record to eventLog", "record", logRecord, "err", err)
		return
	}
	app.Logger().Info("Updated eventLog record", "record", logRecord)
}
