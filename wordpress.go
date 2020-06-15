package main

import (
	xmlrpc "github.com/abcdsxg/go-wordpress-xmlrpc"
	"github.com/gosimple/slug"
	"log"
	"strconv"
	"time"
)

type WpCat struct {
	Description string `json:"description"`
	Filter string `json:"filter"`
	Name string `json:"name"`
	Parent int `json:"parent"`
	Slug string `json:"slug"`
	Taxonomy string `json:"taxonomy"`
	TermGroup int `json:"term_group"`
	TermId int `json:"term_id"`
	TermTaxonomyId int `json:"term_taxonomy_id"`
}

type WpPost struct {
	Id int
	Title string
	Content string
	Date time.Time
	Link string
	Slug string
	Parent int
	Terms []WpCat
}

type Wordpress struct {
	client *xmlrpc.Client
	cnf []interface{}
	err error
}

func (w *Wordpress) Connect(url string, username string, password string, blogId int) *xmlrpc.Client {
	c, err := xmlrpc.NewClient(url, xmlrpc.UserInfo{
		username,
		password,
	})
	if err != nil {
		log.Println(err)
	}
	w.client = c
	w.cnf = []interface{}{
		blogId, username, password,
	}
	return c
}

func (w *Wordpress) PrepareCat(cat map[string]interface{}) WpCat {
	parentId, _ := strconv.Atoi(cat["parent"].(string))
	termGroup, _ := strconv.Atoi(cat["term_group"].(string))
	termId, _ := strconv.Atoi(cat["term_id"].(string))
	termTaxonomyId, _ := strconv.Atoi(cat["term_taxonomy_id"].(string))
	var description string
	if cat["description"] != nil {
		description = cat["description"].(string)
	}
	return WpCat{
		Description:    description,
		Filter:         cat["filter"].(string),
		Name:           cat["name"].(string),
		Parent:         parentId,
		Slug:           cat["slug"].(string),
		Taxonomy:       cat["taxonomy"].(string),
		TermGroup:      termGroup,
		TermId:         termId,
		TermTaxonomyId: termTaxonomyId,
	}
}

func (w *Wordpress) PreparePost(post map[string]interface{}) WpPost {
	parent, _ := strconv.Atoi(post["post_parent"].(string))
	var cats []WpCat
	terms := post["terms"].([]interface{})
	if len(terms) > 0 {
		for _, item := range terms {
			cat := item.(map[string]interface{})
			cats = append(cats, w.PrepareCat(cat))
		}
	}
	id, _ := strconv.Atoi(post["post_id"].(string))
	return WpPost{
		Id: id,
		Link: post["link"].(string),
		Title: post["post_title"].(string),
		Content: post["post_content"].(string),
		Date: post["post_date"].(time.Time),
		Slug: post["post_name"].(string),
		Parent: parent,
		Terms: cats,
	}
}

func (w *Wordpress) GetCats() []WpCat {
	var result interface{}
	err := w.client.Client.Call(`wp.getTerms`, append(
		w.cnf, "category",
	), &result)
	if err != nil {
		log.Println(err)
	}
	res := result.([]interface{})
	var cats []WpCat
	if len(res) > 0 {
		for _, item := range res {
			cat := item.(map[string]interface{})
			cats = append(cats, w.PrepareCat(cat))
		}
	}
	return cats
}

func (w *Wordpress) NewTerm(name string, taxonomy string, slug string, description string, parentId int) int {
	params := map[string]string{
		"name": name,
		"taxonomy": taxonomy,
	}

	if slug != "" {
		params["slug"] = slug
	}

	if description != "" {
		params["description"] = description
	}

	if parentId > 0 {
		params["parent"] = strconv.Itoa(parentId)
	}

	var result interface{}
	err := w.client.Client.Call(`wp.newTerm`, append(
		w.cnf, params,
	), &result)
	if err != nil {
		w.err = err
	}

	return result.(int)
}

func (w *Wordpress) GetPost(id int) WpPost {
	var result interface{}
	err := w.client.Client.Call(`wp.getPost`, append(
		w.cnf, id,
	), &result)
	if err != nil {
		log.Println(err)
	}
	res := result.(map[string]interface{})
	post := w.PreparePost(res)

	return post
}

func (w *Wordpress) EditPost(id int, title string, content string) bool {
	params := map[string]string{}
	if title != "" {
		params["post_title"] = title
	}
	if content != "" {
		params["post_content"] = content
	}
	var result interface{}
	err := w.client.Client.Call(`wp.editPost`, append(
		w.cnf, id, params,
	), &result)
	if err != nil {
		log.Println(err)
	}
	return result.(bool)
}

func (w *Wordpress) NewPost(title string, content string, catId int, photoId int) int {
	params := map[string]interface{}{
		"post_type": "post",
		"post_status": "publish",
	}
	if title != "" {
		params["post_title"] = title
		params["post_name"] = slug.Make(title)
	}
	if content != "" {
		params["post_content"] = content
	}
	if photoId > 0 {
		params["post_thumbnail"] = photoId
	}
	if catId > 0 {
		params["terms"] = map[string][]int{
			"category": {catId},
		}
	}

	var result interface{}
	err := w.client.Client.Call(`wp.newPost`, append(
		w.cnf, params,
	), &result)
	if err != nil {
		w.err = err
	}

	id, _ := strconv.Atoi(result.(string))
	return id
}

func (w *Wordpress) UploadFile(name string, mime string, bits string, overwrite string, postId int) map[string]interface{} {
	params := map[string]interface{}{
		"name": name,
		"type": mime,
		"bits": bits,
	}
	if overwrite != "" {
		params["overwrite"] = overwrite
	}
	if postId != 0 {
		params["post_id"] = postId
	}

	var result map[string]interface{}
	err := w.client.Client.Call(`wp.uploadFile`, append(
		w.cnf, params,
	), &result)
	if err != nil {
		log.Println(err)
	}

	return result
}

func (w *Wordpress) CatIdByName(name string) int {
	var catId int

	// Загружаем список категорий
	cats := w.GetCats()

	// Создавать ли категорию
	create := true

	// Пробегаем по всем категориям
	if len(cats) > 0 {
		for _, cat := range cats {
			// Проверка существования категории
			if cat.Name == name {
				catId = cat.TermId
				create = false
				break
			}
		}
	}

	// Создаём категорию
	if create {
		catId = w.NewTerm(name, "category", slug.Make(name), "", 0)
	}

	return catId
}