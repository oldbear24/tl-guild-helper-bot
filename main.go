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
	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/core"
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
		scheduler.MustAdd("send_item_rolls", "*/1 * * * *", sendItemRolls)
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

	discord.AddHandler(func(s *discordgo.Session, i *discordgo.GuildScheduledEventCreate) {
		if i.EntityType == discordgo.GuildScheduledEventEntityTypeExternal {
			return
		}
		guildRecord, err := getOrCreateGuildRecord(i.GuildID)
		if err != nil {
			return
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
	return int(rand.IntN(100))
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
