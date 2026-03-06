package services

import (
	"context"
	"database/sql"
	"errors"
	"strings"
	"time"

	"github.com/Marst/reminder-app/internal/database"
	"github.com/Marst/reminder-app/internal/models"
	"github.com/Marst/reminder-app/internal/utils"
)

func Register(req *models.RegisterRequest) (*models.User, error) {
	Name := strings.TrimSpace(req.Name)
	Email := strings.TrimSpace(req.Email)
	Password := strings.TrimSpace(req.Password)

	if Name == "" || Email == "" || Password == "" {
		return nil, errors.New("name, email, and password cannot be empty!")
	}

	// ! !. Check if user already exists
	var existingID *int
	err := database.DB.QueryRow("SELECT id from users WHERE email = $1", req.Email).Scan(&existingID)

	if err == nil {
		return nil, errors.New("Email already registered")
	}
	if err != sql.ErrNoRows {
		return nil, err
	}

	// ! 2. Hash passsword
	hashedPassword, err := utils.HashedPassword(req.Password)
	if err != nil {
		return nil, err
	}

	// ! 3. CREATE USER OBJECT
	user := &models.User{
		Name:      req.Name,
		Email:     req.Email,
		CreatedAt: time.Now(),
	}

	// ! 4. INSERT TO DATABASE
	var userID int
	err = database.DB.QueryRow(`INSERT INTO users (name, email, password, created_at)
	VALUES ($1, $2, $3, $4)
	RETURNING id`,
		&req.Name,
		&req.Email,
		string(hashedPassword),
		time.Now()).Scan(&userID)

	if err != nil {
		return nil, err
	}

	return user, nil
}

func Login(req *models.LoginRequest) (*models.AuthResponse, error) {
	Email := strings.TrimSpace(req.Email)
	Password := strings.TrimSpace(req.Password)

	// ! 1. VALIDATE INPUT
	if Email == "" || Password == "" {
		return nil, errors.New("Email and password are required!")
	}

	// ! 2. GET USER FROM DB
	var user models.User
	var hashedPassword string

	err := database.DB.QueryRow(`
	SELECT id, name, email, password
	FROM users
	WHERE email = $1`, req.Email).Scan(&user.ID, &user.Name, &user.Email, &hashedPassword)

	if err == sql.ErrNoRows {
		return nil, errors.New("Invalid email or password")
	} else if err != nil {
		return nil, err
	}

	// ! 3. Compare password and hashed password
	err = utils.CompareHashAndPassword(hashedPassword, req.Password)
	if err != nil {
		return nil, errors.New("Invalid email or password")
	}

	// ! 4. Generate JWT
	token, err := utils.GenerateJWT(user.ID, user.Email)
	if err != nil {
		return nil, err
	}

	// ! 5. Send Response
	return &models.AuthResponse{
		Token: token,
		User:  user,
	}, nil

}

func Profile(ctx context.Context, userID int) (*models.User, error) {
	var user models.User

	err := database.DB.QueryRowContext(ctx, `SELECT id, name, email, phone, bio, avatar_url, created_at
	FROM users
	WHERE id = $1`, userID).Scan(&user.ID, &user.Name, &user.Email, &user.Phone, &user.Bio, &user.AvatarURL, &user.CreatedAt)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.New("Profile not found with that id")
		}

	}

	return &user, nil
}

func RefreshCookies(ctx context.Context, userID int) (*models.AuthResponse, error) {
	var user models.User

	err := database.DB.QueryRowContext(ctx, `SELECT id, name, email FROM users 
	WHERE id = $1 LIMIT 1`, userID).Scan(&user.ID, &user.Name, &user.Email)

	if err != nil {
		return nil, errors.New("User is not found")
	}

	token, err := utils.GenerateJWT(user.ID, user.Email)
	if err != nil {
		return nil, err
	}

	return &models.AuthResponse{
		User:  user,
		Token: token,
	}, nil
}
