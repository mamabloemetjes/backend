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

var defaultParams = &structs.ArgonParams{
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
		as.logger.Error("Failed to find user by email", gecho.Field("error", err), gecho.Field("email", authRequest.Email))
		return nil, lib.MapPgError(err)
	}

	// Check if user was found (First() can return nil, nil for no results)
	if user == nil {
		as.logger.Warn("User not found", gecho.Field("email", authRequest.Email))
		return nil, lib.ErrInvalidCredentials
	}

	// Verify password
	valid, err := as.VerifyPassword(authRequest.Password, user.PasswordHash)
	if err != nil {
		as.logger.Error("Failed to verify password", gecho.Field("error", err))
		return nil, err
	}
	if !valid {
		as.logger.Warn("Invalid password attempt", gecho.Field("email", authRequest.Email))
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
	passwordHash, err := as.HashPassword(registerRequest.Password, defaultParams)
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
	// Check if it a conflict error (e.g., duplicate email)
	if err != nil {
		return nil, lib.MapPgError(err)
	}

	elapsedTime := time.Since(startTime)
	as.logger.Debug("User registered successfully", gecho.Field("user_id", user.Id), gecho.Field("elapsed_time_ms", elapsedTime.Milliseconds()))

	// Remove password hash before returning user
	user.PasswordHash = ""

	// Set user in cache
	cacheErr := as.cacheService.SetUserInCache(user)
	if cacheErr != nil {
		as.logger.Warn("Failed to set user in cache after registration", gecho.Field("error", cacheErr), gecho.Field("user_id", user.Id))
	}

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
