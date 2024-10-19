package migrations

import (
	"encoding/json"

	"github.com/pocketbase/dbx"
	"github.com/pocketbase/pocketbase/daos"
	m "github.com/pocketbase/pocketbase/migrations"
	"github.com/pocketbase/pocketbase/models"
)

func init() {
	m.Register(func(db dbx.Builder) error {
		jsonData := `[
			{
				"id": "_pb_users_auth_",
				"created": "2024-10-09 18:46:25.135Z",
				"updated": "2024-10-18 23:03:44.148Z",
				"name": "users",
				"type": "auth",
				"system": false,
				"schema": [
					{
						"system": false,
						"id": "users_name",
						"name": "name",
						"type": "text",
						"required": false,
						"presentable": false,
						"unique": false,
						"options": {
							"min": null,
							"max": null,
							"pattern": ""
						}
					},
					{
						"system": false,
						"id": "users_avatar",
						"name": "avatar",
						"type": "file",
						"required": false,
						"presentable": false,
						"unique": false,
						"options": {
							"mimeTypes": [
								"image/jpeg",
								"image/png",
								"image/svg+xml",
								"image/gif",
								"image/webp"
							],
							"thumbs": null,
							"maxSelect": 1,
							"maxSize": 5242880,
							"protected": false
						}
					}
				],
				"indexes": [],
				"listRule": "id = @request.auth.id",
				"viewRule": "id = @request.auth.id",
				"createRule": "",
				"updateRule": "id = @request.auth.id",
				"deleteRule": "id = @request.auth.id",
				"options": {
					"allowEmailAuth": true,
					"allowOAuth2Auth": true,
					"allowUsernameAuth": true,
					"exceptEmailDomains": null,
					"manageRule": null,
					"minPasswordLength": 8,
					"onlyEmailDomains": null,
					"onlyVerified": false,
					"requireEmail": false
				}
			},
			{
				"id": "493kl7wl5aa5qvw",
				"created": "2024-10-09 18:47:24.141Z",
				"updated": "2024-10-19 20:06:25.932Z",
				"name": "guilds",
				"type": "base",
				"system": false,
				"schema": [
					{
						"system": false,
						"id": "dpa7td81",
						"name": "guild_id",
						"type": "text",
						"required": true,
						"presentable": false,
						"unique": false,
						"options": {
							"min": null,
							"max": null,
							"pattern": ""
						}
					},
					{
						"system": false,
						"id": "we0kuw0l",
						"name": "itemRollChannelId",
						"type": "text",
						"required": false,
						"presentable": false,
						"unique": false,
						"options": {
							"min": null,
							"max": null,
							"pattern": ""
						}
					},
					{
						"system": false,
						"id": "dtihoomv",
						"name": "eventChannelId",
						"type": "text",
						"required": false,
						"presentable": false,
						"unique": false,
						"options": {
							"min": null,
							"max": null,
							"pattern": ""
						}
					}
				],
				"indexes": [
					"CREATE UNIQUE INDEX ` + "`" + `idx_SyoDMiH` + "`" + ` ON ` + "`" + `guilds` + "`" + ` (` + "`" + `guild_id` + "`" + `)"
				],
				"listRule": null,
				"viewRule": null,
				"createRule": null,
				"updateRule": null,
				"deleteRule": null,
				"options": {}
			},
			{
				"id": "db4psg6zrmmssv2",
				"created": "2024-10-09 18:48:10.760Z",
				"updated": "2024-10-18 23:03:44.149Z",
				"name": "eventAnnouncementsConfig",
				"type": "base",
				"system": false,
				"schema": [
					{
						"system": false,
						"id": "pusa19fv",
						"name": "guild",
						"type": "relation",
						"required": true,
						"presentable": false,
						"unique": false,
						"options": {
							"collectionId": "493kl7wl5aa5qvw",
							"cascadeDelete": true,
							"minSelect": null,
							"maxSelect": 1,
							"displayFields": null
						}
					},
					{
						"system": false,
						"id": "s92rmmvc",
						"name": "channelId",
						"type": "text",
						"required": true,
						"presentable": false,
						"unique": false,
						"options": {
							"min": null,
							"max": null,
							"pattern": ""
						}
					},
					{
						"system": false,
						"id": "n37ikior",
						"name": "annoucementChannelId",
						"type": "text",
						"required": true,
						"presentable": false,
						"unique": false,
						"options": {
							"min": null,
							"max": null,
							"pattern": ""
						}
					}
				],
				"indexes": [
					"CREATE UNIQUE INDEX ` + "`" + `idx_iJwp9tP` + "`" + ` ON ` + "`" + `eventAnnouncementsConfig` + "`" + ` (\n  ` + "`" + `channelId` + "`" + `,\n  ` + "`" + `guild` + "`" + `\n)"
				],
				"listRule": null,
				"viewRule": null,
				"createRule": null,
				"updateRule": null,
				"deleteRule": null,
				"options": {}
			},
			{
				"id": "x7ch4g59gc6mg1u",
				"created": "2024-10-12 01:32:31.380Z",
				"updated": "2024-10-18 23:06:15.472Z",
				"name": "players",
				"type": "base",
				"system": false,
				"schema": [
					{
						"system": false,
						"id": "htqv2nse",
						"name": "guild",
						"type": "relation",
						"required": true,
						"presentable": false,
						"unique": false,
						"options": {
							"collectionId": "493kl7wl5aa5qvw",
							"cascadeDelete": true,
							"minSelect": null,
							"maxSelect": 1,
							"displayFields": null
						}
					},
					{
						"system": false,
						"id": "tgm9q1bg",
						"name": "userId",
						"type": "text",
						"required": true,
						"presentable": false,
						"unique": false,
						"options": {
							"min": null,
							"max": null,
							"pattern": ""
						}
					},
					{
						"system": false,
						"id": "kfikyi76",
						"name": "nickname",
						"type": "text",
						"required": false,
						"presentable": false,
						"unique": false,
						"options": {
							"min": null,
							"max": null,
							"pattern": ""
						}
					}
				],
				"indexes": [
					"CREATE UNIQUE INDEX ` + "`" + `idx_Mg6Lwaq` + "`" + ` ON ` + "`" + `players` + "`" + ` (\n  ` + "`" + `guild` + "`" + `,\n  ` + "`" + `userId` + "`" + `\n)"
				],
				"listRule": null,
				"viewRule": null,
				"createRule": null,
				"updateRule": null,
				"deleteRule": null,
				"options": {}
			},
			{
				"id": "7n8h6g7j86361xn",
				"created": "2024-10-12 15:10:10.289Z",
				"updated": "2024-10-18 23:03:44.150Z",
				"name": "itemRolls",
				"type": "base",
				"system": false,
				"schema": [
					{
						"system": false,
						"id": "jppioodv",
						"name": "guild",
						"type": "relation",
						"required": true,
						"presentable": false,
						"unique": false,
						"options": {
							"collectionId": "493kl7wl5aa5qvw",
							"cascadeDelete": false,
							"minSelect": null,
							"maxSelect": 1,
							"displayFields": null
						}
					},
					{
						"system": false,
						"id": "bighkmvy",
						"name": "messageId",
						"type": "text",
						"required": false,
						"presentable": false,
						"unique": false,
						"options": {
							"min": null,
							"max": null,
							"pattern": ""
						}
					},
					{
						"system": false,
						"id": "efruo9a1",
						"name": "itemName",
						"type": "text",
						"required": true,
						"presentable": false,
						"unique": false,
						"options": {
							"min": null,
							"max": null,
							"pattern": ""
						}
					},
					{
						"system": false,
						"id": "bbv4dnaz",
						"name": "itemDescription",
						"type": "text",
						"required": false,
						"presentable": false,
						"unique": false,
						"options": {
							"min": null,
							"max": null,
							"pattern": ""
						}
					},
					{
						"system": false,
						"id": "7zoxriyj",
						"name": "rollStart",
						"type": "date",
						"required": true,
						"presentable": false,
						"unique": false,
						"options": {
							"min": "",
							"max": ""
						}
					},
					{
						"system": false,
						"id": "zouay7bx",
						"name": "rollEnd",
						"type": "date",
						"required": true,
						"presentable": false,
						"unique": false,
						"options": {
							"min": "",
							"max": ""
						}
					},
					{
						"system": false,
						"id": "ndhsn7nw",
						"name": "status",
						"type": "select",
						"required": true,
						"presentable": false,
						"unique": false,
						"options": {
							"maxSelect": 1,
							"values": [
								"new",
								"in_progress",
								"ended"
							]
						}
					}
				],
				"indexes": [],
				"listRule": null,
				"viewRule": null,
				"createRule": null,
				"updateRule": null,
				"deleteRule": null,
				"options": {}
			},
			{
				"id": "tmsr4yger7lzzgr",
				"created": "2024-10-13 01:30:00.688Z",
				"updated": "2024-10-18 23:03:44.151Z",
				"name": "itemPlayerRolls",
				"type": "base",
				"system": false,
				"schema": [
					{
						"system": false,
						"id": "c78hk94i",
						"name": "roll",
						"type": "relation",
						"required": true,
						"presentable": false,
						"unique": false,
						"options": {
							"collectionId": "7n8h6g7j86361xn",
							"cascadeDelete": true,
							"minSelect": null,
							"maxSelect": 1,
							"displayFields": null
						}
					},
					{
						"system": false,
						"id": "hucvgxba",
						"name": "player",
						"type": "relation",
						"required": true,
						"presentable": false,
						"unique": false,
						"options": {
							"collectionId": "x7ch4g59gc6mg1u",
							"cascadeDelete": true,
							"minSelect": null,
							"maxSelect": 1,
							"displayFields": null
						}
					},
					{
						"system": false,
						"id": "kphks3o3",
						"name": "rolledNumber",
						"type": "number",
						"required": false,
						"presentable": false,
						"unique": false,
						"options": {
							"min": 0,
							"max": 100,
							"noDecimal": true
						}
					}
				],
				"indexes": [
					"CREATE UNIQUE INDEX ` + "`" + `idx_bxfK4qY` + "`" + ` ON ` + "`" + `itemPlayerRolls` + "`" + ` (\n  ` + "`" + `roll` + "`" + `,\n  ` + "`" + `player` + "`" + `\n)"
				],
				"listRule": null,
				"viewRule": null,
				"createRule": null,
				"updateRule": null,
				"deleteRule": null,
				"options": {}
			},
			{
				"id": "mrrj4sgraz37pmn",
				"created": "2024-10-19 18:50:46.113Z",
				"updated": "2024-10-19 18:50:46.113Z",
				"name": "items",
				"type": "base",
				"system": false,
				"schema": [
					{
						"system": false,
						"id": "ebiupd9g",
						"name": "externalId",
						"type": "number",
						"required": false,
						"presentable": false,
						"unique": false,
						"options": {
							"min": null,
							"max": null,
							"noDecimal": false
						}
					},
					{
						"system": false,
						"id": "ywltwx9k",
						"name": "name",
						"type": "text",
						"required": false,
						"presentable": false,
						"unique": false,
						"options": {
							"min": null,
							"max": null,
							"pattern": ""
						}
					},
					{
						"system": false,
						"id": "0af8rkmx",
						"name": "color",
						"type": "text",
						"required": false,
						"presentable": false,
						"unique": false,
						"options": {
							"min": null,
							"max": null,
							"pattern": ""
						}
					},
					{
						"system": false,
						"id": "vgvw9zpl",
						"name": "category",
						"type": "text",
						"required": false,
						"presentable": false,
						"unique": false,
						"options": {
							"min": null,
							"max": null,
							"pattern": ""
						}
					}
				],
				"indexes": [
					"CREATE UNIQUE INDEX ` + "`" + `idx_v6BiERr` + "`" + ` ON ` + "`" + `items` + "`" + ` (` + "`" + `name` + "`" + `)"
				],
				"listRule": null,
				"viewRule": null,
				"createRule": null,
				"updateRule": null,
				"deleteRule": null,
				"options": {}
			},
			{
				"id": "mrxs8i8fudqo2e3",
				"created": "2024-10-19 18:51:47.619Z",
				"updated": "2024-10-19 23:19:33.385Z",
				"name": "eventTypes",
				"type": "base",
				"system": false,
				"schema": [
					{
						"system": false,
						"id": "7wrnh8ku",
						"name": "name",
						"type": "text",
						"required": false,
						"presentable": false,
						"unique": false,
						"options": {
							"min": null,
							"max": null,
							"pattern": ""
						}
					},
					{
						"system": false,
						"id": "qon5tp0v",
						"name": "category",
						"type": "select",
						"required": false,
						"presentable": false,
						"unique": false,
						"options": {
							"maxSelect": 1,
							"values": [
								"boss"
							]
						}
					},
					{
						"system": false,
						"id": "blvrbajc",
						"name": "enabled",
						"type": "bool",
						"required": false,
						"presentable": false,
						"unique": false,
						"options": {}
					}
				],
				"indexes": [],
				"listRule": null,
				"viewRule": null,
				"createRule": null,
				"updateRule": null,
				"deleteRule": null,
				"options": {}
			},
			{
				"id": "j57p12wtgaqqj62",
				"created": "2024-10-19 20:06:01.886Z",
				"updated": "2024-10-19 23:21:09.640Z",
				"name": "eventConfig",
				"type": "base",
				"system": false,
				"schema": [
					{
						"system": false,
						"id": "ywzr9i76",
						"name": "name",
						"type": "text",
						"required": false,
						"presentable": false,
						"unique": false,
						"options": {
							"min": null,
							"max": null,
							"pattern": ""
						}
					},
					{
						"system": false,
						"id": "k4le9e3h",
						"name": "description",
						"type": "text",
						"required": false,
						"presentable": false,
						"unique": false,
						"options": {
							"min": null,
							"max": null,
							"pattern": ""
						}
					},
					{
						"system": false,
						"id": "4uus9oxz",
						"name": "type",
						"type": "relation",
						"required": false,
						"presentable": false,
						"unique": false,
						"options": {
							"collectionId": "mrxs8i8fudqo2e3",
							"cascadeDelete": true,
							"minSelect": null,
							"maxSelect": 1,
							"displayFields": null
						}
					},
					{
						"system": false,
						"id": "cknronyw",
						"name": "eventTime",
						"type": "text",
						"required": false,
						"presentable": false,
						"unique": false,
						"options": {
							"min": null,
							"max": null,
							"pattern": ""
						}
					},
					{
						"system": false,
						"id": "hssrdwkt",
						"name": "image",
						"type": "file",
						"required": false,
						"presentable": false,
						"unique": false,
						"options": {
							"mimeTypes": [],
							"thumbs": [],
							"maxSelect": 1,
							"maxSize": 5242880,
							"protected": false
						}
					},
					{
						"system": false,
						"id": "9cgm5hz3",
						"name": "enabled",
						"type": "bool",
						"required": false,
						"presentable": false,
						"unique": false,
						"options": {}
					}
				],
				"indexes": [],
				"listRule": null,
				"viewRule": null,
				"createRule": null,
				"updateRule": null,
				"deleteRule": null,
				"options": {}
			},
			{
				"id": "o2fn8399e28o9wm",
				"created": "2024-10-19 23:20:47.048Z",
				"updated": "2024-10-19 23:20:58.060Z",
				"name": "events",
				"type": "base",
				"system": false,
				"schema": [
					{
						"system": false,
						"id": "g64xohyj",
						"name": "event",
						"type": "relation",
						"required": false,
						"presentable": false,
						"unique": false,
						"options": {
							"collectionId": "j57p12wtgaqqj62",
							"cascadeDelete": true,
							"minSelect": null,
							"maxSelect": 1,
							"displayFields": null
						}
					},
					{
						"system": false,
						"id": "sydjeyff",
						"name": "startDate",
						"type": "date",
						"required": false,
						"presentable": false,
						"unique": false,
						"options": {
							"min": "",
							"max": ""
						}
					},
					{
						"system": false,
						"id": "b6sxq3jg",
						"name": "status",
						"type": "select",
						"required": false,
						"presentable": false,
						"unique": false,
						"options": {
							"maxSelect": 1,
							"values": [
								"new",
								"end"
							]
						}
					}
				],
				"indexes": [],
				"listRule": null,
				"viewRule": null,
				"createRule": null,
				"updateRule": null,
				"deleteRule": null,
				"options": {}
			}
		]`

		collections := []*models.Collection{}
		if err := json.Unmarshal([]byte(jsonData), &collections); err != nil {
			return err
		}

		return daos.New(db).ImportCollections(collections, true, nil)
	}, func(db dbx.Builder) error {
		return nil
	})
}
