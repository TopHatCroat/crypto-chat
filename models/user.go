package models

import (
	"errors"
	"github.com/TopHatCroat/CryptoChat-server/constants"
	"github.com/TopHatCroat/CryptoChat-server/database"
	"github.com/TopHatCroat/CryptoChat-server/helpers"
	"golang.org/x/crypto/bcrypt"
	"time"
	"github.com/TopHatCroat/CryptoChat-server/protocol"
	"github.com/dgrijalva/jwt-go"
)

var ()

type Entity interface {
	Save() int64
	Delete() int64
}

type User struct {
	id             int64
	Username       string
	PasswordDigest string
	Gcm            string
	PublicKey      string
}

func (u *User) Save() int64 {
	db := database.GetDatabase()
	if u.id == 0 {
		preparedStatement, err := db.Prepare("INSERT INTO users(username, password, gcm, public_key) VALUES(?,?,?,?)")
		helpers.HandleError(err)
		result, err := preparedStatement.Exec(u.Username, u.PasswordDigest, u.Gcm, u.PublicKey)
		helpers.HandleError(err)
		u.id, _ = result.LastInsertId()
	} else {
		preparedStatement, err := db.Prepare("UPDATE users set username = ?, password = ?, gcm = ?, public_key = ? WHERE id = ?")
		helpers.HandleError(err)
		_, err = preparedStatement.Exec(u.Username, u.PasswordDigest, u.Gcm, u.id, u.PublicKey)
		helpers.HandleError(err)
	}
	return u.id
}

func (u *User) LogIn(password string) (string, error) {
	err := bcrypt.CompareHashAndPassword([]byte(u.PasswordDigest), []byte(password))
	if err != nil {
		return "", errors.New(constants.WRONG_CREDS_ERROR)
	}

	claims := protocol.Claims{
		u.Username,
		jwt.StandardClaims{
			IssuedAt: time.Now().UnixNano(),
			Issuer: constants.SERVER_NAME,
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodES384, claims)
	tokenBytes, err := helpers.ReadFromFile("token.pem")
	if err != nil {
		return "", err
	}
	tokenParsedECPrivate, err := jwt.ParseECPrivateKeyFromPEM(tokenBytes)
	if err != nil {
		return "", err
	}
	signedToken, err := token.SignedString(tokenParsedECPrivate)
	if err != nil {
		return "", err
	}

	userSession := UserSession{
		SessionKey:	signedToken,
		UserID: u.id,
		LoginTime: time.Now().UnixNano(),
		LastSeenTime: time.Now().UnixNano(),
	}

	err = userSession.Save()
	if err != nil {
		return "", err
	}

	return signedToken, nil
}

func (u *User) Delete() int64 {
	db := database.GetDatabase()

	preparedStatement, err := db.Prepare("DELETE FROM users WHERE id = ?")
	helpers.HandleError(err)
	result, err := preparedStatement.Exec(u.id)
	helpers.HandleError(err)

	defer db.Close()
	count, _ := result.RowsAffected()

	return count
}

func FindUserById(id int64) (u User) {
	db := database.GetDatabase()

	preparedStatement, err := db.Prepare("SELECT * FROM users WHERE id = ?")
	helpers.HandleError(err)
	row, err := preparedStatement.Query(id)

	row.Next()
	err = row.Scan(&u.id, &u.Username, &u.PasswordDigest, &u.Gcm, &u.PublicKey)
	helpers.HandleError(err)
	row.Close()
	return u
}

func FindUserByCreds(username string) (u User, e error) {
	db := database.GetDatabase()

	preparedStatement, err := db.Prepare("SELECT * FROM users WHERE username = ? ")
	helpers.HandleError(err)
	row, err := preparedStatement.Query(username)
	if err != nil {
		return u, err
	}

	row.Next()
	err = row.Scan(&u.id, &u.Username, &u.PasswordDigest, &u.Gcm, &u.PublicKey)
	defer row.Close()
	if err != nil {
		return u, errors.New(constants.WRONG_CREDS_ERROR)
	}
	return u, nil
}

func usernameExists(username string) (exists bool) {
	db := database.GetDatabase()
	preparedStatement, err := db.Prepare("SELECT COUNT(*) FROM users WHERE username = ?")
	helpers.HandleError(err)
	row, err := preparedStatement.Query(username)
	defer row.Close()
	row.Next()
	var count int
	err = row.Scan(&count)
	if count == 0 {
		return false
	} else {
		return true
	}
}

func CreateUser(nick string, pass string, publicKey string) (u User, e error) {
	if usernameExists(nick) {
		return u, errors.New(constants.ALREADY_EXISTS)
	}

	passwordDigest, err := bcrypt.GenerateFromPassword([]byte(pass), bcrypt.DefaultCost)
	if err != nil {
		return u, err
	}

	user := User{Username: nick, PasswordDigest: string(passwordDigest), Gcm: "0", PublicKey: publicKey}
	id := user.Save()
	user = FindUserById(id)
	return user, nil
}
