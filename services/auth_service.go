package services

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"mamabloemetjes_server/database"
	"mamabloemetjes_server/lib"
	"mamabloemetjes_server/structs"
	"mamabloemetjes_server/structs/tables"
	"time"

	"github.com/MonkyMars/gecho"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/argon2"
)

var DefaultParams = &structs.ArgonParams{
	Memory:  64 * 1024, // 64 MB
	Time:    1,
	Threads: 4,
	KeyLen:  32,
	SaltLen: 16,
}

type AuthService struct {
	logger       *gecho.Logger
	cfg          *structs.Config
	db           *database.DB
	cacheService *CacheService
}

func NewAuthService(cfg *structs.Config, logger *gecho.Logger, db *database.DB) *AuthService {
	return &AuthService{
		logger:       logger,
		cfg:          cfg,
		db:           db,
		cacheService: NewCacheService(logger, cfg),
	}
}

func (as *AuthService) Login(authRequest *structs.AuthRequest) (*tables.User, error) {
	startTime := time.Now()
	user, err := database.Query[tables.User](as.db).Where("email", authRequest.Email).First(context.Background())
	if err != nil {
		// Map database error to user-friendly message
		mappedErr := lib.MapPgError(err)

		// Use debug logging for database query errors (could be legitimate "user not found")
		as.logger.Debug("Database query during login",
			gecho.Field("identifier", authRequest.Email),
			gecho.Field("error_detail", lib.GetDetailForLogging(mappedErr)),
		)

		// Only log as error if it's not a "not found" error
		if !lib.IsNotFound(mappedErr) {
			as.logger.Error("Unexpected database error during login",
				gecho.Field("error", mappedErr),
				gecho.Field("original_error", err),
			)
		}

		// Always return invalid credentials (don't leak user existence)
		return nil, lib.ErrInvalidCredentials
	}

	// Check if user was found (First() can return nil, nil for no results)
	if user == nil {
		as.logger.Debug("User not found during login attempt", gecho.Field("identifier", authRequest.Email))
		return nil, lib.ErrInvalidCredentials
	}

	// Verify password
	valid, err := as.VerifyPassword(authRequest.Password, user.PasswordHash)
	if err != nil {
		as.logger.Error("Failed to verify password hash",
			gecho.Field("error", err),
			gecho.Field("user_id", user.Id),
		)
		return nil, err
	}
	if !valid {
		as.logger.Debug("Invalid password attempt",
			gecho.Field("identifier", authRequest.Email),
			gecho.Field("user_id", user.Id),
		)
		return nil, lib.ErrInvalidCredentials
	}

	elapsedTime := time.Since(startTime)
	as.logger.Debug("User logged in successfully", gecho.Field("user_id", user.Id), gecho.Field("elapsed_time_ms", elapsedTime.Milliseconds()))

	// Remove password hash before returning user
	user.PasswordHash = ""

	// Set user in cache
	cacheErr := as.cacheService.SetUserInCache(user)
	if cacheErr != nil {
		as.logger.Warn("Failed to set user in cache after login", gecho.Field("error", cacheErr), gecho.Field("user_id", user.Id))
	}

	return user, nil
}

func (as *AuthService) Register(registerRequest *structs.RegisterRequest) (*tables.User, error) {
	startTime := time.Now()
	passwordHash, err := as.HashPassword(registerRequest.Password, DefaultParams)
	if err != nil {
		as.logger.Error("Failed to hash password", gecho.Field("error", err))
		return nil, err
	}
	user := &tables.User{
		Username:     registerRequest.Username,
		Email:        registerRequest.Email,
		PasswordHash: passwordHash,
	}
	user, err = database.Query[tables.User](as.db).Insert(context.Background(), user)
	if err != nil {
		// Map the error to a user-friendly message
		mappedErr := lib.MapPgError(err)

		// Log unique violations as warnings (user error)
		if lib.IsUniqueViolation(mappedErr) {
			as.logger.Warn("Registration failed - duplicate user",
				gecho.Field("username", registerRequest.Username),
				gecho.Field("email", registerRequest.Email),
			)
		} else {
			// Log other database errors as errors
			as.logger.Error("Database error during registration",
				gecho.Field("error", mappedErr),
				gecho.Field("username", registerRequest.Username),
			)
		}

		return nil, mappedErr
	}

	elapsedTime := time.Since(startTime)
	as.logger.Debug("User registered successfully", gecho.Field("user_id", user.Id), gecho.Field("elapsed_time_ms", elapsedTime.Milliseconds()))

	// Remove password hash before returning user
	user.PasswordHash = ""

	return user, nil
}

