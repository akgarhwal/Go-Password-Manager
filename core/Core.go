package core

import (
	"context"
	"github.com/akgarhwal/go-password-manager/constant"
	"github.com/akgarhwal/go-password-manager/model"
	"github.com/akgarhwal/go-password-manager/util"
	"github.com/pterm/pterm"
	"strings"
)

func StartInteractiveMode(ctx context.Context, pm *model.PasswordManager) {

	for {
		userInputCmd := util.AskUserInput("Enter command (add, get, delete, list, exit)")

		switch userInputCmd {
		case "add":

			ShowKeyInfoIfNeeded(pm)
			key := util.AskUserInput("Enter Key (nickname) for Password")
			var values []model.KeyValuePair

			pterm.DefaultBasicText.Print(pterm.LightMagenta("Enter values (eg: username=cli_user). Type 'done' to finish."))
			for {
				userInput := util.AskUserInput("Enter value")
				userInput = strings.TrimSpace(userInput)

				if userInput == "done" {
					break
				}

				kv := strings.Split(userInput, "=")
				values = append(values, model.KeyValuePair{Key: kv[0], Value: kv[1]})
			}

			pm.AddPassword(ctx, model.PasswordEntry{Key: key, Values: values})
			pterm.Success.Println("Password Saved.")

		case "get":
			userKey := util.AskUserInput("Enter Key (nickname) for Password")
			pm.GetPassword(userKey)

		case "delete":
			keyToDelete := util.AskUserInput("Enter Key for Password to Delete")
			isDeleted := pm.DeletePassword(ctx, keyToDelete)
			if isDeleted {
				pterm.Success.Println("Password deleted with key=", keyToDelete)
			} else {
				pterm.Warning.Println("Password not found for key=", keyToDelete)
			}

		case "list":
			pm.ListPasswords()

		case "exit":
			pterm.Info.Println("Exiting...")
			return
		default:
			pterm.Warning.Println("Unknown command. Available commands: add, get, delete, list, exit")
		}
	}
}

func ShowKeyInfoIfNeeded(pm *model.PasswordManager) {
	if len(pm.Passwords) == 0 {
		pterm.DefaultBox.
			WithTitle(pterm.LightYellow("What is Key ?")).
			Println(pterm.LightMagenta("Keys are nicknames for passwords. " +
				"\nFor example, you can use 'HDFC Personal Login'" +
				"\nas the key for your HDFC credentials."))
	}
}

func LoadSavedPassword(ctx context.Context) (context.Context, *model.PasswordManager, error) {
	pm := model.NewPasswordManager()
	var err error
	retryCount := 0

	if !util.IsSavedPasswordPresent() {
		ctx = AskUserForMasterKey(ctx)
		err = pm.LoadFromFile(ctx)
		return ctx, pm, err
	}

	for {
		masterKey := util.AskUserInputWithMask("Enter Master Key")
		ctx = context.WithValue(ctx, constant.MasterKey, masterKey)

		err = pm.LoadFromFile(ctx)

		if err != nil && err.Error() == constant.ErrMessageAuthFailed {
			pterm.Error.Println("Master key is incorrect. Please try again.")
			retryCount += 1
		} else if err != nil {
			return ctx, pm, err
		}

		if err == nil {
			break
		}

		if retryCount > constant.MaxRetryCount {
			pterm.Println()
			confirm := util.AskUserInput("Forgot Master Key :( Delete saved passwords and start fresh ? (yes/no)")
			if strings.ToLower(confirm) == "yes" {
				util.ResetSavedPasswords()
				ctx = AskUserForMasterKey(ctx)
				err = pm.LoadFromFile(ctx)
				return ctx, pm, err
			}
		}
	}

	return ctx, pm, err
}

func AskUserForMasterKey(ctx context.Context) context.Context {
	masterKey := ""
	for {
		masterKey = util.AskUserInputWithMask("Please enter the master key (Don't forget it)")

		// Simple Validation for Master Key
		if len(masterKey) < 8 {
			pterm.Error.Println("Master Key should contain at-least 8 chars.")
		} else {
			break
		}
	}

	ctx = context.WithValue(ctx, constant.MasterKey, masterKey)
	return ctx
}
