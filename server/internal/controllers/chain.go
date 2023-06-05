package controllers

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/the-clothing-loop/website/server/internal/app/auth"
	"github.com/the-clothing-loop/website/server/internal/app/goscope"
	"github.com/the-clothing-loop/website/server/internal/models"
	"github.com/the-clothing-loop/website/server/internal/views"

	"github.com/gin-gonic/gin"
	uuid "github.com/satori/go.uuid"
	"gopkg.in/guregu/null.v3/zero"
)

const (
	UnapprovedReasonOther        = "other"
	UnapprovedReasonOutOfAria    = "out_of_aria"
	UnapprovedReasonSizesGenders = "sizes_genders"
)

type ChainCreateRequestBody struct {
	Name             string   `json:"name" binding:"required"`
	Description      string   `json:"description"`
	Address          string   `json:"address" binding:"required"`
	Latitude         float64  `json:"latitude" binding:"required"`
	Longitude        float64  `json:"longitude" binding:"required"`
	Radius           float32  `json:"radius" binding:"required,gte=1.0,lte=70.0"`
	OpenToNewMembers bool     `json:"open_to_new_members" binding:"required"`
	Sizes            []string `json:"sizes" binding:"required"`
	Genders          []string `json:"genders" binding:"required"`
}

func ChainCreate(c *gin.Context) {
	db := getDB(c)
	ok, user, _ := auth.Authenticate(c, db, auth.AuthState1AnyUser, "")
	if !ok {
		return
	}

	var body ChainCreateRequestBody
	if err := c.ShouldBindJSON(&body); err != nil {
		c.String(http.StatusBadRequest, err.Error())
		return
	}
	if ok := models.ValidateAllSizeEnum(body.Sizes); !ok {
		c.String(http.StatusBadRequest, models.ErrSizeInvalid.Error())
		return
	}
	if ok := models.ValidateAllGenderEnum(body.Genders); !ok {
		c.String(http.StatusBadRequest, models.ErrGenderInvalid.Error())
		return
	}

	chain := models.Chain{
		UID:              uuid.NewV4().String(),
		Name:             body.Name,
		Description:      body.Description,
		Address:          body.Address,
		Latitude:         body.Latitude,
		Longitude:        body.Longitude,
		Radius:           body.Radius,
		Published:        true,
		OpenToNewMembers: true,
		Sizes:            body.Sizes,
		Genders:          body.Genders,
		UserChains: []models.UserChain{
			{
				UserID:       user.ID,
				IsChainAdmin: true,
				IsApproved:   true,
				RouteOrder:   0,
			},
		},
	}
	if err := db.Create(&chain).Error; err != nil {
		goscope.Log.Warningf("Unable to create chain: %v", err)
		c.String(http.StatusInternalServerError, "Unable to create chain")
		return
	}

	c.Status(200)
}

func ChainGet(c *gin.Context) {
	db := getDB(c)

	var query struct {
		ChainUID string `form:"chain_uid" binding:"required"`
	}
	if err := c.ShouldBindQuery(&query); err != nil {
		c.String(http.StatusBadRequest, err.Error())
		return
	}

	chain := &models.Chain{}
	err := db.Raw(`SELECT * FROM chains WHERE uid = ? LIMIT 1`, query.ChainUID).Scan(chain).Error
	if err != nil || chain.ID == 0 {
		c.String(http.StatusBadRequest, models.ErrChainNotFound.Error())
		return
	}

	c.JSON(200, gin.H{
		"uid":                 chain.UID,
		"name":                chain.Name,
		"description":         chain.Description,
		"address":             chain.Address,
		"latitude":            chain.Latitude,
		"longitude":           chain.Longitude,
		"radius":              chain.Radius,
		"sizes":               chain.Sizes,
		"genders":             chain.Genders,
		"published":           chain.Published,
		"open_to_new_members": chain.OpenToNewMembers,
	})
}

