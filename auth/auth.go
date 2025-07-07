package auth

import (
	"context"
	"fynance/helpers"
	"fynance/models"
	"fynance/utils"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/dialog"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"golang.org/x/crypto/bcrypt"
)

func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(bytes), err
}

func CheckPasswordHash(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

func Register(username, phone, password string) error {
	if err := helpers.ValidateUsername(username); err != nil {
		return err
	}

	if err := helpers.ValidatePassword(password); err != nil {
		return err
	}

	if err := helpers.ValidatePhoneNumber(phone); err != nil {
		return err
	}

	hashedPassword, err := HashPassword(password)
	if err != nil {
		return err
	}

	userCollection := utils.GetCollection("users")
	_, err = userCollection.InsertOne(context.Background(), models.User{
		ID:       primitive.NewObjectID(), // Generate a new ID for the user
		Username: username,
		Phone:    phone,
		Password: hashedPassword,
	})

	return err
}

// login user
func Login(username, password string, updateProgress func(progress float64)) (*models.User, error) {
	userCollection := utils.GetCollection("users")
	var user models.User

	// Step 1: Update progress for finding the user in the database
	updateProgress(0.3) // 30% progress
	err := userCollection.FindOne(context.Background(), bson.M{"username": username}).Decode(&user)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			updateProgress(0.0) // Reset progress on failure
			return nil, err
		}
		updateProgress(0.0) // Reset progress on failure
		return nil, err
	}

	// Step 2: Update progress for password validation
	updateProgress(0.5) // 70% progress
	if !CheckPasswordHash(password, user.Password) {
		updateProgress(0.0) // Reset progress on failure
		return nil, mongo.ErrNoDocuments
	}

	// Step 3: Finalize progress on successful login
	updateProgress(1.0) // 100% progress
	return &user, nil
}

// UpdateUserPassword updates the user's password in the database.
func UpdateUserPassword(userID primitive.ObjectID, password string, window fyne.Window) error {
	collection := utils.GetCollection("users")

	newHashedPassword, err := HashPassword(password)

	if err != nil {
		return err
	}

	// Update the user's password field in the database.
	_, err = collection.UpdateOne(
		context.Background(),
		bson.M{"_id": userID},
		bson.M{"$set": bson.M{"password": newHashedPassword}},
	)

	if err != nil {
		dialog.ShowError(err, window)
		return err
	}

	return nil
}
