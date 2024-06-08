package models

// Comment - комментарий
type Comment struct {
	ID              uint
	NewsID          uint
	ParentCommentID uint
	Text            string
}
