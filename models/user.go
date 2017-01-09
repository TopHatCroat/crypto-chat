package models

import (
	"errors"
	"github.com/TopHatCroat/CryptoChat-server/constants"
	"github.com/TopHatCroat/CryptoChat-server/database"
	"github.com/TopHatCroat/CryptoChat-server/helpers"
	"github.com/dgrijalva/jwt-go"
	"golang.org/x/crypto/bcrypt"
	"time"
)

type Entity interface {
	Save() int64
	Delete() int64
}

type User struct {
	ID             int64
	Username       string
	PasswordDigest string
	Gcm            string
	PublicKey      string
}

func (u *User) Save() int64 {
	db := database.GetDatabase()
	if u.ID == 0 {
		preparedStatement, err := db.Prepare("INSERT INTO users(username, password, gcm, public_key) VALUES(?,?,?,?)")
		helpers.HandleError(err)
		result, err := preparedStatement.Exec(u.Username, u.PasswordDigest, u.Gcm, u.PublicKey)
		helpers.HandleError(err)
		u.ID, _ = result.LastInsertId()
	} else {
		preparedStatement, err := db.Prepare("UPDATE users set username = ?, password = ?, gcm = ?, public_key = ? WHERE id = ?")
		helpers.HandleError(err)
		_, err = preparedStatement.Exec(u.Username, u.PasswordDigest, u.Gcm, u.PublicKey, u.ID)
		helpers.HandleError(err)
	}
	return u.ID
}

func (u *User) LogIn(password string) (string, error) {
	err := bcrypt.CompareHashAndPassword([]byte(u.PasswordDigest), []byte(password))
	if err != nil {
		return "", errors.New(constants.WRONG_CREDS_ERROR)
	}

	claims := Claims{
		u.Username,
		jwt.StandardClaims{
			IssuedAt: time.Now().Unix(),
			Issuer:   constants.SERVER_NAME,
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodES384, claims)
	tokenBytes, err := helpers.ReadFromFile(constants.TOKEN_KEY_FILE)
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
		SessionKey:   signedToken,
		UserID:       u.ID,
		LoginTime:    time.Now().UnixNano(),
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
	result, err := preparedStatement.Exec(u.ID)
	helpers.HandleError(err)

	defer db.Close()
	count, _ := result.RowsAffected()

	return count
}

func FindUserById(id int64) (u User, err error) {
	db := database.GetDatabase()

	preparedStatement, err := db.Prepare("SELECT * FROM users WHERE id = ?")
	helpers.HandleError(err)
	row, err := preparedStatement.Query(id)

	row.Next()
	err = row.Scan(&u.ID, &u.Username, &u.PasswordDigest, &u.Gcm, &u.PublicKey)
	if err != nil {
		return u, errors.New(constants.NO_SUCH_USER_ERROR)
	}
	row.Close()
	return u, nil
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
	err = row.Scan(&u.ID, &u.Username, &u.PasswordDigest, &u.Gcm, &u.PublicKey)
	defer row.Close()
	if err != nil {
		//TODO: we are returning a wrong error here in case it's a friend request and not a regular user querry
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

func FindUserByToken(token string) (user User, err error) {
	db := database.GetDatabase()

	claims, err := ParseToken(token)
	if err != nil {
		return user, err
	}

	preparedStatement, err := db.Prepare("SELECT * FROM user_sessions WHERE session_key = ? ")
	if err != nil {
		panic(err)
	}
	row, err := preparedStatement.Query(token)
	if err != nil {
		return user, errors.New(constants.INVALID_TOKEN)
	}

	var userSession UserSession
	row.Next()
	err = row.Scan(&userSession.SessionKey, &userSession.UserID, &userSession.LoginTime, &userSession.LastSeenTime)
	row.Close()
	if err != nil {
		return user, errors.New(constants.INVALID_TOKEN)
	}

	if userSession.LastSeenTime < time.Now().Add(-1*time.Minute).UnixNano() ||
		userSession.LoginTime < time.Now().Add(-1*time.Hour).UnixNano() {
		return user, errors.New(constants.OLD_TOKEN)
	}

	userSession.LastSeenTime = time.Now().UnixNano()
	err = userSession.Save()
	if err != nil {
		return user, err
	}

	user, err = FindUserById(userSession.UserID)
	if err != nil {
		return user, err
	}

	if claims.Username != user.Username {
		return user, errors.New(constants.INVALID_TOKEN)
	}

	return user, nil
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
	user, err = FindUserById(id)
	if err != nil {
		return u, err
	}

	return user, nil
}
