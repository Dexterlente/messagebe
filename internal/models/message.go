package models

import (
	"time"
)

type Message struct {
    ID         int       `db:"id"`
    SenderID   int       `db:"sender_id"`
    ReceiverID int       `db:"receiver_id"`
    Content    string    `db:"content"`
    SentAt     time.Time `db:"sent_at"`
}