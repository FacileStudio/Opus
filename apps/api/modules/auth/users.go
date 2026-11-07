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

var _ porte.UserStore = (*UserStore)(nil)

// UpsertFromOIDC matches on (provider, subject) through the identity row porte
// owns, falls back to a verified email, and creates the account when neither
// finds anything. porte has already done the identity lookup; this is what
// links a pre-existing account to a subject signing in for the first time.
// The fallback applies porte v0.10.0's linking rules, which an app with its
// own users table has to enforce itself: link only by a verified address, and
// refuse when the account already holds another subject of the same issuer, so
// a recycled address is not handed to a newcomer.
// Every call rewrites the name, the email, the picture URL and, when the
// claims carry one, the locale, so porte's /auth/sync-profile refreshes them
// through the same statement.
func (s *UserStore) UpsertFromOIDC(ctx context.Context, claims porte.Claims) (int64, error) {
	email := strings.ToLower(strings.TrimSpace(claims.Email))
	if email == "" {
		return 0, errors.Invalid("the identity provider returned no email")
	}

	user, err := s.resolve(ctx, claims, email)
	if err != nil {
		return 0, err
	}

	updates := map[string]any{"email": email, "image": pictureURL(claims.Picture)}
	if displayName := claims.DisplayName(); displayName != "" {
		updates["name"] = displayName
	}
	if claims.Locale != "" {
		updates["locale"] = claims.Locale
	}
	if err := s.orm.WithContext(ctx).Model(&schemas.User{}).Where("id = ?", user.ID).Updates(updates).Error; err != nil {
		return 0, errors.Internal("failed to update the account", err)
	}
	return user.ID, nil
}

// resolve finds the account a sign-in belongs to: by (provider, subject), then
// by a verified email, and otherwise a new one.
func (s *UserStore) resolve(ctx context.Context, claims porte.Claims, email string) (*schemas.User, error) {
	var user schemas.User
	err := s.orm.WithContext(ctx).
		Joins("JOIN porte_identities i ON i.user_id = users.id AND i.provider = ? AND i.subject = ?", claims.Provider, claims.Subject).
		First(&user).Error
	switch {
	case err == nil:
		return &user, nil
	case !isNotFound(err):
		return nil, errors.Internal("failed to resolve the identity", err)
	}

	err = s.orm.WithContext(ctx).Where("email = ?", email).First(&user).Error
	switch {
	case err == nil:
		return &user, s.checkLinkable(ctx, user.ID, claims)
	case isNotFound(err):
		return s.create(ctx, email, claims.DisplayName(), claims.Locale)
	}
	return nil, errors.Internal("failed to look up the account", err)
}

// checkLinkable refuses to attach a sign-in to an existing account by email
// unless the address is verified and the account holds no other subject of the
// same provider.
func (s *UserStore) checkLinkable(ctx context.Context, userID int64, claims porte.Claims) error {
	if !claims.EmailVerified {
		return errors.Conflict("an account with this email already exists and the identity provider did not verify the address")
	}
	var held int64
	err := s.orm.WithContext(ctx).Table("porte_identities").
		Where("user_id = ? AND provider = ?", userID, claims.Provider).
		Count(&held).Error
	if err != nil {
		return errors.Internal("failed to check the account's identities", err)
	}
	if held > 0 {
		return errors.Conflict("an account with this email is already linked to another identity")
	}
	return nil
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

// create is the one place an Opus account comes into existence.
func (s *UserStore) create(ctx context.Context, email, name, locale string) (*schemas.User, error) {
	user := schemas.User{
		Email:  email,
		Name:   name,
		Locale: locale,
	}
	if err := s.orm.WithContext(ctx).Create(&user).Error; err != nil {
		return nil, errors.Internal("failed to create the account", err)
	}
	return &user, nil
}

func isNotFound(err error) bool {
	return stderrors.Is(err, gorm.ErrRecordNotFound)
}
