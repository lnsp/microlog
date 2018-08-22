package models

import (
	"os"
	"regexp"

	"github.com/Sirupsen/logrus"
	"github.com/jinzhu/gorm"
	"github.com/pkg/errors"
	"golang.org/x/crypto/bcrypt"

	// need sqlite3 context for gorm
	_ "github.com/mattn/go-sqlite3"
)

const (
	postTitleMaxLength    = 80
	postContentMaxLength  = 80000
	reportReasonMaxLength = 240
	usernameMaxLength     = 24
	biographyMaxLength    = 240
)

var (
	errPostNotFound             = errors.New("could not find post")
	errUserNotFound             = errors.New("could not find user")
	errIdentityNotFound         = errors.New("could not find identity")
	errIdentityAlreadyConfirmed = errors.New("identity is already confirmed")
	errPostNotOwned             = errors.New("can not edit alien post")
	errValidation               = errors.New("can not validate params")
)

type User struct {
	gorm.Model
	Name       string
	Biography  string
	Posts      []Post     `gorm:"foreignkey:UserID"`
	Identities []Identity `gorm:"foreignkey:UserID"`
}

type Identity struct {
	gorm.Model
	Email     string
	Hash      []byte
	UserID    uint
	Confirmed bool
}

type Report struct {
	gorm.Model
	PostID     uint
	ReporterID uint
	Reason     string
}

type Post struct {
	gorm.Model
	Title   string
	Content string
	UserID  uint
}

type DataSource struct {
	db *gorm.DB
}

var log = &logrus.Logger{
	Out:       os.Stderr,
	Hooks:     make(logrus.LevelHooks),
	Formatter: new(logrus.TextFormatter),
	Level:     logrus.DebugLevel,
}

func Open(path string) (*DataSource, error) {
	log.Infoln("Opening sqlite3 database", path)
	db, err := gorm.Open("sqlite3", path)
	db.SetLogger(log)
	if err != nil {
		return nil, errors.Wrap(err, "could not create data source")
	}
	db.AutoMigrate(&Identity{}, &User{}, &Post{})
	return &DataSource{db}, nil
}

// ResetPassword sets the password of the related identity.
// It returns an error if the action was unsuccessful.
func (data *DataSource) ResetPassword(user uint, email string, password []byte) error {
	hash, err := bcrypt.GenerateFromPassword(password, bcrypt.DefaultCost)
	if err != nil {
		return errors.Wrap(err, "could not generate hash")
	}
	var identity Identity
	data.db.Where("user_id = ? AND email = ?", user, email).First(&identity)
	if identity.UserID != user {
		return errIdentityNotFound
	}
	identity.Hash = hash
	data.db.Save(&identity)
	return nil
}

// HasUser checks if the user identified by the given email and password exists.
// It returns the user ID, the confirmation state of the user's identity and error value.
func (data *DataSource) HasUser(email string, password []byte) (uint, bool, error) {
	var id Identity
	data.db.Where("email = ?", email).First(&id)
	if id.Email != email {
		return 0, false, errIdentityNotFound
	}
	if err := bcrypt.CompareHashAndPassword(id.Hash, password); err != nil {
		return 0, false, errIdentityNotFound
	}
	return id.UserID, id.Confirmed, nil
}

// UpdatePost updates the title and content of a specific post.
// This action can only be applied to posts owned by the given user ID.
// It returns any error if the action is unsuccessful.
func (data *DataSource) UpdatePost(userID, postID uint, title, content string) error {
	var post Post
	data.db.First(&post, postID)
	if post.ID != postID {
		return errPostNotFound
	}
	if post.UserID != userID {
		return errPostNotOwned
	}
	post.Title = title
	post.Content = content
	data.db.Save(&post)
	return nil
}

