package types

// Chat represents a chat.
type Chat struct {
	ID               int64  `json:"id"`
	Type             string `json:"type"`
	Title            string `json:"title,omitempty"`
	Username         string `json:"username,omitempty"`
	FirstName        string `json:"first_name,omitempty"`
	LastName         string `json:"last_name,omitempty"`
	IsForum          bool   `json:"is_forum,omitempty"`
	IsDirectMessages bool   `json:"is_direct_messages,omitempty"`
}

// ChatFullInfo contains full information about a chat.
type ChatFullInfo struct {
	ID                                int64                  `json:"id"`
	Type                              string                 `json:"type"`
	Title                             string                 `json:"title,omitempty"`
	Username                          string                 `json:"username,omitempty"`
	FirstName                         string                 `json:"first_name,omitempty"`
	LastName                          string                 `json:"last_name,omitempty"`
	IsForum                           bool                   `json:"is_forum,omitempty"`
	IsDirectMessages                  bool                   `json:"is_direct_messages,omitempty"`
	AccentColorID                     int                    `json:"accent_color_id,omitempty"`
	MaxReactionCount                  int                    `json:"max_reaction_count,omitempty"`
	Photo                             *ChatPhoto             `json:"photo,omitempty"`
	ActiveUsernames                   []string               `json:"active_usernames,omitempty"`
	Birthdate                         *Birthdate             `json:"birthdate,omitempty"`
	BusinessIntro                     *BusinessIntro         `json:"business_intro,omitempty"`
	BusinessLocation                  *BusinessLocation      `json:"business_location,omitempty"`
	BusinessOpeningHours              *BusinessOpeningHours  `json:"business_opening_hours,omitempty"`
	PersonalChat                      *Chat                  `json:"personal_chat,omitempty"`
	ParentChat                        *Chat                  `json:"parent_chat,omitempty"`
	AvailableReactions                []ReactionType         `json:"available_reactions,omitempty"`
	BackgroundCustomEmojiID           string                 `json:"background_custom_emoji_id,omitempty"`
	ProfileAccentColorID              int                    `json:"profile_accent_color_id,omitempty"`
	ProfileBackgroundCustomEmojiID    string                 `json:"profile_background_custom_emoji_id,omitempty"`
	EmojiStatusCustomEmojiID          string                 `json:"emoji_status_custom_emoji_id,omitempty"`
	EmojiStatusExpirationDate         int                    `json:"emoji_status_expiration_date,omitempty"`
	Bio                               string                 `json:"bio,omitempty"`
	HasPrivateForwards                bool                   `json:"has_private_forwards,omitempty"`
	HasRestrictedVoiceAndVideoMessage bool                   `json:"has_restricted_voice_and_video_messages,omitempty"`
	JoinToSendMessages                bool                   `json:"join_to_send_messages,omitempty"`
	JoinByRequest                     bool                   `json:"join_by_request,omitempty"`
	Description                       string                 `json:"description,omitempty"`
	InviteLink                        string                 `json:"invite_link,omitempty"`
	PinnedMessage                     *Message               `json:"pinned_message,omitempty"`
	Permissions                       *ChatPermissions       `json:"permissions,omitempty"`
	AcceptedGiftTypes                 *AcceptedGiftTypes     `json:"accepted_gift_types,omitempty"`
	CanSendPaidMedia                  bool                   `json:"can_send_paid_media,omitempty"`
	SlowModeDelay                     int                    `json:"slow_mode_delay,omitempty"`
	UnrestrictBoostCount              int                    `json:"unrestrict_boost_count,omitempty"`
	MessageAutoDeleteTime             int                    `json:"message_auto_delete_time,omitempty"`
	HasAggressiveAntiSpamEnabled      bool                   `json:"has_aggressive_anti_spam_enabled,omitempty"`
	HasHiddenMembers                  bool                   `json:"has_hidden_members,omitempty"`
	HasProtectedContent               bool                   `json:"has_protected_content,omitempty"`
	HasVisibleHistory                 bool                   `json:"has_visible_history,omitempty"`
	StickerSetName                    string                 `json:"sticker_set_name,omitempty"`
	CanSetStickerSet                  bool                   `json:"can_set_sticker_set,omitempty"`
	CustomEmojiStickerSetName         string                 `json:"custom_emoji_sticker_set_name,omitempty"`
	LinkedChatID                      int64                  `json:"linked_chat_id,omitempty"`
	Location                          *ChatLocation          `json:"location,omitempty"`
	UniqueGiftColors                  *UniqueGiftColors      `json:"unique_gift_colors,omitempty"`
	PaidMessageStarCount              int                    `json:"paid_message_star_count,omitempty"`
}

// ChatPhoto represents a chat photo.
type ChatPhoto struct {
	SmallFileID       string `json:"small_file_id"`
	SmallFileUniqueID string `json:"small_file_unique_id"`
	BigFileID         string `json:"big_file_id"`
	BigFileUniqueID   string `json:"big_file_unique_id"`
}

// ChatPermissions describes actions that a non-administrator user is allowed to take in a chat.
type ChatPermissions struct {
	CanSendMessages       bool `json:"can_send_messages,omitempty"`
	CanSendAudios         bool `json:"can_send_audios,omitempty"`
	CanSendDocuments      bool `json:"can_send_documents,omitempty"`
	CanSendPhotos         bool `json:"can_send_photos,omitempty"`
	CanSendVideos         bool `json:"can_send_videos,omitempty"`
	CanSendVideoNotes     bool `json:"can_send_video_notes,omitempty"`
	CanSendVoiceNotes     bool `json:"can_send_voice_notes,omitempty"`
	CanSendPolls          bool `json:"can_send_polls,omitempty"`
	CanSendOtherMessages  bool `json:"can_send_other_messages,omitempty"`
	CanAddWebPagePreviews bool `json:"can_add_web_page_previews,omitempty"`
	CanChangeInfo         bool `json:"can_change_info,omitempty"`
	CanInviteUsers        bool `json:"can_invite_users,omitempty"`
	CanPinMessages        bool `json:"can_pin_messages,omitempty"`
	CanManageTopics       bool `json:"can_manage_topics,omitempty"`
}

// ChatLocation represents a location to which a chat is connected.
type ChatLocation struct {
	Location Location `json:"location"`
	Address  string   `json:"address"`
}

// ChatBoostAdded represents a service message about a user boosting a chat.
type ChatBoostAdded struct {
	BoostCount int `json:"boost_count"`
}

// ChatBackground represents a chat background.
type ChatBackground struct {
	Type BackgroundType `json:"type"`
}

// ChatShared contains information about a chat shared with the bot.
type ChatShared struct {
	RequestID int   `json:"request_id"`
	ChatID    int64 `json:"chat_id"`
	Title     string `json:"title,omitempty"`
	Username  string `json:"username,omitempty"`
	Photo     []PhotoSize `json:"photo,omitempty"`
}
