package main

import "github.com/bwmarrin/discordgo"

var commands = []*discordgo.ApplicationCommand{
	{
		Name:        "gamenick",
		Description: "Command for setting in-game nickname",
		Options: []*discordgo.ApplicationCommandOption{

			{
				Type:        discordgo.ApplicationCommandOptionString,
				Name:        "nick",
				Description: "In-game nickname",
				Required:    true,
			},
		},
	},
	{
		Name:                     "setrollchannel",
		Description:              "Sets roll channel for this server",
		DefaultMemberPermissions: &serverManagerPerms,
		Options: []*discordgo.ApplicationCommandOption{
			{
				Type:        discordgo.ApplicationCommandOptionChannel,
				Name:        "channel",
				Description: "Channel for rolls",
				Required:    true,
				ChannelTypes: []discordgo.ChannelType{
					discordgo.ChannelTypeGuildText,
				},
			},
		},
	},
	{
		Name:                     "createroll",
		Description:              "Creates new roll for item",
		DefaultMemberPermissions: &serverManagerPerms,
		Options: []*discordgo.ApplicationCommandOption{
			{
				Type:        discordgo.ApplicationCommandOptionInteger,
				Name:        "expiration",
				Required:    false,
				Description: "Expiration of roll",
				Choices: []*discordgo.ApplicationCommandOptionChoice{
					{
						Name:  "1 day",
						Value: 1,
					},
					{
						Name:  "3 days",
						Value: 3,
					},
					{
						Name:  "7 days",
						Value: 7,
					},
				},
			},
		},
	},
	{
		Name:        "feedback",
		Description: "Send your feedback",
	},
	{
		Name:        "roll",
		Description: "Roll a dice",
	},
}
