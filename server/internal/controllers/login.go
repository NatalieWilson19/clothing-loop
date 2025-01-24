package controllers

import (
	"encoding/base64"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/samber/lo"
	"github.com/the-clothing-loop/website/server/internal/app"
	"github.com/the-clothing-loop/website/server/internal/app/auth"
	"github.com/the-clothing-loop/website/server/internal/models"
	"github.com/the-clothing-loop/website/server/internal/services"
	"github.com/the-clothing-loop/website/server/internal/views"
	ginext "github.com/the-clothing-loop/website/server/pkg/gin_ext"
	"github.com/the-clothing-loop/website/server/sharedtypes"

	"github.com/gin-gonic/gin"
	uuid "github.com/satori/go.uuid"
)

func LoginEmail(c *gin.Context) {
	db := getDB(c)

	var body sharedtypes.LoginEmailRequest
	if err := c.ShouldBindJSON(&body); err != nil {
		c.String(http.StatusBadRequest, "Email required")
		return
	}

	// make sure that this email exists in db
	user, err := models.UserGetByEmail(db, body.Email)
	if err != nil {
		c.String(http.StatusUnauthorized, "Email is not yet registered")
		return
	}

	token, err := auth.OtpCreate(db, user.ID)
	if err != nil {
		ginext.AbortWithErrorInBody(c, http.StatusInternalServerError, err, "Unable to create token")
		return
	}
	if body.Email == app.Config.APPSTORE_REVIEWER_EMAIL {
		c.String(http.StatusOK, token)
		return
	}

	err = views.EmailLoginVerification(c, db, user.Name, *user.Email, token, body.IsApp, body.ChainUID)
	if err != nil {
		ginext.AbortWithErrorInBody(c, http.StatusInternalServerError, err, "Unable to send email")
		return
	}
}

func LoginValidate(c *gin.Context) {
	db := getDB(c)

	var query struct {
		OTP          string `form:"apiKey,required"`
		EmailEncoded string `form:"u,required"`
		ChainUID     string `form:"c" binding:"omitempty,uuid"`
	}
	if err := c.ShouldBindQuery(&query); err != nil {
		c.String(http.StatusBadRequest, "Malformed url: one time password required")
		return
	}
	userEmail, err := base64.StdEncoding.DecodeString(query.EmailEncoded)
	if err != nil {
		c.String(http.StatusBadRequest, "Malformed url: email required")
		return
	}
	user, newToken, err := auth.OtpVerify(db, string(userEmail), query.OTP)
	if err != nil {
		c.String(http.StatusUnauthorized, "Invalid token")
		return
	}

	err = user.AddUserChainsToObject(db)
	if err != nil {
		ginext.AbortWithErrorInBody(c, http.StatusInternalServerError, err, models.ErrAddUserChainsToObject.Error())
		return
	}

	// Is the first time verifying the user account
	if user.Email != nil && !user.IsEmailVerified {
		db.Exec(`UPDATE chains SET published = TRUE WHERE id IN (
			SELECT chain_id FROM user_chains WHERE user_id = ? AND is_chain_admin = TRUE
	   )`, user.ID)

		// Reset joined-at time
		db.Exec(`UPDATE user_chains SET created_at = NOW() WHERE user_id = ?`, user.ID)

		// Add all chains to be notified
		chainIDs := []uint{}
		for _, uc := range user.Chains {
			if !uc.IsChainAdmin {
				chainIDs = append(chainIDs, uc.ChainID)
			}
		}

		if len(chainIDs) > 0 {
			err = services.EmailLoopAdminsOnUserJoin(db, user, chainIDs...)
			if err != nil {
				slog.Error("Unable to send email to associated loop admins", "err", err)
				// This doesn't return because it would be impossible to login if attempting to join a loop without admins.
			}

			chainNames, _ := models.ChainGetNamesByIDs(db, chainIDs...)
			go services.EmailYouSignedUpForLoop(db, user, chainNames...)
		}
	} else if query.ChainUID != "" {
		chainID, found, err := models.ChainCheckIfExist(db, query.ChainUID, true)
		if err != nil {
			ginext.AbortWithErrorInBody(c, http.StatusInternalServerError, err, "Loop does not exist")
			return
		}
		if !found {
			c.String(http.StatusFailedDependency, "Loop does not exist")
			return
		}
		_, found, err = models.UserChainCheckIfRelationExist(db, chainID, user.ID, false)
		if err != nil {
			ginext.AbortWithErrorInBody(c, http.StatusInternalServerError, err, "Loop connection unable to lookup")
			return
		}
		if !found {
			db.Create(&sharedtypes.UserChain{
				UserID:       user.ID,
				ChainID:      chainID,
				IsChainAdmin: false,
				IsApproved:   false,
			})

			chainNames, _ := models.ChainGetNamesByIDs(db, chainID)
			services.EmailYouSignedUpForLoop(db, user, chainNames...)
			services.EmailLoopAdminsOnUserJoin(db, user, chainID)
		}
	}

	// re-add IsEmailVerified, see TokenVerify
	user.IsEmailVerified = true

	// set token as cookie
	auth.CookieSet(c, user.UID, newToken)
	c.JSON(200, gin.H{
		"user":  user,
		"token": newToken,
	})
}

