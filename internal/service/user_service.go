package service

import (
	"context"
	"errors"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"

	"github.com/Over-knight/Lujay-assesment/internal/auth"
	"github.com/Over-knight/Lujay-assesment/internal/models"
)

// UserService handles user-related business logic
type UserService struct {
	collection *mongo.Collection
	jwtManager *auth.JWTManager
}

// NewUserService creates a new user service instance
// collection: MongoDB collection for users
// jwtManager: JWT manager for token operations
func NewUserService(collection *mongo.Collection, jwtManager *auth.JWTManager) *UserService {
	return &UserService{
		collection: collection,
		jwtManager: jwtManager,
	}
}

// Register creates a new user account
// ctx: Context for the operation
// req: Registration request with user details
// Returns authentication response with token and user data, or an error
func (s *UserService) Register(ctx context.Context, req models.RegisterRequest) (*models.AuthResponse, error) {
	// Check if user already exists
	var existingUser models.User
	err := s.collection.FindOne(ctx, bson.M{"email": req.Email}).Decode(&existingUser)
	if err == nil {
		return nil, errors.New("user with this email already exists")
	}
	if err != mongo.ErrNoDocuments {
		return nil, err
	}

	// Hash password
	hashedPassword, err := auth.HashPassword(req.Password)
	if err != nil {
		return nil, errors.New("failed to hash password")
	}

	// Set default role if not provided
	role := req.Role
	if role == "" {
		role = models.RoleBuyer
	}

	// Create user object
	user := models.User{
		ID:        primitive.NewObjectID(),
		Email:     req.Email,
		Password:  hashedPassword,
		FirstName: req.FirstName,
		LastName:  req.LastName,
		Role:      role,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// Insert user into database
	_, err = s.collection.InsertOne(ctx, user)
	if err != nil {
		return nil, errors.New("failed to create user")
	}

	// Generate JWT token
	token, err := s.jwtManager.GenerateToken(user.ID.Hex(), user.Email)
	if err != nil {
		return nil, errors.New("failed to generate token")
	}

	// Return response with token and user data
	return &models.AuthResponse{
		Token: token,
		User:  user,
	}, nil
}

// Login authenticates a user and returns a token
// ctx: Context for the operation
// req: Login request with credentials
// Returns authentication response with token and user data, or an error
func (s *UserService) Login(ctx context.Context, req models.LoginRequest) (*models.AuthResponse, error) {
	// Find user by email
	var user models.User
	err := s.collection.FindOne(ctx, bson.M{"email": req.Email}).Decode(&user)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, errors.New("invalid email or password")
		}
		return nil, err
	}

	// Verify password
	if !auth.ComparePasswords(user.Password, req.Password) {
		return nil, errors.New("invalid email or password")
	}

	// Generate JWT token
	token, err := s.jwtManager.GenerateToken(user.ID.Hex(), user.Email)
	if err != nil {
		return nil, errors.New("failed to generate token")
	}

	// Return response with token and user data
	return &models.AuthResponse{
		Token: token,
		User:  user,
	}, nil
}

// GetUserByID retrieves a user by their ID
// ctx: Context for the operation
// userID: The user's ID as a string
// Returns the user or an error if not found
func (s *UserService) GetUserByID(ctx context.Context, userID string) (*models.User, error) {
	// Convert string ID to ObjectID
	objectID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return nil, errors.New("invalid user ID")
	}

	// Find user by ID
	var user models.User
	err = s.collection.FindOne(ctx, bson.M{"_id": objectID}).Decode(&user)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, errors.New("user not found")
		}
		return nil, err
	}

	return &user, nil
}
