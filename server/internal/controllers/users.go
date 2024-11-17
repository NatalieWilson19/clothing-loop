package controllers

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/the-clothing-loop/website/server/internal/app"
	"github.com/the-clothing-loop/website/server/internal/app/auth"
	"github.com/the-clothing-loop/website/server/internal/models"
	"github.com/the-clothing-loop/website/server/internal/services"
	"github.com/the-clothing-loop/website/server/internal/views"
	"gopkg.in/guregu/null.v3"
	"gorm.io/gorm"

	"github.com/gin-gonic/gin"
)

type UserCreateRequestBody struct {
	Email       string   `json:"email" binding:"required,email"`
	Name        string   `json:"name" binding:"required,min=3"`
	Address     string   `json:"address" binding:"required,min=3"`
	PhoneNumber string   `json:"phone_number" binding:"required,min=3"`
	Newsletter  bool     `json:"newsletter"`
	Sizes       []string `json:"sizes"`
	Latitude    float64  `json:"latitude"`
	Longitude   float64  `json:"longitude"`
}

func UserGet(c *gin.Context) {
	db := getDB(c)

	var query struct {
		UserUID         string `form:"user_uid" binding:"omitempty,uuid"`
		ChainUID        string `form:"chain_uid" binding:"omitempty,uuid"`
		AddApprovedTOH  bool   `form:"add_approved_toh" binding:"omitempty"`
		AddNotification bool   `form:"add_notification" binding:"omitempty"`
	}
	if err := c.ShouldBindQuery(&query); err != nil {
		c.String(http.StatusBadRequest, err.Error())
		return
	}

	// retrieve user from query
	if query.UserUID == "" {
		c.String(http.StatusBadRequest, "Add uid or email to query")
		return
	}

	ok := false
	var authUser *models.User
	if query.ChainUID == "" {
		ok, authUser, _ = auth.Authenticate(c, db, auth.AuthState1AnyUser, "")

		if !ok || query.UserUID != authUser.UID {
			c.String(http.StatusUnauthorized, "For elevated privileges include a chain_uid")
			return
		}
	} else {
		ok, authUser, _ = auth.Authenticate(c, db, auth.AuthState3AdminChainUser, query.ChainUID)
	}
	if !ok {
		return
	}
	isMe := authUser.UID == query.UserUID
	if !isMe && query.AddApprovedTOH && !authUser.IsRootAdmin {
		c.String(http.StatusUnauthorized, "User details requested are not authorized")
		return
	}

	user, err := models.UserGetByUID(db, query.UserUID, true)
	if err != nil {
		c.String(http.StatusBadRequest, "User not found")
		return
	}

	err = user.AddUserChainsToObject(db)
	if err != nil {
		slog.Error(models.ErrAddUserChainsToObject.Error(), "err", err)
		c.String(http.StatusInternalServerError, models.ErrAddUserChainsToObject.Error())
		return
	}

	if query.AddNotification && (isMe || authUser.IsRootAdmin) {
		err := user.AddNotificationChainUIDs(db)
		if err != nil {
			c.AbortWithError(http.StatusInternalServerError, err)
			return
		}
	}

	if query.AddApprovedTOH {
		user.SetAcceptedLegal()
	}

	c.JSON(200, user)
}

