package fssync

type ServiceRequest struct {
	Address  *string `db:"reqaddr1"`
	City     *string `db:"reqcity"`
	Priority *string `db:"priority"`
	Status   *string `db:"status"`
	Source   *string `db:"source"`
	Target   *string `db:"reqtarget"`
	Zip      *string `db:"reqzip"`
}
