package apis

import (
	"gofreta/app"
	"gofreta/app/daos"
	"gofreta/app/models"
	"gofreta/app/utils"
	"net/http"
	"time"

	jwt "github.com/dgrijalva/jwt-go"
	"github.com/globalsign/mgo"
	"github.com/globalsign/mgo/bson"
	routing "github.com/go-ozzo/ozzo-routing"
	"github.com/go-ozzo/ozzo-routing/auth"
)

const (
	// TokenIdentityUser specify "user" token idenity model.
	TokenIdentityUser = "user"

	// TokenIdentityUser specify "key" token idenity model.
	TokenIdentityKey = "key"
)

// AuthApi defines auth api services
type AuthApi struct {
	routeGroup   *routing.RouteGroup
	mongoSession *mgo.Session
	dao          *daos.UserDAO
}

// InitAuthApi sets up the routing of auth endpoints and the corresponding handlers.
func InitAuthApi(rg *routing.RouteGroup, session *mgo.Session) {
	api := AuthApi{
		routeGroup:   rg,
		mongoSession: session,
		dao:          daos.NewUserDAO(session),
	}

	rg.Post("/auth", api.auth)
	rg.Post("/forgotten-password", api.sendResetEmail)
	rg.Post("/reset-password/<hash>", api.resetPassword)
}

// -------------------------------------------------------------------
// • API endpoint handlers
// -------------------------------------------------------------------

// auth api endpoint handler for authenticating users.
func (api *AuthApi) auth(c *routing.Context) error {
	data := &struct {
		Username string `json:"username" form:"username"`
		Password string `json:"password" form:"password"`
	}{}
	if readErr := c.Read(&data); readErr != nil {
		return utils.NewBadRequestError("Oops, something went wrong while creating the auth token.", readErr)
	}

	user, userErr := api.dao.Authenticate(data.Username, data.Password)
	if userErr != nil {
		return utils.NewBadRequestError("Invalid username or password.", nil)
	}

	exp := time.Now().Add(time.Hour * time.Duration(gofreta.App.Config.GetInt64("userTokenExpire"))).Unix()

	token, err := user.NewAuthToken(exp)
	if err != nil {
		return utils.NewBadRequestError("Oops, something went wrong while creating the auth token.", err)
	}

	return c.Write(map[string]interface{}{
		"token":  token,
		"expire": exp,
		"user":   user,
	})
}

// sendResetEmail api endpoint handler for sending a reset user password email.
func (api *AuthApi) sendResetEmail(c *routing.Context) error {
	data := &struct {
		Username string `json:"username" form:"username"`
	}{}
	if readErr := c.Read(&data); readErr != nil {
		return utils.NewBadRequestError("Oops, an error occurred while generating new reset password key.", readErr)
	}

	user, err := api.dao.GetByUsername(data.Username, bson.M{"status": models.UserStatusActive})
	if err != nil {
		return utils.NewNotFoundError("Inactive or missing user.")
	}

	// --- set new user reset password hash
	renewedUser, renewedErr := api.dao.RenewResetPasswordHash(user)
	if renewedErr != nil {
		return utils.NewBadRequestError("Oops, an error occurred while generating new reset password key.", renewedErr)
	}
	// ---

	// --- render mail body
	params := struct {
		RessetPasswordHash string
		SupportEmail       string
	}{
		RessetPasswordHash: renewedUser.ResetPasswordHash,
		SupportEmail:       gofreta.App.Config.GetString("emails.support"),
	}
	body, renderErr := utils.RenderTemplates(params, "app/emails/layout.tmpl", "app/emails/reset_password.tmpl")
	if renderErr != nil {
		return utils.NewBadRequestError("Oops, an error occurred while rendering the email body.", renderErr)
	}
	// ---

	// --- send email
	sendErr := utils.SendEmail(
		gofreta.App.Config.GetString("emails.noreply"),
		renewedUser.Email,
		"Reset password request",
		body,
	)
	if sendErr != nil {
		return utils.NewBadRequestError("Oops, an error occurred while sending the reset password email.", sendErr)
	}
	// ---

	c.Response.WriteHeader(http.StatusNoContent)

	return nil
}

// resetPassword api endpoint handler for resetting a forgotten user password.
func (api *AuthApi) resetPassword(c *routing.Context) error {
	hash := c.Param("hash")

	user, err := api.dao.GetOne(bson.M{
		"status": models.UserStatusActive,
		"$and": []bson.M{
			bson.M{"reset_password_hash": hash},
			bson.M{"reset_password_hash": bson.M{"$exists": true, "$ne": ""}},
		},
	})

	if err != nil || !user.HasValidResetPasswordHash() {
		return utils.NewBadRequestError("Invalid or expired reset password key!", nil)
	}

	form := &models.UserResetPasswordForm{}
	if readErr := c.Read(form); readErr != nil {
		return utils.NewBadRequestError("Oops, an error occurred while changing User model password.", readErr)
	}
	form.Model = user

	updatedUser, updateErr := api.dao.ResetPassword(form)
	if updateErr != nil {
		return utils.NewBadRequestError("Oops, an error occurred while changing User model password.", updateErr)
	}

	return c.Write(updatedUser)
}

