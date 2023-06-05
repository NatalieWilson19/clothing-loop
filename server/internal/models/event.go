package models

import (
	"time"

	"gopkg.in/guregu/null.v3"
	"gopkg.in/guregu/null.v3/zero"
)

type EventPrice struct {
	Value    float64 `json:"value"`
	Currency string  `json:"currency"`
}

type Event struct {
	ID             uint        `json:"-"`
	UID            string      `gorm:"uniqueIndex" json:"uid"`
	Name           string      `json:"name"`
	Description    string      `json:"description"`
	Latitude       float64     `json:"latitude"`
	Longitude      float64     `json:"longitude"`
	Address        string      `json:"address"`
	PriceValue     float64     `json:"price_value"`
	PriceCurrency  zero.String `json:"price_currency"`
	Link           string      `json:"link"`
	Date           time.Time   `json:"date"`
	DateEnd        null.Time   `json:"date_end"`
	Genders        []string    `gorm:"serializer:json" json:"genders"`
	ChainID        zero.Int    `json:"-"`
	ChainUID       zero.String `json:"chain_uid" gorm:"-:migration;<-:false"`
	UserID         uint        `json:"-"`
	CreatedAt      time.Time   `json:"-"`
	UpdatedAt      time.Time   `json:"-"`
	UserUID        zero.String `json:"user_uid" gorm:"-:migration;<-:false"`
	UserName       zero.String `json:"user_name" gorm:"-:migration;<-:false"`
	UserEmail      zero.String `json:"user_email" gorm:"-:migration;<-:false"`
	ImageUrl       string      `json:"image_url"`
	ImageDeleteUrl string      `json:"-"`
	ChainName      zero.String `json:"chain_name" gorm:"-:migration;<-:false"`
}

const EventGetSql = `SELECT
events.id                    AS id,
events.uid                   AS uid,
events.name                  AS name,
events.description           AS description,
events.latitude              AS latitude,
events.longitude             AS longitude,
events.address               AS address,
events.price_value           AS price_value,
events.price_currency        AS price_currency,
events.link                  AS link,
events.date                  AS date,
events.date_end              AS date_end,
events.genders               AS genders,
events.chain_id              AS chain_id,
chains.uid                   AS chain_uid,
events.user_id               AS user_id,
events.created_at            AS created_at,
events.updated_at            AS updated_at,
users.uid                    AS user_uid,
users.name                   AS user_name,
users.email                  AS user_email,
events.image_url             AS image_url,
chains.name                  AS chain_name
FROM events
LEFT JOIN chains ON chains.id = chain_id
LEFT JOIN users ON users.id = user_id
`
