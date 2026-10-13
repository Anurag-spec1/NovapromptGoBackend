package cache

import "fmt"

func KeyCategoriesList() string {
	return "nova:category:list"
}

func KeyCategory(id string) string {
	return fmt.Sprintf("nova:category:%s", id)
}

func KeyAdKey(key string) string {
	return fmt.Sprintf("nova:key:%s", key)
}

func KeyAdKeysList() string {
	return "nova:key:list"
}

const (
	TTLCategories = 60
	TTLAdKeys     = 300
)