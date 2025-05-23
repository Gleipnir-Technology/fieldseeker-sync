package fssync

type Geometry struct {
	X float64 `db:"X"`
	Y float64 `db:"Y"`
}

type ServiceRequest struct {
	Geometry Geometry `db:"geometry"`
	Address  *string  `db:"reqaddr1"`
	City     *string  `db:"reqcity"`
	Priority *string  `db:"priority"`
	Status   *string  `db:"status"`
	Source   *string  `db:"source"`
	Target   *string  `db:"reqtarget"`
	Zip      *string  `db:"reqzip"`
}

type User struct {
	DisplayName      string `db:"display_name"`
	PasswordHashType string `db:"password_hash_type"`
	PasswordHash     string `db:"password_hash"`
	Username         string `db:"username"`
}
