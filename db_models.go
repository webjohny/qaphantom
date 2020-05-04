package main

import (
	"database/sql"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"time"
)

// You will be using this Trainer type later in the program
type Question struct {
	Id interface{} `db:"Id" json:"_id"`
	Log string `db:"Log" json:"log"`
	LogLast string `db:"LogLast" json:"log_last"`
	SiteId int `db:"SiteId" json:"site_id"`
	Cat string `db:"Cat" json:"cat"`
	CatId primitive.ObjectID `db:"CatId" json:"cat_id"`
	TryCount int `db:"TryCount" json:"try_count"`
	ErrorsCount int `db:"ErrorsCount" json:"errors_count"`
	Status int `db:"Status" json:"status"`
	Error string `db:"Error" json:"error"`
	ParserId int `db:"ParserId" json:"parser"`
	Timeout time.Time `db:"Timeout" json:"timeout"`
	Keyword string `db:"Keyword" json:"keyword"`
	FastA string `db:"FastA" json:"fast_a"`
	FastLink string `db:"FastLink" json:"fast_link"`
	FastLinkTitle string `db:"FastLinkTitle" json:"fast_link_title"`
	FastDate time.Time `db:"FastDate" json:"fast_date"`
}

type MysqlSite struct {
	Id sql.NullInt64 `db:"id" json:"id"`
	Language string `db:"language" json:"language"`
	Theme sql.NullString `db:"theme" json:"theme"`
	Domain sql.NullString `db:"domain" json:"domain"`
	Login sql.NullString `db:"login" json:"login"`
	Password sql.NullString `db:"password" json:"password"`
	From sql.NullInt64 `db:"from" json:"from"`
	To sql.NullInt64 `db:"to" json:"to"`
	QstsLimit int `db:"qsts_limit" json:"qsts_limit"`
	Linking int `db:"linking" json:"linking"`
	Header int `db:"header" json:"header"`
	SubHeaders int `db:"subheaders" json:"subheaders"`
	ParseDates int `db:"parse_dates" json:"parse_dates"`
	ParseDoubles int `db:"parse_doubles" json:"parse_doubles"`
	PubImage int `db:"pub_image" json:"pub_image"`
	VideoStep int `db:"video_step" json:"video_step"`
	QaCount int `db:"qa_count" json:"qa_count"`
	QaCountFrom sql.NullInt32 `db:"qa_count_from" json:"qa_count_from"`
	QaCountTo sql.NullInt32 `db:"qa_count_to" json:"qa_count_to"`
	ParseFast sql.NullInt32 `db:"parse_fast" json:"parse_fast"`
	ParseSearch4 sql.NullInt32 `db:"parse_search4" json:"parse_search4"`
	ImageKey int `db:"image_key" json:"image_key"`
	H1 sql.NullInt32 `db:"h1" json:"h1"`
	ShOrder sql.NullInt32 `db:"sh_order" json:"sh_order"`
	ShFormat sql.NullInt32 `db:"sh_format" json:"sh_format"`
	ImageSource int `db:"image_source" json:"image_source"`
	Info sql.NullString `json:"info"`
}

type MysqlCat struct {
	Id sql.NullInt64 `db:"id" json:"language"`
	SiteId sql.NullInt64 `db:"site_id" json:"site_id"`
	Title sql.NullString `db:"title" json:"title"`
}

type MysqlTask struct {
	Id sql.NullInt64 `db:"id" json:"id"`
	Log sql.NullString `db:"log" json:"log"`
	LogLast sql.NullString `db:"log_last" json:"log_last"`
	SiteId sql.NullInt64 `db:"site_id" json:"site_id"`
	CatId sql.NullInt64 `db:"cat_id" json:"cat_id"`
	TryCount sql.NullInt32 `db:"try_count" json:"try_count"`
	ErrorsCount sql.NullInt32 `db:"errors_count" json:"errors_count"`
	Status sql.NullInt32 `db:"status" json:"status"`
	Error sql.NullString `db:"error" json:"error"`
	Parser sql.NullInt64 `db:"parser" json:"parser"`
	Timeout sql.NullString `db:"timeout" json:"timeout"`
	Keyword sql.NullString `db:"keyword" json:"keyword"`
	FastA sql.NullString `db:"fast_a" json:"fast_a"`
	FastLink sql.NullString `db:"fast_link" json:"fast_link"`
	FastLinkTitle sql.NullString `db:"fast_link_title" json:"fast_link_title"`
	FastDate sql.NullString `db:"fast_date" json:"fast_date"`
	ParseDate sql.NullString `db:"parse_date" json:"parse_date"`
}
