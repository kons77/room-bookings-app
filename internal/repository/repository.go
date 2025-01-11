package repository

import "github.com/kons77/room-bookings-app/internal/models"

type DatabaseRepo interface {
	AllUsers() bool // this function is listed in the interface

	InsertReservation(res models.Reservation) (int, error)
	InsertRoomRestriction(res models.RoomRestriction) error
}