// Sizes and Address is set to the user and the chain
func RegisterChainAdmin(c *gin.Context) {
	db := getDB(c)

	var body sharedtypes.RegisterChainAdminRequest
	if err := c.ShouldBindJSON(&body); err != nil {
		c.String(http.StatusBadRequest, err.Error())
		return
	}

	if ok := models.ValidateAllSizeEnum(body.User.Sizes); !ok {
		c.String(http.StatusBadRequest, models.ErrSizeInvalid.Error())
		return
	}
	if ok := models.ValidateAllSizeEnum(body.Chain.Sizes); !ok {
		c.String(http.StatusBadRequest, models.ErrSizeInvalid.Error())
		return
	}

	if ok := models.ValidateAllGenderEnum(body.Chain.Genders); !ok {
		c.String(http.StatusBadRequest, models.ErrGenderInvalid.Error())
		return
	}

	if !body.User.Newsletter {
		c.String(http.StatusBadRequest, "Newsletter-Box must be checked to create a new loop admin.")
		return
	}
	if !body.Chain.AllowTOH {
		c.String(http.StatusBadRequest, ErrAllowTOHFalse)
		return
	}

	chain := &models.Chain{
		UID:              uuid.NewV4().String(),
		Name:             body.Chain.Name,
		Description:      body.Chain.Description,
		Address:          body.Chain.Address,
		Latitude:         body.Chain.Latitude,
		Longitude:        body.Chain.Longitude,
		Radius:           body.Chain.Radius,
		Published:        false,
		OpenToNewMembers: body.Chain.OpenToNewMembers,
		CountryCode:      body.Chain.CountryCode,
		Sizes:            body.Chain.Sizes,
		Genders:          body.Chain.Genders,
		RoutePrivacy:     2, // default route_privacy
	}
	user := &models.User{
		UID:             uuid.NewV4().String(),
		Email:           &(body.User.Email),
		IsEmailVerified: false,
		IsRootAdmin:     false,
		Name:            body.User.Name,
		PhoneNumber:     body.User.PhoneNumber,
		Sizes:           body.User.Sizes,
		Address:         body.User.Address,
		Latitude:        body.User.Latitude,
		Longitude:       body.User.Longitude,
		AcceptedTOH:     true,
		AcceptedDPA:     true,
	}
	if err := db.Create(user).Error; err != nil {
		slog.Warn("User already exists", "err", err)
		c.String(http.StatusConflict, "User already exists")
		return
	}
	chain.UserChains = []sharedtypes.UserChain{{
		UserID:       user.ID,
		IsChainAdmin: true,
		IsApproved:   true,
	}}
	db.Create(chain)

	db.Create(&models.Newsletter{
		Email:    body.User.Email,
		Name:     body.User.Name,
		Verified: false,
	})

	token, err := auth.OtpCreate(db, user.ID)
	if err != nil {
		ginext.AbortWithErrorInBody(c, http.StatusInternalServerError, err, "Unable to create token")
		return
	}

	go views.EmailRegisterVerification(c, db, user.Name, *user.Email, token, chain.UID)
}

func RegisterBasicUser(c *gin.Context) {
	db := getDB(c)

	var body sharedtypes.RegisterBasicUserRequest
	if err := c.ShouldBindJSON(&body); err != nil {
		c.String(http.StatusBadRequest, err.Error())
		return
	}
	if ok := models.ValidateAllSizeEnum(body.User.Sizes); !ok {
		c.String(http.StatusBadRequest, models.ErrSizeInvalid.Error())
		return
	}

	var chainID uint
	if body.ChainUID != "" {
		var row struct {
			ID uint `gorm:"id"`
		}
		err := db.Raw("SELECT id FROM chains WHERE uid = ? AND deleted_at IS NULL AND open_to_new_members = TRUE LIMIT 1", body.ChainUID).Scan(&row).Error
		chainID = row.ID
		if chainID == 0 {
			slog.Warn("Chain does not exist", "err", err)
			c.String(http.StatusBadRequest, "Chain does not exist")
			return
		}
	}

	user := &models.User{
		UID:             uuid.NewV4().String(),
		Email:           &(body.User.Email),
		IsEmailVerified: false,
		IsRootAdmin:     false,
		Name:            body.User.Name,
		PhoneNumber:     body.User.PhoneNumber,
		Sizes:           body.User.Sizes,
		Address:         body.User.Address,
		Latitude:        body.User.Latitude,
		Longitude:       body.User.Longitude,
	}
	if res := db.Create(user); res.Error != nil {
		slog.Warn("User already exists", "err", res.Error)
		c.String(http.StatusConflict, "User already exists")
		return
	}
	if body.ChainUID != "" {
		db.Create(&sharedtypes.UserChain{
			UserID:       user.ID,
			ChainID:      chainID,
			IsChainAdmin: false,
			IsApproved:   false,
		})
	}
	if body.User.Newsletter {
		n := &models.Newsletter{
			Email:    body.User.Email,
			Name:     body.User.Name,
			Verified: false,
		}
		n.CreateOrUpdate(db)
	}

	token, err := auth.OtpCreate(db, user.ID)
	if err != nil {
		ginext.AbortWithErrorInBody(c, http.StatusInternalServerError, err, "Unable to create token")
		return
	}
	views.EmailRegisterVerification(c, db, user.Name, *user.Email, token, body.ChainUID)
}

