package common

import "github.com/labstack/echo/v4"

// EchoLevel is an interface capable of defining children
// groups. Stands for a group or the Echo main object.
type EchoLevel interface {
	Group(prefix string, middleware ...echo.MiddlewareFunc) *echo.Group
}