func ChainGetAll(c *gin.Context) {
	db := getDB(c)

	var query struct {
		FilterSizes     []string `form:"filter_sizes"`
		FilterGenders   []string `form:"filter_genders"`
		FilterPublished bool     `form:"filter_out_unpublished"`
	}
	if err := c.ShouldBindQuery(&query); err != nil && err != io.EOF {
		c.String(http.StatusBadRequest, err.Error())
		return
	}

	if ok := models.ValidateAllSizeEnum(query.FilterSizes); !ok {
		c.String(http.StatusBadRequest, models.ErrSizeInvalid.Error())
		return
	}
	if ok := models.ValidateAllGenderEnum(query.FilterGenders); !ok {
		c.String(http.StatusBadRequest, models.ErrGenderInvalid.Error())
		return
	}

	chains := []models.Chain{}
	sql := "SELECT * FROM chains"
	whereOrSql := []string{}
	args := []any{}

	// filter sizes and genders
	isGendersEmpty := len(query.FilterGenders) == 0
	isSizesEmpty := len(query.FilterSizes) == 0
	if !isSizesEmpty || !isGendersEmpty {
		var whereAndSql []string
		if !isSizesEmpty {
			for _, size := range query.FilterSizes {
				whereAndSql = append(whereAndSql, "chains.sizes LIKE ?")
				args = append(args, fmt.Sprintf("%%%s%%", size))
			}
		}
		if !isGendersEmpty {
			for _, gender := range query.FilterGenders {
				whereAndSql = append(whereAndSql, "chains.genders LIKE ?")
				args = append(args, fmt.Sprintf("%%%s%%", gender))
			}
		}

		whereOrSql = append(whereOrSql, fmt.Sprintf("( %s )", strings.Join(whereAndSql, " AND ")))
	}

	if query.FilterPublished {
		whereOrSql = append(whereOrSql, "chains.published = TRUE")
	}
	if len(whereOrSql) > 0 {
		sql = fmt.Sprintf("%s WHERE %s", sql, strings.Join(whereOrSql, " OR "))
	}
	if err := db.Raw(sql, args...).Scan(&chains).Error; err != nil {
		goscope.Log.Warningf("Chain not found: %v", err)
		c.String(http.StatusBadRequest, models.ErrChainNotFound.Error())
		return
	}

	chainsJson := []*gin.H{}
	for _, chain := range chains {
		chainsJson = append(chainsJson, &gin.H{
			"uid":                 chain.UID,
			"name":                chain.Name,
			"description":         chain.Description,
			"address":             chain.Address,
			"latitude":            chain.Latitude,
			"longitude":           chain.Longitude,
			"radius":              chain.Radius,
			"sizes":               chain.Sizes,
			"genders":             chain.Genders,
			"published":           chain.Published,
			"open_to_new_members": chain.OpenToNewMembers,
		})
	}

	c.JSON(200, chainsJson)
}