func Logout(c *gin.Context) {
	_, ok := auth.TokenReadFromRequest(c)
	if !ok {
		c.String(http.StatusBadRequest, "No token received")
	}

	auth.CookieRemove(c)
}

func RefreshToken(c *gin.Context) {
	db := getDB(c)

	ok, authUser, _ := auth.Authenticate(c, db, auth.AuthState1AnyUser, "")
	if !ok {
		return
	}

	token, err := auth.JwtGenerate(authUser)
	if err != nil {
		c.String(http.StatusUnauthorized, "Invalid token")
		return
	}

	authUser.AddUserChainsToObject(db)

	auth.CookieSet(c, authUser.UID, token)
	c.JSON(http.StatusOK, gin.H{
		"token": token,
		"user":  authUser,
	})
}

func LoginSuperAsGenerateLink(c *gin.Context) {
	db := getDB(c)

	var body sharedtypes.LoginSuperAsGenerateLinkRequest
	if err := c.ShouldBindJSON(&body); err != nil {
		c.String(http.StatusBadRequest, err.Error())
		return
	}

	ok, _, _ := auth.Authenticate(c, db, auth.AuthState4RootUser, "")
	if !ok {
		return
	}

	user, err := models.UserGetByUID(db, body.UserUID, true)
	if err != nil {
		ginext.AbortWithErrorInBody(c, http.StatusInternalServerError, err, "Unable to get user for generating a token")
		return
	}

	token, err := auth.JwtGenerate(user)
	if err != nil {
		ginext.AbortWithErrorInBody(c, http.StatusInternalServerError, err, "Unable to generate token")
		return
	}

	tokenBase64 := base64.URLEncoding.EncodeToString([]byte(token))

	openInPrivateWindowLink := fmt.Sprintf("%s/api/v2/login/super/as?u=%s&t=%s", getBaseUrl(body.IsApp), user.UID, tokenBase64)
	if body.IsApp {
		openInPrivateWindowLink += "&app=true"
	}

	c.JSON(http.StatusOK, openInPrivateWindowLink)
}

func LoginSuperAsRedirect(c *gin.Context) {
	var query struct {
		UserUID string `form:"u" binding:"required,uuid"`
		Token   string `form:"t" binding:"required,base64url"`
		IsApp   bool   `form:"app"`
	}
	if err := c.ShouldBindQuery(&query); err != nil {
		c.String(http.StatusBadRequest, err.Error())
		return
	}

	b, err := base64.URLEncoding.DecodeString(query.Token)
	if err != nil {
		ginext.AbortWithErrorInBody(c, http.StatusBadRequest, err, "Token is not valid base64")
		return
	}
	token := string(b)

	db := getDB(c)

	_, err = models.UserGetByUID(db, query.UserUID, true)
	if err != nil {
		ginext.AbortWithErrorInBody(c, http.StatusExpectationFailed, err, "User not found")
		return
	}

	auth.CookieSet(c, query.UserUID, token)
	c.Redirect(http.StatusTemporaryRedirect, fmt.Sprintf("%s/", getBaseUrl(query.IsApp)))
}

func getBaseUrl(isApp bool) (baseUrl string) {
	switch app.Config.ENV {
	case app.EnvEnumProduction:
		baseUrl = lo.If(isApp, "https://app.clothingloop.org").Else("https://www.clothingloop.org")
	case app.EnvEnumAcceptance:
		baseUrl = lo.If(isApp, "https://app.acc.clothingloop.org").Else("https://acc.clothingloop.org")
	default:
		baseUrl = lo.If(isApp, "http://localhost:5173").Else("http://localhost:3000")
	}
	return baseUrl
}
