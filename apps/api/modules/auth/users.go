// Package auth wires Opus's account table into porte's login. porte verifies
// the credential and hands back an app-side int64 user id; this store is the
// app's half, creating and finding the row porte resolved.
package auth

import (
	"context"
	stderrors "errors"
	"strings"

	"github.com/FacileStudio/Opus/apps/api/schemas"
	"github.com/FacileStudio/porte"
	"github.com/FacileStudio/tronc/errors"

	"gorm.io/gorm"
)

// UserStore resolves a login to a row in Opus's own users table.
//
// It is porte's escape hatch, taken deliberately rather than moving onto
// porte/pg's porte_users: this users row is referenced from workspace members
// and every project/task/comment relation, and the port keeps the profile on
// the table the frontend already reads.
type UserStore struct {
	orm *gorm.DB
}

// NewUserStore returns the user store backed by the given database.
func NewUserStore(orm *gorm.DB) *UserStore {
	return &UserStore{orm: orm}
}

var (
	_ porte.UserStore         = (*UserStore)(nil)
	_ porte.PasswordUserStore = (*UserStore)(nil)
)

// UpsertFromOIDC matches on (provider, subject) through the identity row porte
// owns, falls back to a verified email, and creates the account when neither
// finds anything. porte has already done the identity lookup; this is what
// links a pre-existing account to a subject signing in for the first time.
// Every call rewrites the name, the email and the picture URL from the claims,
// so porte's /auth/sync-profile refreshes them through the same statement.
func (s *UserStore) UpsertFromOIDC(ctx context.Context, claims porte.Claims) (int64, error) {
	email := strings.ToLower(strings.TrimSpace(claims.Email))
	if email == "" {
		return 0, errors.Invalid("the identity provider returned no email")
	}

	var user schemas.User
	err := s.orm.WithContext(ctx).
		Joins("JOIN porte_identities i ON i.user_id = users.id AND i.provider = ? AND i.subject = ?", claims.Provider, claims.Subject).
		First(&user).Error
	switch {
	case err == nil:
	case isNotFound(err):
		err = s.orm.WithContext(ctx).Where("email = ?", email).First(&user).Error
		switch {
		case err == nil:
			if !claims.EmailVerified {
				return 0, errors.Conflict("an account with this email already exists and the identity provider did not verify the address")
			}
		case isNotFound(err):
			created, createErr := s.create(ctx, email, claims.DisplayName())
			if createErr != nil {
				return 0, createErr
			}
			user = *created
		default:
			return 0, errors.Internal("failed to look up the account", err)
		}
	default:
		return 0, errors.Internal("failed to resolve the identity", err)
	}

	updates := map[string]any{"email": email, "image": pictureURL(claims.Picture)}
	if displayName := claims.DisplayName(); displayName != "" {
		updates["name"] = displayName
	}
	if err := s.orm.WithContext(ctx).Model(&schemas.User{}).Where("id = ?", user.ID).Updates(updates).Error; err != nil {
		return 0, errors.Internal("failed to update the account", err)
	}
	return user.ID, nil
}

// BackfillPictures points every account with an identity under issuer at the
// picture URL that issuer serves for its subject, <issuer>/avatars/<subject>.
//
// Registre emits that URL as the picture claim for every human, so this is the
// value the next sign-in would write anyway. Running it at boot gives accounts
// that have not signed in since the claim was stored their picture now, and
// replaces the dead image URLs older providers left in the column. It only
// touches rows whose image differs, so it is idempotent and cheap on reboot.
// Local-password accounts have no identity row and are never touched.
func (s *UserStore) BackfillPictures(ctx context.Context, issuer string) (int64, error) {
	prefix := strings.TrimRight(issuer, "/") + "/avatars/"
	if pictureURL(prefix) == "" {
		return 0, nil
	}
	result := s.orm.WithContext(ctx).Exec(
		`UPDATE users SET image = ? || i.subject
		FROM porte_identities i
		WHERE i.user_id = users.id AND i.provider = ? AND COALESCE(users.image, '') <> ? || i.subject`,
		prefix, issuer, prefix,
	)
	if result.Error != nil {
		return 0, errors.Internal("failed to backfill the picture URLs", result.Error)
	}
	return result.RowsAffected, nil
}

// pictureURL keeps the picture claim when it is an https URL and drops
// anything else, such as a data: placeholder or a plain-http link the browser
// would block as mixed content. The browser fetches the URL itself; Opus never
// downloads the bytes.
func pictureURL(claim string) string {
	if strings.HasPrefix(strings.ToLower(claim), "https://") {
		return claim
	}
	return ""
}

// CreateFromPassword creates a user row for a local registration. porte has
// already validated the address and hashed the password.
func (s *UserStore) CreateFromPassword(ctx context.Context, email, name string) (int64, error) {
	user, err := s.create(ctx, email, name)
	if err != nil {
		return 0, err
	}
	return user.ID, nil
}

// FindByEmail returns the user id for an address, or porte.ErrNotFound.
func (s *UserStore) FindByEmail(ctx context.Context, email string) (int64, error) {
	var user schemas.User
	err := s.orm.WithContext(ctx).Select("id").Where("email = ?", email).First(&user).Error
	if isNotFound(err) {
		return 0, porte.ErrNotFound
	}
	if err != nil {
		return 0, errors.Internal("failed to look up the account", err)
	}
	return user.ID, nil
}

// CountUsers backs porte/local's registration gate.
func (s *UserStore) CountUsers(ctx context.Context) (int64, error) {
	var count int64
	if err := s.orm.WithContext(ctx).Model(&schemas.User{}).Count(&count).Error; err != nil {
		return 0, errors.Internal("failed to count users", err)
	}
	return count, nil
}

// create is the one place an Opus account comes into existence.
func (s *UserStore) create(ctx context.Context, email, name string) (*schemas.User, error) {
	user := schemas.User{
		Email: email,
		Name:  name,
	}
	if err := s.orm.WithContext(ctx).Create(&user).Error; err != nil {
		return nil, errors.Internal("failed to create the account", err)
	}
	return &user, nil
}

func isNotFound(err error) bool {
	return stderrors.Is(err, gorm.ErrRecordNotFound)
}