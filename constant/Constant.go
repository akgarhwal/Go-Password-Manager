package constant

// TODO: write this data in hoem dire of user
const JSON_FILE_PATH = ".passwords.xyz"

const MaxRetryCount = 3

const ErrMessageAuthFailed = "cipher: message authentication failed"

// Define Master Key for Global Context
type contextKey string

const MasterKey contextKey = "masterKey"
