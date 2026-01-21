package types

// MessageEntity represents one special entity in a text message.
type MessageEntity struct {
	Type          string `json:"type"`
	Offset        int    `json:"offset"`
	Length        int    `json:"length"`
	URL           string `json:"url,omitempty"`
	User          *User  `json:"user,omitempty"`
	Language      string `json:"language,omitempty"`
	CustomEmojiID string `json:"custom_emoji_id,omitempty"`
}

// LinkPreviewOptions describes options for link preview generation.
type LinkPreviewOptions struct {
	IsDisabled       bool   `json:"is_disabled,omitempty"`
	URL              string `json:"url,omitempty"`
	PreferSmallMedia bool   `json:"prefer_small_media,omitempty"`
	PreferLargeMedia bool   `json:"prefer_large_media,omitempty"`
	ShowAboveText    bool   `json:"show_above_text,omitempty"`
}

// TextQuote contains information about the quoted part of a message that is replied to.
type TextQuote struct {
	Text     string          `json:"text"`
	Entities []MessageEntity `json:"entities,omitempty"`
	Position int             `json:"position"`
	IsManual bool            `json:"is_manual,omitempty"`
}

// ExternalReplyInfo contains information about a message that is being replied to.
type ExternalReplyInfo struct {
	Origin             MessageOrigin       `json:"origin"`
	Chat               *Chat               `json:"chat,omitempty"`
	MessageID          int                 `json:"message_id,omitempty"`
	LinkPreviewOptions *LinkPreviewOptions `json:"link_preview_options,omitempty"`
	Animation          *Animation          `json:"animation,omitempty"`
	Audio              *Audio              `json:"audio,omitempty"`
	Document           *Document           `json:"document,omitempty"`
	PaidMedia          *PaidMediaInfo      `json:"paid_media,omitempty"`
	Photo              []PhotoSize         `json:"photo,omitempty"`
	Sticker            *Sticker            `json:"sticker,omitempty"`
	Story              *Story              `json:"story,omitempty"`
	Video              *Video              `json:"video,omitempty"`
	VideoNote          *VideoNote          `json:"video_note,omitempty"`
	Voice              *Voice              `json:"voice,omitempty"`
	HasMediaSpoiler    bool                `json:"has_media_spoiler,omitempty"`
	Contact            *Contact            `json:"contact,omitempty"`
	Dice               *Dice               `json:"dice,omitempty"`
	Game               *Game               `json:"game,omitempty"`
	Giveaway           *Giveaway           `json:"giveaway,omitempty"`
	GiveawayWinners    *GiveawayWinners    `json:"giveaway_winners,omitempty"`
	Invoice            *Invoice            `json:"invoice,omitempty"`
	Location           *Location           `json:"location,omitempty"`
	Poll               *Poll               `json:"poll,omitempty"`
	Venue              *Venue              `json:"venue,omitempty"`
}

// MessageOrigin describes the origin of a message.
type MessageOrigin struct {
	Type            string `json:"type"`
	Date            int    `json:"date"`
	SenderUser      *User  `json:"sender_user,omitempty"`
	SenderUserName  string `json:"sender_user_name,omitempty"`
	SenderChat      *Chat  `json:"sender_chat,omitempty"`
	AuthorSignature string `json:"author_signature,omitempty"`
	Chat            *Chat  `json:"chat,omitempty"`
	MessageID       int    `json:"message_id,omitempty"`
}

// MessageAutoDeleteTimerChanged represents a service message about a change in auto-delete timer settings.
type MessageAutoDeleteTimerChanged struct {
	MessageAutoDeleteTime int `json:"message_auto_delete_time"`
}

// MaybeInaccessibleMessage describes a message that might or might not be accessible.
type MaybeInaccessibleMessage struct {
	Chat      Chat `json:"chat"`
	MessageID int  `json:"message_id"`
	Date      int  `json:"date"`
}

// DirectMessagesTopic represents a topic in the direct messages chat.
type DirectMessagesTopic struct {
	ID         int    `json:"id"`
	ThreadID   int    `json:"thread_id"`
	Name       string `json:"name"`
	IconColor  int    `json:"icon_color,omitempty"`
	IconCustomEmojiID string `json:"icon_custom_emoji_id,omitempty"`
}

// SuggestedPostInfo contains information about a suggested post.
type SuggestedPostInfo struct {
	SuggestID             int     `json:"suggest_id"`
	SenderChat            Chat    `json:"sender_chat"`
	ScheduleDate          int     `json:"schedule_date,omitempty"`
	RequestedStarCount    int     `json:"requested_star_count,omitempty"`
	AwardedStarCount      int     `json:"awarded_star_count,omitempty"`
}

// Checklist represents a checklist.
type Checklist struct {
	Title string          `json:"title"`
	Tasks []ChecklistTask `json:"tasks"`
}

// ChecklistTask represents a checklist task.
type ChecklistTask struct {
	ID        int    `json:"id"`
	Text      string `json:"text"`
	Completed bool   `json:"completed"`
}

// ChecklistTasksDone represents service message about checklist tasks done.
type ChecklistTasksDone struct {
	TasksCompleted int `json:"tasks_completed"`
}

// ChecklistTasksAdded represents service message about checklist tasks added.
type ChecklistTasksAdded struct {
	TasksAdded int `json:"tasks_added"`
}

// DirectMessagePriceChanged represents service message about direct message price change.
type DirectMessagePriceChanged struct {
	OldPrice int `json:"old_price"`
	NewPrice int `json:"new_price"`
}

// PaidMessagePriceChanged represents service message about paid message price change.
type PaidMessagePriceChanged struct {
	StarCount int `json:"star_count"`
}

// SuggestedPostApproved represents service message about a suggested post being approved.
type SuggestedPostApproved struct{}

// SuggestedPostApprovalFailed represents service message about a suggested post approval failure.
type SuggestedPostApprovalFailed struct {
	ErrorMessage string `json:"error_message"`
}

// SuggestedPostDeclined represents service message about a suggested post being declined.
type SuggestedPostDeclined struct{}

// SuggestedPostPaid represents service message about a suggested post being paid.
type SuggestedPostPaid struct {
	StarCount int `json:"star_count"`
}

// SuggestedPostRefunded represents service message about a suggested post being refunded.
type SuggestedPostRefunded struct {
	StarCount int `json:"star_count"`
}

// VoiceChatScheduled represents a service message about a voice chat scheduled in the chat.
type VoiceChatScheduled struct {
	StartDate int `json:"start_date"`
}

// VoiceChatStarted represents a service message about a voice chat started in the chat.
type VoiceChatStarted struct{}

// VoiceChatEnded represents a service message about a voice chat ended in the chat.
type VoiceChatEnded struct {
	Duration int `json:"duration"`
}

// VoiceChatParticipantsInvited represents a service message about new members invited to a voice chat.
type VoiceChatParticipantsInvited struct {
	Users []User `json:"users,omitempty"`
}
