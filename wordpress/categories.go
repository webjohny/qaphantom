package wordpress

import (
	"fmt"
	"net/http"
)
//"id": 1,
//"count": 21,
//"description": "",
//"link": "http://philli.beget.tech/category/%d0%b1%d0%b5%d0%b7-%d1%80%d1%83%d0%b1%d1%80%d0%b8%d0%ba%d0%b8/",
//"name": "Без рубрики",
//"slug": "%d0%b1%d0%b5%d0%b7-%d1%80%d1%83%d0%b1%d1%80%d0%b8%d0%ba%d0%b8",
//"taxonomy": "category",
//"parent": 0,
//"meta": [],
type Category struct {
	collection *CategoriesCollection `json:"-"`

	ID            int     `json:"id,omitempty"`
	Count         int  `json:"date,omitempty"`
	Description   string  `json:"description,omitempty"`
	Link          string    `json:"link,omitempty"`
	Name          string  `json:"name,omitempty"`
	Slug          string  `json:"slug,omitempty"`
	Taxonomy      string  `json:"taxonomy,omitempty"`
	Parent            int  `json:"parent,omitempty"`
	Meta           interface{}  `json:"meta,omitempty"`
}

func (entity *Category) setCollection(col *CategoriesCollection) {
	entity.collection = col
}

func (entity *Category) Populate(params interface{}) (*Category, *http.Response, []byte, error) {
	return entity.collection.Get(entity.ID, params)
}

type CategoriesCollection struct {
	client    *Client
	url       string
	entityURL string
}

func (col *CategoriesCollection) List(params interface{}) ([]Category, *http.Response, []byte, error) {
	var cats []Category
	resp, body, err := col.client.List(col.url, params, &cats)

	// set collection object for each entity which has sub-collection
	for _, p := range cats {
		p.setCollection(col)
	}

	return cats, resp, body, err
}
func (col *CategoriesCollection) Create(new *Category) (*Category, *http.Response, []byte, error) {
	var created Category
	resp, body, err := col.client.Create(col.url, new, &created)

	created.setCollection(col)

	return &created, resp, body, err
}
func (col *CategoriesCollection) Get(id int, params interface{}) (*Category, *http.Response, []byte, error) {
	var entity Category
	entityURL := fmt.Sprintf("%v/%v", col.url, id)
	resp, body, err := col.client.Get(entityURL, params, &entity)

	// set collection object for each entity which has sub-collection
	entity.setCollection(col)

	return &entity, resp, body, err
}
func (col *CategoriesCollection) Entity(id int) *Category {
	entity := Category{
		collection: col,
		ID:         id,
	}
	return &entity
}

func (col *CategoriesCollection) Update(id int, category *Category) (*Category, *http.Response, []byte, error) {
	var updated Category
	entityURL := fmt.Sprintf("%v/%v", col.url, id)
	resp, body, err := col.client.Update(entityURL, category, &updated)

	// set collection object for each entity which has sub-collection
	updated.setCollection(col)

	return &updated, resp, body, err
}
func (col *CategoriesCollection) Delete(id int, params interface{}) (*Category, *http.Response, []byte, error) {
	var deleted Category
	entityURL := fmt.Sprintf("%v/%v", col.url, id)

	resp, body, err := col.client.Delete(entityURL, params, &deleted)

	// set collection object for each entity which has sub-collection
	deleted.setCollection(col)

	return &deleted, resp, body, err
}
