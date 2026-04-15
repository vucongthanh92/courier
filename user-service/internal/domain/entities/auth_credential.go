package entities

import (
	"context"
	"errors"
	"time"

	"golang.org/x/crypto/bcrypt"
)

type AuthCredential struct {
	ID              uint64    `gorm:"column:id;primaryKey;autoIncrement:true;check:id>0" json:"id"`
	UserID          uint64    `gorm:"column:user_id" json:"user_id"`
	PasswordHash    string    `gorm:"column:password_hash;type:text;" json:"-"`
	PasswordAlgo    string    `gorm:"column:password_algo;type:text;check:password_algo IN ('sha256','bcrypt','scrypt')" json:"-"`
	MFAEnabled      bool      `gorm:"column:mfa_enabled;not null;default:false" json:"mfa_enabled"`
	PasswordVersion int16     `gorm:"column:password_version;not null;default:0" json:"password_version"`
	CreatedAt       time.Time `gorm:"column:created_at;type:timestamptz;not null;default:now()" json:"created_at"`
	UpdatedAt       time.Time `gorm:"column:updated_at;type:timestamptz;autoUpdateTime" json:"updated_at"`
}

func (AuthCredential) TableName() string {
	return `"user-service".auth_credential`
}

// ComparePwdHashWithAlgo compares the provided password with the stored password hash
func (a *AuthCredential) ComparePwdHashWithAlgo(ctx context.Context, comparePwd string) error {
	switch a.PasswordAlgo {
	case "sha256":
		return errors.New("unsupported password hashing algorithm")
	case "bcrypt":
		err := bcrypt.CompareHashAndPassword([]byte(a.PasswordHash), []byte(comparePwd))
		if err != nil {
			return err
		}
	case "scrypt":
		return errors.New("unsupported password hashing algorithm")
	default:
		return errors.New("unsupported password hashing algorithm")
	}

	return nil
}

// GeneratePasswordHash generates the password hash
// based on the password algorithm specified in the AuthCredential struct.
func (a *AuthCredential) GeneratePwdHashWithAlgo(pwd string) error {
	switch a.PasswordAlgo {
	case "sha256":
		return errors.New("unsupported password hashing algorithm")
	case "bcrypt":
		hash, err := bcrypt.GenerateFromPassword([]byte(pwd), 12)
		if err != nil {
			return err
		}
		a.PasswordHash = string(hash)
	case "scrypt":
		return errors.New("unsupported password hashing algorithm")
	default:
		return errors.New("unsupported password hashing algorithm")
	}

	return nil
}

// MappingToAuthCredEntity maps the password and password algorithm to the AuthCredential entity
func (a *AuthCredential) MappingToAuthCredEntity(algo string, pwd string, version int16) {
	a.PasswordAlgo = algo
	a.MFAEnabled = false
	a.PasswordVersion = version
	a.GeneratePwdHashWithAlgo(pwd)
}