func ChainUpdate(c *gin.Context) {
	db := getDB(c)

	var body struct {
		UID              string    `json:"uid" binding:"required"`
		Name             *string   `json:"name,omitempty"`
		Description      *string   `json:"description,omitempty"`
		Address          *string   `json:"address,omitempty"`
		Latitude         *float32  `json:"latitude,omitempty"`
		Longitude        *float32  `json:"longitude,omitempty"`
		Radius           *float32  `json:"radius,omitempty" binding:"omitempty,gte=1.0,lte=70.0"`
		Sizes            *[]string `json:"sizes,omitempty"`
		Genders          *[]string `json:"genders,omitempty"`
		Route            *[]string `json:"route,omitempty"`
		Published        *bool     `json:"published,omitempty"`
		OpenToNewMembers *bool     `json:"open_to_new_members,omitempty"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.String(http.StatusBadRequest, err.Error())
		return
	}

	if body.Sizes != nil {
		if ok := models.ValidateAllSizeEnum(*(body.Sizes)); !ok {
			c.String(http.StatusBadRequest, models.ErrSizeInvalid.Error())
			return
		}
	}
	if body.Genders != nil {
		if ok := models.ValidateAllGenderEnum(*(body.Genders)); !ok {
			c.String(http.StatusBadRequest, models.ErrGenderInvalid.Error())
			return
		}
	}

	ok, _, chain := auth.Authenticate(c, db, auth.AuthState3AdminChainUser, body.UID)
	if !ok {
		return
	}

	valuesToUpdate := map[string]any{}
	if body.Name != nil {
		valuesToUpdate["name"] = *(body.Name)
	}
	if body.Description != nil {
		valuesToUpdate["description"] = *(body.Description)
	}
	if body.Address != nil {
		valuesToUpdate["address"] = *(body.Address)
	}
	if body.Latitude != nil {
		valuesToUpdate["latitude"] = *(body.Latitude)
	}
	if body.Longitude != nil {
		valuesToUpdate["longitude"] = *(body.Longitude)
	}
	if body.Radius != nil {
		valuesToUpdate["radius"] = *(body.Radius)
	}
	if body.Sizes != nil {
		j, _ := json.Marshal(body.Sizes)
		valuesToUpdate["sizes"] = string(j)
	}
	if body.Genders != nil {
		j, _ := json.Marshal(body.Genders)
		valuesToUpdate["genders"] = string(j)
	}
	if body.Published != nil {
		valuesToUpdate["published"] = *(body.Published)
	}
	if body.OpenToNewMembers != nil {
		valuesToUpdate["open_to_new_members"] = *(body.OpenToNewMembers)
	}

	err := db.Model(chain).Updates(valuesToUpdate).Error
	if err != nil {
		goscope.Log.Errorf("Unable to update loop values: %v", err)
		c.String(http.StatusInternalServerError, "Unable to update loop values")
	}
}

func ChainAddUser(c *gin.Context) {
	db := getDB(c)

	var body struct {
		UserUID      string `json:"user_uid" binding:"required,uuid"`
		ChainUID     string `json:"chain_uid" binding:"required,uuid"`
		IsChainAdmin bool   `json:"is_chain_admin"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.String(http.StatusBadRequest, err.Error())
		return
	}

	var ok bool
	var chain *models.Chain
	if body.IsChainAdmin {
		ok, _, chain = auth.Authenticate(c, db, auth.AuthState3AdminChainUser, body.ChainUID)
	} else {
		ok, _, _, chain = auth.AuthenticateUserOfChain(c, db, body.ChainUID, body.UserUID)
	}
	if !ok {
		return
	}
	if !chain.OpenToNewMembers {
		c.String(http.StatusConflict, "Loop is not open to new members")
		return
	}

	user := models.User{}
	db.Raw(`
SELECT * FROM users
WHERE uid = ? AND is_email_verified = TRUE
LIMIT 1
	`, body.UserUID).Scan(&user)
	if user.ID == 0 {
		c.String(http.StatusBadRequest, models.ErrUserNotFound.Error())
		return
	}

	userChain := &models.UserChain{}
	db.Raw(`
SELECT * FROM user_chains
WHERE user_id = ? AND chain_id = ?
LIMIT 1
	`, user.ID, chain.ID).Scan(userChain)
	if userChain.ID != 0 {
		if (!userChain.IsChainAdmin && body.IsChainAdmin) || (userChain.IsChainAdmin && !body.IsChainAdmin) {
			userChain.IsChainAdmin = body.IsChainAdmin
			db.Save(userChain)
		}
	} else {
		if err := db.Create(&models.UserChain{
			UserID:       user.ID,
			ChainID:      chain.ID,
			IsChainAdmin: false,
		}).Error; err != nil {
			goscope.Log.Errorf("User could not be added to chain: %v", err)
			c.String(http.StatusInternalServerError, "User could not be added to chain due to unknown error")
			return
		}

		// find admin users related to the chain to email
		results := []struct {
			Name  string
			Email zero.String
			Chain string
		}{}
		db.Raw(`
SELECT users.name AS name, users.email AS email, chains.name AS chain
FROM user_chains AS uc
LEFT JOIN users ON uc.user_id = users.id 
LEFT JOIN chains ON chains.id = uc.chain_id
WHERE uc.chain_id = ?
	AND uc.is_chain_admin = TRUE
	AND users.is_email_verified = TRUE
`, chain.ID).Scan(&results)

		if len(results) == 0 {
			goscope.Log.Errorf("Empty chain that is still public: ChainID: %d", chain.ID)
			c.String(http.StatusInternalServerError, "No admins exist for this loop")
			return
		}

		for _, result := range results {
			if result.Email.Valid {
				go views.EmailAParticipantJoinedTheLoop(
					c,
					db,
					result.Email.String,
					result.Name,
					result.Chain,
					user.Name,
					user.Email.String,
					user.PhoneNumber,
					user.Address,
				)
			}
		}
	}
}

