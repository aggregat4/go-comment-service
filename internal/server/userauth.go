package server

import (
	"aggregat4/go-commentservice/internal/domain"
	"net/http"

	"github.com/aggregat4/go-baselib/lang"
	"github.com/gorilla/sessions"
	"github.com/labstack/echo-contrib/session"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

var authenticatedUserCookieName = "commentservice-authenticated-user"

func getUserIdFromSession(c echo.Context) (int, error) {
	sess, err := session.Get(authenticatedUserCookieName, c)
	if err != nil {
		return -1, err
	}
	if sess.Values["userid"] != nil {
		return sess.Values["userid"].(int), nil
	} else {
		return -1, lang.ErrNotFound
	}
}

func getUserRolesFromSession(c echo.Context) ([]string, error) {
	sess, err := session.Get(authenticatedUserCookieName, c)
	if err != nil {
		return nil, err
	}
	if sess.Values["userroles"] != nil {
		return sess.Values["userroles"].([]string), nil
	} else {
		return nil, lang.ErrNotFound
	}
}

func createSessionCookie(c echo.Context, userId int, userRoles []string) error {
	// we have a valid user, we can now create a session and redirect to the original request
	sess, err := session.Get(authenticatedUserCookieName, c)
	if err != nil {
		return err
	}
	sess.Options = &sessions.Options{
		Path: "/", // TODO: this path is not context path safe
		// 30 days
		MaxAge:   3600 * 24 * 30,
		HttpOnly: true,
		Secure:   false,
		SameSite: http.SameSiteLaxMode,
	}
	sess.Values["userid"] = userId
	sess.Values["userroles"] = userRoles
	err = sess.Save(c.Request(), c.Response())
	if err != nil {
		return sendInternalError(c, err)
	}
	return nil
}

func CreateUserAuthenticationMiddleware(skipper middleware.Skipper) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			if skipper(c) {
				return next(c)
			}
			_, err := getUserIdFromSession(c)
			if err != nil {
				// user is not authenticated, redirect him to the authentication token link generation form
				return c.Redirect(http.StatusFound, "/userauthentication/")
			} else {
				return next(c)
			}
		}
	}
}

func getUserFromSession(c echo.Context, controller *Controller) (domain.User, error) {
	userId, err := getUserIdFromSession(c)
	if err != nil {
		return domain.User{}, err
	}
	user, err := controller.Store.FindUserById(userId)
	if err != nil {
		return domain.User{}, err
	}
	return user, nil
}
