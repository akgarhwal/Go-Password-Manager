package model

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/akgarhwal/go-password-manager/constant"
	"github.com/akgarhwal/go-password-manager/util"
	"github.com/pterm/pterm"
	"os"
	"strings"
)

type KeyValuePair struct {
	Key   string
	Value string
}

type PasswordEntry struct {
	Key    string
	Values []KeyValuePair
}

type PasswordManager struct {
	Passwords []PasswordEntry
}

func NewPasswordManager() *PasswordManager {
	return &PasswordManager{}
}

func (pm *PasswordManager) AddPassword(ctx context.Context, newPasswordEntry PasswordEntry) {

	// Check if the PasswordEntry with the same key already exists
	for i, entry := range pm.Passwords {
		if entry.Key == newPasswordEntry.Key {

			// Update the existing entry values one by one
			for _, newKV := range newPasswordEntry.Values {
				updated := false
				for j, existingKV := range entry.Values {
					if existingKV.Key == newKV.Key {
						pm.Passwords[i].Values[j].Value = newKV.Value
						updated = true
						break
					}
				}
				if !updated {
					pm.Passwords[i].Values = append(pm.Passwords[i].Values, newKV)
				}
			}

			pm.SaveToFile(ctx)
			return
		}
	}

	// If not found, append the new PasswordEntry
	pm.Passwords = append(pm.Passwords, newPasswordEntry)
	pm.SaveToFile(ctx)
}

func (pm *PasswordManager) ListPasswords() {
	if len(pm.Passwords) == 0 {
		pterm.DefaultBasicText.Println("No passwords found")
		return
	}

	title := pterm.LightGreen("Keys")
	keysFound := ""
	counter := 0
	for _, p := range pm.Passwords {
		counter += 1
		keysFound += fmt.Sprintf("%d. %s\n", counter, p.Key)
	}

	pterm.DefaultBox.WithTitle(title).Println(keysFound)
}

func (pm *PasswordManager) GetPassword(searchKey string) {
	found := false
	for _, p := range pm.Passwords {
		if strings.HasPrefix(p.Key, searchKey) {
			found = true
			title := pterm.LightGreen(p.Key)

			matchPasswordEntry := ""
			for _, kv := range p.Values {
				matchPasswordEntry += fmt.Sprintf("%s: %s\n", kv.Key, kv.Value)
			}

			pterm.DefaultBox.WithTitle(title).Println(matchPasswordEntry)
		}
	}

	if !found {
		pterm.Warning.Println("Password not found for Key=", searchKey)
	}
}

func (pm *PasswordManager) DeletePassword(ctx context.Context, key string) bool {
	isDeleted := false
	for i, p := range pm.Passwords {
		if p.Key == key {
			pm.Passwords = append(pm.Passwords[:i], pm.Passwords[i+1:]...)
			isDeleted = true
			break
		}
	}
	pm.SaveToFile(ctx)
	return isDeleted
}

func (pm *PasswordManager) SaveToFile(ctx context.Context) {
	data, _ := json.Marshal(pm.Passwords)
	cipherText := util.Encrypt(data, ctx.Value(constant.MasterKey).(string))
	encodedCipherText := base64.StdEncoding.EncodeToString(cipherText)
	_ = os.WriteFile(constant.JSON_FILE_PATH, []byte(encodedCipherText), 0600)
}

func (pm *PasswordManager) LoadFromFile(ctx context.Context) error {

	err := util.CreateIfFileNotExists()
	if err != nil {
		pterm.Error.Println("Not able to create file to save password. Error: ", err)
		return err
	}

	ciphertext, err := os.ReadFile(constant.JSON_FILE_PATH)
	if ciphertext == nil || err != nil {
		pterm.Error.Println("Error while reading saved password. Error: ", err)
		return nil
	}

	decodedCipherText, _ := base64.StdEncoding.DecodeString(string(ciphertext))
	if len(decodedCipherText) == 0 {
		pterm.Info.Println(pterm.LightYellow("No saved passwords found"))
		return nil
	}

	data, err := util.Decrypt(decodedCipherText, ctx.Value(constant.MasterKey).(string))
	if err != nil {
		return err
	}

	err = json.Unmarshal(data, &pm.Passwords)
	if err != nil {
		pterm.Error.Println("Could not read decrypted passwords. Error: ", err)
		return err
	}

	return nil
}