// AddUser creates a new user with a new default identity.
// It returns the user ID and an error if the action is unsuccessful.
func (data *DataSource) AddUser(name, email string, password []byte) (uint, error) {
	if !data.ValidateName(name) || !data.ValidateEmail(email) || !data.ValidatePassword(string(password)) {
		return 0, errValidation
	}
	hash, err := bcrypt.GenerateFromPassword(password, bcrypt.DefaultCost)
	if err != nil {
		return 0, errors.Wrap(err, "could not generate hash")
	}
	id := Identity{
		Email:     email,
		Hash:      hash,
		Confirmed: false,
	}
	user := User{
		Name:       name,
		Identities: []Identity{id},
	}
	data.db.Create(&user)
	return user.ID, nil
}

// ValidateReportReason checks if the given reason string satisfies the conditions for report reasons.
// It returns true if the report reason is valid.
func (data *DataSource) ValidateReportReason(reason string) bool {
	return len(reason) <= reportReasonMaxLength
}

// AddReport creates a new report of the given post with a specific reason.
// It returns an error if the report is unsuccessful.
func (data *DataSource) AddReport(postID, reporterID uint, reason string) error {
	if !data.ValidateReportReason(reason) {
		return errValidation
	}
	var reporter User
	data.db.First(&reporter, reporterID)
	if reporter.ID != reporterID {
		return errUserNotFound
	}
	var post Post
	data.db.First(&post, postID)
	if post.ID != postID {
		return errPostNotFound
	}
	report := Report{
		PostID:     postID,
		ReporterID: reporterID,
		Reason:     reason,
	}
	data.db.Create(&report)
	return nil
}

// EmailExists checks if the email already is used by anotherr identity.
func (data *DataSource) EmailExists(email string) bool {
	var count int
	data.db.Model(&Identity{}).Where("email = ?", email).Count(&count)
	return count > 0
}

// NameExists checks if the name already is used by another user.
func (data *DataSource) NameExists(name string) bool {
	var count int
	data.db.Model(&User{}).Where("name = ?", name).Count(&count)
	return count > 0
}

// GetUser retrieves the user with the specified ID.
// It returns a reference to the user instance and an error if unsuccessful.
func (data *DataSource) GetUser(id uint) (*User, error) {
	var user User
	data.db.First(&user, id)
	if user.ID != id {
		return nil, errUserNotFound
	}
	return &user, nil
}

// GetUserByName returns the user identified by the given username.
// It returns a reference to the user instance and an error if unsuccessful.
func (data *DataSource) GetUserByName(name string) (*User, error) {
	var user User
	data.db.Where("name = ?", name).First(&user)
	if user.Name != name {
		return nil, errUserNotFound
	}
	return &user, nil
}

// DeleteUser deletes a user from the database.
func (data *DataSource) DeleteUser(id uint) {
	data.db.Delete(&Identity{}, "user_id = ?", id)
	data.db.Delete(&Post{}, "user_id = ?", id)
	data.db.Delete(&User{}, "id = ?", id)
}

// DeletePost deletes a specific post.
func (data *DataSource) DeletePost(user, id uint) error {
	var post Post
	data.db.First(&post, id)
	if post.ID != id {
		return errPostNotFound
	}
	if post.UserID != user {
		return errPostNotOwned
	}
	data.db.Delete(&post)
	return nil
}

const (
	namePattern       = `^[a-zA-Z0-9]+$`
	passwordMinLength = 8
)

var nameRegexp = regexp.MustCompile(namePattern)

// ValidateName checks if the name satifies the alphanumeric characters-only and length condition.
func (data *DataSource) ValidateName(name string) bool {
	return nameRegexp.Match([]byte(name)) && len(name) <= usernameMaxLength
}

// ValidatePassword checks if the password is longer than the given minimum length.
func (data *DataSource) ValidatePassword(password string) bool {
	return len(password) >= passwordMinLength
}

// ValidateEmail checks if the given address may be a real email.
func (data *DataSource) ValidateEmail(email string) bool {
	for _, c := range email {
		if c == '@' {
			return true
		}
	}
	return false
}

