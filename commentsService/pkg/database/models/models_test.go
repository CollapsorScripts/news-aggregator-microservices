package models

import (
	"commentsService/cmd/config"
	"commentsService/pkg/utilities"
	"reflect"
	"testing"
)

func loadConfig() {
	config.Init("C:\\Users\\Alex\\Desktop\\final project\\commentsService\\.env")
}

func TestComment_Create(t *testing.T) {
	loadConfig()
	type fields struct {
		ID              uint
		NewsID          uint
		ParentCommentID uint
		Text            string
		Acceptable      bool
	}
	tests := []struct {
		name    string
		fields  fields
		wantErr bool
	}{
		{
			name: "Создание комментария",
			fields: fields{
				NewsID:          1,
				ParentCommentID: 0,
				Text:            "Это комментарий",
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Comment{
				ID:              tt.fields.ID,
				NewsID:          tt.fields.NewsID,
				ParentCommentID: tt.fields.ParentCommentID,
				Text:            tt.fields.Text,
				Acceptable:      tt.fields.Acceptable,
			}
			if err := c.Create(); (err != nil) != tt.wantErr {
				t.Errorf("Create() error = %v, wantErr %v", err, tt.wantErr)
			}

			t.Logf("Комментарий создан")
		})
	}
}

func TestComment_Delete(t *testing.T) {
	type fields struct {
		ID              uint
		NewsID          uint
		ParentCommentID uint
		Text            string
		Acceptable      bool
	}
	tests := []struct {
		name    string
		fields  fields
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Comment{
				ID:              tt.fields.ID,
				NewsID:          tt.fields.NewsID,
				ParentCommentID: tt.fields.ParentCommentID,
				Text:            tt.fields.Text,
				Acceptable:      tt.fields.Acceptable,
			}
			if err := c.Delete(); (err != nil) != tt.wantErr {
				t.Errorf("Delete() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestComment_FindByID(t *testing.T) {
	type fields struct {
		ID              uint
		NewsID          uint
		ParentCommentID uint
		Text            string
		Acceptable      bool
	}
	tests := []struct {
		name    string
		fields  fields
		wantErr bool
	}{
		{
			name: "Поиск комментария по ID",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Comment{
				ID:              tt.fields.ID,
				NewsID:          tt.fields.NewsID,
				ParentCommentID: tt.fields.ParentCommentID,
				Text:            tt.fields.Text,
				Acceptable:      tt.fields.Acceptable,
			}
			if err := c.FindByID(); (err != nil) != tt.wantErr {
				t.Errorf("FindByID() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestDeleteByID(t *testing.T) {
	type args struct {
		id uint
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := DeleteByID(tt.args.id); (err != nil) != tt.wantErr {
				t.Errorf("DeleteByID() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestFindByNewsID(t *testing.T) {
	loadConfig()
	type args struct {
		id uint
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name:    "Поиск комментария по ID новости",
			args:    args{id: 1},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := FindByNewsID(tt.args.id)
			if (err != nil) != tt.wantErr {
				t.Errorf("FindByNewsID() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			t.Logf("Комментарии: %s", utilities.ToJSON(got))
		})
	}
}

func TestFindByParentID(t *testing.T) {
	type args struct {
		id uint
	}
	tests := []struct {
		name    string
		args    args
		want    []*Comment
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := FindByParentID(tt.args.id)
			if (err != nil) != tt.wantErr {
				t.Errorf("FindByParentID() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("FindByParentID() got = %v, want %v", got, tt.want)
			}
		})
	}
}
