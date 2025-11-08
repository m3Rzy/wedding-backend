// internal/guest/dto/guest_dto.go
package dto

type GuestDto struct {
	Fio       string `json:"fio" validate:"required,min=2,max=100"`
	Telephone string `json:"telephone" validate:"required,min=5,max=20"`
	Transport string `json:"transport" validate:"required,oneof=transfer car"`
	CarNumber string `json:"car_number" validate:"omitempty,min=2,max=20"`
}