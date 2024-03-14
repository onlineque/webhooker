package cmd

import (
	"context"
	"crypto/sha512"
	"fmt"
	"github.com/sethvargo/go-password/password"
	"github.com/spf13/cobra"
	"go.mongodb.org/mongo-driver/bson"
	"log"
	"webhooker/database"
)

// createTokenCmd represents the createToken command
var createTokenCmd = &cobra.Command{
	Use:   "createToken",
	Short: "create a webhook token",
	Long:  `Creates webhook token which can be used by the client to post data to this webhook server`,
	Run: func(cmd *cobra.Command, args []string) {
		dbUri, _ := cmd.Flags().GetString("dburi")

		dbObj := &database.DbConnection{}
		dbObj.NewDbConnection(dbUri)
		err := dbObj.Connect()
		if err != nil {
			log.Fatalf("error connecting to database: %s", err)
		}
		tokens := dbObj.GetCollection("tokens")

		password, err := password.Generate(64, 10, 2, false, false)
		if err != nil {
			log.Fatalf("error generating the token: %s", err)
		}
		shaObj := sha512.New()
		shaObj.Write([]byte(password))
		userTokenHash := fmt.Sprintf("%x", shaObj.Sum(nil))

		_, err = tokens.InsertOne(context.TODO(), bson.D{
			bson.E{Key: "token", Value: userTokenHash},
		})
		if err != nil {
			log.Fatalf("error inserting the token into database: %s", err)
		}
		log.Printf("Please write the token down to a secure place, it is the last time it is displayed. "+
			"Token: %s", password)
	},
}

func init() {
	rootCmd.AddCommand(createTokenCmd)
	createTokenCmd.PersistentFlags().String("dburi", "", "MongoDB database connection URI")
	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// createTokenCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// createTokenCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
