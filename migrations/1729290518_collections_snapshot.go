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
				"updated": "2024-10-16 21:56:44.302Z",
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
				"updated": "2024-10-16 21:56:44.303Z",
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
				"updated": "2024-10-16 21:56:44.304Z",
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
				"updated": "2024-10-16 21:56:44.304Z",
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
				"updated": "2024-10-17 22:17:00.545Z",
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
				"updated": "2024-10-16 21:56:44.306Z",
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
