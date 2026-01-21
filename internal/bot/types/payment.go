package types

// Invoice contains basic information about an invoice.
type Invoice struct {
	Title          string `json:"title"`
	Description    string `json:"description"`
	StartParameter string `json:"start_parameter"`
	Currency       string `json:"currency"`
	TotalAmount    int    `json:"total_amount"`
}

// SuccessfulPayment contains basic information about a successful payment.
type SuccessfulPayment struct {
	Currency                   string     `json:"currency"`
	TotalAmount                int        `json:"total_amount"`
	InvoicePayload             string     `json:"invoice_payload"`
	SubscriptionExpirationDate int        `json:"subscription_expiration_date,omitempty"`
	IsRecurring                bool       `json:"is_recurring,omitempty"`
	IsFirstRecurring           bool       `json:"is_first_recurring,omitempty"`
	ShippingOptionID           string     `json:"shipping_option_id,omitempty"`
	OrderInfo                  *OrderInfo `json:"order_info,omitempty"`
	TelegramPaymentChargeID    string     `json:"telegram_payment_charge_id"`
	ProviderPaymentChargeID    string     `json:"provider_payment_charge_id"`
}

// RefundedPayment contains basic information about a refunded payment.
type RefundedPayment struct {
	Currency                string `json:"currency"`
	TotalAmount             int    `json:"total_amount"`
	InvoicePayload          string `json:"invoice_payload"`
	TelegramPaymentChargeID string `json:"telegram_payment_charge_id"`
	ProviderPaymentChargeID string `json:"provider_payment_charge_id,omitempty"`
}

// ShippingQuery contains information about an incoming shipping query.
type ShippingQuery struct {
	ID              string          `json:"id"`
	From            User            `json:"from"`
	InvoicePayload  string          `json:"invoice_payload"`
	ShippingAddress ShippingAddress `json:"shipping_address"`
}

// PreCheckoutQuery contains information about an incoming pre-checkout query.
type PreCheckoutQuery struct {
	ID               string     `json:"id"`
	From             User       `json:"from"`
	Currency         string     `json:"currency"`
	TotalAmount      int        `json:"total_amount"`
	InvoicePayload   string     `json:"invoice_payload"`
	ShippingOptionID string     `json:"shipping_option_id,omitempty"`
	OrderInfo        *OrderInfo `json:"order_info,omitempty"`
}

// OrderInfo represents information about an order.
type OrderInfo struct {
	Name            string           `json:"name,omitempty"`
	PhoneNumber     string           `json:"phone_number,omitempty"`
	Email           string           `json:"email,omitempty"`
	ShippingAddress *ShippingAddress `json:"shipping_address,omitempty"`
}

// ShippingAddress represents a shipping address.
type ShippingAddress struct {
	CountryCode string `json:"country_code"`
	State       string `json:"state"`
	City        string `json:"city"`
	StreetLine1 string `json:"street_line1"`
	StreetLine2 string `json:"street_line2"`
	PostCode    string `json:"post_code"`
}

// ShippingOption represents one shipping option.
type ShippingOption struct {
	ID     string         `json:"id"`
	Title  string         `json:"title"`
	Prices []LabeledPrice `json:"prices"`
}

// LabeledPrice represents a portion of the price for goods or services.
type LabeledPrice struct {
	Label  string `json:"label"`
	Amount int    `json:"amount"`
}

// StarTransaction describes a Telegram Star transaction.
type StarTransaction struct {
	ID       string                     `json:"id"`
	Amount   int                        `json:"amount"`
	NanostarAmount int                  `json:"nanostar_amount,omitempty"`
	Date     int                        `json:"date"`
	Source   *TransactionPartner        `json:"source,omitempty"`
	Receiver *TransactionPartner        `json:"receiver,omitempty"`
}

// TransactionPartner describes the source of a transaction or its recipient.
type TransactionPartner struct {
	Type              string      `json:"type"`
	User              *User       `json:"user,omitempty"`
	AffiliateInfo     interface{} `json:"affiliate_info,omitempty"`
	InvoicePayload    string      `json:"invoice_payload,omitempty"`
	SubscriptionPeriod int        `json:"subscription_period,omitempty"`
	PaidMedia         []PaidMedia `json:"paid_media,omitempty"`
	PaidMediaPayload  string      `json:"paid_media_payload,omitempty"`
	Gift              *Gift       `json:"gift,omitempty"`
	WithdrawalState   interface{} `json:"withdrawal_state,omitempty"`
}

// Gift represents a gift that can be sent by the bot.
type Gift struct {
	ID             string   `json:"id"`
	Sticker        Sticker  `json:"sticker"`
	StarCount      int      `json:"star_count"`
	TotalCount     int      `json:"total_count,omitempty"`
	RemainingCount int      `json:"remaining_count,omitempty"`
}

// GiftInfo contains information about a gift.
type GiftInfo struct {
	Gift               Gift            `json:"gift"`
	OwnedGiftID        string          `json:"owned_gift_id,omitempty"`
	ConvertStarCount   int             `json:"convert_star_count,omitempty"`
	PrepaidUpgradeStarCount int        `json:"prepaid_upgrade_star_count,omitempty"`
	CanBeUpgraded      bool            `json:"can_be_upgraded,omitempty"`
	Text               string          `json:"text,omitempty"`
	Entities           []MessageEntity `json:"entities,omitempty"`
	IsPrivate          bool            `json:"is_private,omitempty"`
}

// UniqueGiftInfo contains information about a unique gift.
type UniqueGiftInfo struct {
	Gift       UniqueGift      `json:"gift"`
	Origin     string          `json:"origin"`
	OwnedGiftID string         `json:"owned_gift_id,omitempty"`
	TransferStarCount int      `json:"transfer_star_count,omitempty"`
	Text       string          `json:"text,omitempty"`
	Entities   []MessageEntity `json:"entities,omitempty"`
	IsPrivate  bool            `json:"is_private,omitempty"`
}

// UniqueGift represents a unique gift.
type UniqueGift struct {
	BaseName       string `json:"base_name"`
	Name           string `json:"name"`
	Number         int    `json:"number"`
	Model          UniqueGiftModel `json:"model"`
	Symbol         UniqueGiftSymbol `json:"symbol"`
	Backdrop       UniqueGiftBackdrop `json:"backdrop"`
}

// UniqueGiftModel represents a unique gift model.
type UniqueGiftModel struct {
	Name        string  `json:"name"`
	Sticker     Sticker `json:"sticker"`
	RarityPermille int  `json:"rarity_permille"`
}

// UniqueGiftSymbol represents a unique gift symbol.
type UniqueGiftSymbol struct {
	Name        string  `json:"name"`
	Sticker     Sticker `json:"sticker"`
	RarityPermille int  `json:"rarity_permille"`
}

// UniqueGiftBackdrop represents a unique gift backdrop.
type UniqueGiftBackdrop struct {
	Name           string           `json:"name"`
	Colors         UniqueGiftColors `json:"colors"`
	RarityPermille int              `json:"rarity_permille"`
}

// UniqueGiftColors represents unique gift colors (defined in common.go).
