package models

import "time"

type WalletBalanceResponse struct {
	WalletID       *string    `json:"wallet_id"`
	Currency       string     `json:"currency"`
	Status         string     `json:"status"`
	AvailableMinor int64      `json:"available_minor"`
	PendingMinor   int64      `json:"pending_minor"`
	HeldMinor      int64      `json:"held_minor"`
	UpdatedAt      *time.Time `json:"updated_at"`
}
