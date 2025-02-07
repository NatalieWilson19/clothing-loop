package controllers

import (
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"

	"github.com/the-clothing-loop/website/server/internal/app"
	"github.com/the-clothing-loop/website/server/internal/models"
	ginext "github.com/the-clothing-loop/website/server/pkg/gin_ext"
	"github.com/the-clothing-loop/website/server/sharedtypes"

	"github.com/gin-gonic/gin"
	"github.com/stripe/stripe-go/v73"
	stripe_session "github.com/stripe/stripe-go/v73/checkout/session"
	stripe_webhook "github.com/stripe/stripe-go/v73/webhook"
	"gopkg.in/guregu/null.v3/zero"
)

// Rewrite of https://github.com/CollActionteam/clothing-loop/blob/e5d09d38d72bb42f531d0dc0ec7a5b18459bcbcd/firebase/functions/src/payments.ts#L18
func PaymentsInitiate(c *gin.Context) {
	var body sharedtypes.PaymentsInitiateRequest
	if err := c.ShouldBindJSON(&body); err != nil {
		c.String(http.StatusBadRequest, "Email required")
		return
	}

	db := getDB(c)

	name := "Donation"
	successURL := app.Config.SITE_BASE_URL_FE + "/donate/thankyou"
	cancelURL := app.Config.SITE_BASE_URL_FE + "/donate/cancel"

	checkout := new(stripe.CheckoutSessionParams)

	if body.IsRecurring {
		checkout.PaymentMethodTypes = stripe.StringSlice([]string{
			string(stripe.PaymentMethodTypeSEPADebit),
			string(stripe.PaymentMethodTypeCard),
		})
		checkout.Mode = stripe.String(string(stripe.CheckoutSessionModeSubscription))
		checkout.LineItems = []*stripe.CheckoutSessionLineItemParams{
			{
				Price:    &body.PriceID,
				Quantity: stripe.Int64(1),
			},
		}
	} else {
		checkout.PaymentMethodTypes = stripe.StringSlice([]string{
			string(stripe.PaymentMethodTypeIDEAL),
			string(stripe.PaymentMethodTypeCard),
		})
		checkout.Mode = stripe.String(string(stripe.CheckoutSessionModePayment))
		checkout.LineItems = []*stripe.CheckoutSessionLineItemParams{
			{
				PriceData: &stripe.CheckoutSessionLineItemPriceDataParams{
					Currency: stripe.String(string(stripe.CurrencyEUR)),
					ProductData: &stripe.CheckoutSessionLineItemPriceDataProductDataParams{
						Name: &name,
					},
					UnitAmount: stripe.Int64(body.PriceCents),
				},
				Quantity: stripe.Int64(1),
			},
		}
	}

	checkout.SuccessURL = &successURL
	checkout.CancelURL = &cancelURL

	if body.Email != "" {
		checkout.CustomerEmail = stripe.String(body.Email)
	}

	session, err := stripe_session.New(checkout)
	if err != nil {
		slog.Warn("Something went wrong when processing your checkout request", "err", err)
		c.String(http.StatusUnavailableForLegalReasons, "Something went wrong when processing your checkout request...")
		return
	}

	if err := db.Create(&models.Payment{
		Amount:          float32(session.AmountTotal) / 100,
		Email:           body.Email,
		IsRecurring:     body.IsRecurring,
		SessionStripeID: zero.StringFrom(session.ID),
		Status:          string(session.Status),
	}).Error; err != nil {
		ginext.AbortWithErrorInBody(c, http.StatusInternalServerError, err, "Unable to add payment to database")
		return
	}

	c.JSON(200, sharedtypes.PaymentsInitiateResponse{
		SessionID: session.ID,
	})
}

// Rewrite of https://github.com/CollActionteam/clothing-loop/blob/e5d09d38d72bb42f531d0dc0ec7a5b18459bcbcd/firebase/functions/src/payments.ts#L99
func PaymentsWebhook(c *gin.Context) {
	const MaxBodyBytes = int64(65536)

	signature := c.Request.Header.Get("stripe-signature")
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, MaxBodyBytes)
	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		slog.Warn("Body does not exist", "err", err)
		c.String(400, "Body does not exist")
		return
	}
	event, err := stripe_webhook.ConstructEvent(body, signature, app.Config.STRIPE_WEBHOOK)
	if err != nil {
		slog.Warn("Webhook Error", "err", err)
		c.String(400, fmt.Sprintf("Webhook Error: %s", err))
		return
	}

	slog.Info(fmt.Sprintf("event.Type: %+v", event.Type))

	switch event.Type {
	case "checkout.session.completed":
		paymentsWebhookCheckoutSessionCompleted(c, event)
	default:
		c.JSON(200, gin.H{"received": true})
	}
}

func paymentsWebhookCheckoutSessionCompleted(c *gin.Context, event stripe.Event) {
	db := getDB(c)

	session := new(stripe.CheckoutSession)
	err := json.Unmarshal(event.Data.Raw, session)
	if err != nil {
		ginext.AbortWithErrorInBody(c, http.StatusInternalServerError, err, "Incorrect response from stripe")
		return
	}

	err = db.Model(&models.Payment{}).Where("session_stripe_id = ?", session.ID).UpdateColumns(&models.Payment{
		Status: string(session.Status),
		Email:  session.CustomerEmail,
	}).Error
	if err != nil {
		ginext.AbortWithErrorInBody(c, http.StatusInternalServerError, err, "Unable to update payment in database")
		return
	}

	c.JSON(200, gin.H{"received": true})
}
