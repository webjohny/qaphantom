package main

import (
	"database/sql"
)

type MysqlSite struct {
	Id sql.NullInt64 `db:"id" json:"id"`
	Language sql.NullString `db:"language" json:"language"`
	Theme sql.NullString `db:"theme" json:"theme"`
	Domain sql.NullString `db:"domain" json:"domain"`
	Login sql.NullString `db:"login" json:"login"`
	Password sql.NullString `db:"password" json:"password"`
	From sql.NullInt64 `db:"from" json:"from"`
	To sql.NullInt64 `db:"to" json:"to"`
	QstsLimit sql.NullInt64 `db:"qsts_limit" json:"qsts_limit"`
	Linking sql.NullInt64 `db:"linking" json:"linking"`
	Header sql.NullInt64 `db:"header" json:"header"`
	SubHeaders sql.NullInt64 `db:"subheaders" json:"subheaders"`
	ParseDates sql.NullInt64 `db:"parse_dates" json:"parse_dates"`
	ParseDoubles sql.NullInt64 `db:"parse_doubles" json:"parse_doubles"`
	PubImage sql.NullInt64 `db:"pub_image" json:"pub_image"`
	VideoStep sql.NullInt64 `db:"video_step" json:"video_step"`
	QaCount sql.NullInt64 `db:"qa_count" json:"qa_count"`
	QaCountFrom sql.NullInt32 `db:"qa_count_from" json:"qa_count_from"`
	QaCountTo sql.NullInt32 `db:"qa_count_to" json:"qa_count_to"`
	ParseFast sql.NullInt32 `db:"parse_fast" json:"parse_fast"`
	ParseSearch4 sql.NullInt32 `db:"parse_search4" json:"parse_search4"`
	ImageKey sql.NullInt64 `db:"image_key" json:"image_key"`
	H1 sql.NullInt32 `db:"h1" json:"h1"`
	ShOrder sql.NullInt32 `db:"sh_order" json:"sh_order"`
	ShFormat sql.NullInt32 `db:"sh_format" json:"sh_format"`
	ImageSource sql.NullInt64 `db:"image_source" json:"image_source"`
	Info sql.NullString `db:"info" json:"info"`
	MoreTags sql.NullString `db:"more_tags" json:"more_tags"`
	SymbMicroMarking sql.NullString `db:"symb_micro_marking" json:"symb_micro_marking"`
	CountRows sql.NullInt64 `db:"count_rows" json:"count_rows"`
}

type MysqlConfig struct {
	Id sql.NullInt64 `db:"id" json:"id"`
	FlickrKey sql.NullString `db:"flickr_key" json:"flickr_key"`
	FlickrSecret sql.NullString `db:"flickr_secret" json:"flickr_secret"`
	Antigate sql.NullString `db:"antigate" json:"antigate"`
	Language sql.NullString `db:"language" json:"language"`
	Variants sql.NullString `db:"variants" json:"variants"`
}

type MysqlCat struct {
	Id sql.NullInt64 `db:"id" json:"id"`
	SiteId sql.NullInt64 `db:"site_id" json:"site_id"`
	Title sql.NullString `db:"title" json:"title"`
}

type MysqlUagent struct {
	Id sql.NullInt64 `db:"id" json:"id"`
	Sign sql.NullString `db:"sign" json:"sign"`
	Status sql.NullInt32 `db:"status" json:"status"`
	Timeout sql.NullString `db:"timeout" json:"timeout"`
}

type MysqlResult struct {
	Id sql.NullInt64 `db:"id" json:"id"`
	TaskId sql.NullInt64 `db:"task_id" json:"task_id"`
	Q sql.NullString `db:"q" json:"q"`
	A sql.NullString `db:"a" json:"a"`
	Link sql.NullString `db:"link" json:"link"`
	LinkTitle sql.NullString `db:"link_title" json:"link_title"`
	CreateDate sql.NullString `db:"create_date" json:"create_date"`
	QaDate sql.NullString `db:"qa_date" json:"qa_date"`
}

