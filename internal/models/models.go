package models

import (
	"os"
	"regexp"

	"github.com/Sirupsen/logrus"
	"github.com/jinzhu/gorm"
	_ "github.com/mattn/go-sqlite3"
	"github.com/pkg/errors"
	"golang.org/x/crypto/bcrypt"
)

const (
	postTitleMaxLength    = 80
	postContentMaxLength  = 80000
	reportReasonMaxLength = 240
	usernameMaxLength     = 24
	biographyMaxLength    = 240
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

func (data *DataSource) HasUser(email string, password []byte) (uint, error) {
	var id Identity
	data.db.Where("email = ?", email).First(&id)
	if id.Email != email {
		return 0, errors.New("email does not exist")
	}
	if err := bcrypt.CompareHashAndPassword(id.Hash, password); err != nil {
		return 0, errors.New("passwords do not match")
	}
	return id.UserID, nil
}

func (data *DataSource) UpdatePost(userID, postID uint, title, content string) error {
	var post Post
	data.db.First(&post, postID)
	if post.ID != postID {
		return errors.New("post does not exist")
	}
	if post.UserID != userID {
		return errors.New("post not owned by user")
	}
	post.Title = title
	post.Content = content
	data.db.Save(&post)
	return nil
}

func (data *DataSource) AddUser(name, email string, password []byte) error {
	hash, err := bcrypt.GenerateFromPassword(password, bcrypt.DefaultCost)
	if err != nil {
		return errors.Wrap(err, "could not generate hash")
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
	return nil
}

func (data *DataSource) ValidateReportReason(reason string) bool {
	return len(reason) <= reportReasonMaxLength
}

func (data *DataSource) AddReport(postID, reporterID uint, reason string) error {
	if !data.ValidateReportReason(reason) {
		return errors.New("reason is not valid")
	}
	var reporter User
	data.db.First(&reporter, reporterID)
	if reporter.ID != reporterID {
		return errors.New("user does not exist")
	}
	var post Post
	data.db.First(&post, postID)
	if post.ID != postID {
		return errors.New("post does not exist")
	}
	report := Report{
		PostID:     postID,
		ReporterID: reporterID,
		Reason:     reason,
	}
	data.db.Create(&report)
	return nil
}

func (data *DataSource) EmailExists(email string) bool {
	var count int
	data.db.Model(&Identity{}).Where("email = ?", email).Count(&count)
	return count > 0
}

func (data *DataSource) NameExists(name string) bool {
	var count int
	data.db.Model(&User{}).Where("name = ?", name).Count(&count)
	return count > 0
}

func (data *DataSource) GetUser(id uint) (*User, error) {
	var user User
	data.db.First(&user, id)
	if user.ID != id {
		return nil, errors.New("user does not exist")
	}
	return &user, nil
}

func (data *DataSource) GetUserByName(name string) (*User, error) {
	var user User
	data.db.Where("name = ?", name).First(&user)
	if user.Name != name {
		return nil, errors.New("user does not exist")
	}
	return &user, nil
}

func (data *DataSource) DeleteUser(id uint) {
	data.db.Delete(&Identity{}, "user_id = ?", id)
	data.db.Delete(&Post{}, "user_id = ?", id)
	data.db.Delete(&User{}, "id = ?", id)
}

func (data *DataSource) DeletePost(user, id uint) error {
	var post Post
	data.db.First(&post, id)
	if post.ID != id {
		return errors.New("post does not exist")
	}
	if post.UserID != user {
		return errors.New("user not egligible for deletion")
	}
	data.db.Delete(&post)
	return nil
}

var nameRegexp = regexp.MustCompile(`^[a-zA-Z0-9]+$`)

func (data *DataSource) ValidateName(name string) bool {
	return nameRegexp.Match([]byte(name)) && len(name) <= usernameMaxLength
}

const passwordMinLength = 8

func (data *DataSource) ValidatePassword(password string) bool {
	return len(password) >= passwordMinLength
}

func (data *DataSource) ValidateEmail(email string) bool {
	for _, c := range email {
		if c == '@' {
			return true
		}
	}
	return false
}

func (data *DataSource) GetPosts(id uint) ([]Post, error) {
	var posts []Post
	data.db.Where("user_id = ?", id).Find(&posts)
	return posts, nil
}

func (data *DataSource) GetPostsDesc(id uint) ([]Post, error) {
	var posts []Post
	data.db.Where("user_id = ?", id).Order("created_at DESC").Find(&posts)
	return posts, nil
}

func (data *DataSource) GetRecentPosts(count int) ([]Post, error) {
	var posts []Post
	data.db.Order("created_at DESC").Limit(count).Find(&posts)
	return posts, nil
}

func (data *DataSource) GetRecentUsers(count int) ([]User, error) {
	var users []User
	data.db.Order("created_at DESC").Limit(count).Find(&users)
	return users, nil
}

func (data *DataSource) SetBiography(id uint, biography string) error {
	if !data.ValidateBiography(biography) {
		return errors.New("invalid biography")
	}
	var user User
	data.db.First(&user, id)
	if user.ID != id {
		return errors.New("user does not exist")
	}
	user.Biography = biography
	data.db.Save(&user)
	return nil
}

func (data *DataSource) ValidateBiography(biography string) bool {
	return len(biography) <= biographyMaxLength
}

func (data *DataSource) ValidatePostTitle(title string) bool {
	return len(title) <= postTitleMaxLength
}

func (data *DataSource) ValidatePostContent(content string) bool {
	return len(content) <= postContentMaxLength
}

func (data *DataSource) AddPost(author uint, title, content string) (uint, error) {
	if !data.ValidatePostTitle(title) || !data.ValidatePostContent(content) {
		return 0, errors.New("invalid post title or content")
	}
	post := Post{
		UserID:  author,
		Title:   title,
		Content: content,
	}
	data.db.Create(&post)
	return post.ID, nil
}

func (data *DataSource) GetPost(id uint) (*Post, error) {
	var post Post
	data.db.First(&post, id)
	if post.ID != id {
		return nil, errors.New("could not find post")
	}
	return &post, nil
}
