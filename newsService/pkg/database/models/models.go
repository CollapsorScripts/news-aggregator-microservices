package models

import (
	"newsService/pkg/database"
	"regexp"
)

// Размер записей на страницу
const pageSize = 15

type News struct {
	ID      int    // номер записи
	Title   string // заголовок публикации
	Content string `gorm:"unique"` // содержание публикации
	PubTime int64  // время публикации
	Link    string // ссылка на источник
}

// FindAll - все записи
func FindAll() ([]*News, error) {
	db := database.GetDB()
	query := "SELECT * FROM news"
	rows, err := db.Raw(query).Rows()
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var news []*News

	for rows.Next() {
		row := &News{}
		err := rows.Scan(&row.ID, &row.Title, &row.Content, &row.PubTime, &row.Link)
		if err != nil {
			return nil, err
		}
		news = append(news, row)
	}

	return news, nil
}

// FindAllWithoutContent - все записи без контента
func FindAllWithoutContent() ([]*News, error) {
	db := database.GetDB()
	query := "SELECT id, title, pub_time, link FROM news"
	rows, err := db.Raw(query).Rows()
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var news []*News

	for rows.Next() {
		row := &News{}
		err := rows.Scan(&row.ID, &row.Title, &row.PubTime, &row.Link)
		if err != nil {
			return nil, err
		}
		news = append(news, row)
	}

	return news, nil
}

// FindByTitleWithoutContent - поиск записи по title без контента
func FindByTitleWithoutContent(title string) ([]*News, error) {
	db := database.GetDB()

	var news []*News
	result := db.Where("title LIKE ?", "%"+title+"%").Select("id, title, pub_time, link").Find(&news)

	return news, result.Error
}

// Create - создание новой записи
func (n *News) Create() error {
	db := database.GetDB()
	// Регулярное выражение для удаления HTML тегов
	htmlTagRegex := regexp.MustCompile("<[^>]*>")

	// Удаление HTML тегов из строки
	title := htmlTagRegex.ReplaceAllString(n.Title, "")
	content := htmlTagRegex.ReplaceAllString(n.Content, "")

	query := "INSERT INTO news (title, content, pub_time, link) VALUES (?, ?, ?, ?)"
	result := db.Exec(query, title, content, n.PubTime, n.Link)
	return result.Error
}

// FindOne - поиск записи по id
func (n *News) FindOne(id int) error {
	db := database.GetDB()
	query := "SELECT * FROM news WHERE (id = ?)"
	result := db.Raw(query, id).Scan(&n)
	return result.Error
}

// FindLimit - поиск последних N записей
func FindLimit(limit int) ([]*News, error) {
	db := database.GetDB()
	query := "SELECT * FROM news ORDER BY id DESC LIMIT ?;"
	var news []*News

	rows, err := db.Raw(query, limit).Rows()
	if err != nil {
		return nil, err
	}

	for rows.Next() {
		row := &News{}
		err := rows.Scan(&row.ID, &row.Title, &row.Content, &row.PubTime, &row.Link)
		if err != nil {
			return nil, err
		}
		news = append(news, row)
	}

	return news, nil
}

// Paginator - рассчитывает количество страниц в бд (делим кол-во записей на pageSize)
func Paginator() (int, error) {
	db := database.GetDB()
	var count int
	query := "SELECT COUNT(*) FROM news;"
	result := db.Raw(query).Scan(&count)

	if result.Error != nil {
		return -1, result.Error
	}

	pages := count / pageSize
	return pages, nil
}

// PaginationLimitWithoutContent - возвращает новости на определенной странице в размере pageSize (15) без контента
func PaginationLimitWithoutContent(page uint) ([]*News, error) {
	db := database.GetDB()
	var news []*News

	offset := (page - 1) * pageSize

	query := "SELECT id, title, pub_time, link FROM news LIMIT ? OFFSET ?;"

	rows, err := db.Raw(query, pageSize, offset).Rows()
	if err != nil {
		return nil, err
	}

	for rows.Next() {
		row := &News{}
		err := rows.Scan(&row.ID, &row.Title, &row.PubTime, &row.Link)
		if err != nil {
			return nil, err
		}
		news = append(news, row)
	}

	return news, nil
}

// Delete - удаление новости по ID
func Delete(id uint) error {
	db := database.GetDB()
	query := "DELETE FROM news WHERE id = ?"
	result := db.Exec(query, id)

	return result.Error
}
