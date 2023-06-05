package auth

import (
	"strings"
	"time"

	"github.com/GGP1/atoll"
	"github.com/gin-gonic/gin"
	uuid "github.com/satori/go.uuid"
	"github.com/the-clothing-loop/website/server/internal/models"
	"gorm.io/gorm"
)

func TokenReadFromRequest(c *gin.Context) (string, bool) {
	token, ok := cookieRead(c)
	if !ok {
		prefix := "Bearer "
		// fallback to bearer token for testing
		a := c.Request.Header.Get("Authorization")
		_, t, ok := strings.Cut(a, prefix)
		if !ok {
			return "", false
		}
		token = t
	}

	return token, true
}

func TokenCreateUnverified(db *gorm.DB, userID uint) (string, error) {
	// create token
	tokenB, err := atoll.NewPassword(8, []atoll.Level{atoll.Digit})
	if err != nil {
		return "", err
	}
	token := string(tokenB)

	// set token in database
	res := db.Create(&models.UserToken{
		Token:    token,
		Verified: false,
		UserID:   userID,
	})
	if res.Error != nil {
		return "", res.Error
	}

	return token, nil
}

// Returns the user before it was verified
func TokenVerify(db *gorm.DB, token string) (bool, *models.User, string) {
	newToken := uuid.NewV4().String()
	if res := db.Exec(`
UPDATE user_tokens
SET user_tokens.verified = TRUE,
	user_tokens.token = ?
WHERE user_tokens.token = ?
	AND user_tokens.verified = FALSE
	AND user_tokens.created_at > ADDDATE(NOW(), INTERVAL -24 HOUR)
	`, newToken, token); res.Error != nil || res.RowsAffected == 0 {
		return false, nil, ""
	}

	user := &models.User{}
	db.Raw(`
SELECT users.*
FROM user_tokens
LEFT JOIN users ON user_tokens.user_id = users.id
WHERE user_tokens.token = ?
LIMIT 1
	`, newToken).Scan(user)
	if user.ID == 0 {
		return false, nil, ""
	}

	if res := db.Exec(`
UPDATE users
SET is_email_verified = TRUE
WHERE id = ?
	`, user.ID); res.Error != nil {
		return false, nil, ""
	}

	if user.Email.Valid {
		if res := db.Exec(`
UPDATE newsletters
SET verified = TRUE
WHERE email = ?
	`, user.Email.String); res.Error != nil {
			return false, nil, ""
		}
	}

	return true, user, newToken
}

func TokenAuthenticate(db *gorm.DB, token string) (user *models.User, ok bool) {
	user = &models.User{}
	err := db.Raw(`
SELECT users.*
FROM user_tokens
LEFT JOIN users ON user_tokens.user_id = users.id
WHERE user_tokens.token = ? AND user_tokens.verified = TRUE
LIMIT 1
	`, token).Scan(user).Error
	if err != nil || user.ID == 0 {
		return nil, false
	}

	shouldUpdateLastSignedInAt := true
	if user.LastSignedInAt.Valid {
		shouldUpdateLastSignedInAt = user.LastSignedInAt.Time.Before(time.Now().Add(time.Duration(-1 * time.Hour)))
	}
	if shouldUpdateLastSignedInAt {
		db.Exec(`
UPDATE users
SET users.last_signed_in_at = NOW()
WHERE users.id = ?
	`, user.ID)
	}

	return user, true
}

func TokenDelete(db *gorm.DB, token string) {
	db.Exec(`
DELETE FROM user_tokens
WHERE token = ?
	`, token)
}
