package domain

import (
	"context"
	"time"
)

type Booking struct {
	ID               int       `json:"bookingId"`
	UserID           int       `json:"userId"`
	PlaceID          int       `json:"placeId"`
	Status           string    `json:"bookingStatus"`
	StartTime        time.Time `json:"bookingStartTime"`
	EndTime          time.Time `json:"bookingEndTime"`
	CreatedAt        time.Time `json:"createdAt"`
}

type CreateBookingRequest struct {
	PlaceID   int       `json:"placeId"` // Берем из URL
	UserID    int       `json:"userId"`  // Берем из JWT токена
	StartTime time.Time `json:"bookingStartTime"`
	EndTime   time.Time `json:"bookingEndTime"`
}

type BookingRepository interface {
	CreateBookingTx(ctx context.Context, req CreateBookingRequest) (*Booking, error)
}

type EventPublisher interface {
	PublishEvent(ctx context.Context, eventType string, payload interface{}) error
}

type BookingService interface {
	CreateBooking(ctx context.Context, req CreateBookingRequest) (*Booking, error)
}