// HashPassword hashes a plain-text password and returns a string and possible error
func (as *AuthService) HashPassword(password string, p *structs.ArgonParams) (string, error) {
	salt, err := generateSalt(p.SaltLen)
	if err != nil {
		return "", err
	}
	hash := argon2.IDKey([]byte(password), salt, p.Time, p.Memory, p.Threads, p.KeyLen)
	b64Salt := base64.RawStdEncoding.EncodeToString(salt)
	b64Hash := base64.RawStdEncoding.EncodeToString(hash)
	// format: $argon2id$v=19$m=65536,t=1,p=4$<salt>$<hash>
	params := fmt.Sprintf("m=%d,t=%d,p=%d", p.Memory, p.Time, p.Threads)
	encoded := fmt.Sprintf("$argon2id$v=19$%s$%s$%s", params, b64Salt, b64Hash)
	return encoded, nil
}

func generateSalt(n uint32) ([]byte, error) {
	b := make([]byte, n)
	_, err := rand.Read(b)
	return b, err
}

// VerifyPassword verifies a plain-text password against a hashed password
func (as *AuthService) VerifyPassword(password, hashedPassword string) (bool, error) {
	parts, err := lib.DecodeArgon2Hash(hashedPassword)
	if err != nil {
		return false, err
	}

	// Hash the input password with the same parameters
	hash := argon2.IDKey([]byte(password), parts.Salt, parts.Time, parts.Memory, parts.Threads, parts.KeyLen)

	// Compare the hashes
	return lib.SecureCompare(hash, parts.Hash), nil
}

// GenerateAccessToken generates a JWT access token for the given user
func (as *AuthService) GenerateAccessToken(user *tables.User) (string, error) {
	secret := as.cfg.Auth.AccessTokenSecret

	now := time.Now()
	exp := as.GetAccessTokenExpiration()

	claims := &structs.AuthClaims{
		Sub:   user.Id,
		Email: user.Username,
		Role:  user.Role,
		Iat:   now,
		Exp:   exp,
		Jti:   uuid.New(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub":   claims.Sub.String(),
		"email": claims.Email,
		"role":  claims.Role,
		"iat":   claims.Iat.Unix(),
		"exp":   claims.Exp.Unix(),
		"jti":   claims.Jti.String(),
	})
	return token.SignedString([]byte(secret))
}

// GetAccessTokenExpiration returns the expiration time for access tokens
func (as *AuthService) GetAccessTokenExpiration() time.Time {
	return time.Now().Add(time.Duration(as.cfg.Auth.AccessTokenExpiry))
}

// GenerateRefreshToken generates a JWT refresh token for the given user
func (as *AuthService) GenerateRefreshToken(user *tables.User) (string, error) {
	secret := as.cfg.Auth.RefreshTokenSecret

	now := time.Now()
	exp := as.GetRefreshTokenExpiration()

	claims := &structs.AuthClaims{
		Sub:   user.Id,
		Email: user.Username,
		Role:  user.Role,
		Iat:   now,
		Exp:   exp,
		Jti:   uuid.New(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub":   claims.Sub.String(),
		"email": claims.Email,
		"role":  claims.Role,
		"iat":   claims.Iat.Unix(),
		"exp":   claims.Exp.Unix(),
		"jti":   claims.Jti.String(),
	})
	return token.SignedString([]byte(secret))
}

// GetRefreshTokenExpiration returns the expiration time for refresh tokens
func (as *AuthService) GetRefreshTokenExpiration() time.Time {
	return time.Now().Add(time.Duration(as.cfg.Auth.RefreshTokenExpiry))
}

