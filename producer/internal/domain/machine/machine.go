package machine

import "time"

type Machine struct {
	ID              string    `db:"id" json:"id"`
	Name            string    `db:"name" json:"name"`
	FingerprintHash *string   `db:"fingerprint_hash" json:"-"`
	TokenHash       string    `db:"token_hash" json:"-"`
	Status          string    `db:"status" json:"status"`
	CreatedAt       time.Time `db:"created_at" json:"created_at"`
	UpdatedAt       time.Time `db:"updated_at" json:"updated_at"`
	LastSeenAt      time.Time `db:"last_seen_at" json:"last_seen_at"`
}

type RegisterResult struct {
	Machine      *Machine `json:"machine"`
	MachineToken string   `json:"machine_token"`
}
