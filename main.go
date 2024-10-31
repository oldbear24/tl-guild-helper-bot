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
	"github.com/pocketbase/pocketbase/core"
	"github.com/pocketbase/pocketbase/forms"
	"github.com/pocketbase/pocketbase/models"
	"github.com/pocketbase/pocketbase/plugins/migratecmd"
	"github.com/pocketbase/pocketbase/tools/cron"
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

	app.OnBeforeServe().Add(func(e *core.ServeEvent) error {
		scheduler := cron.New()
		scheduler.MustAdd("send_item_rolls", "* * * * *", sendItemRolls)
		scheduler.MustAdd("close_item_rolls", "* * * * *", closeItemRolls)
		scheduler.MustAdd("create_events", "* * * * *", createEvents)
		scheduler.MustAdd("get_guilds_members", "0 * * * *", refreshGuildsMembers)

		scheduler.Start()
		return nil
	})

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

		members, _ := discord.GuildMembers(i.Guild.ID, "", 1000)

		_, err = app.Dao().DB().Update("players", dbx.Params{"active": false}, dbx.NewExp("guild={:guild}", dbx.Params{"guild": guildRecord.Id})).Execute()
		if err != nil {
			return
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

			_, err := getOrCreatePlayer(i.Guild.ID, v.User, map[string]any{"serverNick": nick, "active": true})
			if err != nil {
				return
			}
		}

	})
	discord.AddHandler(func(s *discordgo.Session, r *discordgo.Ready) {
		app.Logger().Info("Discord bot: Bot is up!")
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
		guildRecord, err := getOrCreateGuildRecordById(i.GuildID)
		if err != nil {
			return
		}
		createOrUpdateEventLogRecord(guildRecord, i.ID, i.Name, i.ScheduledStartTime)
	})

	discord.AddHandler(func(s *discordgo.Session, i *discordgo.GuildScheduledEventCreate) {
		if i.EntityType == discordgo.GuildScheduledEventEntityTypeExternal {
			return
		}
		guildRecord, err := getOrCreateGuildRecordById(i.GuildID)
		if err != nil {
			return
		}
		createOrUpdateEventLogRecord(guildRecord, i.ID, i.Name, i.ScheduledStartTime)
		targetChannel := getTargetEventChannel(i.ChannelID, guildRecord.Id)
		if targetChannel == "" {
			return
		}
		_, err = s.ChannelMessageSend(targetChannel, "https://discord.com/events/"+i.GuildID+"/"+i.ID)
		if err != nil {
			app.Logger().Error("Could not sent discord message!", "channel", targetChannel, "guild", i.GuildID, "error", err)
			return
		}
		app.Logger().Info("Sent event info", "Guild", i.GuildID, "event", i, "targetChannelId", targetChannel)

	})

	discord.AddHandler(func(s *discordgo.Session, i *discordgo.GuildScheduledEventUserAdd) {

		el, err := app.Dao().FindFirstRecordByData("eventLogs", "eventId", i.GuildScheduledEventID)
		if err != nil {
			return
		}
		member, err := discord.GuildMember(i.GuildID, i.UserID)
		if err != nil {
			return
		}
		pl, err := getOrCreatePlayer(i.GuildID, member.User, map[string]any{})
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
			"status":   "registered",
		})
		form.Submit()

	})
	discord.AddHandler(func(s *discordgo.Session, i *discordgo.GuildScheduledEventUserRemove) {
		el, err := app.Dao().FindFirstRecordByData("eventLogs", "eventId", i.GuildScheduledEventID)
		if err != nil {
			return
		}
		member, err := discord.GuildMember(i.GuildID, i.UserID)
		if err != nil {
			return
		}
		pl, err := getOrCreatePlayer(i.GuildID, member.User, map[string]any{})
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
			"status":   "unregistered",
		})
		form.Submit()
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

	return rand.IntN(100)
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
	gRecord, _ := getOrCreateGuildRecordById(i.GuildID)
	form := forms.NewRecordUpsert(app, gRecord)
	form.LoadData(map[string]any{
		channelDbName: channelId,
	})

	if err := form.Submit(); err != nil {
		app.Logger().Error("Cannot save guild chanel", "guildId", i.GuildID, "channel_db_name", channelDbName, "channel", channelId, "error", err)
		return err
	}
	app.Logger().Info("Save guild chanel", "guildId", i.GuildID, "channel_db_name", channelDbName, "channel", channelId)
	return nil
}
func createOrUpdateEventLogRecord(guildRecord *models.Record, id, name string, start time.Time) {
	logRecord, err := app.Dao().FindFirstRecordByData("eventLogs", "eventId", id)
	if err != nil {
		collection, _ := app.Dao().FindCollectionByNameOrId("eventLogs")
		logRecord = models.NewRecord(collection)
	}
	form := forms.NewRecordUpsert(app, logRecord)

	form.LoadData(map[string]any{
		"eventName": name,
		"guild":     guildRecord.Id,
		"eventId":   id,
		"start":     start,
	})
	form.Submit()
}
