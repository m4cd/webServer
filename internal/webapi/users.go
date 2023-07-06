package webapi

import (
	"errors"
	"fmt"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

// getUsers returns all users in the database
func (db *DB) getUsers() ([]User, error) {
	dbUsers, err := db.loadDB()
	if err == nil {
		mapUsers := []User{}

		for _, u := range dbUsers.Users {
			mapUsers = append(mapUsers, u)
		}

		return mapUsers, nil
	}
	return []User{}, err
}

func (db *DB) CreateUser(pwd string, email string) (ResponseUser, error) {
	dbContents, _ := db.loadDB()
	dbUsers, err := db.getUsers()

	if err != nil {
		fmt.Println("Error while loading db")
	}

	UserFound, err := db.findUserByEmail(email)
	if err == nil {
		return ResponseUser{ID: UserFound.ID, Email: UserFound.Email}, errors.New("User already exists")
	}

	// find max id
	max := 0
	for _, u := range dbContents.Users {
		if u.ID > max {
			max = u.ID
		}
	}

	newID := max + 1
	hash, err := bcrypt.GenerateFromPassword([]byte(pwd), 0)
	if err != nil {
		fmt.Println("Password Hashing Error: ", err)
	}
	user := User{
		ID:       newID,
		Email:    email,
		Password: string(hash),
	}

	dbUsers = append(dbUsers, user)
	dbUsersMap := make(map[int]User)

	for _, u := range dbUsers {
		us := User{
			ID:       u.ID,
			Email:    u.Email,
			Password: u.Password,
		}
		dbUsersMap[u.ID] = us
	}
	dbUsersStruct := dbContents
	dbUsersStruct.Users = dbUsersMap

	db.writeDB(dbUsersStruct)

	return ResponseUser{ID: user.ID, Email: user.Email}, nil

}

func (db *DB) findUserByEmail(email string) (User, error) {
	dbUsers, err := db.getUsers()

	if err != nil {
		fmt.Println("Error while loading db")
	}

	for _, u := range dbUsers {
		if u.Email == email {
			return u, nil
		}
	}

	return User{}, errors.New("User not found")
}

func (db *DB) findUserById(id int) (User, error) {
	dbUsers, err := db.getUsers()

	if err != nil {
		fmt.Println("Error while loading db")
	}

	for _, u := range dbUsers {
		if u.ID == id {
			return u, nil
		}
	}

	return User{}, errors.New("User not found")
}

func (db *DB) VerifyCredentials(email string, password string) (ResponseUser, error) {

	UserFound, err := db.findUserByEmail(email)
	if err != nil {
		return ResponseUser{}, errors.New("User not found")
	}

	PasswordIsCorrect := bcrypt.CompareHashAndPassword([]byte(UserFound.Password), []byte(password))

	if PasswordIsCorrect == nil {
		return ResponseUser{ID: UserFound.ID, Email: UserFound.Email, ChirpyRed: UserFound.ChirpyRed}, nil
	}

	return ResponseUser{}, errors.New("Incorrect password.")

}

func (db *DB) UpdateUser(id int, email string, password string) (ResponseUser, error) {
	dbContents, _ := db.loadDB()
	dbUsers, _ := db.getUsers()

	UserFound, err := db.findUserById(id)

	if err != nil {
		return ResponseUser{}, errors.New("User not found")
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), 0)
	if err != nil {
		fmt.Println("Password Hashing Error: ", err)
	}
	UserFound = User{
		ID:       id,
		Email:    email,
		Password: string(hash),
	}

	dbUsers = append(dbUsers, UserFound)
	dbUsersMap := make(map[int]User)

	for _, u := range dbUsers {
		us := User{
			ID:       u.ID,
			Email:    u.Email,
			Password: u.Password,
		}
		dbUsersMap[u.ID] = us
	}
	dbUsersStruct := dbContents
	dbUsersStruct.Users = dbUsersMap

	db.writeDB(dbUsersStruct)

	return ResponseUser{ID: UserFound.ID, Email: UserFound.Email}, nil
}

func (db *DB) RevokeToken(tokenString string, RevokeTime jwt.NumericDate) (RevokedToken, error) {
	dbContents, _ := db.loadDB()

	RevokedTokens := make(map[string]jwt.NumericDate)

	for token, time := range dbContents.RevokedTokens {
		RevokedTokens[token] = time
	}

	RevokedTokens[tokenString] = RevokeTime
	dbContents.RevokedTokens = RevokedTokens

	db.writeDB(dbContents)

	return RevokedToken{TokenString: tokenString, RevokeTime: RevokeTime}, nil
}

func (db *DB) CheckToken(tokenString string) bool {
	dbContents, _ := db.loadDB()

	_, ok := dbContents.RevokedTokens[tokenString]
	if ok {
		return true

	}
	return false

}

func (db *DB) ChirpyRedUpdateUser(userID int) error {
	dbContents, _ := db.loadDB()

	dbUser, err := db.findUserById(userID)

	if err != nil {
		return errors.New("User not found")
	}

	dbUser.ChirpyRed = true

	dbContents.Users[userID] = dbUser

	db.writeDB(dbContents)
	return nil
}
