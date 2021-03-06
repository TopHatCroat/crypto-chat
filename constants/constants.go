package constants

var (
	USERNAME              = "username"
	PASSWORD              = "password"
	LOGIN_SUCCESS         = "Successful login"
	REGISTER_SUCCESS      = "Registration has been successful"
	NO_SUCH_USER_ERROR    = "No such user found"
	NO_SUCH_SETTING_ERROR = "No such setting found"
	WRONG_CREDS_ERROR     = "Wrong username and/or password"
	ALREADY_EXISTS        = "Username already taken"
	WRONG_REQUEST         = "Incorrect request"
	INVALID_TOKEN         = "Invalid token"
	OLD_TOKEN             = "Expired token"
	MESSAGE_SENT          = "Message sent successfully"
	PRIVATE_KEY           = "private_key"
	PUBLIC_KEY            = "public_key"
	TOKEN_KEY             = "token_key"
	GENERATE_NONCE_ERROR  = "Error generating unique key"
	NONCE_ERROR           = "Error validating unique key"
	ENCRYPT_ERROR         = "Error encrypting message"
	DECRYPT_ERROR         = "Error decrypting message"
	REQUEST_REJECTED      = "Request rejected"
	NO_SUCH_KEY_ERROR     = "Not such key found"
	KEY_SUBMIT_SUCCESS    = "Key submited successfully"
	KEY_FOUND_SUCCESS     = "Key found"
	INVALID_DECYPT_ERROR  = "Invalid data for decryption"

	SERVER_NAME    = "localhost:44333"
	TODO_SECRET    = "TODO: secret"
	TOKEN_KEY_FILE = "token_key.pem"
	EDITION_VAR    = "SECURECHAT"
	SERVER_EDITION = "server"
	CLIENT_EDITION = "client"
)

const (
	KEY_SIZE   = 32
	NONCE_SIZE = 24
)
