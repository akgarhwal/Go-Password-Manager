package util

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"github.com/akgarhwal/go-password-manager/constant"
	"github.com/pterm/pterm"
	"github.com/pterm/pterm/putils"
	"golang.org/x/crypto/pbkdf2"
	"io"
	"os"
	"time"
)

func createHash(key string, salt []byte) []byte {
	return pbkdf2.Key([]byte(key), salt, 367848, 32, sha256.New)
}

func Encrypt(data []byte, passphrase string) []byte {
	salt := make([]byte, 16)

	if _, err := io.ReadFull(rand.Reader, salt); err != nil {
		panic(err.Error())
	}
	key := createHash(passphrase, salt)
	block, err := aes.NewCipher(key)
	if err != nil {
		panic(err.Error())
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		panic(err.Error())
	}
	nonce := make([]byte, gcm.NonceSize())
	if _, err = io.ReadFull(rand.Reader, nonce); err != nil {
		panic(err.Error())
	}
	ciphertext := gcm.Seal(nonce, nonce, data, nil)
	return append(salt, ciphertext...)
}

func Decrypt(data []byte, passphrase string) ([]byte, error) {
	salt := data[:16]
	data = data[16:]
	key := createHash(passphrase, salt)
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}
	nonceSize := gcm.NonceSize()
	nonce, ciphertext := data[:nonceSize], data[nonceSize:]
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return nil, err
	}
	return plaintext, nil
}

func CreateIfFileNotExists() error {
	if !fileExists(constant.JSON_FILE_PATH) {
		file, err := os.Create(constant.JSON_FILE_PATH)
		if err != nil {
			pterm.Error.Println("Error creating file:", err)
			return err
		}

		defer func(file *os.File) {
			_ = file.Close()
		}(file)
	}

	return nil
}

func fileExists(filename string) bool {
	_, err := os.Stat(filename)
	if err == nil {
		return true
	}
	if os.IsNotExist(err) {
		return false
	}
	return false
}

func ShowWelcomeText() {
	pterm.Println()
	err := pterm.DefaultBigText.WithLetters(
		putils.LettersFromStringWithStyle("GO-", pterm.FgCyan.ToStyle()),
		putils.LettersFromStringWithStyle("PW-MGR", pterm.FgGreen.ToStyle())).
		Render()
	if err != nil {
		return
	}
}

func ResetSavedPasswords() {
	currentDate := time.Now().Format("2006-01-02")
	_ = os.Rename(constant.JSON_FILE_PATH, constant.JSON_FILE_PATH+"."+currentDate)
}

func AskUserInput(inputPrompt string) string {
	pterm.Println()
	userInput, _ := pterm.DefaultInteractiveTextInput.Show(inputPrompt)
	return userInput
}

func AskUserInputWithMask(inputPrompt string) string {
	pterm.Println()
	userInput, _ := pterm.DefaultInteractiveTextInput.WithMask("*").Show(inputPrompt)
	return userInput
}

func IsSavedPasswordPresent() bool {
	if fileExists(constant.JSON_FILE_PATH) {
		text, err := os.ReadFile(constant.JSON_FILE_PATH)
		if text == nil || err != nil {
			return false
		}

		decodedCipherText, _ := base64.StdEncoding.DecodeString(string(text))
		if len(decodedCipherText) > 0 {
			return true
		}
	}
	return false
}
