package e

import (
	"github.com/labstack/echo/v4"
	"gitlab.broearn.net/common-go/library/code"
	"gitlab.broearn.net/common-go/library/log"
	"net/http"
)

func HTTPErrHandler(err error, c echo.Context) {
	ctx := c.Request().Context()

	if c.Response().Committed {
		return
	}

	switch typeErr := err.(type) {
	case *echo.HTTPError:
		if err := c.JSON(typeErr.Code, typeErr); err != nil {
			log.WithContext(ctx).Error(err)
		}
	case HttpErr:
		typeErr.SendErrorMsg(ctx)
		if err := c.JSON(http.StatusUnprocessableEntity, code.Error(ctx, typeErr)); err != nil {
			log.WithContext(ctx).Error(err)
		}
	default:
		SendMessage(ctx, typeErr)
		httpErr := NewHttpErr(SystemErr, err)
		c.JSON(http.StatusUnprocessableEntity, code.Error(ctx, httpErr))
	}

}