// GetPostsByUser retrieves the posts by the given user.
// It returns the slice of posts and an error if something unexpected occurs.
func (data *DataSource) GetPostsByUser(user uint) ([]Post, error) {
	var posts []Post
	data.db.Where("user_id = ?", user).Find(&posts)
	return posts, nil
}

// GetRecentPosts fetches the most recent posts.
// It takes the maximum number of posts to fetch as a parameter.
// It returns the slice of posts sorted and an error if something unexpected occurs.
func (data *DataSource) GetRecentPosts(count int) ([]Post, error) {
	var posts []Post
	data.db.Order("created_at DESC").Limit(count).Find(&posts)
	return posts, nil
}

// GetRecentUsers fetches the most recent users.
// It takes the maximum number of users to fetch as a parameter.
// It returns the slice of users in descending order and an error if something unexpected occurs.
func (data *DataSource) GetRecentUsers(count int) ([]User, error) {
	var users []User
	data.db.Order("created_at DESC").Limit(count).Find(&users)
	return users, nil
}

// SetBiography changes the biography of the given user.
// It returns an error if the biography is not valid or the user does not exist.
func (data *DataSource) SetBiography(id uint, biography string) error {
	if !data.ValidateBiography(biography) {
		return errValidation
	}
	var user User
	data.db.First(&user, id)
	if user.ID != id {
		return errUserNotFound
	}
	user.Biography = biography
	data.db.Save(&user)
	return nil
}

// ValidateBiography checks if the biography text is valid.
func (data *DataSource) ValidateBiography(biography string) bool {
	return len(biography) <= biographyMaxLength
}

// ValidatePostTitle checks if the post title is valid.
func (data *DataSource) ValidatePostTitle(title string) bool {
	return len(title) <= postTitleMaxLength
}

// ValidatePostContent checks if the post content is valid.
func (data *DataSource) ValidatePostContent(content string) bool {
	return len(content) <= postContentMaxLength
}

// AddPost creates a new post by the given user and with the given title and content.
// It returns the ID of the post and an error if the params are invalid.
func (data *DataSource) AddPost(author uint, title, content string) (uint, error) {
	if !data.ValidatePostTitle(title) || !data.ValidatePostContent(content) {
		return 0, errValidation
	}
	post := Post{
		UserID:  author,
		Title:   title,
		Content: content,
	}
	data.db.Create(&post)
	return post.ID, nil
}

// GetPost returns the post identified by the given unique ID.
// It returns the post and an error if the ID does not exist.
func (data *DataSource) GetPost(id uint) (*Post, error) {
	var post Post
	data.db.First(&post, id)
	if post.ID != id {
		return nil, errPostNotFound
	}
	return &post, nil
}

// GetIdentities retrieves the identities associated with the given user.
// It returns a slice of identities and an error if no identity can be found.
func (data *DataSource) GetIdentities(user uint) ([]Identity, error) {
	var identities []Identity
	data.db.Where("user_id = ?", user).Find(&identities)
	if len(identities) == 0 {
		return nil, errIdentityNotFound
	}
	return identities, nil
}

// GetIdentityByEmail retrieves the identity associated with the given email.
// It returns the identity and an error if no identity can be found.
func (data *DataSource) GetIdentityByEmail(email string) (*Identity, error) {
	var identity Identity
	data.db.Where("email = ?", email).First(&identity)
	if identity.Email != email {
		return nil, errIdentityNotFound
	}
	return &identity, nil
}

// ConfirmIdentity confirms a previously unconfirmed identity.
// It returns an error if the identity does not exist or is already confirmed.
func (data *DataSource) ConfirmIdentity(user uint, email string) error {
	var identity Identity
	data.db.Where("user_id = ? AND email = ?", user, email).First(&identity)
	if identity.UserID != user {
		return errIdentityNotFound
	}
	if identity.Confirmed {
		return errIdentityAlreadyConfirmed
	}
	identity.Confirmed = true
	data.db.Save(&identity)
	return nil
}
