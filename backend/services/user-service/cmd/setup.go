package cmd

import (
	"log"

	"github.com/spf13/cobra"
	"github.com/wutthichod/sa-connext/services/user-service/internal/models"
	"github.com/wutthichod/sa-connext/services/user-service/pkg/database"
)

// setupCmd represents the setup command
var setupCmd = &cobra.Command{
	Use:   "setup",
	Short: "Run initial setup for the application",
	RunE: func(cmd *cobra.Command, args []string) error {

		cfg, err := getConfigFromCmd(cmd)
		if err != nil {
			return err
		}

		db, err := database.InitDatabase(cfg.Database())
		if err != nil {
			return err
		}

		// TODO: run migrations
		// db.Migrator().DropTable(&models.Province{},&models.District{},&models.SubDistrict{})
		err = db.Migrator().DropTable(
			&models.User{},
			&models.Contact{},
			&models.Education{},
			&models.Interest{},
		)
		if err != nil {
			log.Fatalf("failed to drop tables: %v", err)
		}
		log.Println("All tables dropped successfully")
		err = db.AutoMigrate(
			&models.User{},
			&models.Contact{},
			&models.Education{},
			&models.Interest{},
		)
		if err != nil {
			log.Fatalf("failed to migrate tables: %v", err)
		}
		return nil

	},
}
