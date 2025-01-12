package migrations

import (
	"github.com/pocketbase/dbx"
	"github.com/pocketbase/pocketbase/core"
	m "github.com/pocketbase/pocketbase/migrations"
)

func init() {
	m.Register(func(app core.App) error {
		_, err := app.DB().Update("plannedEvents", dbx.Params{"week": "all"}, dbx.HashExp{"week": ""}).Execute()
		if err != nil {
			return err
		}
		return nil
	}, func(app core.App) error {
		return nil
	})
}