func UserGetAllOfChain(c *gin.Context) {
	db := getDB(c)

	var query struct {
		ChainUID string `form:"chain_uid" binding:"required,uuid"`
	}
	if err := c.ShouldBindQuery(&query); err != nil {
		c.String(http.StatusBadRequest, err.Error())
		return
	}

	ok, authUser, chain := auth.Authenticate(c, db, auth.AuthState2UserOfChain, query.ChainUID)
	if !ok {
		return
	}

	_, isAuthUserChainAdmin := authUser.IsPartOfChain(chain.UID)
	isAuthState3AdminChainUser := isAuthUserChainAdmin || authUser.IsRootAdmin

	// retrieve user from query
	tx := db.Begin()
	allUserChains, err := models.UserChainGetIndirectByChain(tx, chain.ID)

	if err != nil {
		slog.Error("Unable to retrieve associations between a loop and its users", "err", err)
		c.String(http.StatusInternalServerError, "Unable to retrieve associations between a loop and its users")
		return
	}
	users, errUsersByChain := models.UserGetAllUsersByChain(tx, chain.ID)
	if errUsersByChain != nil {
		slog.Error("Unable to retrieve associated users of a loop", "err", err)
		c.String(http.StatusInternalServerError, "Unable to retrieve associated users of a loop")
		return
	}
	tx.Commit()

	for i, user := range users {
		thisUserChains := []models.UserChain{}
		for ii := range allUserChains {
			userChain := (allUserChains)[ii]
			if userChain.UserID == user.ID {
				// goscope.Log.Infof("userchain is added (userChain.ID: %d -> user.ID: %d)\n", userChain.ID, user.ID)
				thisUserChains = append(thisUserChains, userChain)
			}
		}
		users[i].Chains = thisUserChains
	}

	// omit user data from participants
	if !isAuthState3AdminChainUser {
		users, err = models.UserOmitData(db, chain, users, authUser.ID)

		if err != nil {
			slog.Error("Unable to omit user data", "err", err)
			c.String(http.StatusInternalServerError, "Internal error hiding user information")
			return
		}
	}

	c.JSON(200, users)
}

func UserHasNewsletter(c *gin.Context) {
	db := getDB(c)

	var query struct {
		UserUID  string `form:"user_uid" binding:"required,uuid"`
		ChainUID string `form:"chain_uid" binding:"omitempty,uuid"`
	}
	if err := c.ShouldBindQuery(&query); err != nil {
		c.String(http.StatusBadRequest, err.Error())
		return
	}

	ok, user, _, _ := auth.AuthenticateUserOfChain(c, db, query.ChainUID, query.UserUID)
	if !ok {
		return
	}

	hasNewsletter := 0
	db.Raw(`SELECT COUNT(*) FROM newsletters WHERE email = ? LIMIT 1`, user.Email.String).Scan(&hasNewsletter)

	c.JSON(200, hasNewsletter > 0)
}

