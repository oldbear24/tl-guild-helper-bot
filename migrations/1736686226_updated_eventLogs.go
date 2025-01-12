package migrations

import (
	"github.com/pocketbase/pocketbase/core"
	m "github.com/pocketbase/pocketbase/migrations"
)

func init() {
	m.Register(func(app core.App) error {
		collection, err := app.FindCollectionByNameOrId("xmrfb0lc2lbf182")
		if err != nil {
			return err
		}

		// add field
		if err := collection.Fields.AddMarshaledJSONAt(10, []byte(`{
			"autogeneratePattern": "",
			"hidden": false,
			"id": "text1786863759",
			"max": 0,
			"min": 0,
			"name": "reminderMessageId",
			"pattern": "",
			"presentable": false,
			"primaryKey": false,
			"required": false,
			"system": false,
			"type": "text"
		}`)); err != nil {
			return err
		}

		// add field
		if err := collection.Fields.AddMarshaledJSONAt(11, []byte(`{
			"autogeneratePattern": "",
			"hidden": false,
			"id": "text3996573572",
			"max": 0,
			"min": 0,
			"name": "reminderMessageChannelId",
			"pattern": "",
			"presentable": false,
			"primaryKey": false,
			"required": false,
			"system": false,
			"type": "text"
		}`)); err != nil {
			return err
		}

		return app.Save(collection)
	}, func(app core.App) error {
		collection, err := app.FindCollectionByNameOrId("xmrfb0lc2lbf182")
		if err != nil {
			return err
		}

		// remove field
		collection.Fields.RemoveById("text1786863759")

		// remove field
		collection.Fields.RemoveById("text3996573572")

		return app.Save(collection)
	})
}
