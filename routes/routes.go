package routes

import "github.com/tedsuo/rata"

const (
	CreateUser = "CreateUser"
	GetUser    = "GetUser"
	UpdateUser = "UpdateUser"
	DeleteUser = "DeleteUser"
)

var Routes = rata.Routes{
	{Path: "/v1/users", Method: "POST", Name: CreateUser},
	{Path: "/v1/users/:account_id", Method: "GET", Name: GetUser},
	{Path: "/v1/users/:account_id", Method: "PUT", Name: UpdateUser},
	{Path: "/v1/users/:account_id", Method: "DELETE", Name: DeleteUser},
}