func UserUpdate(c *gin.Context) {
	db := getDB(c)

	var body struct {
		ChainUID      string     `json:"chain_uid,omitempty" binding:"omitempty,uuid"`
		UserUID       string     `json:"user_uid,omitempty" binding:"uuid"`
		Name          *string    `json:"name,omitempty"`
		PhoneNumber   *string    `json:"phone_number,omitempty"`
		Newsletter    *bool      `json:"newsletter,omitempty"`
		PausedUntil   *time.Time `json:"paused_until,omitempty"`
		ChainPaused   *bool      `json:"chain_paused,omitempty"`
		Sizes         *[]string  `json:"sizes,omitempty"`
		Address       *string    `json:"address,omitempty"`
		I18n          *string    `json:"i18n,omitempty"`
		Latitude      *float64   `json:"latitude,omitempty"`
		Longitude     *float64   `json:"longitude,omitempty"`
		AcceptedLegal *bool      `json:"accepted_legal,omitempty"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.String(http.StatusBadRequest, err.Error())
		return
	}

	ok, user, authUser, chain := auth.AuthenticateUserOfChain(c, db, body.ChainUID, body.UserUID)
	if !ok {
		return
	}
	isAnyChainAdmin := user.IsAnyChainAdmin()

	if body.Sizes != nil {
		if ok := models.ValidateAllSizeEnum(*body.Sizes); !ok {
			c.String(http.StatusBadRequest, "Invalid size enum")
			return
		}
	}

	userChanges := map[string]interface{}{}
	{
		if body.Name != nil {
			userChanges["name"] = *body.Name
		}
		if body.PhoneNumber != nil {
			userChanges["phone_number"] = *body.PhoneNumber
		}
		if body.Address != nil {
			userChanges["address"] = *body.Address
		}
		if body.Latitude != nil {
			userChanges["latitude"] = *body.Latitude
		}
		if body.Longitude != nil {
			userChanges["longitude"] = *body.Longitude
		}
		if body.PausedUntil != nil {
			if body.PausedUntil.After(time.Now()) {
				userChanges["paused_until"] = body.PausedUntil
			} else {
				userChanges["paused_until"] = null.Time{}
				if authUser.ID == user.ID {
					db.Exec(`UPDATE user_chains SET is_paused = FALSE WHERE user_id = ?`, user.ID)
				}
			}
		}
		if body.ChainPaused != nil && chain != nil {
			db.Exec(`UPDATE user_chains SET is_paused = ? WHERE user_id = ? AND chain_id = ?`, *body.ChainPaused, user.ID, chain.ID)
		}
		if body.Sizes != nil {
			j, _ := json.Marshal(body.Sizes)
			userChanges["sizes"] = string(j)
		}
		if body.I18n != nil {
			userChanges["i18n"] = *body.I18n
		}
		if body.AcceptedLegal != nil {
			userChanges["accepted_toh"] = *body.AcceptedLegal
			userChanges["accepted_dpa"] = *body.AcceptedLegal
			if !*body.AcceptedLegal {
				// set as participant for all connected chains
				db.Exec(`UPDATE user_chains SET is_chain_admin = FALSE WHERE user_id = ?`, user.ID)
				// find chains that don't have 1+ host and if connected to this user set to draft
				db.Exec(`
UPDATE chains AS c
JOIN user_chains AS uc ON uc.chain_id = c.id
SET c.open_to_new_members = FALSE, c.published = FALSE
WHERE uc.user_id = ? AND c.id IN (
	SELECT UNIQUE(c2.id)
	FROM chains AS c2
	LEFT JOIN user_chains AS uc2 ON uc2.chain_id = c2.id AND uc2.is_chain_admin = TRUE
	WHERE c2.published = TRUE
	GROUP BY c2.id
	HAVING COUNT(uc2.id) = 0
)
				`, user.ID)
			}
		}
		if len(userChanges) > 0 {
			if err := db.Model(user).Updates(userChanges).Error; err != nil {
				slog.Error("Unable to update user", "err", err)
				c.String(http.StatusInternalServerError, "Unable to update user")
				return
			}
		}
	}

	if body.Newsletter != nil {
		if *body.Newsletter {
			n := &models.Newsletter{
				Email:    user.Email.String,
				Name:     user.Name,
				Verified: true,
			}

			err := n.CreateOrUpdate(db)
			if err != nil {
				slog.Error("", "err", err)
				c.String(http.StatusInternalServerError, "Internal Server Error")
				return
			}
		} else {
			if isAnyChainAdmin {
				c.String(http.StatusConflict, "Newsletter-Box must be checked to create a new loop admin.")
				return
			}

			err := db.Exec("DELETE FROM newsletters WHERE email = ?", user.Email).Error
			if err != nil {
				slog.Error("", "err", err)
				c.String(http.StatusInternalServerError, "Internal Server Error")
				return
			}
			if app.Brevo != nil && user.Email.Valid {
				app.Brevo.DeleteContact(c.Request.Context(), user.Email.String)
			}
		}
	}
}

func UserPurge(c *gin.Context) {
	db := getDB(c)

	var query struct {
		UserUID           string   `form:"user_uid" binding:"required,uuid"`
		ReasonsForLeaving []string `form:"reasons_for_leaving"`
	}
	if err := c.ShouldBindQuery(&query); err != nil {
		c.String(http.StatusBadRequest, err.Error())
		return
	}

	ok, user, _ := auth.Authenticate(c, db, auth.AuthState1AnyUser, "")
	if !ok {
		return
	}
	if user.UID != query.UserUID {
		if user.IsRootAdmin {
			err := db.Raw(`SELECT * FROM users WHERE uid = ? LIMIT 1`, query.UserUID).Scan(user).Error
			if err != nil {
				c.String(http.StatusNotFound, "User not found")
				return
			}
		} else {
			c.String(http.StatusUnauthorized, "Only you can delete your account")
			return
		}
	}

	amountOfBags, err := user.CountAttachedBags(db)
	if err != nil {
		slog.Error("Error getting bag count", "err", err)
	}
	if amountOfBags != 0 {
		c.String(http.StatusConflict, "Please give your bags to someone else, or delete them from the app")
		return
	}

	err = user.AddUserChainsToObject(db)
	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	// find chains where user is the last chain admin
	chainIDsToDelete := []uint{}
	db.Raw(`
SELECT uc.chain_id
FROM  user_chains AS uc
WHERE uc.chain_id IN (
	SELECT uc2.chain_id
	FROM user_chains AS uc2
	WHERE uc2.is_chain_admin = TRUE AND uc2.user_id = ?
) AND uc.is_chain_admin = TRUE
GROUP BY uc.chain_id
HAVING COUNT(uc.id) = 1
	`, user.ID).Scan(&chainIDsToDelete)
	participantsToBeOrphaned := int64(0)
	db.Raw("SELECT COUNT(id) FROM user_chains WHERE chain_id IN ? AND is_chain_admin = FALSE AND is_approved = TRUE AND user_id != ?", chainIDsToDelete, user.ID).Count(&participantsToBeOrphaned)
	if participantsToBeOrphaned > 0 {
		c.String(http.StatusConflict, "Set someone else as host or delete the loop first before deleting your account")
		return
	}
	fmt.Print("im here", query.ReasonsForLeaving)

	deletedUser := models.DeletedUser{
		Email:     user.Email.String,
		CreatedAt: user.CreatedAt,
		DeletedAt: time.Now(),
	}
	if ok := models.ValidateAllReasonsEnum(query.ReasonsForLeaving); !ok {
		c.String(http.StatusBadRequest, models.ErrReasonInvalid.Error())
		return
	}

	for _, reason := range query.ReasonsForLeaving {
		reasons := strings.Split(reason, ",")
		if err := deletedUser.SetReasons(reasons); err != nil {
			c.String(http.StatusBadRequest, err.Error())
			return
		}
	}

	if err := db.Create(&deletedUser).Error; err != nil {
		c.String(http.StatusInternalServerError, "Failed to add deleted user to database")
		return
	}
	tx := db.Begin()

	err = user.DeleteUserChainDependenciesAllChains(tx)
	if err != nil {
		tx.Rollback()
		slog.Error("UserPurge", "err", err)
		c.String(http.StatusInternalServerError, "Unable to disconnect bag connections")
		return
	}

	err = tx.Exec(`
UPDATE events SET user_id = (
	SELECT id FROM users WHERE is_root_admin = 1 LIMIT 1
) WHERE user_id = ?
	`, user.ID).Error
	if err != nil {
		tx.Rollback()
		slog.Error("UserPurge: Unable to remove event connections", "err", err)
		c.String(http.StatusInternalServerError, "Unable to remove event connections")
		return
	}
	err = tx.Exec(`DELETE FROM user_chains WHERE user_id = ?`, user.ID).Error
	if err != nil {
		tx.Rollback()
		slog.Error("UserPurge: Unable to remove loop connections", "err", err)
		c.String(http.StatusInternalServerError, "Unable to remove loop connections")
		return
	}
	err = tx.Exec(`DELETE FROM user_tokens WHERE user_id = ?`, user.ID).Error
	if err != nil {
		tx.Rollback()
		slog.Error("UserPurge: Unable to remove token connections", "err", err)
		c.String(http.StatusInternalServerError, "Unable to remove token connections")
		return
	}
	err = tx.Exec(`DELETE FROM user_onesignals WHERE user_id = ?`, user.ID).Error
	if err != nil {
		tx.Rollback()
		slog.Error("UserPurge: Unable to remove onesignal connections", "err", err)
		c.String(http.StatusInternalServerError, "Unable to remove onesignal connections")
		return
	}
	err = tx.Exec(`DELETE FROM users WHERE id = ?`, user.ID).Error
	if err != nil {
		tx.Rollback()
		slog.Error("UserPurge: Unable to user", "err", err)
		c.String(http.StatusInternalServerError, "Unable to user")
		return
	}

	slog.Info("Purging chains", "chainIDsToDelete", chainIDsToDelete)
	if len(chainIDsToDelete) > 0 {
		err := tx.Exec(`DELETE FROM bags WHERE user_chain_id IN (
			SELECT id FROM user_chains WHERE chain_id IN ?
		)`, chainIDsToDelete).Error
		if err != nil {
			tx.Rollback()
			slog.Error("UserPurge", "err", err)
			c.String(http.StatusInternalServerError, "Unable to disconnect all loop bag connections")
			return
		}
		err = tx.Exec(`DELETE FROM user_chains WHERE chain_id IN ?`, chainIDsToDelete).Error
		if err != nil {
			tx.Rollback()
			slog.Error("UserPurge: Unable to remove hosted loop connections", "err", err)
			c.String(http.StatusInternalServerError, "Unable to remove hosted loop connections")
			return
		}
		err = tx.Exec(`DELETE FROM chains WHERE id IN ?`, chainIDsToDelete).Error
		if err != nil {
			tx.Rollback()
			slog.Error("UserPurge: Unable to remove hosted loop", "err", err)
			c.String(http.StatusInternalServerError, "Unable to remove hosted loop")
			return
		}
	}

	if user.Email.Valid {
		err = tx.Exec(`DELETE FROM newsletters WHERE email = ?`, user.Email.String).Error
		if err != nil {
			tx.Rollback()
			slog.Error("UserPurge: Unable to remove newsletter", "err", err)
			c.String(http.StatusInternalServerError, "Unable to remove newsletter")
			return
		}

		views.EmailAccountDeletedSuccessfully(db, user.I18n, user.Name, user.Email.String)

		if app.Brevo != nil {
			app.Brevo.DeleteContact(c.Request.Context(), user.Email.String)
		}
	}

	tx.Commit()

	// notify connected hosts, send email to chain admins
	chainIDs := []uint{}
	for _, uc := range user.Chains {
		chainIDs = append(chainIDs, uc.ChainID)
	}

	services.EmailLoopAdminsOnUserLeft(db,
		user.Name,
		user.Email.String,
		user.Email.String,
		chainIDs...)

	auth.CookieRemove(c)
}

func UserTransferChain(c *gin.Context) {
	db := getDB(c)

	var body struct {
		TransferUserUID string `json:"transfer_user_uid" binding:"required,uuid"`
		FromChainUID    string `json:"from_chain_uid" binding:"required,uuid"`
		ToChainUID      string `json:"to_chain_uid" binding:"required,uuid"`
		IsCopy          bool   `json:"is_copy"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.String(http.StatusBadRequest, err.Error())
		return
	}

	ok, authUser, authChain := auth.Authenticate(c, db, auth.AuthState3AdminChainUser, body.FromChainUID)
	if !ok {
		return
	}

	if !authUser.IsRootAdmin {
		_, isChainAdmin := authUser.IsPartOfChain(body.ToChainUID)
		if !isChainAdmin {
			c.String(http.StatusUnauthorized, "you must be a host of both loops")
			return
		}
	}
	// finished authentication

	handleError := func(tx *gorm.DB, err error) {
		tx.Rollback()
		slog.Error("UserTransferChain", "err", err)
		responseBody := "Unable transfer user from loop to loop"
		if body.IsCopy {
			responseBody = "Unable copy user from loop to loop"
		}
		c.String(http.StatusInternalServerError, responseBody)
	}
	var err error
	// run in a queue with the ability to rollback on failure, race conditions are mitigated as well
	tx := db.Begin()

	var result struct {
		UserID              uint     `gorm:"user_id"`
		FromChainID         uint     `gorm:"from_chain_id"`
		ToChainID           uint     `gorm:"to_chain_id"`
		ToUserChainIDExists null.Int `gorm:"to_user_chain_exists"`
	}
	row := tx.Raw(`
SELECT u.id as user_id, uc.chain_id as from_chain_id, c2.id as to_chain_id, (
	SELECT uc_dest.id FROM user_chains AS uc_dest
	WHERE uc_dest.chain_id = c2.id AND uc_dest.user_id = u.id
) as to_user_chain_exists
FROM users AS u
JOIN user_chains AS uc ON uc.user_id = u.id AND uc.chain_id = ?
JOIN chains AS c2 ON c2.uid = ?
WHERE u.uid = ?
LIMIT 1
	`, authChain.ID, body.ToChainUID, body.TransferUserUID).Row()
	// For some stupid reason gorm doesn't handle this properly with a simple Scan function
	err = row.Scan(&result.UserID, &result.FromChainID, &result.ToChainID, &result.ToUserChainIDExists)
	if result.UserID == 0 && err == nil {
		err = fmt.Errorf("User %s not found", body.TransferUserUID)
	}
	if err != nil {
		handleError(tx, err)
		return
	}

	uc := &models.UserChain{}
	err = tx.Raw(`SELECT * FROM user_chains WHERE chain_id = ? AND user_id = ? LIMIT 1`, result.FromChainID, result.UserID).Scan(uc).Error

	if uc.ID == 0 || err != nil {
		handleError(tx, fmt.Errorf("User %s not found", body.TransferUserUID))
		return
	}

	// If the user already exists in the destination chain:
	// - on copy instruction:     do nothing
	// - on transfer instruction: remove from source chain
	if result.ToUserChainIDExists.Valid {
		// remove source user_chain and move it's dependencies to destination
		if !body.IsCopy {
			err = tx.Exec(`UPDATE bags SET user_chain_id = ? WHERE user_chain_id = ?`, result.ToUserChainIDExists.Int64, uc.ID).Error
			if err != nil {
				handleError(tx, err)
				return
			}
			err = tx.Exec(`UPDATE bulky_items SET user_chain_id = ? WHERE user_chain_id = ?`, result.ToUserChainIDExists.Int64, uc.ID).Error
			if err != nil {
				handleError(tx, err)
				return
			}
			err = tx.Exec(`DELETE FROM user_chains WHERE user_id = ? AND chain_id = ?`, result.UserID, result.FromChainID).Error
			if err != nil {
				handleError(tx, err)
				return
			}
		}

		err = tx.Commit().Error
		if err != nil {
			handleError(tx, err)
		}
		return
	} else if body.IsCopy {
		// Copy from one chain to another

		err = tx.Create(&models.UserChain{
			UserID:       result.UserID,
			ChainID:      result.ToChainID,
			IsChainAdmin: uc.IsChainAdmin,
			IsApproved:   uc.IsApproved,
		}).Error
		if err != nil {
			tx.Rollback()
			slog.Error("User could not be added to chain", "err", err)
			c.String(http.StatusInternalServerError, "User could not be added to chain due to unknown error")
			return
		}
	} else {
		// Transfer from one chain to another

		err = tx.Exec(`UPDATE user_chains SET chain_id = ?, route_order = 0 WHERE id = ?`, result.ToChainID, uc.ID).Error
		if err != nil {
			handleError(tx, err)
			return
		}
	}

	err = tx.Commit().Error
	if err != nil {
		handleError(tx, err)
		return
	}
}

func UserCheckIfEmailExists(c *gin.Context) {
	db := getDB(c)

	var query struct {
		Email string `form:"email" binding:"required,email"`
	}
	if err := c.ShouldBindQuery(&query); err != nil {
		c.String(http.StatusBadRequest, err.Error())
		return
	}

	_, found, err := models.UserCheckEmail(db, query.Email)
	if err != nil {
		c.String(http.StatusInternalServerError, "Error Checking user email")
		return
	}
	c.JSON(200, found)
}
