package migrations

import (
	"github.com/pocketbase/pocketbase/core"
	m "github.com/pocketbase/pocketbase/migrations"
)

func init() {
	m.Register(func(app core.App) error {
		collection, err := app.FindCollectionByNameOrId("00kdyrpyd873da9")
		if err != nil {
			return err
		}

		// update field
		if err := collection.Fields.AddMarshaledJSONAt(9, []byte(`{
			"hidden": false,
			"id": "select1532651968",
			"maxSelect": 1,
			"name": "week",
			"presentable": false,
			"required": true,
			"system": false,
			"type": "select",
			"values": [
				"even",
				"odd",
				"all"
			]
		}`)); err != nil {
			return err
		}

		return app.Save(collection)
	}, func(app core.App) error {
		collection, err := app.FindCollectionByNameOrId("00kdyrpyd873da9")
		if err != nil {
			return err
		}

		// update field
		if err := collection.Fields.AddMarshaledJSONAt(9, []byte(`{
			"hidden": false,
			"id": "select1532651968",
			"maxSelect": 1,
			"name": "week",
			"presentable": false,
			"required": false,
			"system": false,
			"type": "select",
			"values": [
				"even",
				"odd",
				"all"
			]
		}`)); err != nil {
			return err
		}

		return app.Save(collection)
	})
}
