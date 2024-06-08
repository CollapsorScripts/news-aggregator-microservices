package models

import (
	"commentsService/pkg/database"
	"commentsService/pkg/logger"
)

// Comment - комментарий
type Comment struct {
	ID              uint
	NewsID          uint
	ParentCommentID uint       `gorm:"default:0"`
	ChildComments   []*Comment `gorm:"-:all"`
	Text            string
}

func (c *Comment) Create() error {
	dbase := database.GetDB()

	result := dbase.Create(&c)

	return result.Error
}

func (c *Comment) FindByID() error {
	dbase := database.GetDB()
	result := dbase.Where(&Comment{ID: c.ID}).Find(&c)

	return result.Error
}

func FindByNewsID(id uint) ([]*Comment, error) {
	dbase := database.GetDB()
	var comments []*Comment
	result := dbase.Where(&Comment{NewsID: id}).Find(&comments)

	if result.Error != nil {
		return nil, result.Error
	}

	var childComments []*Comment
	//Подгружаем дочерние комментарии
	for i, comment := range comments {
		result = dbase.Where(&Comment{ParentCommentID: comment.ID}).Find(&childComments)
		if result.Error != nil {
			logger.Error("Ошибка при поиске дочерних комментариев: %v", result.Error)
			continue
		}
		if len(childComments) > 0 {
			comments[i].ChildComments = childComments
		}
	}

	return comments, result.Error
}

func FindByParentID(id uint) ([]*Comment, error) {
	dbase := database.GetDB()
	var childComments []*Comment

	result := dbase.Where(&Comment{ParentCommentID: id}).Find(&childComments)

	return childComments, result.Error
}

func (c *Comment) Delete() error {
	dbase := database.GetDB()
	result := dbase.Delete(&c)

	return result.Error
}

func DeleteByID(id uint) error {
	dbase := database.GetDB()
	result := dbase.Delete(&Comment{}, id)

	return result.Error
}
