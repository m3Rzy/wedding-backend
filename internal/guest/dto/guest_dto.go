// internal/guest/dto/guest_dto.go
package dto

type GuestDto struct {
	Fio       string `json:"fio" validate:"required,min=2,max=100"`
	Telephone string `json:"telephone" validate:"required,min=5,max=20"`
}