func (as *AuthService) RefreshAccessToken(refreshToken string) (*tables.AuthResponse, error) {
	claims, err := lib.ParseToken(refreshToken, false, as.cfg.Auth.RefreshTokenSecret)
	if err != nil {
		as.logger.Error("Failed to parse refresh token", gecho.Field("error", err))
		return nil, lib.ErrInvalidToken
	}

	if time.Now().After(claims.Exp) {
		as.logger.Warn("Refresh token has expired", gecho.Field("exp", claims.Exp))
		return nil, lib.ErrExpiredToken
	}

	// TODO: check for blacklisted/revoked tokens
	isBlacklisted, err := as.cacheService.IsTokenBlacklisted(claims.Jti)
	if err != nil {
		as.logger.Error("Failed to check if token is blacklisted", gecho.Field("error", err), gecho.Field("jti", claims.Jti))
		return nil, err
	}

	if isBlacklisted {
		as.logger.Warn("Refresh token is blacklisted", gecho.Field("jti", claims.Jti))
		return nil, lib.ErrInvalidToken
	}

	// get user
	user, err := as.GetUserByID(claims.Sub)
	if err != nil {
		as.logger.Error("Failed to get user by ID during token refresh", gecho.Field("error", err), gecho.Field("user_id", claims.Sub))
		return nil, err
	}

	// generate new tokens
	newAccessToken, err := as.GenerateAccessToken(user)
	if err != nil {
		as.logger.Error("Failed to generate new access token during refresh", gecho.Field("error", err), gecho.Field("user_id", user.Id))
		return nil, err
	}

	newRefreshToken, err := as.GenerateRefreshToken(user)
	if err != nil {
		as.logger.Error("Failed to generate new refresh token during refresh", gecho.Field("error", err), gecho.Field("user_id", user.Id))
		return nil, err
	}

	return &tables.AuthResponse{
		User:         user,
		AccessToken:  newAccessToken,
		RefreshToken: newRefreshToken,
	}, nil
}

func (as *AuthService) GetUserByID(userId uuid.UUID) (*tables.User, error) {
	// Try to get user from cache first
	cachedUser, err := as.cacheService.GetUserFromCache(userId)
	if err != nil {
		as.logger.Warn("Failed to get user from cache", gecho.Field("error", err), gecho.Field("user_id", userId))
	} else if cachedUser != nil {
		as.logger.Debug("User retrieved from cache", gecho.Field("user_id", userId))
		return cachedUser, nil
	}

	// Cache miss - fetch user from database
	user, err := database.Query[tables.User](as.db).Where("id", userId).First(context.Background())
	if err != nil {
		as.logger.Error("Failed to find user by ID", gecho.Field("error", err), gecho.Field("user_id", userId))
		return nil, lib.MapPgError(err)
	}

	// Cache the user asynchronously
	go func() {
		if err := as.cacheService.SetUserInCache(user); err != nil {
			as.logger.Warn("Failed to cache user after DB fetch", gecho.Field("error", err), gecho.Field("user_id", userId))
		}
	}()

	return user, nil
}

func (as *AuthService) GetAccessTokenSecret() string {
	secret := as.cfg.Auth.AccessTokenSecret
	return secret
}

func (as *AuthService) GetRefreshTokenSecret() string {
	secret := as.cfg.Auth.RefreshTokenSecret
	return secret
}

func (as *AuthService) UpdateLastLogin(userId uuid.UUID) error {
	updates := map[string]any{
		"last_login": time.Now(),
	}
	_, err := database.Query[tables.User](as.db).Where("id", userId).Update(context.Background(), updates)
	if err != nil {
		return lib.MapPgError(err)
	}
	return nil
}

func (as *AuthService) VerifyEmail(userId uuid.UUID, token string) error {
	// Get verification record
	verification, err := database.Query[tables.EmailVerification](as.db).
		Where("user_id", userId).
		Where("token", token).
		First(context.Background())
	if err != nil {
		as.logger.Error("Failed to find email verification record", gecho.Field("error", err), gecho.Field("user_id", userId))
		return lib.MapPgError(err)
	}
	if verification == nil {
		as.logger.Warn("Email verification record not found", gecho.Field("user_id", userId))
		return lib.ErrInvalidToken
	}

	// Check if token is expired
	if time.Now().After(verification.ExpiresAt) {
		as.logger.Warn("Email verification token has expired", gecho.Field("user_id", userId), gecho.Field("expires_at", verification.ExpiresAt))
		return lib.ErrExpiredToken
	}

	if token != verification.Token {
		as.logger.Warn("Email verification token does not match", gecho.Field("user_id", userId))
		return lib.ErrInvalidToken
	}

	// Update user to set email as verified
	updates := map[string]any{
		"email_verified": true,
	}
	_, err = database.Query[tables.User](as.db).Where("id", userId).Update(context.Background(), updates)
	if err != nil {
		as.logger.Error("Failed to update user email verification status", gecho.Field("error", err), gecho.Field("user_id", userId))
		return lib.MapPgError(err)
	}

	// Delete verification record
	_, err = database.Query[tables.EmailVerification](as.db).Where("id", verification.Id).Delete(context.Background())
	if err != nil {
		as.logger.Error("Failed to delete email verification record", gecho.Field("error", err), gecho.Field("user_id", userId))
		return lib.MapPgError(err)
	}

	as.logger.Info("Email verified successfully", gecho.Field("user_id", userId))
	return nil
}

// GetDB returns the database instance (helper method for accessing db)
func (as *AuthService) GetDB() *database.DB {
	return as.db
}
