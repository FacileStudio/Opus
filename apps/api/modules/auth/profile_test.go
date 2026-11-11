package auth

import (
	"context"
	"log/slog"
	"net/http/httptest"
	"testing"

	"github.com/FacileStudio/Opus/apps/api/internal/database/dbtest"
	"github.com/FacileStudio/Opus/apps/api/schemas"
)

func newProfileService(t *testing.T) *Service {
	t.Helper()
	db := dbtest.Open(t)
	return NewService(db, nil, slog.New(slog.DiscardHandler))
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
