package main

import (
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

type Site struct {
	Language string `db:"Language" json:"language"`
	Theme string `db:"Theme" json:"theme"`
	Domain string `db:"Domain" json:"domain"`
	Login string `db:"Login" json:"login"`
	Password string `db:"Password" json:"password"`
	From int `db:"From" json:"from"`
	To int `db:"To" json:"to"`
	Linking int `db:"Linking" json:"linking"`
	Header int `db:"Header" json:"header"`
	ParseDates int `db:"ParseDates" json:"parse_dates"`
	ParseDoubles int `db:"ParseDoubles" json:"parse_doubles"`
	PubImage int `db:"PubImage" json:"pub_image"`
	VideoStep int `db:"VideoStep" json:"video_step"`
	QaCountFrom int `db:"QaCountFrom" json:"qa_count_from"`
	QaCountTo int `db:"QaCountTo" json:"qa_count_to"`
	ImageKey int `db:"ImageKey" json:"image_key"`
	H1 int `db:"H1" json:"h1"`
	ShOrder int `db:"ShOrder" json:"sh_order"`
	ShFormat int `db:"ShFormat" json:"sh_format"`
	ImageSource int `db:"ImageSource" json:"image_source"`
}
