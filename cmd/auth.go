package cmd

import (
	"context"
	"fmt"
	"log"
	"net"

	"github.com/AlecAivazis/survey/v2"
	"github.com/n4vxn/FileMove/db"
	"github.com/n4vxn/FileMove/types"
	"github.com/spf13/cobra"
	"golang.org/x/crypto/bcrypt"
)

type contextKey = string

const userContextKey contextKey = "username"

type UserConn struct {
	Conn net.Conn
	Username string
	Ctx context.Context
}

var activeConns = make(map[string]*UserConn)

var SignUpCmd = &cobra.Command{
	Use:   "signup",
	Short: "Sign up a new user.",
	Run: func(cmd *cobra.Command, args []string) {
		var username, password string

		survey.AskOne(&survey.Input{Message: "Enter Username:"}, &username)
		survey.AskOne(&survey.Input{Message: "Enter Password:"}, &password)

		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), 10)
		if err != nil {
			log.Fatal(err)
		}
		user := types.User{
			Username: username,
			Password: string(hashedPassword),
		}

		db.SaveUsers(&user)
	},
}

var LoginCmd = &cobra.Command{
	Use:   "login",
	Short: "Login an existing user.",
	RunE: func(cmd *cobra.Command, args []string) error {
		var username, password string

		survey.AskOne(&survey.Input{Message: "Enter Username: "}, &username)
		survey.AskOne(&survey.Input{Message: "Enter Password: "}, &password)

		hashedpassword, err := db.RetrieveHashedPassword(username)
		if err != nil {
			return fmt.Errorf("invalid credentials: %v", err)
		}
		err = bcrypt.CompareHashAndPassword([]byte(hashedpassword), []byte(password))
		if err != nil {
			return fmt.Errorf("invalid credentials")
		}

		log.Printf("Login successful! Welcome, %s.\n", username)
		return nil
	},
}

func init() {
	// Add the 'signup' and 'login' commands to the root command
	rootCmd.AddCommand(SignUpCmd)

}