type MysqlTask struct {
	Id sql.NullInt64 `db:"id" json:"id"`
	Log sql.NullString `db:"log" json:"log"`
	LogLast sql.NullString `db:"log_last" json:"log_last"`
	ParentId sql.NullInt64 `db:"parent_id" json:"parent_id"`
	SiteId sql.NullInt64 `db:"site_id" json:"site_id"`
	CatId sql.NullInt64 `db:"cat_id" json:"cat_id"`
	Keyword sql.NullString `db:"keyword" json:"keyword"`
	Cat sql.NullString `db:"cat" json:"cat"`
	TryCount sql.NullInt32 `db:"try_count" json:"try_count"`
	ErrorsCount sql.NullInt32 `db:"errors_count" json:"errors_count"`
	Status sql.NullInt32 `db:"status" json:"status"`
	Error sql.NullString `db:"error" json:"error"`
	Parser sql.NullInt64 `db:"parser" json:"parser"`
	Timeout sql.NullString `db:"timeout" json:"timeout"`
	FastA sql.NullString `db:"fast_a" json:"fast_a"`
	FastLink sql.NullString `db:"fast_link" json:"fast_link"`
	FastLinkTitle sql.NullString `db:"fast_link_title" json:"fast_link_title"`
	FastDate sql.NullString `db:"fast_date" json:"fast_date"`
	ParseDate sql.NullString `db:"parse_date" json:"parse_date"`
}

type MysqlFreeTask struct {
	Id int `db:"id" json:"id"`
	SiteId int `db:"site_id" json:"site_id"`
	CatId int `db:"cat_id" json:"cat_id"`
	Keyword string `db:"keyword" json:"keyword"`
	Cat string `db:"cat" json:"cat"`
	TryCount int `db:"try_count" json:"try_count"`
	Log []string `db:"log" json:"log"`

	Language string `db:"language" json:"language"`
	Theme string `db:"theme" json:"theme"`
	Domain string `db:"domain" json:"domain"`
	Login string `db:"login" json:"login"`
	Password string `db:"password" json:"password"`
	From int `db:"from" json:"from"`
	To int `db:"to" json:"to"`
	QstsLimit int `db:"qsts_limit" json:"qsts_limit"`
	Linking int `db:"linking" json:"linking"`
	Header int `db:"header" json:"header"`
	SubHeaders int `db:"subheaders" json:"subheaders"`
	ParseDates int `db:"parse_dates" json:"parse_dates"`
	ParseDoubles int `db:"parse_doubles" json:"parse_doubles"`
	PubImage int `db:"pub_image" json:"pub_image"`
	VideoStep int `db:"video_step" json:"video_step"`
	QaCountFrom int `db:"qa_count_from" json:"qa_count_from"`
	QaCountTo int `db:"qa_count_to" json:"qa_count_to"`
	ParseFast int `db:"parse_fast" json:"parse_fast"`
	ParseSearch4 int `db:"parse_search4" json:"parse_search4"`
	ImageKey int `db:"image_key" json:"image_key"`
	H1 int `db:"h1" json:"h1"`
	ShOrder int `db:"sh_order" json:"sh_order"`
	ShFormat int `db:"sh_format" json:"sh_format"`
	ImageSource int `db:"image_source" json:"image_source"`
	MoreTags string `db:"more_tags" json:"more_tags"`
	SymbMicroMarking string `db:"symb_micro_marking" json:"symb_micro_marking"`
	CountRows int `db:"count_rows" json:"count_rows"`
	SavingAvailable bool `json:"saving_available"`
}

type MysqlProxy struct {
	Id sql.NullInt64 `json:"id"`
	Type sql.NullString `json:"type"`
	Host sql.NullString `json:"host"`
	Port sql.NullString `json:"port"`
	Login sql.NullString `json:"login"`
	Password sql.NullString `json:"password"`
	Agent sql.NullString `json:"agent"`
	Status sql.NullInt64 `json:"status"`
	Parser sql.NullString `json:"parser"`
	Timeout sql.NullString `json:"timeout"`
}

