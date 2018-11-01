// Package models contains all code related to data models.
package models

import (
	"os"
	"regexp"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/jinzhu/gorm"
	"github.com/pkg/errors"
	"golang.org/x/crypto/bcrypt"

	_ "github.com/jinzhu/gorm/dialects/postgres"
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

var (
	unavailableNames = []string{
		"microlog", "legal", "auth", "changelog", "profile", "post", "explore",
	}
)

// User stores the name, biography, posts and identities of a user.
type User struct {
	gorm.Model
	Name       string
	Biography  string
	Admin      bool
	Posts      []Post     `gorm:"foreignkey:UserID"`
	Identities []Identity `gorm:"foreignkey:UserID"`
	Likes      []Like     `gorm:"foreignkey:UserID"`
}

// Identity stores the email, password hash and user.
type Identity struct {
	gorm.Model
	Email     string
	Hash      []byte
	UserID    uint
	Confirmed bool
}

// Report stores the post, report author and reason for report.
type Report struct {
	gorm.Model
	PostID     uint
	ReporterID uint
	Reason     string
}

// Post stores the title, content and author of a post.
type Post struct {
	gorm.Model
	Title    string
	Content  string
	ParentID uint
	UserID   uint
	Likes    []Like `gorm:"foreignkey:PostID"`
}

// Like stores the user and post that got liked.
type Like struct {
	gorm.Model
	UserID uint
	PostID uint
}

// DataSource is a generic source of data.
type DataSource struct {
	db *gorm.DB
}

var log = &logrus.Logger{
	Out:       os.Stderr,
	Hooks:     make(logrus.LevelHooks),
	Formatter: new(logrus.JSONFormatter),
	Level:     logrus.DebugLevel,
}

// Open instantiates a new data source with the given file as a backend.
func Open(path string) (*DataSource, error) {
	log.WithFields(logrus.Fields{
		"path": path,
		"type": "postgres",
	}).Info("accessing database")
	db, err := gorm.Open("postgres", path)
	db.SetLogger(log)
	if err != nil {
		return nil, errors.Wrap(err, "could not create data source")
	}
	db.AutoMigrate(&Identity{}, &User{}, &Post{}, &Report{}, &Like{})
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

// User retrieves the user with the specified ID.
// It returns a reference to the user instance and an error if unsuccessful.
func (data *DataSource) User(id uint) (*User, error) {
	var user User
	data.db.First(&user, id)
	if user.ID != id {
		return nil, errUserNotFound
	}
	return &user, nil
}

// UserByName returns the user identified by the given username.
// It returns a reference to the user instance and an error if unsuccessful.
func (data *DataSource) UserByName(name string) (*User, error) {
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
	namePattern       = `^[a-z0-9]+$`
	passwordMinLength = 8
)

var nameRegexp = regexp.MustCompile(namePattern)

// ValidateName checks if the name satifies the alphanumeric characters-only and length condition.
func (data *DataSource) ValidateName(name string) bool {
	for _, n := range unavailableNames {
		if n == name {
			return false
		}
	}
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

// PostsByUser retrieves the posts by the given user.
// It returns the slice of posts and an error if something unexpected occurs.
func (data *DataSource) PostsByUser(user uint) ([]Post, error) {
	var posts []Post
	data.db.Where("user_id = ?", user).Find(&posts)
	return posts, nil
}

// RecentPosts fetches the most recent posts.
// It takes the maximum number of posts to fetch as a parameter.
// It returns the slice of posts sorted and an error if something unexpected occurs.
func (data *DataSource) RecentPosts(count int) ([]Post, error) {
	var posts []Post
	data.db.Order("created_at DESC").Limit(count).Find(&posts)
	return posts, nil
}

// RecentUsers fetches the most recent users.
// It takes the maximum number of users to fetch as a parameter.
// It returns the slice of users in descending order and an error if something unexpected occurs.
func (data *DataSource) RecentUsers(count int) ([]User, error) {
	var users []User
	data.db.Order("created_at DESC").Limit(count).Find(&users)
	return users, nil
}

// UpdateBiography changes the biography of the given user.
// It returns an error if the biography is not valid or the user does not exist.
func (data *DataSource) UpdateBiography(id uint, biography string) error {
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

// Post returns the post identified by the given unique ID.
// It returns the post and an error if the ID does not exist.
func (data *DataSource) Post(id uint) (*Post, error) {
	var post Post
	data.db.First(&post, id)
	if post.ID != id {
		return nil, errPostNotFound
	}
	return &post, nil
}

// CommentsOn returns the comments on the given post.
// It returns the slice of posts and never an error.
func (data *DataSource) CommentsOn(id uint) ([]Post, error) {
	var comments []Post
	data.db.Where("parent_id = ?", id).Find(&comments)
	return comments, nil
}

// Identities retrieves the identities associated with the given user.
// It returns a slice of identities and an error if no identity can be found.
func (data *DataSource) Identities(user uint) ([]Identity, error) {
	var identities []Identity
	data.db.Where("user_id = ?", user).Find(&identities)
	if len(identities) == 0 {
		return nil, errIdentityNotFound
	}
	return identities, nil
}

// IdentityByEmail retrieves the identity associated with the given email.
// It returns the identity and an error if no identity can be found.
func (data *DataSource) IdentityByEmail(email string) (*Identity, error) {
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

// NumberOfLikes retrieves the number of likes a post has received.
// It returns the count and an error if something unexpected occurs.
func (data *DataSource) NumberOfLikes(id uint) (int, error) {
	var count int
	data.db.Model(&Like{}).Where("post_id = ?", id).Count(&count)
	return count, nil
}

// Likes retrieves the likes of a user.
// It returns a slice of likes and an error if something unexpected occurs.
func (data *DataSource) Likes(id uint) ([]Like, error) {
	var likes []Like
	data.db.Where("user_id = ?", id).Find(&likes)
	return likes, nil
}

// ToggleLike deletes an already existing like and adds a missing one.
// It returns an error if something unexpected occurs.
func (data *DataSource) ToggleLike(user uint, post uint) error {
	var like Like
	data.db.Where("user_id = ? AND post_id = ?", user, post).First(&like)
	if like.PostID != post {
		data.db.Create(&Like{
			PostID: post,
			UserID: user,
		})
	} else {
		data.db.Delete(&like)
	}
	return nil
}

// HasLiked checks if the user has liked the post.
// It returns true if at least one matching like exists.
func (data *DataSource) HasLiked(user, post uint) bool {
	var count int
	data.db.Model(&Like{}).Where("user_id = ? AND post_id = ?", user, post).Count(&count)
	return count > 0
}

const rankingQuery = `
SELECT id, created_at, updated_at, title, content, user_id
FROM posts
LEFT JOIN (
	SELECT post_id, COUNT(*) as votes
	FROM likes
	WHERE deleted_at IS NULL GROUP BY user_id)
AS ranking
ON posts.id = ranking.post_id
WHERE deleted_at IS NULL AND created_at > ?
ORDER BY ranking.votes DESC, created_at DESC
LIMIT ?`

// PopularPosts returns the posts created since the given time ranked by their vote count.
// It returns the slice of posts
func (data *DataSource) PopularPosts(since time.Time, count int) ([]Post, error) {
	var posts []Post
	data.db.Raw(rankingQuery, since, count).Scan(&posts)
	return posts, nil
}
