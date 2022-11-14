package mysql

import (
	"context"
	"testing"
	"time"
)

func TestMain(m *testing.M) {
	configs := map[string]*Conf{
		"default-master": &Conf{
			Addr:        "127.0.0.1",
			Username:    "root",
			Password:    "123456",
			Database:    "db_test",
			DsnOptions:  "charset=utf8mb4&parseTime=True",
			MaxIdle:     16,
			MaxOpen:     128,
			MaxLifeTime: time.Second * 300,
			SlowTime:    1 * time.Second,
			SlowLogger:  "slow",
		},

		"default-slave": &Conf{
			Addr:        "127.0.0.1",
			Username:    "root",
			Password:    "123456",
			Database:    "db_test",
			DsnOptions:  "charset=utf8mb4&parseTime=True",
			MaxIdle:     16,
			MaxOpen:     128,
			MaxLifeTime: time.Second * 300,
			SlowTime:    1 * time.Second,
			SlowLogger:  "slow",
		},
	}

	if err := Init(context.TODO(), configs); err != nil {
		panic(err)
	}

	m.Run()
}

type UserModel struct {
	ID         int64  `gorm:"column:id;autoIncrement;primary_key;type:bigint unsigned;comment:ID"`
	Name       string `gorm:"column:name;type:varchar(32) NOT NULL default 'default';comment:Name"`
	Age        int64  `gorm:"column:age;type:int unsigned NOT NULL default '0';comment:Age"`
	CreateTime int64  `gorm:"column:create_time;type:int unsigned NOT NULL default '0';comment:创建时间"`
	UpdateTime int64  `gorm:"column:update_time;type:int unsigned NOT NULL default '0';comment:更新时间"`
}

func (um *UserModel) TableName() string {
	return "user"
}

func TestMigrate(t *testing.T) {
	if Get("default-master").Migrator().HasTable(&UserModel{}) {
		if err := Get("default-master").Migrator().DropTable(&UserModel{}); err != nil {
			t.Fatal(err)
		}
	}
	if err := Get("default-master").Set("gorm:table_options", "ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='用户表'").AutoMigrate(&UserModel{}); err != nil {
		t.Fatal(err)
	}
}
