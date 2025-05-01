package v1

import (
	"errors"
	"net/http"

	"github.com/atnomoverflow/auth-server/internal/entity"
	"github.com/atnomoverflow/auth-server/internal/service"
	"github.com/atnomoverflow/auth-server/pkg/helpers/response"
	"github.com/atnomoverflow/auth-server/pkg/token"
	validation "github.com/atnomoverflow/auth-server/pkg/validator"
)

func (h *Handler) initAuthRoutes(api *http.ServeMux) {

	api.HandleFunc("GET /auth/sign-in", h.signUp)
	api.HandleFunc("POST /auth/login", h.login)
	api.HandleFunc("GET /auth/verify", h.verifyUserEmail)
	api.HandleFunc("POST /auth/2fa/{methode}/verify", h.VerifyTwoFAuth)

	/**
	*
	* TODO: Enable 2FAuth
	* Note: need to be authentificated to be used !!!
	*
	**/
	api.HandleFunc("POST /auth/2fa/{methode}/enable", h.signUp)

	/**
	*
	* TODO: Start the OTP flow
	*
	**/
	api.HandleFunc("POST /auth/2fa/{methode}", h.signUp)

}

func (h *Handler) signUp(w http.ResponseWriter, r *http.Request) {
	var signUp = service.SignUpInput{}
	err := validation.Validate(r.Body, signUp)
	if err != nil {
		resp, responseErr := response.ValidationErrors(err)
		if responseErr != nil {
			h.logger.Error("unkown error from validation %v", err)
			response.InternalServerError(w)
			return
		}
		resp.Write(w)
		return
	}
	err = h.services.AuthService.SignUp(r.Context(), signUp)
	if err != nil {
		if errors.Is(err, service.ErrorEmailUsed) {
			response.ConflictError(w, err.Error())
			return
		}
		response.InternalServerError(w)
		return
	}
	response.OK(w, "success")
}

func (h *Handler) login(w http.ResponseWriter, r *http.Request) {
	var cridentials = service.CredentialsInput{}
	err := validation.Validate(r.Body, cridentials)
	if err != nil {
		resp, responseErr := response.ValidationErrors(err)
		if responseErr != nil {
			h.logger.Error("unkown error from validation %v", err)
			response.InternalServerError(w)
			return
		}
		resp.Write(w)
		return
	}
	cridToken, twoFactAuthToken, err := h.services.AuthService.Login(r.Context(), cridentials)
	if err != nil {
		if errors.Is(err, service.ErrorBadCridentials) {
			message := "Bad cridentials!"
			response.UnauthorizedError(w, &message)
			return
		}
		if errors.Is(err, service.ErrorRequire2Fauth) {
			var requireAuth response.JSONResponse
			requireAuth.StatusCode = http.StatusForbidden
			requireAuth.Message = struct {
				Message             string `json:"message"`
				Two_fact_auth_token string `json:"2fA_token"`
			}{
				Message:             err.Error(),
				Two_fact_auth_token: *twoFactAuthToken,
			}
			requireAuth.Write(w)
			return
		}
		response.InternalServerError(w)
		return
	}

	response.OK(w, cridToken)
}

func (h *Handler) verifyUserEmail(w http.ResponseWriter, r *http.Request) {
	queryParams := r.URL.Query()
	token := queryParams.Get("token")
	err := h.services.AuthService.Verify(r.Context(), token)
	if err != nil {
		if errors.Is(err, service.ErrorServerFailed) {
			response.InternalServerError(w)
			return
		}
		message := err.Error()
		response.UnauthorizedError(w, &message)
		return
	}
	response.OK(w, "Congrats! your account has been validated.")
}
func (h *Handler) VerifyTwoFAuth(w http.ResponseWriter, r *http.Request) {

	methode := r.PathValue("methode")

	switch methode {
	case entity.TWOFAUTH_METHODE_TOTP:
		{
			h.HandleTotp(w, r)
			return
		}
	default:
		{
			response.NotFoundError(w)
			return
		}
	}
}

func (h *Handler) HandleTotp(w http.ResponseWriter, r *http.Request) {
	rawToken, err := token.ExtractToken(r)

	totp := service.TOTPVerifyInput{
		Token: *rawToken,
	}
	err = validation.Validate(r.Body, totp)
	if err != nil {
		resp, responseErr := response.ValidationErrors(err)
		if responseErr != nil {
			h.logger.Error("unkown error from validation %v", err)
			response.InternalServerError(w)
			return
		}
		resp.Write(w)
		return
	}
	token, err := h.services.AuthService.VerifyTOTP(r.Context(), totp)
	if err != nil {
		if errors.Is(err, service.ErrorServerFailed) {
			response.InternalServerError(w)
			return
		}
		msg := err.Error()
		response.UnauthorizedError(w, &msg)
		return
	}
	response.OK(w, token)
}
