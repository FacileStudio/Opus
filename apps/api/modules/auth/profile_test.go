package auth

import (
	"context"
	"log/slog"
	"net/http/httptest"
	"testing"

	"github.com/FacileStudio/Opus/apps/api/schemas"
	"github.com/FacileStudio/tronc/errors"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func newProfileService(t *testing.T) *Service {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	if err := db.AutoMigrate(&schemas.User{}); err != nil {
		t.Fatalf("migrate user: %v", err)
	}
	return NewService(db, nil, nil, slog.New(slog.DiscardHandler))
}

func seedProfileUser(t *testing.T, service *Service) int64 {
	t.Helper()
	user := schemas.User{Email: "camille@facile.studio", Name: "Camille", Color: "#ff0000", Locale: "fr"}
	if err := service.orm.Create(&user).Error; err != nil {
		t.Fatalf("seed user: %v", err)
	}
	return user.ID
}

func TestUpdateProfilePartial(t *testing.T) {
	service := newProfileService(t)
	userID := seedProfileUser(t, service)

	name := "Camille Martel"
	user, err := service.UpdateProfile(context.Background(), userID, UpdateProfileRequest{Name: &name})
	if err != nil {
		t.Fatalf("UpdateProfile: %v", err)
	}
	if user.Name != name {
		t.Fatalf("name = %q, want %q", user.Name, name)
	}
	if user.Color != "#ff0000" || user.Locale != "fr" {
		t.Fatalf("untouched fields wiped: color=%q locale=%q", user.Color, user.Locale)
	}
}

func TestUpdateProfileEmptyIsNoop(t *testing.T) {
	service := newProfileService(t)
	userID := seedProfileUser(t, service)

	user, err := service.UpdateProfile(context.Background(), userID, UpdateProfileRequest{})
	if err != nil {
		t.Fatalf("UpdateProfile: %v", err)
	}
	if user.Name != "Camille" || user.Color != "#ff0000" || user.Locale != "fr" {
		t.Fatalf("empty patch changed the profile: %+v", user)
	}
}

func TestControllerUpdateProfileEchoesBearer(t *testing.T) {
	service := newProfileService(t)
	userID := seedProfileUser(t, service)
	request := httptest.NewRequest("PATCH", "/api/auth/me", nil)
	request.Header.Set("Authorization", "Bearer the-token")

	locale := "en"
	resp, err := service.controller.updateProfile(httptest.NewRecorder(), request, userID, &UpdateProfileRequest{Locale: &locale})
	if err != nil {
		t.Fatalf("updateProfile: %v", err)
	}
	if resp.Token != "the-token" {
		t.Fatalf("token = %q, want the echoed bearer", resp.Token)
	}
	if resp.User == nil || resp.User.Locale != "en" {
		t.Fatalf("user = %+v, want locale en", resp.User)
	}
}

func TestControllerChangePasswordValidatesBody(t *testing.T) {
	service := newProfileService(t)
	request := httptest.NewRequest("POST", "/api/auth/change-password", nil)

	_, err := service.controller.changePassword(httptest.NewRecorder(), request, 1, &ChangePasswordRequest{CurrentPassword: "", NewPassword: "new"})
	if errors.Status(err) != 400 {
		t.Fatalf("empty currentPassword status = %d, want 400", errors.Status(err))
	}
	_, err = service.controller.changePassword(httptest.NewRecorder(), request, 1, &ChangePasswordRequest{CurrentPassword: "old", NewPassword: ""})
	if errors.Status(err) != 400 {
		t.Fatalf("empty newPassword status = %d, want 400", errors.Status(err))
	}
}