func ChainRemoveUser(c *gin.Context) {
	db := getDB(c)

	var body struct {
		UserUID  string `json:"user_uid" binding:"required,uuid"`
		ChainUID string `json:"chain_uid" binding:"required,uuid"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.String(http.StatusBadRequest, err.Error())
		return
	}

	ok, user, authUser, chain := auth.AuthenticateUserOfChain(c, db, body.ChainUID, body.UserUID)
	if !ok {
		return
	}

	if authUser.ID == user.ID && !authUser.IsRootAdmin {
		if _, isChainAdmin := user.IsPartOfChain(chain.UID); isChainAdmin {
			amountChainAdmins := -1
			err := db.Raw(`SELECT COUNT(*) FROM user_chains WHERE chain_id = ? AND is_chain_admin = TRUE`, chain.ID).Scan(&amountChainAdmins).Error

			if amountChainAdmins == 1 {
				goscope.Log.Warningf("Unable to remove last host of loop: %v", err)
				c.String(http.StatusConflict, "Unable to remove last host of loop")
				return
			}
		}
	}

	err := chain.RemoveUser(db, user.ID)
	if err != nil {
		goscope.Log.Errorf("User could not be removed from chain: %v", err)
		c.String(http.StatusInternalServerError, "User could not be removed from chain due to unknown error")
		return
	}
}

func ChainApproveUser(c *gin.Context) {
	db := getDB(c)

	var body struct {
		UserUID  string `json:"user_uid" binding:"required,uuid"`
		ChainUID string `json:"chain_uid" binding:"required,uuid"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.String(http.StatusBadRequest, err.Error())
		return
	}

	ok, user, _, chain := auth.AuthenticateUserOfChain(c, db, body.ChainUID, body.UserUID)
	if !ok {
		return
	}

	db.Exec(`
UPDATE user_chains
SET is_approved = TRUE, created_at = NOW()
WHERE user_id = ? AND chain_id = ?
	`, user.ID, chain.ID)

	if user.Email.Valid {
		views.EmailToLoopParticipant(c, db, user.Name, user.Email.String, chain.Name,
			"",
			"an_admin_approved_your_join_request",
			"an_admin_approved_your_join_request.gohtml",
		)
	}
}

func ChainDeleteUnapproved(c *gin.Context) {
	db := getDB(c)
	var query struct {
		UserUID  string `form:"user_uid" binding:"required,uuid"`
		ChainUID string `form:"chain_uid" binding:"required,uuid"`
		Reason   string `form:"reason" binding:"required,oneof='other' 'too_far_away' 'sizes_genders'"`
	}
	if err := c.ShouldBindQuery(&query); err != nil {
		c.String(http.StatusBadRequest, err.Error())
		return
	}

	ok, user, _, chain := auth.AuthenticateUserOfChain(c, db, query.ChainUID, query.UserUID)
	if !ok {
		return
	}

	err := chain.RemoveUserUnapproved(db, user.ID)
	if err != nil {
		goscope.Log.Errorf("User could not be removed from chain: %v", err)
		c.String(http.StatusInternalServerError, "User could not be removed from chain due to unknown error")
		return
	}

	if user.Email.Valid {
		views.EmailToLoopParticipant(c, db, user.Name, user.Email.String, chain.Name,
			query.Reason,
			"an_admin_denied_your_join_request",
			"an_admin_denied_your_join_request.gohtml",
		)
	}

}
