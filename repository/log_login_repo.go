package repository

import (
	"green-house-api/model"

	"gorm.io/gorm"
)

type logLoginRepo struct {
	DB        *gorm.DB
	TableName string
}

type LogLoginRepoInterface interface {
	FindOneByID(ID uint) (model.LogLoginModel, error)
	FindAllSearchLimitOffsetOrder(param model.LogLoginModel, search string, limit int, offset int, order string) ([]model.LogLoginModel, error)
	CountSearch(param model.LogLoginModel, search string, delete bool) (int64, error)
	UpdateData(ID uint, dataModel model.LogLoginModel) (model.LogLoginModel, error)
	Create(data model.LogLoginModel) (model.LogLoginModel, error)
}

func NewLogLoginRepo(db *gorm.DB) LogLoginRepoInterface {
	var model model.LogLoginModel
	return &logLoginRepo{
		DB:        db,
		TableName: model.TableName(),
	}
}

func (t *logLoginRepo) FindOneByID(ID uint) (model.LogLoginModel, error) {
	data := model.LogLoginModel{}
	err := t.DB.Where("id = ?", ID).First(&data)
	if err.Error != nil {
		return data, err.Error
	}
	return data, nil
}

func (t *logLoginRepo) FindAllSearchLimitOffsetOrder(param model.LogLoginModel, search string, limit int, offset int, order string) ([]model.LogLoginModel, error) {
	if limit == 0 {
		limit = 10
	}

	data := []model.LogLoginModel{}
	db := t.DB

	if search != "" {
		q := "("
		q += "to_char(date, 'YYYY-MM-DD') like ? "
		q += ")"
		db = db.Where(q, "%"+search+"%")
	}

	err := db.Limit(limit).Offset(offset).Order(order).Find(&data)
	if err.Error != nil {
		return data, err.Error
	}

	return data, nil
}

func (t *logLoginRepo) CountSearch(param model.LogLoginModel, search string, delete bool) (int64, error) {
	var result int64
	db := t.DB.Table(t.TableName)
	if delete {
		db = db.Where("deleted_at is null")
	}

	if search != "" {
		q := "("
		q += "to_char(date, 'YYYY-MM-DD') like ? "
		q += ")"
		db = db.Where(q, "%"+search+"%")
	}
	err := db.Count(&result)
	if err.Error != nil {
		return 0, err.Error
	}
	return result, nil
}

func (t *logLoginRepo) Create(data model.LogLoginModel) (model.LogLoginModel, error) {
	err := t.DB.Create(&data)
	if err.Error != nil {
		return data, err.Error
	}
	return data, nil
}

func (t *logLoginRepo) UpdateData(ID uint, dataModel model.LogLoginModel) (model.LogLoginModel, error) {
	// now := time.Now()
	err := t.DB.Where("id = ?", ID).Updates(dataModel)
	data := model.LogLoginModel{}
	t.DB.First(&data, dataModel.ID)
	if err.Error != nil {
		return data, err.Error
	}
	return data, nil
}