// -------------------------------------------------------------------
// • Routing handlers
// -------------------------------------------------------------------

// authenticateToken checks whether the request access token is valid and optionally
// checks whether the token identity has access to a route action.
func authenticateToken(session *mgo.Session, checkRoutePair ...string) routing.Handler {
	return func(c *routing.Context) error {
		handler := auth.JWT(gofreta.App.Config.GetString("jwt.verificationKey"), auth.JWTOptions{
			SigningMethod: gofreta.App.Config.GetString("jwt.signingMethod"),
			TokenHandler:  tokenHandler(session),
		})

		err := handler(c)

		if err != nil {
			return utils.NewApiError(http.StatusUnauthorized, err.Error(), nil)
		}

		if len(checkRoutePair) == 2 && checkRoutePair[0] != "" && checkRoutePair[1] != "" {
			return canAccess(c, checkRoutePair[0], checkRoutePair[1])
		}

		return nil
	}
}

// usersOnly checks whether the request is authenticated via a user token.
// NB! Need to be called after the token is read and validated (aka. after `authenticateToken()`).
func usersOnly(c *routing.Context) error {
	isUserIdentity := isUser(c)
	if isUserIdentity {
		// access granted
		return nil
	}

	// access forbidden
	return utils.NewApiError(http.StatusForbidden, "Oops, it seems that you don't have access rights to perform this request.", nil)
}

// -------------------------------------------------------------------
// • Auth helpers
// -------------------------------------------------------------------

// tokenHandler takes care for initializing basic user info on successful JWT token authentication.
func tokenHandler(session *mgo.Session) auth.JWTTokenHandler {
	return func(c *routing.Context, j *jwt.Token) error {
		claims := j.Claims.(jwt.MapClaims)

		var id, model string

		if claims["id"] != nil {
			id, _ = claims["id"].(string)
		}

		if claims["model"] != nil {
			model, _ = claims["model"].(string)
		}

		var accessData map[string][]string

		if model == TokenIdentityKey { // key
			keyDAO := daos.NewKeyDAO(session)

			key, err := keyDAO.GetByID(id)
			if err != nil {
				return utils.NewApiError(http.StatusForbidden, "Access rules can not be fetched.", nil)
			}

			accessData = key.Access
		} else { // user
			model = TokenIdentityUser // reset

			userDAO := daos.NewUserDAO(session)

			user, err := userDAO.GetByID(id, bson.M{"status": models.UserStatusActive})
			if err != nil {
				return utils.NewApiError(http.StatusForbidden, "Access rules can not be fetched.", nil)
			}

			accessData = user.Access
		}

		c.Set("identityID", id)
		c.Set("identityModel", model)
		c.Set("identityAccess", accessData)

		return nil
	}
}

// canAccess checks whether the authenticated identity is allowed to access a request group and action.
func canAccess(c *routing.Context, group string, action string) error {
	accessData := map[string][]string{}
	if c.Get("identityAccess") != nil {
		accessData, _ = c.Get("identityAccess").(map[string][]string)
	}

	allowedActions, hasGroup := accessData[group]

	if hasGroup && utils.StringInSlice(action, allowedActions) {
		// access granted
		return nil
	}

	// access forbidden
	return utils.NewApiError(http.StatusForbidden, "Oops, it seems that you don't have access rights to perform this request.", nil)
}

// getAccessColectionIds returns slice with allowed access collection ids.
func getAccessColectionIds(c *routing.Context, actions ...string) []bson.ObjectId {
	var ids []bson.ObjectId

	accessData := map[string][]string{}
	if c.Get("identityAccess") != nil {
		accessData, _ = c.Get("identityAccess").(map[string][]string)
	}

GROUP_LOOP:
	for group, allowedActions := range accessData {
		if !bson.IsObjectIdHex(group) {
			continue
		}

		if len(actions) > 0 {
			// each provided action must be allowed
			for _, action := range actions {
				if !utils.StringInSlice(action, allowedActions) {
					continue GROUP_LOOP
				}
			}
		}

		ids = append(ids, bson.ObjectIdHex(group))
	}

	return ids
}

// isUser checks whether the current request token represents a user or an api key.
func isUser(c *routing.Context) bool {
	if c.Get("identityModel") != nil {
		identityModel, _ := c.Get("identityModel").(string)

		return identityModel == TokenIdentityUser
	}

	return false
}
