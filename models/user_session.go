package models

import (
	"github.com/TopHatCroat/CryptoChat-server/database"
	"github.com/TopHatCroat/CryptoChat-server/protocol"
	"github.com/dgrijalva/jwt-go"
	"errors"
	"github.com/TopHatCroat/CryptoChat-server/helpers"
	"github.com/TopHatCroat/CryptoChat-server/constants"
)

type UserSession struct {
	SessionKey   string
	UserID       int64
	LoginTime    int64
	LastSeenTime int64
}

func (ses *UserSession) Save() error {
	db := database.GetDatabase()

	preparedStatement, err := db.Prepare("INSERT OR REPLACE INTO user_sessions (session_key, user_id, login_time," +
		" last_seen_time) VALUES(?,?,?,?)")
	if err != nil {
		return err
	}
	_, err = preparedStatement.Exec(ses.SessionKey, ses.UserID, ses.LoginTime, ses.LastSeenTime)
	if err != nil {
		return err
	}

	return nil
}

func (ses *UserSession) Delete() error {
	db := database.GetDatabase()
	defer db.Close()

	preparedStatement, err := db.Prepare("DELETE FROM user_sessions WHERE session_key = ?")
	if err != nil {
		return err
	}

	_, err = preparedStatement.Exec(ses.SessionKey)
	if err != nil {
		return err
	}

	return nil
}

func ParseToken(tokenString string) (cl protocol.Claims, err error) {
	tokenBytes, err := helpers.ReadFromFile(constants.TOKEN_KEY_FILE)
	if err != nil {
		return cl, err
	}
	tokenParsedECPrivate, err := jwt.ParseECPrivateKeyFromPEM(tokenBytes)
	if err != nil {
		return cl, err
	}

	token, err := jwt.ParseWithClaims(tokenString, &protocol.Claims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodECDSA); !ok {
			return nil, errors.New(constants.INVALID_TOKEN)
		}

		return &tokenParsedECPrivate.PublicKey, nil
	})

	if err != nil {
		return cl, err
	}

	if claims, ok := token.Claims.(*protocol.Claims); ok && token.Valid {
		return *claims, nil
  	} else {
		return *claims, errors.New(constants.INVALID_TOKEN)
	}

}
