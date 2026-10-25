package vault

type ProfileRequest struct {
	Profile string `json:"profile"`
}
type LoginRequest struct {
	Profile  string `json:"profile"`
	Server   string `json:"server"`
	Email    string `json:"email"`
	Password []byte `json:"password"`
	AuthID   string `json:"authId,omitempty"`
}
type UnlockRequest struct {
	Profile  string `json:"profile"`
	Password []byte `json:"password"`
	AuthID   string `json:"authId,omitempty"`
}

type ReadRequest struct {
	Profile string `json:"profile"`
	ID      string `json:"id"`
	Field   string `json:"field"`
}
type Empty struct{}
type ReadResponse struct {
	Value string `json:"value"`
}
type StatusResponse struct {
	State string `json:"state"`
}
