package telegram

import (
	"strings"
	"time"
)

// Default configuration values for the poller.
const (
	DefaultPollInterval    = 100 * time.Millisecond
	DefaultTimeout         = 30 * time.Second
	DefaultMaxRetries      = 5
	DefaultInitialBackoff  = 1 * time.Second
	DefaultMaxBackoff      = 60 * time.Second
	DefaultBackoffFactor   = 2.0
	DefaultUpdatesChanSize = 100
)

// AllowedUpdateType represents the types of updates the bot can receive.
type AllowedUpdateType string

// Allowed update types as per Telegram Bot API.
const (
	UpdateTypeMessage              AllowedUpdateType = "message"
	UpdateTypeEditedMessage        AllowedUpdateType = "edited_message"
	UpdateTypeChannelPost          AllowedUpdateType = "channel_post"
	UpdateTypeEditedChannelPost    AllowedUpdateType = "edited_channel_post"
	UpdateTypeBusinessConnection   AllowedUpdateType = "business_connection"
	UpdateTypeBusinessMessage      AllowedUpdateType = "business_message"
	UpdateTypeEditedBusinessMsg    AllowedUpdateType = "edited_business_message"
	UpdateTypeDeletedBusinessMsgs  AllowedUpdateType = "deleted_business_messages"
	UpdateTypeMessageReaction      AllowedUpdateType = "message_reaction"
	UpdateTypeMessageReactionCount AllowedUpdateType = "message_reaction_count"
	UpdateTypeInlineQuery          AllowedUpdateType = "inline_query"
	UpdateTypeChosenInlineResult   AllowedUpdateType = "chosen_inline_result"
	UpdateTypeCallbackQuery        AllowedUpdateType = "callback_query"
	UpdateTypeShippingQuery        AllowedUpdateType = "shipping_query"
	UpdateTypePreCheckoutQuery     AllowedUpdateType = "pre_checkout_query"
	UpdateTypePurchasedPaidMedia   AllowedUpdateType = "purchased_paid_media"
	UpdateTypePoll                 AllowedUpdateType = "poll"
	UpdateTypePollAnswer           AllowedUpdateType = "poll_answer"
	UpdateTypeMyChatMember         AllowedUpdateType = "my_chat_member"
	UpdateTypeChatMember           AllowedUpdateType = "chat_member"
	UpdateTypeChatJoinRequest      AllowedUpdateType = "chat_join_request"
	UpdateTypeChatBoost            AllowedUpdateType = "chat_boost"
	UpdateTypeRemovedChatBoost     AllowedUpdateType = "removed_chat_boost"
)

// String returns a string representation of AllowedUpdateType.
func (t AllowedUpdateType) String() string {
	return string(t)
}

// AllAllowedUpdates returns a slice of all supported update types.
func AllAllowedUpdates() []AllowedUpdateType {
	return []AllowedUpdateType{
		UpdateTypeMessage,
		UpdateTypeEditedMessage,
		UpdateTypeChannelPost,
		UpdateTypeEditedChannelPost,
		UpdateTypeBusinessConnection,
		UpdateTypeBusinessMessage,
		UpdateTypeEditedBusinessMsg,
		UpdateTypeDeletedBusinessMsgs,
		UpdateTypeMessageReaction,
		UpdateTypeMessageReactionCount,
		UpdateTypeInlineQuery,
		UpdateTypeChosenInlineResult,
		UpdateTypeCallbackQuery,
		UpdateTypeShippingQuery,
		UpdateTypePreCheckoutQuery,
		UpdateTypePurchasedPaidMedia,
		UpdateTypePoll,
		UpdateTypePollAnswer,
		UpdateTypeMyChatMember,
		UpdateTypeChatMember,
		UpdateTypeChatJoinRequest,
		UpdateTypeChatBoost,
		UpdateTypeRemovedChatBoost,
	}
}

// CommonAllowedUpdates returns a slice of commonly used update types.
func CommonAllowedUpdates() []AllowedUpdateType {
	return []AllowedUpdateType{
		UpdateTypeMessage,
		UpdateTypeEditedMessage,
		UpdateTypeCallbackQuery,
		UpdateTypeInlineQuery,
		UpdateTypeChosenInlineResult,
	}
}

// ParseAllowedUpdates parses a comma-separated string into AllowedUpdateTypes.
func ParseAllowedUpdates(s string) []AllowedUpdateType {
	if s == "" {
		return nil
	}

	parts := strings.Split(s, ",")
	result := make([]AllowedUpdateType, 0, len(parts))

	for _, part := range parts {
		trimmed := strings.TrimSpace(part)
		if trimmed != "" {
			result = append(result, AllowedUpdateType(trimmed))
		}
	}

	return result
}
