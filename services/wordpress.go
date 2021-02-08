package services

import (
	"encoding/base64"
	"fmt"
	wpXmlrpc "github.com/abcdsxg/go-wordpress-xmlrpc"
	"github.com/gosimple/slug"
	"github.com/h2non/filetype"
	"github.com/kolo/xmlrpc"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"path"
	"reflect"
	"strconv"
	"strings"
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
	client *wpXmlrpc.Client
	cnf []interface{}
	err error
}

type WpImage struct {
	Id int
	Url string
	UrlMedium string
}


func isNil(i interface{}) bool {
	return i == nil || reflect.ValueOf(i).IsNil()
}

func randStringRunes(n int) string {
	letterRunes := []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

	b := make([]rune, n)
	for i := range b {
		b[i] = letterRunes[rand.Intn(len(letterRunes))]
	}
	return string(b)
}

func toInt(value string) int {
	var integer int = 0
	if value != "" {
		integer, _ = strconv.Atoi(value)
	}
	return integer
}

func (w *Wordpress) Connect(url string, username string, password string, blogId int) *wpXmlrpc.Client {
	c, err := wpXmlrpc.NewClient(url, wpXmlrpc.UserInfo{
		username,
		password,
	})
	if err != nil {
		w.err = err
		log.Println("Wordpress.Connect.HasError", err)
		return nil
	}
	w.client = c
	w.cnf = []interface{}{
		blogId, username, password,
	}
	return c
}

func (w *Wordpress) GetError() error {
	return w.err
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

	wpPost := WpPost{
		Id: id,
		Date: post["post_date"].(time.Time),
		Parent: parent,
		Terms: cats,
	}
	if !isNil(post["post_content"]){
		wpPost.Content = post["post_content"].(string)
	}
	if !isNil(post["post_content"]){
		wpPost.Title = post["post_title"].(string)
	}
	if !isNil(post["post_name"]){
		wpPost.Slug = post["post_name"].(string)
	}
	if !isNil(post["link"]){
		wpPost.Link = post["link"].(string)
	}
	return wpPost
}

func (w *Wordpress) GetCats() []WpCat {
	var result interface{}
	err := w.client.Client.Call(`wp.getTerms`, append(
		w.cnf, "category",
	), &result)
	if err != nil {
		w.err = err
		log.Println("Wordpress.GetCats.HasError", err)
	}
	var cats []WpCat
	if result != nil {
		res := result.([]interface{})
		if len(res) > 0 {
			for _, item := range res {
				cat := item.(map[string]interface{})
				cats = append(cats, w.PrepareCat(cat))
			}
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
		log.Println("Wordpress.NewTerm.HasError", err)
		return 0
	}

	return result.(int)
}

func (w *Wordpress) GetPost(id int) WpPost {
	var result interface{}
	err := w.client.Client.Call(`wp.getPost`, append(
		w.cnf, id,
	), &result)
	if err != nil {
		w.err = err
		log.Println("Wordpress.GetPost.HasError", err)
		return WpPost{}
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
		w.err = err
		log.Println("Wordpress.EditPost.HasError", err)
		return false
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
		log.Println("Wordpress.NewPost.HasError", err)
		return 0
	}

	id, _ := strconv.Atoi(result.(string))
	return id
}

func (w *Wordpress) CheckConn() bool {
	return w.client != nil
}

func (w *Wordpress) UploadFile(url string, postId int, encoded bool) (WpImage, error) {
	var image WpImage
	var bytes []byte
	var err error
	var name string

	if !encoded {
		resp, _ := http.Get(url)
		defer resp.Body.Close()

		bytes, err = ioutil.ReadAll(resp.Body)
		if err != nil {
			log.Println("Wordpress.UploadFile.HasError", err)
			return image, err
		}
		name = path.Base(url)
	}else{
		bytes, err = base64.StdEncoding.DecodeString(url)
		if err != nil {
			log.Println("Wordpress.UploadFile.HasError.1", err)
			return image, err
		}
		kind, _ := filetype.Match(bytes)
		if kind == filetype.Unknown {
			fmt.Println("Wordpress.UploadFile.HasError.2", "Unknown file type")
			return image, nil
		}

		name = randStringRunes(20) + "." + kind.Extension
	}

	mime := http.DetectContentType(bytes)
	if !strings.Contains(mime, "image") {
		return image, nil
	}

	encodedImg := base64.StdEncoding.EncodeToString(bytes)

	params := map[string]interface{}{
		"overwrite": true,
		"name": name,
		"type": mime,
		"bits": xmlrpc.Base64(encodedImg),
	}

	if postId != 0 {
		params["post_id"] = postId
	}

	var response map[string]interface{}
	err = w.client.Client.Call(`wp.uploadFile`, append(
		w.cnf, params,
	), &response)
	if err != nil {
		log.Println("Wordpress.UploadFile.2.HasError", err)
		w.err = err
	}else if response != nil{
		image.Id = toInt(response["id"].(string))
		image.Url = response["link"].(string)
		title := path.Base(response["url"].(string))
		image.UrlMedium = response["link"].(string)
		if response["metadata"] != nil {
			metadata := response["metadata"].(map[string]interface{})
			if metadata["sizes"] != nil {
				sizes := metadata["sizes"].(map[string]interface{})
				if sizes["medium"] != nil {
					medium := sizes["medium"].(map[string]interface{})
					if medium["file"] != nil {
						file := medium["file"].(string)
						image.UrlMedium = strings.Replace(image.UrlMedium, title, file, 1)
					}
				}
			}
		}
		return image, err
	}

	return image, nil
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